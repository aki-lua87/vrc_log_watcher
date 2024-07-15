<script>
    export let logs;

    import { onMount, afterUpdate } from "svelte";

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
        {#each logs as log}
            <li class="mb-2 text-sm text-rose-50">{log}</li>
        {/each}
    </ul>
</footer>
