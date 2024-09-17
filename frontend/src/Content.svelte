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
</script>

<div
    class="content flex flex-col gap-2 p-4 bg-gray-300 border border-gray-300 rounded flex-grow"
>
    <label class="text text-black font-bold text-left" for="Title">Title</label>
    <input
        id="Title"
        type="text"
        placeholder="Title"
        bind:value={content.title}
        on:change={(e) => handleInput(e, "title")}
        class="p-2 text-base"
    />
    <label class="text text-black font-bold text-left" for="Details">説明</label
    >
    <input
        id="Details"
        type="text"
        placeholder="Details"
        bind:value={content.details}
        on:change={(e) => handleInput(e, "details")}
        class="p-2 text-base"
    />
    <label class="text text-black font-bold text-left" for="RegExp"
        >正規表現抽出</label
    >
    <input
        id="RegExp"
        type="text"
        placeholder="RegExp"
        bind:value={content.regexp}
        on:change={(e) => handleInput(e, "regexp")}
        class="p-2 text-base"
    />

    <label class="text text-black font-bold text-left" for="Exclude"
        >抽出結果が以下と一致する場合はスキップ</label
    >
    <input
        id="Exclude"
        type="text"
        placeholder="Exclude Text"
        bind:value={content.exclude}
        on:change={(e) => handleInput(e, "exclude")}
        class="p-2 text-base"
    />

    <label class="text text-black font-bold text-left" for="EventType"
        >イベントタイプ</label
    >
    <select bind:value={content.type} on:change={(e) => handleInput(e, "type")}>
        <option value="WebRequest">Web Request</option>
        <option value="SendXSOverlay">Send XSOverlay</option>
        <option value="SendDiscordWebHook">Send Discord WebHook</option>
        <option value="Disable">Disable</option>
    </select>
    {#if content.type === "WebRequest" || content.type === "SendDiscordWebHook"}
        <label class="text" for="url-input">URL</label>
        <input
            id="EventType"
            type="text"
            placeholder="URL"
            bind:value={content.url}
            on:change={(e) => handleInput(e, "url")}
            class="p-2 text-base"
        />
    {/if}

    <button
        class="delete-button mt-4 p-2 bg-red-500 text-white border-none rounded cursor-pointer w-[7.5rem] self-end hover:bg-red-700"
        on:click={deleteContent}>DELETE</button
    >
</div>
