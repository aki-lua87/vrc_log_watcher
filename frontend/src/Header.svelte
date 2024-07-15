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

<header class="flex justify-between items-center p-2">
    <p class="text-white">監視ファイル名: {filename}</p>
    <button
        class="h-6 w-40 align-middle text-black rounded bg-white"
        on:click={handleButtonClick}>フォルダを指定</button
    >
</header>
