<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";

    export let regexp: string = "";
    export let simpleMode: boolean = false;
    export let simpleBlocks: any[] = [];
    export let isReadOnly: boolean = false;

    const dispatch = createEventDispatcher();
    
    // 初期化
    onMount(() => {
        if (!simpleBlocks) {
            simpleBlocks = [];
        }
    });

    // 簡易パターンのタイプ
    const patternTypes = [
        {
            id: "between",
            label: "[開始]と[終了]の間を抽出",
            fields: ["start", "end"],
        },
        {
            id: "contains",
            label: "[キーワード]を含む行全体を抽出",
            fields: ["keyword"],
        },
        {
            id: "after",
            label: "[プレフィックス]の後ろを抽出",
            fields: ["prefix"],
        },
    ];

    // 簡易モードと正規表現モードの切り替え
    function toggleMode() {
        simpleMode = !simpleMode;
        
        // simpleBlocksが未初期化の場合は初期化
        if (!simpleBlocks) {
            simpleBlocks = [];
        }
        
        if (simpleMode && simpleBlocks.length === 0) {
            // 簡易モードに切り替えた際、初期ブロックを追加
            addBlock();
        } else {
            // 切り替えだけの場合もイベントを発火
            dispatch("change", { regexp, simpleMode, simpleBlocks });
        }
    }

    // ブロックを追加
    function addBlock() {
        simpleBlocks = [
            ...simpleBlocks,
            { type: "between", start: "", end: "" },
        ];
        updateRegex();
    }

    // ブロックを削除
    function removeBlock(index: number) {
        simpleBlocks = simpleBlocks.filter((_, i) => i !== index);
        updateRegex();
    }

    // ブロックタイプの変更
    function updateBlockType(index: number, newType: string) {
        const block = simpleBlocks[index];
        const newBlock = { type: newType };

        // タイプに応じてフィールドを初期化
        const patternType = patternTypes.find((pt) => pt.id === newType);
        if (patternType) {
            patternType.fields.forEach((field) => {
                newBlock[field] = block[field] || "";
            });
        }

        simpleBlocks[index] = newBlock;
        simpleBlocks = [...simpleBlocks];
        updateRegex();
    }

    // ブロックのフィールド更新
    function updateBlockField(index: number, field: string, value: string) {
        simpleBlocks[index][field] = value;
        simpleBlocks = [...simpleBlocks];
        updateRegex();
    }

    // 簡易入力から正規表現を生成
    function updateRegex() {
        if (!simpleMode || !simpleBlocks || simpleBlocks.length === 0) {
            return;
        }

        let generatedRegex = "";

        for (let i = 0; i < simpleBlocks.length; i++) {
            const block = simpleBlocks[i];
            let blockRegex = "";

            switch (block.type) {
                case "between":
                    if (block.start && block.end) {
                        const escapedStart = escapeRegex(block.start);
                        const escapedEnd = escapeRegex(block.end);
                        blockRegex = `${escapedStart}(.+?)${escapedEnd}`;
                    }
                    break;
                case "contains":
                    if (block.keyword) {
                        const escapedKeyword = escapeRegex(block.keyword);
                        blockRegex = `.*${escapedKeyword}(.*)`;
                    }
                    break;
                case "after":
                    if (block.prefix) {
                        const escapedPrefix = escapeRegex(block.prefix);
                        blockRegex = `${escapedPrefix}(.+)`;
                    }
                    break;
            }

            if (blockRegex) {
                generatedRegex = blockRegex; // 複数ブロックの組み合わせは将来の拡張として
            }
        }

        regexp = generatedRegex;
        dispatch("change", { regexp, simpleMode, simpleBlocks });
    }

    // 正規表現の特殊文字をエスケープ
    function escapeRegex(str: string): string {
        return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    }

    // 正規表現を直接編集
    function handleRegexChange(event: Event) {
        const target = event.target as HTMLInputElement;
        regexp = target.value;
        dispatch("change", { regexp, simpleMode, simpleBlocks });
    }
</script>

<div class="regex-helper space-y-4">
    <!-- モード切替ボタン -->
    <div class="flex items-center justify-between">
        <h3 class="text-sm font-semibold text-gray-300">正規表現パターン</h3>
        {#if !isReadOnly}
            <button
                type="button"
                on:click={toggleMode}
                class="flex items-center gap-2 px-3 py-1.5 bg-dark-200 hover:bg-dark-300 text-white rounded-md transition-all duration-200 text-sm"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    class="h-4 w-4"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4"
                    />
                </svg>
                {simpleMode ? "正規表現モード" : "簡易入力モード"}
            </button>
        {/if}
    </div>

    {#if simpleMode}
        <!-- 簡易入力モード -->
        <div
            class="space-y-3 p-4 bg-purple-900/20 border border-purple-700/50 rounded-lg"
        >
            <div class="flex items-center justify-between mb-2">
                <span class="text-xs text-purple-300 font-medium">簡易入力</span
                >
                {#if !isReadOnly}
                    <button
                        type="button"
                        on:click={addBlock}
                        class="text-xs px-2 py-1 bg-purple-600 hover:bg-purple-700 text-white rounded transition-all duration-200"
                    >
                        + パターン追加
                    </button>
                {/if}
            </div>

            {#each simpleBlocks || [] as block, index}
                <div class="p-3 bg-dark-200 rounded-md space-y-2">
                    <div class="flex items-center justify-between">
                        <select
                            bind:value={block.type}
                            on:change={(e) =>
                                updateBlockType(index, e.currentTarget.value)}
                            disabled={isReadOnly}
                            class="flex-1 px-2 py-1 bg-dark-300 border border-gray-700 rounded text-white text-sm focus:ring-2 focus:ring-purple-500 {isReadOnly ? 'opacity-60 cursor-not-allowed' : ''}"
                        >
                            {#each patternTypes as patternType}
                                <option value={patternType.id}
                                    >{patternType.label}</option
                                >
                            {/each}
                        </select>
                        {#if simpleBlocks.length > 1 && !isReadOnly}
                            <button
                                type="button"
                                on:click={() => removeBlock(index)}
                                class="ml-2 px-2 py-1 bg-red-600 hover:bg-red-700 text-white rounded text-xs transition-all duration-200"
                            >
                                削除
                            </button>
                        {/if}
                    </div>

                    <!-- パターンタイプに応じた入力フィールド -->
                    {#if block.type === "between"}
                        <div class="grid grid-cols-2 gap-2">
                            <input
                                type="text"
                                placeholder="開始文字列"
                                bind:value={block.start}
                                on:input={() =>
                                    updateBlockField(
                                        index,
                                        "start",
                                        block.start,
                                    )}
                                disabled={isReadOnly}
                                class="px-2 py-1.5 bg-dark-300 border border-gray-700 rounded text-white text-sm focus:ring-2 focus:ring-purple-500 {isReadOnly ? 'opacity-60 cursor-not-allowed' : ''}"
                            />
                            <input
                                type="text"
                                placeholder="終了文字列"
                                bind:value={block.end}
                                on:input={() =>
                                    updateBlockField(index, "end", block.end)}
                                disabled={isReadOnly}
                                class="px-2 py-1.5 bg-dark-300 border border-gray-700 rounded text-white text-sm focus:ring-2 focus:ring-purple-500 {isReadOnly ? 'opacity-60 cursor-not-allowed' : ''}"
                            />
                        </div>
                    {:else if block.type === "contains"}
                        <input
                            type="text"
                            placeholder="キーワード"
                            bind:value={block.keyword}
                            on:input={() =>
                                updateBlockField(
                                    index,
                                    "keyword",
                                    block.keyword,
                                )}
                            disabled={isReadOnly}
                            class="w-full px-2 py-1.5 bg-dark-300 border border-gray-700 rounded text-white text-sm focus:ring-2 focus:ring-purple-500 {isReadOnly ? 'opacity-60 cursor-not-allowed' : ''}"
                        />
                    {:else if block.type === "after"}
                        <input
                            type="text"
                            placeholder="プレフィックス"
                            bind:value={block.prefix}
                            on:input={() =>
                                updateBlockField(index, "prefix", block.prefix)}
                            disabled={isReadOnly}
                            class="w-full px-2 py-1.5 bg-dark-300 border border-gray-700 rounded text-white text-sm focus:ring-2 focus:ring-purple-500 {isReadOnly ? 'opacity-60 cursor-not-allowed' : ''}"
                        />
                    {/if}
                </div>
            {/each}

            <!-- 生成されたプレビュー -->
            {#if regexp}
                <div
                    class="mt-3 p-2 bg-dark-300 rounded border border-purple-700/50"
                >
                    <p class="text-xs text-purple-300 mb-1">
                        生成された正規表現:
                    </p>
                    <code class="text-xs text-white font-mono break-all"
                        >{regexp}</code
                    >
                </div>
            {/if}
        </div>
    {:else}
        <!-- 正規表現モード -->
        <div class="space-y-2">
            <input
                type="text"
                placeholder="正規表現を入力 (例: \[Behaviour\] Entering Room: (.+))"
                bind:value={regexp}
                on:input={handleRegexChange}
                disabled={isReadOnly}
                class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200 font-mono text-sm {isReadOnly ? 'opacity-60 cursor-not-allowed' : ''}"
            />
            <p class="text-xs text-gray-500">
                ログから情報を抽出するための正規表現パターンを入力してください
            </p>
        </div>
    {/if}
</div>

<style>
    /* カスタムスタイル */
</style>
