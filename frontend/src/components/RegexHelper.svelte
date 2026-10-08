<script lang="ts">
    import { createEventDispatcher, onMount } from "svelte";
    import Icon from "../lib/Icon.svelte";

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
            label: "2つの文字の間を取り出す",
            fields: ["start", "end"],
        },
        {
            id: "after",
            label: "ある文字のうしろを取り出す",
            fields: ["prefix"],
        },
        {
            id: "contains",
            label: "ある文字を含む行を取り出す",
            fields: ["keyword"],
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
        simpleBlocks = [...simpleBlocks, { type: "between", start: "", end: "" }];
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

<div class="rh">
    {#if isReadOnly}
        <div class="ro">
            <span class="lb">このパターンに一致する行</span>
            <code>{regexp}</code>
        </div>
    {:else if simpleMode}
        {#each simpleBlocks || [] as block, index}
            <div class="blk">
                <div class="blk-h">
                    <select
                        class="inp"
                        value={block.type}
                        on:change={(e) => updateBlockType(index, e.currentTarget.value)}
                        aria-label="取り出し方"
                    >
                        {#each patternTypes as patternType}
                            <option value={patternType.id}>{patternType.label}</option>
                        {/each}
                    </select>
                    {#if simpleBlocks.length > 1}
                        <button type="button" class="ib" title="このパターンを削除" on:click={() => removeBlock(index)}>
                            <Icon name="x" />
                        </button>
                    {/if}
                </div>

                <!-- パターンタイプに応じた入力フィールド -->
                {#if block.type === "between"}
                    <div class="pair">
                        <input
                            class="inp mono"
                            type="text"
                            placeholder="前の文字（例: Entering Room: ）"
                            value={block.start ?? ""}
                            on:input={(e) => updateBlockField(index, "start", e.currentTarget.value)}
                        />
                        <span class="cap"><Icon name="scan" size={13} />取り出す</span>
                        <input
                            class="inp mono"
                            type="text"
                            placeholder="うしろの文字"
                            value={block.end ?? ""}
                            on:input={(e) => updateBlockField(index, "end", e.currentTarget.value)}
                        />
                    </div>
                {:else if block.type === "contains"}
                    <input
                        class="inp mono"
                        type="text"
                        placeholder="含まれる文字"
                        value={block.keyword ?? ""}
                        on:input={(e) => updateBlockField(index, "keyword", e.currentTarget.value)}
                    />
                {:else if block.type === "after"}
                    <div class="pair two">
                        <input
                            class="inp mono"
                            type="text"
                            placeholder="この文字のうしろ（例: OnPlayerLeft ）"
                            value={block.prefix ?? ""}
                            on:input={(e) => updateBlockField(index, "prefix", e.currentTarget.value)}
                        />
                        <span class="cap"><Icon name="scan" size={13} />取り出す</span>
                    </div>
                {/if}
            </div>
        {/each}
        {#if regexp}
            <p class="hint">使われる正規表現: <span class="mono">{regexp}</span></p>
        {/if}
    {:else}
        <input
            class="inp mono"
            type="text"
            placeholder="例: \[Behaviour\] Entering Room: (.+)"
            value={regexp}
            on:input={handleRegexChange}
            aria-label="正規表現"
        />
        <p class="hint">( ) で囲んだ部分が「取り出した値」になります。</p>
    {/if}

    {#if !isReadOnly}
        <div class="tools">
            {#if simpleMode}
                <button type="button" class="btn sm ghost" on:click={addBlock}><Icon name="plus" />パターンを追加</button>
            {/if}
            <button type="button" class="btn sm ghost" on:click={toggleMode}>
                {simpleMode ? "正規表現で書く" : "かんたん入力に戻す"}
            </button>
        </div>
    {/if}
</div>

<style>
    .rh {
        display: flex;
        flex-direction: column;
        gap: 8px;
    }
    .ro {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }
    .ro code {
        font-family: var(--font-data);
        font-size: 12.5px;
        padding: 8px 12px;
        border-radius: var(--radius-s);
        background: var(--elev);
        border: 1px solid var(--line);
        overflow-wrap: anywhere;
    }
    .blk {
        display: flex;
        flex-direction: column;
        gap: 8px;
        padding: 12px;
        border-radius: var(--radius-s);
        background: var(--bg);
        border: 1px solid var(--line);
    }
    .blk-h {
        display: flex;
        gap: 6px;
        align-items: center;
    }
    .blk-h .inp {
        flex: 1;
    }
    .pair {
        display: grid;
        grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
        gap: 8px;
        align-items: center;
    }
    .pair.two {
        grid-template-columns: minmax(0, 1fr) auto;
    }
    .cap {
        display: inline-flex;
        gap: 5px;
        align-items: center;
        font-size: 12px;
        padding: 5px 10px;
        border-radius: var(--radius-s);
        background: var(--accent-soft);
        color: var(--accent-ink);
        white-space: nowrap;
    }
    .tools {
        display: flex;
        gap: 4px;
        flex-wrap: wrap;
    }
    .hint .mono {
        overflow-wrap: anywhere;
    }
</style>
