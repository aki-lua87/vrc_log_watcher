<script lang="ts">
    import { dndzone } from "svelte-dnd-action";
    import { flip } from "svelte/animate";
    import RegexHelper from "./components/RegexHelper.svelte";
    import { models } from "../wailsjs/go/models";

    export let settings: models.Setting[] = [];
    export let onBack: () => void;
    export let onUpdate: (setting: models.Setting) => void;
    export let onDelete: (setting: models.Setting) => void;
    export let onAdd: () => void;
    export let onReorder: (newSettings: models.Setting[]) => void;

    let selectedSetting: models.Setting | null = null;
    const flipDurationMs = 200;
    let showDeleteConfirm = false;
    let settingToDelete: models.Setting | null = null;

    // 設定を選択
    function selectSetting(setting: models.Setting) {
        selectedSetting = setting;
    }

    // 新しく追加された設定を自動選択（最後の設定を選択）
    $: if (settings.length > 0 && !selectedSetting) {
        selectedSetting = settings[settings.length - 1];
    }

    // フィールドを更新
    function updateField(field: string, value: any) {
        if (!selectedSetting) return;
        selectedSetting[field] = value;
        selectedSetting = selectedSetting; // リアクティビティをトリガー
        onUpdate(selectedSetting);
    }

    // 正規表現ヘルパーからの変更
    function handleRegexChange(event: CustomEvent) {
        if (!selectedSetting) return;

        const { regexp, simpleMode, simpleBlocks } = event.detail;

        // 値を更新
        selectedSetting.regexp = regexp;
        selectedSetting.simpleMode = simpleMode;
        selectedSetting.simpleBlocks = simpleBlocks;

        // リアクティビティをトリガー
        selectedSetting = selectedSetting;

        // 保存処理を実行
        onUpdate(selectedSetting);
    }

    // 追加フィールドの管理
    function addExtraField() {
        if (!selectedSetting) return;
        if (!selectedSetting.extraFields) {
            selectedSetting.extraFields = {};
        }
        const newKey = `field${Object.keys(selectedSetting.extraFields).length + 1}`;
        selectedSetting.extraFields[newKey] = "";
        selectedSetting = selectedSetting; // リアクティビティをトリガー
        onUpdate(selectedSetting);
    }

    function updateExtraField(
        oldKey: string,
        newKey: string,
        newValue: string,
    ) {
        if (!selectedSetting || !selectedSetting.extraFields) return;

        if (oldKey !== newKey) {
            delete selectedSetting.extraFields[oldKey];
            selectedSetting.extraFields[newKey] = newValue;
        } else {
            selectedSetting.extraFields[oldKey] = newValue;
        }
        selectedSetting = selectedSetting; // リアクティビティをトリガー
        onUpdate(selectedSetting);
    }

    function removeExtraField(key: string) {
        if (!selectedSetting || !selectedSetting.extraFields) return;
        delete selectedSetting.extraFields[key];
        selectedSetting = selectedSetting; // リアクティビティをトリガー
        onUpdate(selectedSetting);
    }

    // ドラッグ&ドロップ
    function handleDndConsider(e: CustomEvent) {
        settings = e.detail.items;
    }

    function handleDndFinalize(e: CustomEvent) {
        settings = e.detail.items;
        onReorder(settings);
    }

    // タイプ値を表示用ラベルに変換
    function getTypeLabel(type: string) {
        const labels: Record<string, string> = {
            WebRequest: "Web リクエスト",
            SendXSOverlay: "XSOverlay へ送信",
            SendDiscordWebHook: "Discord WebHook へ送信",
            OutputTextFile: "テキストへ出力",
            LogOnly: "ログにのみ出力",
            Disable: "何もしない",
        };
        return labels[type] || type;
    }

    // タイプに応じたアイコン
    function getTypeIcon(type: string) {
        const icons = {
            WebRequest: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
            </svg>`,
            SendXSOverlay: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>`,
            SendDiscordWebHook: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
            </svg>`,
            OutputTextFile: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>`,
            LogOnly: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>`,
            Disable: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
            </svg>`,
        };
        return icons[type] || icons.Disable;
    }
</script>

<div class="settings-view flex flex-col h-screen bg-dark-200">
    <!-- ヘッダー -->
    <header class="flex justify-center items-center p-3 bg-dark-100 shadow-lg">
        <button
            on:click={onBack}
            class="flex items-center gap-2 px-6 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-full transition-all duration-200 shadow-md"
            title="設定を閉じる"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-5 w-5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M19 9l-7 7-7-7"
                />
            </svg>
            <span class="font-medium">閉じる</span>
        </button>
    </header>

    <!-- メインコンテンツ: 3カラムレイアウト -->
    <main class="flex-1 flex overflow-hidden">
        <!-- 左カラム: 設定リスト -->
        <aside class="w-64 bg-dark-100 border-r border-gray-800 flex flex-col">
            <div class="p-4 border-b border-gray-800">
                <button
                    on:click={onAdd}
                    class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-secondary-600 hover:bg-secondary-700 text-white rounded-lg transition-all duration-200 shadow-md"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-5 w-5"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M12 4v16m8-8H4"
                        />
                    </svg>
                    新しい設定
                </button>
            </div>

            <div class="flex-1 overflow-y-auto p-4 pb-64 settings-scroll">
                <h3
                    class="text-xs text-gray-400 font-semibold uppercase tracking-wider mb-3"
                >
                    設定一覧
                </h3>

                <div
                    class="space-y-2"
                    use:dndzone={{ items: settings, flipDurationMs }}
                    on:consider={handleDndConsider}
                    on:finalize={handleDndFinalize}
                >
                    {#each settings as setting (setting.id)}
                        <div animate:flip={{ duration: flipDurationMs }}>
                            <button
                                on:click={() => selectSetting(setting)}
                                class="w-full text-left p-3 rounded-lg transition-all duration-200 {selectedSetting?.id ===
                                setting.id
                                    ? setting.isCore
                                        ? 'bg-purple-700 border border-purple-500'
                                        : 'bg-primary-700 border border-primary-500'
                                    : setting.isCore
                                      ? 'bg-purple-900/30 hover:bg-purple-800/40 border border-purple-700/50 hover:border-purple-600/70'
                                      : 'bg-dark-200 hover:bg-dark-300 border border-transparent hover:border-gray-700'}"
                            >
                                <div class="flex items-center gap-2 mb-1">
                                    <div
                                        class="w-6 h-6 flex items-center justify-center bg-dark-300 rounded text-gray-400"
                                    >
                                        {@html getTypeIcon(setting.type)}
                                    </div>
                                    <div class="flex-1 min-w-0">
                                        <p
                                            class="text-sm font-medium truncate {selectedSetting?.id ===
                                            setting.id
                                                ? 'text-white'
                                                : setting.isCore
                                                  ? 'text-purple-200'
                                                  : 'text-gray-200'}"
                                        >
                                            {setting.title || "無題"}
                                        </p>
                                    </div>
                                </div>
                                <div class="flex items-center gap-2 text-xs">
                                    <span
                                        class={setting.isCore
                                            ? "text-purple-300"
                                            : "text-gray-400"}
                                        >{getTypeLabel(setting.type)}</span
                                    >
                                </div>
                            </button>
                        </div>
                    {/each}
                </div>

                {#if settings.length === 0}
                    <div class="text-center text-gray-500 text-sm mt-8">
                        <p>設定がありません</p>
                    </div>
                {/if}
            </div>
        </aside>

        {#if selectedSetting}
            <!-- 中央カラム: Hook設定 -->
            <div
                class="flex-1 bg-dark-200 border-r border-gray-800 flex flex-col overflow-hidden"
            >
                <div
                    class="flex-1 overflow-y-auto settings-scroll-hidden p-6 pb-64"
                >
                    <div class="max-w-2xl">
                        <div class="mb-6">
                            <div class="flex items-center gap-2 mb-4">
                                <div
                                    class="w-8 h-8 bg-purple-600 rounded-lg flex items-center justify-center"
                                >
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        class="h-5 w-5 text-white"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        stroke="currentColor"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M13 10V3L4 14h7v7l9-11h-7z"
                                        />
                                    </svg>
                                </div>
                                <h2 class="text-lg font-bold text-white">
                                    トリガー設定
                                </h2>
                            </div>
                            <p class="text-sm text-gray-400">
                                ログから情報を抽出するトリガー条件を設定します
                            </p>
                        </div>

                        <div class="space-y-5">
                            <!-- 基本機能の警告 -->
                            {#if selectedSetting.isCore}
                                <div
                                    class="p-3 bg-purple-900/20 border border-purple-700/50 rounded-lg"
                                >
                                    <p
                                        class="text-sm text-purple-300 flex items-center gap-2"
                                    >
                                        <svg
                                            xmlns="http://www.w3.org/2000/svg"
                                            class="h-5 w-5"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            stroke="currentColor"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                            />
                                        </svg>
                                        基本機能のため、一部の設定は編集できません
                                    </p>
                                </div>
                            {/if}

                            <!-- スクリーンショット専用の注釈 -->
                            {#if selectedSetting.id === "core-screenshot"}
                                <div class="flex items-center gap-2 px-3 py-2 bg-blue-900/20 border border-blue-700/40 rounded-lg text-xs text-blue-300">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                    </svg>
                                    <span>📸 スクリーンショット特殊機能あり — 詳しくはヘルプを参照</span>
                                </div>
                            {/if}

                            <!-- タイトル -->
                            <div class="space-y-2">
                                <label
                                    class="block text-sm font-medium text-gray-300"
                                    >タイトル</label
                                >
                                <input
                                    type="text"
                                    placeholder="設定のタイトル"
                                    bind:value={selectedSetting.title}
                                    on:input={() =>
                                        updateField(
                                            "title",
                                            selectedSetting.title,
                                        )}
                                    disabled={selectedSetting.isCore}
                                    class="w-full p-3 bg-dark-300 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition-all duration-200 {selectedSetting.isCore
                                        ? 'opacity-60 cursor-not-allowed'
                                        : ''}"
                                />
                            </div>

                            <!-- 説明 -->
                            <div class="space-y-2">
                                <label
                                    class="block text-sm font-medium text-gray-300"
                                    >説明</label
                                >
                                <input
                                    type="text"
                                    placeholder="この設定の説明"
                                    bind:value={selectedSetting.details}
                                    on:input={() =>
                                        updateField(
                                            "details",
                                            selectedSetting.details,
                                        )}
                                    disabled={selectedSetting.isCore}
                                    class="w-full p-3 bg-dark-300 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition-all duration-200 {selectedSetting.isCore
                                        ? 'opacity-60 cursor-not-allowed'
                                        : ''}"
                                />
                            </div>

                            <!-- 正規表現ヘルパー -->
                            <RegexHelper
                                regexp={selectedSetting.regexp}
                                simpleMode={selectedSetting.simpleMode}
                                simpleBlocks={selectedSetting.simpleBlocks}
                                isReadOnly={selectedSetting.isCore}
                                on:change={handleRegexChange}
                            />

                            <!-- 除外条件 -->
                            <div class="space-y-2">
                                <label
                                    class="block text-sm font-medium text-gray-300"
                                    >除外条件</label
                                >
                                <input
                                    type="text"
                                    placeholder="この文字列と一致する場合はスキップ"
                                    bind:value={selectedSetting.exclude}
                                    on:input={() =>
                                        updateField(
                                            "exclude",
                                            selectedSetting.exclude,
                                        )}
                                    class="w-full p-3 bg-dark-300 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition-all duration-200"
                                />
                                <p class="text-xs text-gray-500">
                                    抽出結果がこのテキストと一致する場合、アクションをスキップします
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- 右カラム: Action設定 -->
            <div class="flex-1 bg-dark-200 flex flex-col overflow-hidden">
                <div
                    class="flex-1 overflow-y-auto settings-scroll-hidden p-6 pb-64"
                >
                    <div class="max-w-2xl">
                        <div class="mb-6">
                            <div class="flex items-center gap-2 mb-4">
                                <div
                                    class="w-8 h-8 bg-green-600 rounded-lg flex items-center justify-center"
                                >
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        class="h-5 w-5 text-white"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        stroke="currentColor"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            stroke-width="2"
                                            d="M5 13l4 4L19 7"
                                        />
                                    </svg>
                                </div>
                                <h2 class="text-lg font-bold text-white">
                                    アクション設定
                                </h2>
                            </div>
                            <p class="text-sm text-gray-400">
                                トリガー条件が満たされた際の実行アクションを設定します
                            </p>
                        </div>

                        <div class="space-y-5">
                            <!-- アクションタイプ -->
                            <div class="space-y-2">
                                <label
                                    class="block text-sm font-medium text-gray-300"
                                    >アクションタイプ</label
                                >
                                <select
                                    bind:value={selectedSetting.type}
                                    on:change={() =>
                                        updateField(
                                            "type",
                                            selectedSetting.type,
                                        )}
                                    class="w-full p-3 bg-dark-300 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-500 focus:border-green-500 transition-all duration-200"
                                >
                                    <option value="WebRequest"
                                        >Web リクエスト</option
                                    >
                                    <option value="SendXSOverlay"
                                        >XSOverlay へ送信</option
                                    >
                                    <option value="SendDiscordWebHook"
                                        >Discord WebHook へ送信</option
                                    >
                                    <option value="OutputTextFile"
                                        >テキストへ出力</option
                                    >
                                    <option value="LogOnly">ログに出力</option>
                                    <option value="Disable">何もしない</option>
                                </select>
                            </div>

                            {#if selectedSetting.type === "WebRequest" || selectedSetting.type === "SendDiscordWebHook"}
                                <!-- URL -->
                                <div class="space-y-2">
                                    <label
                                        class="block text-sm font-medium text-gray-300"
                                        >URL</label
                                    >
                                    <input
                                        type="text"
                                        placeholder="https://example.com/api"
                                        bind:value={selectedSetting.url}
                                        on:input={() =>
                                            updateField(
                                                "url",
                                                selectedSetting.url,
                                            )}
                                        class="w-full p-3 bg-dark-300 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-all duration-200 font-mono"
                                    />
                                </div>
                            {/if}

                            {#if selectedSetting.type === "OutputTextFile"}
                                <!-- テキストファイル出力先 -->
                                <div class="space-y-2">
                                    <label
                                        class="block text-sm font-medium text-gray-300"
                                        >出力先フォルダ</label
                                    >
                                    <div class="flex gap-2">
                                        <div
                                            class="flex-1 p-3 bg-dark-300 border border-gray-700 rounded-lg text-gray-400 flex items-center"
                                        >
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                class="h-5 w-5 mr-2"
                                                fill="none"
                                                viewBox="0 0 24 24"
                                                stroke="currentColor"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
                                                />
                                            </svg>
                                            <span class="text-sm"
                                                >./OutputText/{selectedSetting.id}/</span
                                            >
                                        </div>
                                        <button
                                            type="button"
                                            on:click={async () => {
                                                const AppModule = await import(
                                                    "../wailsjs/go/main/App"
                                                );
                                                const folderPath =
                                                    await AppModule.GetOutputFolderPath(
                                                        selectedSetting.id,
                                                    );
                                                if (folderPath) {
                                                    await AppModule.OpenInExplorer(
                                                        folderPath,
                                                    );
                                                } else {
                                                    alert(
                                                        "出力フォルダがまだ作成されていません。ログが出力されると自動的に作成されます。",
                                                    );
                                                }
                                            }}
                                            class="px-4 py-3 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-all duration-200 flex items-center gap-2"
                                        >
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                class="h-5 w-5"
                                                fill="none"
                                                viewBox="0 0 24 24"
                                                stroke="currentColor"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                                                />
                                            </svg>
                                            開く
                                        </button>
                                    </div>
                                    <p class="text-xs text-gray-500">
                                        ファイルは実行フォルダ内の OutputText
                                        フォルダに設定IDごとに自動的に分類されます
                                    </p>
                                </div>
                            {/if}

                            {#if selectedSetting.type === "WebRequest"}
                                <!-- メッセージキー -->
                                <div class="space-y-2">
                                    <label
                                        class="block text-sm font-medium text-gray-300"
                                        >メッセージキー名</label
                                    >
                                    <input
                                        type="text"
                                        placeholder="message"
                                        bind:value={selectedSetting.messageKey}
                                        on:input={() =>
                                            updateField(
                                                "messageKey",
                                                selectedSetting.messageKey,
                                            )}
                                        class="w-full p-3 bg-dark-300 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-green-500 focus:border-green-500 transition-all duration-200 font-mono"
                                    />
                                    <p class="text-xs text-gray-500">
                                        JSONペイロードでログメッセージを送信するキー名（空白の場合は"message"）
                                    </p>
                                </div>

                                <!-- 追加フィールド -->
                                <div class="space-y-2">
                                    <label
                                        class="block text-sm font-medium text-gray-300"
                                        >追加フィールド</label
                                    >
                                    <div class="space-y-2">
                                        {#if selectedSetting.extraFields && Object.keys(selectedSetting.extraFields).length > 0}
                                            {#each Object.entries(selectedSetting.extraFields) as [originalKey, originalValue], index (originalKey)}
                                                {@const key = Object.keys(
                                                    selectedSetting.extraFields,
                                                )[index]}
                                                {@const value =
                                                    selectedSetting.extraFields[
                                                        key
                                                    ]}
                                                <div class="flex gap-2">
                                                    <input
                                                        type="text"
                                                        placeholder="キー名"
                                                        value={key}
                                                        on:input={(e) =>
                                                            updateExtraField(
                                                                originalKey,
                                                                e.currentTarget
                                                                    .value,
                                                                value,
                                                            )}
                                                        class="flex-1 p-2 bg-dark-300 border border-gray-700 rounded text-white focus:ring-2 focus:ring-green-500 font-mono text-sm"
                                                    />
                                                    <input
                                                        type="text"
                                                        placeholder="値"
                                                        {value}
                                                        on:input={(e) =>
                                                            updateExtraField(
                                                                key,
                                                                key,
                                                                e.currentTarget
                                                                    .value,
                                                            )}
                                                        class="flex-1 p-2 bg-dark-300 border border-gray-700 rounded text-white focus:ring-2 focus:ring-green-500 text-sm"
                                                    />
                                                    <button
                                                        type="button"
                                                        on:click={() =>
                                                            removeExtraField(
                                                                key,
                                                            )}
                                                        class="px-3 py-2 bg-red-600 hover:bg-red-700 text-white rounded transition-all duration-200"
                                                    >
                                                        削除
                                                    </button>
                                                </div>
                                            {/each}
                                        {/if}
                                        <button
                                            type="button"
                                            on:click={addExtraField}
                                            class="w-full p-2 bg-green-600 hover:bg-green-700 text-white rounded transition-all duration-200 flex items-center justify-center gap-1"
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
                                                    d="M12 6v6m0 0v6m0-6h6m-6 0H6"
                                                />
                                            </svg>
                                            フィールドを追加
                                        </button>
                                    </div>
                                    <p class="text-xs text-gray-500">
                                        リクエストに追加で送信するキーと値のペア
                                    </p>
                                </div>
                            {/if}

                            <!-- 削除ボタン（基本機能の場合は非表示） -->
                            {#if !selectedSetting.isCore}
                                <div class="pt-4 border-t border-gray-700">
                                    <button
                                        on:click={() => {
                                            settingToDelete = selectedSetting;
                                            showDeleteConfirm = true;
                                        }}
                                        class="w-full flex items-center justify-center gap-2 px-4 py-3 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-all duration-200"
                                    >
                                        <svg
                                            xmlns="http://www.w3.org/2000/svg"
                                            class="h-5 w-5"
                                            fill="none"
                                            viewBox="0 0 24 24"
                                            stroke="currentColor"
                                        >
                                            <path
                                                stroke-linecap="round"
                                                stroke-linejoin="round"
                                                stroke-width="2"
                                                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                                            />
                                        </svg>
                                        この設定を削除
                                    </button>
                                </div>
                            {:else}
                                <div class="pt-4 border-t border-gray-700">
                                    <div
                                        class="p-3 bg-purple-900/20 border border-purple-700/50 rounded-lg"
                                    >
                                        <p
                                            class="text-sm text-purple-300 flex items-center gap-2"
                                        >
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                class="h-5 w-5"
                                                fill="none"
                                                viewBox="0 0 24 24"
                                                stroke="currentColor"
                                            >
                                                <path
                                                    stroke-linecap="round"
                                                    stroke-linejoin="round"
                                                    stroke-width="2"
                                                    d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                                                />
                                            </svg>
                                            基本機能のため削除できません。
                                        </p>
                                    </div>
                                </div>
                            {/if}
                        </div>
                    </div>
                </div>
            </div>
        {:else}
            <div
                class="flex-1 flex items-center justify-center bg-dark-200 text-gray-500"
            >
                <div class="text-center">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-20 w-20 mx-auto mb-4 text-gray-600"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                    >
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="1.5"
                            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                        />
                        <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="1.5"
                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                        />
                    </svg>
                    <p class="text-lg">設定を選択してください</p>
                    <p class="text-sm mt-2">
                        左側のリストから設定を選択するか、新しい設定を作成してください
                    </p>
                </div>
            </div>
        {/if}
    </main>

    <!-- 削除確認モーダル -->
    {#if showDeleteConfirm && settingToDelete}
        <div
            class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50"
            on:click={() => {
                showDeleteConfirm = false;
                settingToDelete = null;
            }}
        >
            <div
                class="bg-dark-100 rounded-xl shadow-2xl max-w-md w-full mx-4 border border-gray-700"
                on:click={(e) => e.stopPropagation()}
            >
                <!-- ヘッダー -->
                <div class="p-6 border-b border-gray-700">
                    <div class="flex items-center gap-3">
                        <div
                            class="w-12 h-12 bg-red-600/20 rounded-full flex items-center justify-center"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                class="h-6 w-6 text-red-500"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    stroke-width="2"
                                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                                />
                            </svg>
                        </div>
                        <div>
                            <h3 class="text-lg font-bold text-white">
                                設定を削除
                            </h3>
                            <p class="text-sm text-gray-400 mt-1">
                                この操作は取り消せません
                            </p>
                        </div>
                    </div>
                </div>

                <!-- 本文 -->
                <div class="p-6">
                    <p class="text-gray-300">
                        以下の設定を削除してもよろしいですか？
                    </p>
                    <div
                        class="mt-4 p-4 bg-dark-200 rounded-lg border border-gray-700"
                    >
                        <p class="text-white font-medium">
                            {settingToDelete.title || "無題"}
                        </p>
                        {#if settingToDelete.details}
                            <p class="text-sm text-gray-400 mt-1">
                                {settingToDelete.details}
                            </p>
                        {/if}
                    </div>
                </div>

                <!-- ボタン -->
                <div class="p-6 border-t border-gray-700 flex gap-3">
                    <button
                        on:click={() => {
                            showDeleteConfirm = false;
                            settingToDelete = null;
                        }}
                        class="flex-1 px-4 py-3 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-all duration-200 font-medium"
                    >
                        キャンセル
                    </button>
                    <button
                        on:click={() => {
                            if (settingToDelete) {
                                onDelete(settingToDelete);
                                showDeleteConfirm = false;
                                settingToDelete = null;
                            }
                        }}
                        class="flex-1 px-4 py-3 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-all duration-200 font-medium flex items-center justify-center gap-2"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            class="h-5 w-5"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                        >
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                        </svg>
                        削除する
                    </button>
                </div>
            </div>
        </div>
    {/if}
</div>

<style>
    /* カスタムスクロールバー - 左サイドバー用 */
    .settings-scroll::-webkit-scrollbar {
        width: 6px;
    }

    .settings-scroll::-webkit-scrollbar-track {
        background: transparent;
    }

    .settings-scroll::-webkit-scrollbar-thumb {
        background: rgba(148, 163, 184, 0.3);
        border-radius: 3px;
        transition: background 0.2s ease;
    }

    .settings-scroll::-webkit-scrollbar-thumb:hover {
        background: rgba(148, 163, 184, 0.6);
    }

    /* スクロールバー非表示 - Hook設定とAction設定用 */
    .settings-scroll-hidden {
        overflow-y: auto;
        scrollbar-width: none; /* Firefox */
        -ms-overflow-style: none; /* IE and Edge */
    }

    .settings-scroll-hidden::-webkit-scrollbar {
        display: none; /* Chrome, Safari, Opera */
    }
</style>
