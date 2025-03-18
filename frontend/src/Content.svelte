<script>
    import { createEventDispatcher } from "svelte";
    export let content;
    const dispatch = createEventDispatcher();

    function updateContent(key, value) {
        content = { ...content, [key]: value };
        dispatch("updateContent", content);
    }

    function handleInput(event, key) {
        const target = event.target;
        updateContent(key, target.value);
    }

    function deleteContent() {
        dispatch("deleteContent", content);
    }

    // イベントタイプに基づいてアイコンを取得する関数
    function getTypeIcon(type) {
        switch(type) {
            case 'WebRequest':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>`;
            case 'SendXSOverlay':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>`;
            case 'SendDiscordWebHook':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                </svg>`;
            case 'OutputTextFile':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>`;
            default:
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>`;
        }
    }
</script>

<div class="content flex flex-col gap-4 p-6 bg-dark-100 rounded-lg shadow-card flex-grow overflow-y-auto">
    <div class="flex items-center mb-2">
        <div class="flex-shrink-0 w-10 h-10 flex items-center justify-center bg-primary-700 rounded-md text-white mr-3">
            {@html getTypeIcon(content.type)}
        </div>
        <div>
            <h2 class="text-xl font-bold text-white">{content.title || "Untitled"}</h2>
            <p class="text-sm text-gray-400">{content.type}</p>
        </div>
    </div>

    <div class="grid grid-cols-1 gap-5">
        <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300" for="Title">タイトル</label>
            <input
                id="Title"
                type="text"
                placeholder="タイトルを入力"
                bind:value={content.title}
                on:change={(e) => handleInput(e, "title")}
                class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200"
            />
        </div>

        <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300" for="Details">説明</label>
            <input
                id="Details"
                type="text"
                placeholder="説明を入力"
                bind:value={content.details}
                on:change={(e) => handleInput(e, "details")}
                class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200"
            />
        </div>

        <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300" for="RegExp">
                <div class="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                    </svg>
                    正規表現抽出
                </div>
            </label>
            <input
                id="RegExp"
                type="text"
                placeholder="正規表現を入力"
                bind:value={content.regexp}
                on:change={(e) => handleInput(e, "regexp")}
                class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200 font-mono"
            />
            <p class="text-xs text-gray-500">ログから情報を抽出するための正規表現パターンを入力してください</p>
        </div>

        <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300" for="Exclude">
                <div class="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                    </svg>
                    除外条件
                </div>
            </label>
            <input
                id="Exclude"
                type="text"
                placeholder="除外するテキストを入力"
                bind:value={content.exclude}
                on:change={(e) => handleInput(e, "exclude")}
                class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200"
            />
            <p class="text-xs text-gray-500">抽出結果がこのテキストと一致する場合、アクションをスキップします</p>
        </div>

        <div class="space-y-2">
            <label class="block text-sm font-medium text-gray-300" for="EventType">
                <div class="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                    アクションタイプ
                </div>
            </label>
            <select 
                bind:value={content.type} 
                on:change={(e) => handleInput(e, "type")}
                class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200"
            >
                <option value="WebRequest">Web Request</option>
                <option value="SendXSOverlay">Send XSOverlay</option>
                <option value="SendDiscordWebHook">Send Discord WebHook</option>
                <option value="OutputTextFile">Output Text File</option>
                <option value="Disable">何もしない</option>
            </select>
        </div>

        {#if content.type === "WebRequest" || content.type === "SendDiscordWebHook"}
            <div class="space-y-2">
                <label class="block text-sm font-medium text-gray-300" for="url-input">
                    <div class="flex items-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                        </svg>
                        URL
                    </div>
                </label>
                <input
                    id="url-input"
                    type="text"
                    placeholder="https://example.com/api"
                    bind:value={content.url}
                    on:change={(e) => handleInput(e, "url")}
                    class="w-full p-3 bg-dark-200 border border-gray-700 rounded-lg text-white focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-all duration-200 font-mono"
                />
                <p class="text-xs text-gray-500">
                    {content.type === "WebRequest" ? "POSTリクエストを送信するURL" : "Discord WebhookのURL"}
                </p>
            </div>
        {/if}
    </div>

    <div class="flex justify-end mt-6">
        <button
            class="flex items-center px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-all duration-200 shadow-md"
            on:click={deleteContent}
        >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            削除
        </button>
    </div>
</div>
