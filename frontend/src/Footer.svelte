<script lang="ts">
    import { onMount, afterUpdate } from "svelte";
    // import { main } from "wailsjs/go/models";
    import { createEventDispatcher } from "svelte";

    const dispatch = createEventDispatcher();

    function ClipboardData(text: string) {
        dispatch("clipboardData", text);
    }

    export let noticeLogs = [];

    let footerElement;

    // afterUpdateライフサイクルを使用してフッターが更新された後にスクロール
    afterUpdate(() => {
        footerElement.scrollTop = footerElement.scrollHeight;
    });
</script>

<footer
    class="p-4 bg-black bg-opacity-70 h-[150px] overflow-y-auto text-left"
    bind:this={footerElement}
>
    <ul class="list-none p-0 m-0">
        {#each noticeLogs as noticeLog}
            <li class="flex items-center mb-2 text-sm text-rose-50">
                <span>{noticeLog.text}</span>
                {#if noticeLog.canCopy}
                    <button
                        class="ml-2 px-2 py-1 text-xs bg-blue-700 rounded"
                        on:click={() => ClipboardData(noticeLog.metaData)}
                    >
                        Copy
                    </button>
                {/if}
            </li>
        {/each}
    </ul>
</footer>
