<script lang="ts">
    export let onBack: () => void;

    const templateVars = [
        { name: "{time}", desc: "ログ行のタイムスタンプ", example: "2024.03.30 12:34:56" },
        { name: "{title}", desc: "ルールのタイトル", example: "ワールド参加" },
        { name: "{description}", desc: "ルールの説明文", example: "ワールドに参加したとき" },
        { name: "{matched_text}", desc: "正規表現で抽出されたテキスト", example: "Grumpy Gamer's World" },
        { name: "{log_file}", desc: "読み取り中のログファイル名", example: "output_log_2024-03-30.txt" },
    ];

    const actionTypes = [
        {
            name: "Webリクエスト",
            type: "WebRequest",
            color: "text-sky-300",
            desc: "指定したURLへJSONをPOST送信します。URLと追加フィールドにテンプレート変数が使えます。",
        },
        {
            name: "Discord Webhook",
            type: "SendDiscordWebHook",
            color: "text-indigo-300",
            desc: "Discord WebhookへメッセージをPOSTします。スクリーンショット時は画像ファイルをマルチパートで送信します。",
        },
        {
            name: "XSOverlay通知",
            type: "SendXSOverlay",
            color: "text-purple-300",
            desc: "XSOverlay（VR HUD）へ通知を送信します。localhost:42070のWebSocketを使用します。",
        },
        {
            name: "テキストファイル出力",
            type: "OutputTextFile",
            color: "text-emerald-300",
            desc: "OutputText/{ルールID}/ フォルダにタイムスタンプ付きの.txtファイルを出力します。",
        },
        {
            name: "ログ出力のみ",
            type: "LogOnly",
            color: "text-gray-300",
            desc: "アプリ内のログ欄に表示するだけで、外部への送信は行いません。動作確認に便利です。",
        },
    ];

    let activeSection: string = "template";
</script>

<div class="flex flex-col h-full">
    <!-- ヘッダー -->
    <div class="flex items-center justify-between px-6 py-4 border-b border-gray-700 flex-shrink-0">
        <div class="flex items-center gap-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-sky-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.746 0 3.332.477 4.5 1.253v13C19.832 18.477 18.246 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
            <h2 class="text-lg font-bold text-white">ヘルプ</h2>
        </div>
        <button
            on:click={onBack}
            class="flex items-center gap-2 px-3 py-1.5 text-sm text-gray-400 hover:text-white hover:bg-gray-700 rounded-lg transition-colors"
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            閉じる
        </button>
    </div>

    <div class="flex flex-1 overflow-hidden">
        <!-- サイドナビ -->
        <nav class="w-48 flex-shrink-0 border-r border-gray-700 py-4 overflow-y-auto">
            <button
                class="w-full text-left px-4 py-2 text-sm transition-colors {activeSection === 'template' ? 'text-sky-300 bg-sky-900/20 border-r-2 border-sky-400' : 'text-gray-400 hover:text-white hover:bg-gray-800'}"
                on:click={() => (activeSection = "template")}
            >
                テンプレート変数
            </button>
            <button
                class="w-full text-left px-4 py-2 text-sm transition-colors {activeSection === 'screenshot' ? 'text-sky-300 bg-sky-900/20 border-r-2 border-sky-400' : 'text-gray-400 hover:text-white hover:bg-gray-800'}"
                on:click={() => (activeSection = "screenshot")}
            >
                スクリーンショット機能
            </button>
            <button
                class="w-full text-left px-4 py-2 text-sm transition-colors {activeSection === 'actions' ? 'text-sky-300 bg-sky-900/20 border-r-2 border-sky-400' : 'text-gray-400 hover:text-white hover:bg-gray-800'}"
                on:click={() => (activeSection = "actions")}
            >
                アクションタイプ
            </button>
        </nav>

        <!-- コンテンツ -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6">

            <!-- テンプレート変数 -->
            {#if activeSection === "template"}
                <div>
                    <h3 class="text-base font-semibold text-white mb-1">テンプレート変数</h3>
                    <p class="text-xs text-gray-400 mb-4">
                        WebリクエストのURL・追加フィールドの値に <code class="px-1 py-0.5 bg-dark-300 rounded text-sky-300">{"{変数名}"}</code> 形式で埋め込めます。
                    </p>

                    <div class="rounded-lg overflow-hidden border border-gray-700 mb-6">
                        <table class="w-full text-xs">
                            <thead>
                                <tr class="bg-dark-300 text-gray-400 uppercase text-left">
                                    <th class="px-3 py-2">変数</th>
                                    <th class="px-3 py-2">内容</th>
                                    <th class="px-3 py-2">例</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each templateVars as v, i}
                                    <tr class="{i % 2 === 0 ? 'bg-dark-200' : 'bg-dark-300/50'} border-t border-gray-700/50">
                                        <td class="px-3 py-2">
                                            <code class="text-sky-300 font-mono">{v.name}</code>
                                        </td>
                                        <td class="px-3 py-2 text-gray-300">{v.desc}</td>
                                        <td class="px-3 py-2 text-gray-500 font-mono">{v.example}</td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>

                    <h4 class="text-sm font-semibold text-gray-300 mb-2">使用例</h4>
                    <div class="bg-dark-300 rounded-lg p-4 text-xs font-mono space-y-1 border border-gray-700">
                        <p class="text-gray-500">// URL</p>
                        <p class="text-gray-200">https://example.com/hook?src=<span class="text-sky-300">{"{log_file}"}</span></p>
                        <p class="mt-3 text-gray-500">// 追加フィールド</p>
                        <p><span class="text-purple-300">"timestamp"</span>: <span class="text-green-300">"{"{time}"}"</span></p>
                        <p><span class="text-purple-300">"label"</span>: <span class="text-green-300">"[{"{title}"}] {"{matched_text}"}"</span></p>
                        <p><span class="text-purple-300">"source_file"</span>: <span class="text-green-300">"{"{log_file}"}"</span></p>
                    </div>

                    <div class="mt-4 p-3 bg-amber-900/20 border border-amber-700/40 rounded-lg text-xs text-amber-300">
                        変数はURLと追加フィールドの<span class="font-semibold">値</span>にのみ展開されます。フィールド名（キー）には展開されません。
                    </div>
                </div>
            {/if}

            <!-- スクリーンショット -->
            {#if activeSection === "screenshot"}
                <div>
                    <h3 class="text-base font-semibold text-white mb-1">スクリーンショット特殊機能</h3>
                    <p class="text-xs text-gray-400 mb-4">
                        <code class="px-1 py-0.5 bg-dark-300 rounded text-purple-300">core-screenshot</code> ルールはVRChatのスクリーンショット検知に特化した追加機能があります。
                    </p>

                    <div class="space-y-4">
                        <div class="p-4 bg-dark-300 rounded-lg border border-gray-700">
                            <p class="text-xs font-semibold text-green-300 mb-2">PNGメタデータの自動抽出</p>
                            <p class="text-xs text-gray-300">
                                スクリーンショットのPNGファイルから <span class="text-green-300 font-medium">World ID</span> と <span class="text-green-300 font-medium">World Display Name</span> を自動的に読み取ります。
                            </p>
                        </div>

                        <div class="p-4 bg-dark-300 rounded-lg border border-gray-700">
                            <p class="text-xs font-semibold text-sky-300 mb-2">Webリクエスト</p>
                            <p class="text-xs text-gray-300 mb-2">画像をBase64エンコードしてJSONで送信します。自動で付与されるフィールド：</p>
                            <div class="flex flex-wrap gap-1.5">
                                {#each ["image", "filename", "mimetype"] as f}
                                    <code class="px-1.5 py-0.5 bg-dark-200 rounded text-purple-300 text-xs">{f}</code>
                                {/each}
                                {#each ["world_id", "world_name"] as f}
                                    <code class="px-1.5 py-0.5 bg-dark-200 rounded text-green-300 text-xs">{f}</code>
                                {/each}
                            </div>
                        </div>

                        <div class="p-4 bg-dark-300 rounded-lg border border-gray-700">
                            <p class="text-xs font-semibold text-indigo-300 mb-2">Discord Webhook</p>
                            <p class="text-xs text-gray-300">
                                画像ファイルをマルチパートフォームで送信します。投稿文に <span class="text-green-300 font-medium">World Display Name</span>（太字）を自動表示します。
                            </p>
                        </div>
                    </div>
                </div>
            {/if}

            <!-- アクションタイプ -->
            {#if activeSection === "actions"}
                <div>
                    <h3 class="text-base font-semibold text-white mb-1">アクションタイプ</h3>
                    <p class="text-xs text-gray-400 mb-4">
                        ログパターンにマッチしたとき実行するアクションを選択できます。
                    </p>

                    <div class="space-y-3">
                        {#each actionTypes as action}
                            <div class="p-4 bg-dark-300 rounded-lg border border-gray-700">
                                <div class="flex items-center gap-2 mb-1.5">
                                    <span class="text-sm font-semibold {action.color}">{action.name}</span>
                                    <code class="text-xs text-gray-500 font-mono">{action.type}</code>
                                </div>
                                <p class="text-xs text-gray-300">{action.desc}</p>
                            </div>
                        {/each}
                    </div>
                </div>
            {/if}

        </div>
    </div>
</div>
