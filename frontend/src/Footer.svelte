<script lang="ts">
    import { onMount, afterUpdate } from "svelte";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    function ClipboardData(text: string) {
        dispatch("clipboardData", text);
    }

    export let noticeLogs = [];

    let footerElement;
    let isExpanded = false;

    // afterUpdateライフサイクルを使用してフッターが更新された後にスクロール
    afterUpdate(() => {
        footerElement.scrollTop = footerElement.scrollHeight;
    });

    function toggleExpand() {
        isExpanded = !isExpanded;
    }

    // ログの種類に基づいてアイコンとカラーを取得
    function getLogTypeStyle(title: string) {
        if (title.includes("SYSTEM")) {
            return {
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>`,
                bgColor: "bg-blue-600"
            };
        } else if (title.includes("ERROR")) {
            return {
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>`,
                bgColor: "bg-red-600"
            };
        } else if (title.includes("WARNING")) {
            return {
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>`,
                bgColor: "bg-yellow-600"
            };
        } else if (title.includes("DISCORD")) {
            return {
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                </svg>`,
                bgColor: "bg-purple-600"
            };
        } else if (title.includes("XS")) {
            return {
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>`,
                bgColor: "bg-green-600"
            };
        } else {
            return {
                icon: `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>`,
                bgColor: "bg-gray-600"
            };
        }
    }
</script>

<footer
    class="{isExpanded ? 'h-[300px]' : 'h-[150px]'} transition-all duration-300 ease-in-out p-4 bg-dark-100 overflow-y-auto text-left relative shadow-lg"
    bind:this={footerElement}
>
    <div class="flex justify-between items-center mb-2">
        <h3 class="text-sm font-semibold text-white flex items-center">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1 text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            アプリケーションログ
        </h3>
        <button 
            class="text-gray-400 hover:text-white transition-colors duration-200"
            on:click={toggleExpand}
        >
            {#if isExpanded}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                </svg>
            {:else}
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
                </svg>
            {/if}
        </button>
    </div>
    
    <div class="space-y-2">
        {#each noticeLogs as noticeLog}
            <div class="flex items-start gap-2 text-sm text-gray-200 p-2 rounded-md bg-dark-200 hover:bg-dark-300 transition-colors duration-200 {noticeLog.title.includes('WARNING') ? 'border-l-4 border-yellow-500' : noticeLog.title.includes('ERROR') ? 'border-l-4 border-red-500' : ''}">
                <div class="flex-shrink-0">
                    <div class="w-6 h-6 rounded-full flex items-center justify-center {getLogTypeStyle(noticeLog.title).bgColor}">
                        {@html getLogTypeStyle(noticeLog.title).icon}
                    </div>
                </div>
                <div class="flex-grow overflow-hidden">
                    <div class="flex items-center">
                        <span class="font-medium text-xs text-primary-300">{noticeLog.title}</span>
                    </div>
                    <p class="text-sm break-words">{noticeLog.text}</p>
                </div>
                {#if noticeLog.canCopy}
                    <button
                        class="flex-shrink-0 ml-2 px-2 py-1 text-xs bg-primary-700 hover:bg-primary-600 rounded-md transition-colors duration-200 flex items-center"
                        on:click={() => ClipboardData(noticeLog.metaData)}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                        </svg>
                        コピー
                    </button>
                {/if}
            </div>
        {/each}
        
        {#if noticeLogs.length === 0}
            <div class="flex justify-center items-center h-20 text-gray-500">
                <p>ログはまだありません</p>
            </div>
        {/if}
    </div>
</footer>
