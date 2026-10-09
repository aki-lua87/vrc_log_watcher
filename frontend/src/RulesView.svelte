<script lang="ts">
    import { flip } from "svelte/animate";
    import type { DndEvent } from "svelte-dnd-action";
    import { dndzone } from "svelte-dnd-action";
    import { GetOutputFolderPath, OpenInExplorer } from "../wailsjs/go/main/App";
    import type { models } from "../wailsjs/go/models";
    import RegexHelper from "./components/RegexHelper.svelte";
    import { DESTINATIONS, destinationOf, TEMPLATE_VARS } from "./lib/destinations";
    import Icon from "./lib/Icon.svelte";
    import { toast } from "./lib/toast";

    export let settings: models.Setting[] = [];
    export let focusId = "";
    export let onUpdate: (setting: models.Setting) => void;
    export let onDelete: (setting: models.Setting) => void;
    export let onAdd: () => void;
    export let onReorder: (newSettings: models.Setting[]) => void;

    const flipDurationMs = 200;
    let selectedId = "";
    let confirmDelete = false;

    // 別の画面から指定されたルール（新規追加・エラーからの移動）を選ぶ
    $: if (focusId && settings.some((s) => s.id === focusId)) {
        selectedId = focusId;
        focusId = "";
        confirmDelete = false;
    }
    $: if (!settings.some((s) => s.id === selectedId) && settings.length > 0) {
        selectedId = settings[0].id;
    }
    $: selected = settings.find((s) => s.id === selectedId) ?? null;

    function select(id: string) {
        selectedId = id;
        confirmDelete = false;
    }

    // ---- 保存 ----
    let savedFlash = false;
    let savedTimer: number | undefined;
    function save() {
        if (!selected) return;
        settings = settings; // 一覧の表示も更新する
        onUpdate(selected);
        savedFlash = true;
        clearTimeout(savedTimer);
        savedTimer = window.setTimeout(() => (savedFlash = false), 1600);
    }

    function updateField<K extends keyof models.Setting>(field: K, value: models.Setting[K]) {
        if (!selected) return;
        selected[field] = value;
        save();
    }

    // 正規表現ヘルパーからの変更
    function handleRegexChange(
        event: CustomEvent<{ regexp: string; simpleMode: boolean; simpleBlocks: models.SimpleBlock[] }>,
    ) {
        if (!selected) return;
        const { regexp, simpleMode, simpleBlocks } = event.detail;
        selected.regexp = regexp;
        selected.simpleMode = simpleMode;
        selected.simpleBlocks = simpleBlocks;
        save();
    }

    // ---- ログの1行で試す ----
    const samples: Record<string, string> = {};
    let sample = "";
    $: sample = selected ? (samples[selected.id] ?? "") : "";
    function setSample(value: string) {
        if (selected) samples[selected.id] = value;
        sample = value;
    }
    $: testResult = (() => {
        if (!selected || !sample) return null;
        try {
            const m = new RegExp(selected.regexp).exec(sample);
            if (!selected.regexp || !m) return { state: "ng", text: "この行には一致しません" };
            const value = m[1] ?? m[0];
            if (selected.exclude && value === selected.exclude)
                return { state: "ng", text: "一致しましたが、除外条件に当たるため通知しません" };
            return { state: "ok", text: "一致しました。取り出した値", value };
        } catch {
            return { state: "bad", text: "正規表現の書き方に誤りがあります" };
        }
    })();

    // ---- 追加で送る項目（extraFields） ----
    let extraRows: { key: string; value: string }[] = [];
    let extraFor = "";
    $: if (selected && selected.id !== extraFor) {
        extraFor = selected.id;
        extraRows = Object.entries(selected.extraFields ?? {}).map(([key, value]) => ({ key, value }));
    }
    function commitExtra() {
        if (!selected) return;
        selected.extraFields = Object.fromEntries(extraRows.filter((r) => r.key !== "").map((r) => [r.key, r.value]));
        save();
    }
    function addExtraField() {
        let n = extraRows.length + 1;
        while (extraRows.some((r) => r.key === `field${n}`)) n++;
        extraRows = [...extraRows, { key: `field${n}`, value: "" }];
        commitExtra();
    }
    function removeExtraField(index: number) {
        extraRows = extraRows.filter((_, i) => i !== index);
        commitExtra();
    }

    // ---- テンプレート変数の挿入 ----
    let lastField: HTMLInputElement | null = null;
    function rememberField(e: FocusEvent) {
        lastField = e.currentTarget as HTMLInputElement;
    }
    function insertVar(v: string) {
        const f = lastField;
        if (!f || !document.body.contains(f)) {
            toast("先に入力欄を選んでください");
            return;
        }
        const start = f.selectionStart ?? f.value.length;
        const end = f.selectionEnd ?? start;
        f.value = f.value.slice(0, start) + v + f.value.slice(end);
        f.focus();
        f.setSelectionRange(start + v.length, start + v.length);
        f.dispatchEvent(new Event("input", { bubbles: true }));
    }

    async function openOutputFolder() {
        if (!selected) return;
        const folderPath = await GetOutputFolderPath(selected.id);
        if (folderPath) {
            await OpenInExplorer(folderPath);
        } else {
            toast("まだ出力フォルダがありません。最初の書き出しで作られます");
        }
    }

    function deleteSelected() {
        if (!selected) return;
        const target = selected;
        confirmDelete = false;
        onDelete(target);
        toast(`「${target.title || "名前のないルール"}」を削除しました`);
    }

    // ---- ドラッグ&ドロップ ----
    function handleDndConsider(e: CustomEvent<DndEvent<models.Setting>>) {
        settings = e.detail.items;
    }
    function handleDndFinalize(e: CustomEvent<DndEvent<models.Setting>>) {
        settings = e.detail.items;
        onReorder(settings);
    }
</script>

<div class="rules">
    <nav class="rlist" aria-label="ルール一覧">
        <button class="btn pri block" on:click={onAdd}><Icon name="plus" />新しいルール</button>
        <p class="hint grp">ドラッグで並び替えできます</p>
        <div
            class="items"
            use:dndzone={{ items: settings, flipDurationMs, dropTargetStyle: {} }}
            on:consider={handleDndConsider}
            on:finalize={handleDndFinalize}
        >
            {#each settings as setting (setting.id)}
                {@const dest = destinationOf(setting.type)}
                <div animate:flip={{ duration: flipDurationMs }}>
                    <button
                        class="ritem"
                        class:off={setting.type === "Disable"}
                        aria-current={selectedId === setting.id}
                        on:click={() => select(setting.id)}
                    >
                        <span class="rico"><Icon name={dest.icon} /></span>
                        <span class="rtext">
                            <b>{setting.title || "名前のないルール"}</b>
                            <small>{dest.label}</small>
                        </span>
                        <span class="grip" title={setting.isCore ? "標準ルール" : "ドラッグで並び替え"}>
                            <Icon name={setting.isCore ? "lock" : "grip"} size={14} />
                        </span>
                    </button>
                </div>
            {/each}
        </div>
    </nav>

    <div class="editor">
        {#if selected}
            {#key selected.id}
                <div class="ed">
                    <div class="ed-h">
                        <div class="t">
                            {#if selected.isCore}
                                <div class="ttl">{selected.title}</div>
                                <div class="dsc">{selected.details}</div>
                            {:else}
                                <input
                                    class="ttl"
                                    value={selected.title}
                                    placeholder="ルール名"
                                    aria-label="ルール名"
                                    on:input={(e) => updateField("title", e.currentTarget.value)}
                                >
                                <input
                                    class="dsc"
                                    value={selected.details}
                                    placeholder="説明を追加（任意）"
                                    aria-label="説明"
                                    on:input={(e) => updateField("details", e.currentTarget.value)}
                                >
                            {/if}
                        </div>
                        <div class="meta">
                            {#if selected.isCore}
                                <span class="badge"><Icon name="lock" size={11} />標準ルール</span>
                            {/if}
                            <span class="saved" class:on={savedFlash}>
                                <Icon name="check" size={13} />{savedFlash ? "保存しました" : "自動保存"}
                            </span>
                        </div>
                    </div>

                    {#if selected.isCore}
                        <div class="note">
                            <Icon
                                name="lock"
                                size={14}
                            />標準ルールは条件を変えられません。通知先はいつでも変更できます。
                        </div>
                    {/if}
                    {#if selected.id === "core-screenshot"}
                        <div class="note">
                            <Icon name="image" size={14} />撮影した画像は Discord や
                            Webリクエストで画像付きで送られます。詳しくはヘルプを見てください。
                        </div>
                    {/if}

                    <section class="card">
                        <h3><Icon name="zap" />通知条件</h3>
                        <p class="lead">VRChatのログに次の内容が出たときに動きます。</p>
                        <RegexHelper
                            regexp={selected.regexp}
                            simpleMode={selected.simpleMode}
                            simpleBlocks={selected.simpleBlocks}
                            isReadOnly={selected.isCore}
                            on:change={handleRegexChange}
                        />
                        <div class="test">
                            <label class="lb" for="rule-sample">ログの1行で試す</label>
                            <input
                                id="rule-sample"
                                class="inp mono"
                                placeholder="VRChatのログから1行貼り付けると、一致するか確認できます"
                                value={sample}
                                on:input={(e) => setSample(e.currentTarget.value)}
                            >
                            {#if testResult}
                                <div class="result {testResult.state}">
                                    <Icon
                                        name={testResult.state === "ok"
                                            ? "check"
                                            : testResult.state === "bad"
                                              ? "alert"
                                              : "x"}
                                        size={14}
                                    />
                                    {testResult.text}
                                    {#if testResult.value !== undefined}
                                        <code>{testResult.value}</code>
                                    {/if}
                                </div>
                            {/if}
                        </div>
                    </section>

                    <section class="card">
                        <h3><Icon name="bell" />通知先</h3>
                        <p class="lead">条件に合ったときの送り先を選びます。</p>
                        <div class="dests">
                            {#each DESTINATIONS as d}
                                <button
                                    class="dest"
                                    aria-pressed={selected.type === d.type}
                                    on:click={() => updateField("type", d.type)}
                                >
                                    <Icon name={d.icon} size={18} />
                                    <b>{d.label}</b>
                                    <small>{d.desc}</small>
                                </button>
                            {/each}
                        </div>

                        {#if selected.type === "SendDiscordWebHook"}
                            <div class="field">
                                <label class="lb" for="rule-url">Webhook URL</label>
                                <input
                                    id="rule-url"
                                    class="inp mono"
                                    placeholder="https://discord.com/api/webhooks/…"
                                    value={selected.url}
                                    on:input={(e) => updateField("url", e.currentTarget.value)}
                                >
                                <p class="hint">
                                    Discordのチャンネル設定 › 連携サービス › ウェブフック で作成したURLを貼り付けます。
                                </p>
                            </div>
                        {:else if selected.type === "WebRequest"}
                            <div class="field">
                                <label class="lb" for="rule-url">送信先URL</label>
                                <input
                                    id="rule-url"
                                    class="inp mono"
                                    placeholder="https://example.com/api"
                                    value={selected.url}
                                    on:focus={rememberField}
                                    on:input={(e) => updateField("url", e.currentTarget.value)}
                                >
                            </div>
                            <div class="field">
                                <label class="lb" for="rule-mk">本文のキー名</label>
                                <input
                                    id="rule-mk"
                                    class="inp mono"
                                    placeholder="message"
                                    value={selected.messageKey}
                                    on:input={(e) => updateField("messageKey", e.currentTarget.value)}
                                >
                                <p class="hint">JSONのどのキーにメッセージを入れるか。空なら "message" になります。</p>
                            </div>
                            <div class="field">
                                <span class="lb">追加で送る項目</span>
                                {#each extraRows as row, i (i)}
                                    <div class="kv">
                                        <input
                                            class="inp mono"
                                            placeholder="キー"
                                            aria-label="キー"
                                            bind:value={row.key}
                                            on:input={commitExtra}
                                        >
                                        <input
                                            class="inp"
                                            placeholder="値"
                                            aria-label="値"
                                            bind:value={row.value}
                                            on:focus={rememberField}
                                            on:input={commitExtra}
                                        >
                                        <button class="ib" title="この項目を削除" on:click={() => removeExtraField(i)}>
                                            <Icon name="x" />
                                        </button>
                                    </div>
                                {/each}
                                <div>
                                    <button class="btn sm" on:click={addExtraField}>
                                        <Icon name="plus" />項目を追加
                                    </button>
                                </div>
                            </div>
                            <div class="field">
                                <span class="lb">使える変数</span>
                                <div class="vars">
                                    {#each TEMPLATE_VARS as v}
                                        <button
                                            class="var"
                                            title={v.desc}
                                            on:mousedown|preventDefault
                                            on:click={() => insertVar(v.name)}
                                        >
                                            {v.name}
                                        </button>
                                    {/each}
                                </div>
                                <p class="hint">
                                    クリックすると、URLか項目の値の入力欄（最後に選んだほう）に挿入します。
                                </p>
                            </div>
                        {:else if selected.type === "OutputTextFile"}
                            <div class="field">
                                <span class="lb">書き出し先</span>
                                <div class="inline">
                                    <input
                                        class="inp mono"
                                        value="./OutputText/{selected.id}/"
                                        readonly
                                        aria-label="書き出し先"
                                    >
                                    <button class="btn" on:click={openOutputFolder}><Icon name="folder" />開く</button>
                                </div>
                                <p class="hint">
                                    アプリと同じ場所の OutputText フォルダに、ルールごとに1件ずつファイルを作ります。
                                </p>
                            </div>
                        {:else if selected.type === "SendXSOverlay"}
                            <div class="field">
                                <div class="note">
                                    <Icon name="monitor" size={14} />XSOverlay
                                    が起動していれば、VRの中に通知が表示されます。
                                </div>
                            </div>
                        {:else if selected.type === "LogOnly"}
                            <div class="field">
                                <div class="note">
                                    <Icon name="eye" size={14} />このアプリのログ画面にだけ表示し、外には送りません。
                                </div>
                            </div>
                        {:else if selected.type === "Disable"}
                            <div class="field">
                                <div class="note">
                                    <Icon
                                        name="ban"
                                        size={14}
                                    />このルールは停止中です。ログにも表示されず、通知もしません。
                                </div>
                            </div>
                        {/if}
                    </section>

                    <details class="card adv">
                        <summary>
                            <Icon name="sliders" />詳細設定 <span class="hint">除外条件・ルールID</span
                            ><span class="chev"><Icon name="chev" /></span>
                        </summary>
                        <div class="field">
                            <label class="lb" for="rule-ex">除外する値</label>
                            <input
                                id="rule-ex"
                                class="inp"
                                placeholder="例: 自分のユーザー名"
                                value={selected.exclude}
                                on:input={(e) => updateField("exclude", e.currentTarget.value)}
                            >
                            <p class="hint">取り出した値がこれと同じときは通知しません。</p>
                        </div>
                        <div class="field">
                            <span class="lb">ルールID</span>
                            <span class="mono idv">{selected.id}</span>
                        </div>
                    </details>

                    {#if !selected.isCore}
                        <div class="danger">
                            {#if confirmDelete}
                                <div class="confirm">
                                    <span
                                        >「{selected.title || "名前のないルール"}」を削除します。元に戻せません。</span
                                    >
                                    <button class="btn sm" on:click={() => (confirmDelete = false)}>キャンセル</button>
                                    <button class="btn sm danger fill" on:click={deleteSelected}>
                                        <Icon name="trash" />削除する
                                    </button>
                                </div>
                            {:else}
                                <p>このルールが不要になったら削除できます。</p>
                                <button class="btn danger" on:click={() => (confirmDelete = true)}>
                                    <Icon name="trash" />ルールを削除
                                </button>
                            {/if}
                        </div>
                    {/if}
                </div>
            {/key}
        {:else}
            <div class="empty">
                <Icon name="bell" size={28} /><span>左の一覧からルールを選ぶか、新しいルールを作ってください</span>
            </div>
        {/if}
    </div>
</div>

<style>
    .rules {
        display: grid;
        grid-template-columns: 270px minmax(0, 1fr);
        height: 100%;
        min-height: 0;
    }
    .rlist {
        border-right: 1px solid var(--line);
        overflow: auto;
        padding: 14px 10px 30px;
        display: flex;
        flex-direction: column;
        gap: 2px;
        min-height: 0;
    }
    .grp {
        padding: 10px 10px 4px;
    }
    .items {
        display: flex;
        flex-direction: column;
        gap: 2px;
        outline: none;
    }
    .ritem {
        display: grid;
        grid-template-columns: 30px minmax(0, 1fr) auto;
        gap: 10px;
        align-items: center;
        padding: 8px 10px;
        border-radius: var(--radius-s);
        border: 0;
        background: none;
        text-align: left;
        cursor: pointer;
        width: 100%;
        color: var(--fg);
    }
    .ritem:hover {
        background: var(--hover);
    }
    .ritem[aria-current="true"] {
        background: var(--elev);
        box-shadow: inset 0 0 0 1px var(--line-strong);
    }
    .rico {
        width: 30px;
        height: 30px;
        border-radius: var(--radius-s);
        background: var(--elev);
        display: grid;
        place-items: center;
        color: var(--muted);
        clip-path: var(--cut);
    }
    .ritem[aria-current="true"] .rico {
        color: var(--accent);
        background: var(--accent-soft);
    }
    .rtext {
        min-width: 0;
    }
    .rtext b {
        font-weight: 500;
        font-size: 13.5px;
        display: block;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    .rtext small {
        font-size: 11.5px;
        color: var(--faint);
        display: block;
    }
    .ritem.off b {
        color: var(--faint);
    }
    .grip {
        color: var(--faint);
        opacity: 0.5;
    }
    .ritem:hover .grip {
        opacity: 1;
    }

    .editor {
        overflow: auto;
        padding: 22px 28px 64px;
        min-height: 0;
    }
    .ed {
        max-width: 760px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }
    .ed-h {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        flex-wrap: wrap;
    }
    .t {
        flex: 1 1 280px;
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 2px;
    }
    .ttl {
        font-family: var(--font-display);
        font-size: 21px;
        font-weight: 600;
        background: transparent;
        border: 0;
        border-bottom: 1px solid transparent;
        width: 100%;
        padding: 2px 0;
        color: var(--fg);
    }
    input.ttl:hover {
        border-bottom-color: var(--line-strong);
    }
    input.ttl:focus {
        outline: none;
        border-bottom-color: var(--accent);
    }
    .dsc {
        background: transparent;
        border: 0;
        color: var(--muted);
        width: 100%;
        padding: 2px 0;
        font-size: 13px;
    }
    input.dsc:focus {
        outline: none;
        color: var(--fg);
    }
    .meta {
        display: flex;
        gap: 8px;
        align-items: center;
        padding-top: 6px;
    }
    .saved {
        font-size: 12px;
        color: var(--faint);
        display: inline-flex;
        gap: 5px;
        align-items: center;
        white-space: nowrap;
        transition: color 0.2s;
    }
    .saved.on {
        color: var(--ok);
    }

    .test {
        margin-top: 16px;
        padding-top: 14px;
        border-top: 1px solid var(--line);
        display: flex;
        flex-direction: column;
        gap: 6px;
    }
    .result {
        font-size: 12.5px;
        display: flex;
        gap: 6px;
        align-items: center;
        flex-wrap: wrap;
    }
    .result.ok {
        color: var(--ok);
    }
    .result.ng {
        color: var(--muted);
    }
    .result.bad {
        color: var(--err);
    }
    .result code {
        font-family: var(--font-data);
        color: var(--fg);
        background: var(--elev);
        padding: 1px 7px;
        border-radius: 4px;
        overflow-wrap: anywhere;
    }

    .dests {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
        gap: 8px;
    }
    .dest {
        display: flex;
        flex-direction: column;
        gap: 5px;
        align-items: flex-start;
        padding: 12px;
        border-radius: var(--radius-s);
        border: 1px solid var(--line);
        background: var(--bg);
        text-align: left;
        cursor: pointer;
        color: var(--fg);
        clip-path: var(--cut);
    }
    .dest:hover {
        border-color: var(--line-strong);
    }
    .dest[aria-pressed="true"] {
        border-color: var(--accent);
        box-shadow: inset 0 0 0 1px var(--accent);
        background: var(--accent-soft);
    }
    .dest :global(.ic) {
        color: var(--muted);
    }
    .dest[aria-pressed="true"] :global(.ic) {
        color: var(--accent);
    }
    .dest b {
        font-size: 13px;
        font-weight: 600;
    }
    .dest small {
        font-size: 11.5px;
        color: var(--muted);
        line-height: 1.45;
    }
    .inline {
        display: flex;
        gap: 8px;
        align-items: center;
        flex-wrap: wrap;
    }
    .inline .inp {
        flex: 1 1 220px;
        width: auto;
    }
    .kv {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(0, 1.5fr) auto;
        gap: 6px;
    }
    .vars {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
    }
    .var {
        font-family: var(--font-data);
        font-size: 11.5px;
        padding: 3px 9px;
        border-radius: 999px;
        border: 1px dashed var(--line-strong);
        background: none;
        color: var(--muted);
        cursor: pointer;
    }
    .var:hover {
        color: var(--accent-ink);
        border-color: var(--accent);
    }

    .adv summary {
        cursor: pointer;
        list-style: none;
        display: flex;
        align-items: center;
        gap: 8px;
        font-weight: 600;
        font-size: 14px;
        font-family: var(--font-display);
    }
    .adv summary::-webkit-details-marker {
        display: none;
    }
    .adv summary > :global(.ic:first-child) {
        color: var(--accent);
    }
    .adv summary .hint {
        font-weight: 400;
        font-family: var(--font-ui);
    }
    .chev {
        margin-left: auto;
        transition: transform 0.2s;
        color: var(--muted);
        display: inline-flex;
    }
    .adv[open] .chev {
        transform: rotate(180deg);
    }
    .idv {
        font-size: 12.5px;
        overflow-wrap: anywhere;
    }

    .danger {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 12px;
        flex-wrap: wrap;
    }
    .danger p {
        margin: 0;
        font-size: 12.5px;
        color: var(--muted);
    }
    .confirm {
        display: flex;
        gap: 8px;
        align-items: center;
        flex-wrap: wrap;
        background: var(--err-soft);
        padding: 10px 12px;
        border-radius: var(--radius-s);
        width: 100%;
    }
    .confirm span {
        flex: 1 1 200px;
        font-size: 13px;
    }
    .empty {
        padding: 60px 20px;
        text-align: center;
        color: var(--muted);
        display: flex;
        flex-direction: column;
        gap: 10px;
        align-items: center;
    }
</style>
