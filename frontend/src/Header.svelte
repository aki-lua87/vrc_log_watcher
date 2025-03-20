<script>
    import { createEventDispatcher } from "svelte";
    export let filename;
    const dispatch = createEventDispatcher();

    function handleButtonClick() {
        dispatch(
            "getLogFolderPath",
            `Header button clicked at ${new Date().toLocaleTimeString()}`,
        );
    }
</script>

<header class="flex justify-between items-center p-4 bg-dark-100 shadow-md">
    <div class="flex items-center">
        <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-6 w-6 text-primary-400 mr-2"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
        >
            <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
        </svg>
        <div>
            <h1 class="text-xl font-bold text-white">VRC Log Watcher</h1>
            {#if filename}
                <p class="text-sm text-gray-400">監視中: {filename}</p>
            {:else}
                <p class="text-sm text-red-400 flex items-center">
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        class="h-4 w-4 mr-1"
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
                    ログファイルが見つかりません
                </p>
            {/if}
        </div>
    </div>
    <button
        class="flex items-center px-4 py-2 bg-primary-600 hover:bg-primary-700 text-white rounded-lg transition-all duration-200 shadow-md"
        on:click={handleButtonClick}
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
        フォルダを指定
    </button>
</header>
