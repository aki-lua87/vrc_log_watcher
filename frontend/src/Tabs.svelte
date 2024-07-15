<script>
    import { createEventDispatcher } from "svelte";
    export let contents;
    const dispatch = createEventDispatcher();

    function handleSelectContent(content) {
        dispatch("selectContent", content);
    }

    function handleAddContent() {
        dispatch("addContent");
    }
</script>

<nav class="tabs flex flex-col gap-2 p-2 overflow-y-auto w-[200px]">
    <button
        class="bg-green-500 text-white p-2 rounded add-button hover:bg-green-600"
        on:click={handleAddContent}>+</button
    >
    {#each contents as content}
        <button
            class="bg-blue-500 text-white p-2 rounded hover:bg-blue-700"
            on:click={() => handleSelectContent(content)}
        >
            {content.title || "Untitled"}
        </button>
    {/each}
</nav>
