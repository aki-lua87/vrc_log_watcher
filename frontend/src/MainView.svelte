<script lang="ts">
    import { fade, fly } from "svelte/transition";
    import { flip } from "svelte/animate";
    import { ClipboardSetText, BrowserOpenURL } from "../wailsjs/runtime";
    import { GetWatcherDebugInfo } from "../wailsjs/go/main/App";
    import { models } from "../wailsjs/go/models";

    export let noticeLogs: models.NoticeLog[] = [];
    export let vrcLogFileName: string = "";
    export let logFolderPath: string = "";
    export let onOpenSettings: () => void;
    export let onGetLogFolder: () => void;
    export let onOpenHelp: () => void;

    // デバッグ情報
    let showDebug = false;
    let debugInfo = {
        lastReadLine: "",
        lastOffset: 0,
        fileSize: 0,
        linesRead: 0,
        lastReadAt: "",
        lastNewLinesAt: "",
        isRunning: false,
        skipCount: 0,
        consecutiveNoProgress: 0,
        refreshCount: 0,
    };

    // デバッグ情報を定期更新
    let debugIntervalId: number | null = null;
    function toggleDebug() {
        showDebug = !showDebug;
        if (showDebug && !debugIntervalId) {
            fetchDebugInfo();
            debugIntervalId = setInterval(fetchDebugInfo, 2000);
        } else if (!showDebug && debugIntervalId) {
            clearInterval(debugIntervalId);
            debugIntervalId = null;
        }
    }
    async function fetchDebugInfo() {
        try {
            debugInfo = await GetWatcherDebugInfo();
        } catch (e) {
            console.error("デバッグ情報取得失敗:", e);
        }
    }

    // サムネイルキャッシュ（ファイルパス → Promise<string>）
    const thumbnailPromises = new Map<string, Promise<string>>();

    function getThumbnail(filePath: string): Promise<string> {
        if (!thumbnailPromises.has(filePath)) {
            const promise = import("../wailsjs/go/main/App").then((AppModule) =>
                AppModule.GetImageThumbnail(filePath)
                    .then((b64) => `data:image/jpeg;base64,${b64}`)
                    .catch(() => "")
            );
            thumbnailPromises.set(filePath, promise);
        }
        return thumbnailPromises.get(filePath)!;
    }

    // URL判定
    function isURL(text: string): boolean {
        try {
            const url = new URL(text);
            return url.protocol === "http:" || url.protocol === "https:";
        } catch {
            return false;
        }
    }

    // コピー機能
    function copyToClipboard(text: string) {
        ClipboardSetText(text)
            .then(() => {
                console.log("クリップボードにコピーしました");
            })
            .catch((err) => {
                console.error("コピーに失敗しました:", err);
            });
    }

    // フィルター
    let filterSettingId: string = "";
    let filterText: string = "";

    $: uniqueSettings = [
        ...new Map(
            noticeLogs
                .filter((l) => l.settingId)
                .map((l) => [l.settingId, { id: l.settingId, title: l.title }])
        ).values(),
    ];

    $: filteredLogs = noticeLogs.filter((log) => {
        if (filterSettingId && log.settingId !== filterSettingId) return false;
        if (filterText && !(log.metaData ?? "").toLowerCase().includes(filterText.toLowerCase())) return false;
        return true;
    });

    $: isFiltering = filterSettingId !== "" || filterText !== "";

    function clearFilters() {
        filterSettingId = "";
        filterText = "";
    }

    // ログタイプに応じたスタイルとアイコンを取得
    function getLogStyle(log: models.NoticeLog) {
        if (log.isError) {
            return {
                bgColor: "bg-red-900/20",
                borderColor: "border-red-700/50",
                iconBg: "bg-red-600",
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>`,
            };
        } else if (log.isSystem) {
            return {
                bgColor: "bg-blue-900/20",
                borderColor: "border-blue-700/50",
                iconBg: "bg-blue-600",
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>`,
            };
        } else if (log.actionSuccess) {
            return {
                bgColor: "bg-green-900/20",
                borderColor: "border-green-700/50",
                iconBg: "bg-green-600",
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>`,
            };
        } else {
            return {
                bgColor: "bg-gray-900/20",
                borderColor: "border-gray-700/50",
                iconBg: "bg-gray-600",
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>`,
            };
        }
    }
</script>

<div class="main-view flex flex-col h-screen bg-dark-200">
    <!-- ヘッダー -->
    <header class="flex justify-between items-center p-4 bg-dark-100 shadow-lg border-b border-gray-800">
        <div class="flex items-center gap-4">
            <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                    </svg>
                </div>
                <div>
                    <h1 class="text-xl font-bold text-white">VRC Log Watcher</h1>
                    <p class="text-xs text-gray-400">リアルタイムログモニター</p>
                </div>
            </div>
            
        </div>

        <!-- ログファイル情報（右側） -->
        <div class="flex items-center gap-3">
            <div class="flex items-center gap-2 px-3 py-2 bg-dark-200 rounded-lg border border-gray-700">
                <div class="text-xs">
                    {#if vrcLogFileName}
                        <div class="text-green-400 font-medium flex items-center gap-1">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                            {vrcLogFileName}
                        </div>
                        {#if logFolderPath}
                            <div class="text-gray-500 truncate max-w-xs" title={logFolderPath}>
                                {logFolderPath}
                            </div>
                        {/if}
                    {:else}
                        <div class="text-red-400 flex items-center gap-1">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                            </svg>
                            ログファイルなし
                        </div>
                    {/if}
                </div>
            </div>
            
            <button
                on:click={onGetLogFolder}
                class="flex items-center gap-2 px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-all duration-200 shadow-md text-sm"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
                <span class="font-medium">ログフォルダ選択</span>
            </button>

            <button
                on:click={onOpenHelp}
                class="px-2 py-2 rounded-lg transition-all duration-200 text-sm bg-dark-200 hover:bg-dark-300 text-gray-500 hover:text-sky-400 border border-gray-700"
                title="ヘルプ"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
            </button>

            <button
                on:click={toggleDebug}
                class="px-2 py-2 rounded-lg transition-all duration-200 text-sm {showDebug ? 'bg-yellow-600 hover:bg-yellow-700 text-white' : 'bg-dark-200 hover:bg-dark-300 text-gray-500 border border-gray-700'}"
                title="デバッグ情報"
            >
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
            </button>
        </div>

    </header>

    <!-- デバッグパネル -->
    {#if showDebug}
        <div class="bg-dark-100 border-b border-yellow-700/50 px-4 py-3 font-mono text-xs">
            <div class="max-w-5xl mx-auto">
                <div class="flex items-center gap-2 mb-2 flex-wrap">
                    <span class="text-yellow-400 font-bold">WATCHER DEBUG</span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">更新: {debugInfo.lastReadAt || "---"}</span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">offset: <span class="text-cyan-400">{debugInfo.lastOffset}</span></span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">fileSize: <span class="text-cyan-400">{debugInfo.fileSize}</span></span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">前回読行数: <span class="text-green-400">{debugInfo.linesRead}</span></span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">最終成功: <span class="text-green-400">{debugInfo.lastNewLinesAt || "---"}</span></span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">実行中: <span class={debugInfo.isRunning ? "text-red-400" : "text-green-400"}>{debugInfo.isRunning ? "Yes" : "No"}</span></span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">スキップ: <span class={debugInfo.skipCount > 0 ? "text-red-400" : "text-gray-400"}>{debugInfo.skipCount}</span></span>
                    <span class="text-gray-500">|</span>
                    <span class="text-gray-400">停滞: <span class={debugInfo.consecutiveNoProgress > 10 ? "text-red-400" : debugInfo.consecutiveNoProgress > 0 ? "text-yellow-400" : "text-gray-400"}>{debugInfo.consecutiveNoProgress}</span></span>
                    {#if debugInfo.refreshCount > 0}
                        <span class="text-gray-500">|</span>
                        <span class="text-orange-400">リフレッシュ: {debugInfo.refreshCount}回</span>
                    {/if}
                </div>
                <div class="text-gray-400">
                    最終行: <span class="text-white break-all">{debugInfo.lastReadLine || "(未読み取り)"}</span>
                </div>
            </div>
        </div>
    {/if}

    <!-- フィルターバー -->
    <div class="bg-dark-100 border-b border-gray-800 px-4 py-1.5">
        <div class="max-w-5xl mx-auto flex items-center justify-end gap-3 flex-wrap">
            <span class="text-xs text-gray-500 shrink-0">フィルター:</span>

            <!-- 設定フィルター -->
            <select
                bind:value={filterSettingId}
                class="text-xs bg-dark-200 border border-gray-700 text-gray-300 rounded-md px-2 py-1 focus:outline-none focus:border-primary-500 min-w-32"
            >
                <option value="">すべての設定</option>
                {#each uniqueSettings as s}
                    <option value={s.id}>{s.title}</option>
                {/each}
            </select>

            <!-- テキストフィルター -->
            <div class="relative min-w-40 max-w-xs">
                <input
                    type="text"
                    bind:value={filterText}
                    placeholder="抽出データで絞り込み..."
                    class="w-full text-xs bg-dark-200 border border-gray-700 text-gray-300 rounded-md px-2 py-1 pr-6 focus:outline-none focus:border-primary-500 placeholder-gray-600"
                />
                {#if filterText}
                    <button
                        on:click={() => (filterText = "")}
                        class="absolute right-1.5 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300"
                        title="クリア"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                {/if}
            </div>

            <!-- フィルター中の件数表示・クリアボタン -->
            {#if isFiltering}
                <span class="text-xs text-primary-400">
                    {filteredLogs.length} / {noticeLogs.length} 件
                </span>
                <button
                    on:click={clearFilters}
                    class="text-xs text-gray-500 hover:text-gray-300 underline"
                >
                    クリア
                </button>
            {/if}
        </div>
    </div>

    <!-- ログフィード -->
    <main class="flex-1 overflow-y-auto p-6 main-scroll">
        <div class="max-w-5xl mx-auto space-y-4">
            {#if noticeLogs.length === 0}
                <div class="flex flex-col items-center justify-center h-96 text-gray-500" transition:fade>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-20 w-20 mb-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <p class="text-lg font-medium">ログはまだありません</p>
                    <p class="text-sm mt-2">VRChatのログ監視が開始されると、ここにイベントが表示されます</p>
                </div>
            {:else if filteredLogs.length === 0}
                <div class="flex flex-col items-center justify-center h-96 text-gray-500" transition:fade>
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-20 w-20 mb-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2a1 1 0 01-.293.707L13 13.414V19a1 1 0 01-.553.894l-4 2A1 1 0 017 21v-7.586L3.293 6.707A1 1 0 013 6V4z" />
                    </svg>
                    <p class="text-lg font-medium">条件に一致するログがありません</p>
                    <button on:click={clearFilters} class="text-sm mt-2 text-primary-400 hover:underline">フィルターをクリア</button>
                </div>
            {:else}
                {#each filteredLogs.slice().reverse() as log, index (log.timestamp + log.text + index)}
                    {@const style = getLogStyle(log)}
                    <div
                        class="log-card p-4 rounded-lg border {style.bgColor} {style.borderColor} hover:shadow-lg transition-all duration-200"
                        animate:flip={{ duration: 300 }}
                        in:fly={{ y: -20, duration: 500 }}
                    >
                        <div class="flex items-center gap-4">
                            <!-- アイコン -->
                            <div class="flex-shrink-0 w-10 h-10 rounded-full {style.iconBg} flex items-center justify-center text-white">
                                {@html style.icon}
                            </div>

                            <!-- コンテンツ -->
                            <div class="flex-1 min-w-0">
                                <!-- タイトルとタイムスタンプ -->
                                <div class="flex items-center justify-between mb-2">
                                    <h3 class="text-sm font-semibold text-primary-300">{log.title}</h3>
                                    {#if log.timestamp}
                                        <span class="text-xs text-gray-500 flex items-center gap-1">
                                            <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                                            </svg>
                                            {log.timestamp}
                                        </span>
                                    {/if}
                                </div>

                                <!-- スクリーンショットサムネイル -->
                                {#if log.settingId === "core-screenshot" && log.metaData}
                                    {#await getThumbnail(log.metaData) then src}
                                        {#if src}
                                            <div class="mt-3">
                                                <button
                                                    on:click={async () => {
                                                        const AppModule = await import("../wailsjs/go/main/App");
                                                        try {
                                                            await AppModule.OpenFile(log.metaData);
                                                        } catch (e) {
                                                            console.error("画像を開けませんでした:", e);
                                                        }
                                                    }}
                                                    class="block rounded-md overflow-hidden border border-purple-700/50 hover:border-purple-400 transition-all duration-200 hover:scale-[1.02] focus:outline-none"
                                                    title="クリックで画像を開く"
                                                >
                                                    <img
                                                        {src}
                                                        alt="スクリーンショット"
                                                        class="max-w-80 max-h-80 w-auto h-auto"
                                                    />
                                                </button>
                                            </div>
                                        {/if}
                                    {:catch}
                                        <!-- 読み込み失敗時は非表示 -->
                                    {/await}
                                {/if}

                                <!-- メタデータ（抽出データ） -->
                                {#if log.metaData && log.canCopy}
                                    <div class="flex items-center gap-2 mt-3 p-3 bg-dark-200 rounded-md border border-gray-700">
                                        <div class="flex-1 min-w-0">
                                            <p class="text-xs text-gray-400 mb-1">抽出データ:</p>
                                            <p class="text-base font-mono text-white truncate">{log.metaData}</p>
                                        </div>
                                        <div class="flex gap-2">
                                            {#if isURL(log.metaData)}
                                                <button
                                                    on:click={() => BrowserOpenURL(log.metaData)}
                                                    class="flex-shrink-0 px-3 py-2 bg-sky-600 hover:bg-sky-700 text-white rounded-md transition-all duration-200 flex items-center gap-1.5 text-sm"
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                                                    </svg>
                                                    ブラウザで開く
                                                </button>
                                            {:else if log.settingId === "core-screenshot"}
                                                <button
                                                    on:click={async () => {
                                                        const AppModule = await import("../wailsjs/go/main/App");
                                                        try {
                                                            await AppModule.OpenFileInExplorer(log.metaData);
                                                        } catch (error) {
                                                            console.error("エクスプローラーで開けませんでした。ファイルを直接開きます:", error);
                                                            try {
                                                                await AppModule.OpenFile(log.metaData);
                                                            } catch (openError) {
                                                                console.error("ファイルを開けませんでした:", openError);
                                                            }
                                                        }
                                                    }}
                                                    class="flex-shrink-0 px-3 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-md transition-all duration-200 flex items-center gap-1.5 text-sm"
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                                    </svg>
                                                    画像を表示
                                                </button>
                                            {:else if log.text.includes("[Log Output]") && log.settingId}
                                                <button
                                                    on:click={async () => {
                                                        const AppModule = await import("../wailsjs/go/main/App");
                                                        const folderPath = await AppModule.GetOutputFolderPath(log.settingId);
                                                        if (folderPath) {
                                                            await AppModule.OpenInExplorer(folderPath);
                                                        }
                                                    }}
                                                    class="flex-shrink-0 px-3 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md transition-all duration-200 flex items-center gap-1.5 text-sm"
                                                >
                                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                                                    </svg>
                                                    フォルダを開く
                                                </button>
                                            {/if}
                                            <button
                                                on:click={() => copyToClipboard(log.metaData)}
                                                class="flex-shrink-0 px-3 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-md transition-all duration-200 flex items-center gap-1.5 text-sm"
                                            >
                                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                                                </svg>
                                                コピー
                                            </button>
                                        </div>
                                    </div>
                                    
                                    <!-- ログ出力メッセージ（囲いの外） -->
                                    {#if log.text.includes("[Log Output]")}
                                        <p class="text-xs text-green-400 mt-2">
                                            {log.text.split(':')[0]}
                                        </p>
                                    {/if}
                                {:else}
                                    <!-- メインテキスト（抽出データがない場合） -->
                                    <p class="text-sm text-gray-200 mb-2 mt-2">{log.text}</p>
                                {/if}
                            </div>
                        </div>
                    </div>
                {/each}
            {/if}
        </div>
    </main>
    
    <!-- 設定ボタン（画面下部中央） -->
    <div class="fixed bottom-6 left-1/2 transform -translate-x-1/2 z-40">
        <button
            on:click={onOpenSettings}
            class="flex items-center gap-2 px-6 py-3 bg-primary-600 hover:bg-primary-700 text-white rounded-full transition-all duration-200 shadow-2xl hover:shadow-primary-900/50 hover:scale-105"
            title="設定を開く"
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
            </svg>
            <span class="font-semibold">設定</span>
        </button>
    </div>
</div>

<style>
    .log-card {
        backdrop-filter: blur(10px);
    }

    /* カスタムスクロールバー */
    .main-scroll::-webkit-scrollbar {
        width: 8px;
    }

    .main-scroll::-webkit-scrollbar-track {
        background: transparent;
    }

    .main-scroll::-webkit-scrollbar-thumb {
        background: rgba(148, 163, 184, 0.3);
        border-radius: 4px;
        transition: background 0.2s ease;
    }

    .main-scroll::-webkit-scrollbar-thumb:hover {
        background: rgba(148, 163, 184, 0.6);
    }
</style>
