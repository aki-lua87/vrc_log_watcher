<script>
    import { createEventDispatcher } from "svelte";
    export let contents;
    const dispatch = createEventDispatcher();

    // ドラッグ関連の状態
    let draggedItem = null;
    let draggedOverItem = null;

    function handleSelectContent(content) {
        dispatch("selectContent", content);
    }

    function handleAddContent() {
        dispatch("addContent");
    }

    // ドラッグ開始時のイベントハンドラ
    function handleDragStart(event, content) {
        draggedItem = content;
        event.dataTransfer.effectAllowed = "move";
        event.dataTransfer.setData("text/plain", content.id);
        // ドラッグ中の要素のスタイルを設定
        event.target.classList.add("dragging");
    }

    // ドラッグ終了時のイベントハンドラ
    function handleDragEnd(event) {
        // ドラッグ中のスタイルをリセット
        event.target.classList.remove("dragging");
        draggedItem = null;
        draggedOverItem = null;
    }

    // ドラッグオーバー時のイベントハンドラ
    function handleDragOver(event, content) {
        event.preventDefault();
        if (draggedItem && draggedItem.id !== content.id) {
            draggedOverItem = content;
        }
    }

    // ドロップ時のイベントハンドラ
    function handleDrop(event, targetContent) {
        event.preventDefault();
        if (!draggedItem || draggedItem.id === targetContent.id) return;

        // 元の配列のインデックスを取得
        const draggedIndex = contents.findIndex(item => item.id === draggedItem.id);
        const targetIndex = contents.findIndex(item => item.id === targetContent.id);
        
        if (draggedIndex !== -1 && targetIndex !== -1) {
            // 配列から要素を削除して新しい位置に挿入
            const newContents = [...contents];
            const [removed] = newContents.splice(draggedIndex, 1);
            newContents.splice(targetIndex, 0, removed);
            
            // 親コンポーネントに順序変更を通知
            dispatch("reorderContents", newContents);
        }
        
        // ドラッグ状態をリセット
        draggedItem = null;
        draggedOverItem = null;
    }

    // ドラッグエンター時のイベントハンドラ
    function handleDragEnter(event, content) {
        if (draggedItem && draggedItem.id !== content.id) {
            draggedOverItem = content;
            event.target.classList.add("drag-over");
        }
    }

    // ドラッグリーブ時のイベントハンドラ
    function handleDragLeave(event) {
        event.target.classList.remove("drag-over");
    }

    // 設定タイプに基づいてアイコンを取得する関数
    function getTypeIcon(type) {
        switch (type) {
            case "WebRequest":
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>`;
            case "SendXSOverlay":
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>`;
            case "SendDiscordWebHook":
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                </svg>`;
            case "OutputTextFile":
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>`;
            default:
                return `<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>`;
        }
    }
</script>

<nav
    class="tabs flex flex-col gap-2 p-3 overflow-y-auto w-[240px] bg-dark-100 rounded-lg shadow-card"
>
    <button
        class="flex items-center justify-center bg-secondary-600 text-white p-3 rounded-lg add-button hover:bg-secondary-700 transition-all duration-200 shadow-md"
        on:click={handleAddContent}
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
                d="M12 4v16m8-8H4"
            />
        </svg>
        新しい設定
    </button>

    <div
        class="mt-2 text-xs text-gray-400 font-semibold uppercase tracking-wider pl-2"
    >
        設定一覧
    </div>

    <div class="space-y-1.5 mt-1">
        {#each contents as content}
            <button
                class="w-full flex items-center text-left p-3 rounded-lg hover:bg-dark-200 transition-all duration-200 group border border-transparent hover:border-primary-700"
                on:click={() => handleSelectContent(content)}
                draggable="true"
                on:dragstart={(e) => handleDragStart(e, content)}
                on:dragend={handleDragEnd}
                on:dragover={(e) => handleDragOver(e, content)}
                on:drop={(e) => handleDrop(e, content)}
                on:dragenter={(e) => handleDragEnter(e, content)}
                on:dragleave={handleDragLeave}
                class:dragging={draggedItem && draggedItem.id === content.id}
                class:drag-over={draggedOverItem && draggedOverItem.id === content.id}
            >
                <div
                    class="flex-shrink-0 w-8 h-8 flex items-center justify-center bg-primary-800 rounded-md text-white mr-3"
                >
                    {@html getTypeIcon(content.type)}
                </div>
                <div class="overflow-hidden">
                    <div class="font-medium truncate">
                        {content.title || "Untitled"}
                    </div>
                    <div class="text-xs text-gray-400 truncate">
                        {content.type}
                    </div>
                </div>
            </button>
        {/each}
    </div>

    {#if contents.length === 0}
        <div
            class="flex flex-col items-center justify-center text-center p-4 mt-4 text-gray-500 border border-dashed border-gray-700 rounded-lg"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                class="h-10 w-10 mb-2 text-gray-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                />
            </svg>
            <p class="text-sm">設定がありません</p>
            <p class="text-xs mt-1">
                「新しい設定」ボタンをクリックして作成してください
            </p>
        </div>
    {/if}
</nav>

<style>
    /* ドラッグ中の要素のスタイル */
    .dragging {
        opacity: 0.5;
        border: 2px dashed #4f46e5 !important;
        background-color: rgba(79, 70, 229, 0.1) !important;
    }

    /* ドラッグオーバー中の要素のスタイル */
    .drag-over {
        border: 2px dashed #4f46e5 !important;
        background-color: rgba(79, 70, 229, 0.1) !important;
    }
</style>
