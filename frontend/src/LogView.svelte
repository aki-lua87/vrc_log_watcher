<script lang="ts">
    import { fly } from "svelte/transition";
    import {
        GetImageThumbnail,
        GetOutputFolderPath,
        OpenFile,
        OpenFileInExplorer,
        OpenInExplorer,
    } from "../wailsjs/go/main/App";
    import type { models } from "../wailsjs/go/models";
    import { BrowserOpenURL, ClipboardSetText } from "../wailsjs/runtime";
    import Icon from "./lib/Icon.svelte";
    import { toast } from "./lib/toast";

    export let noticeLogs: models.NoticeLog[] = [];
    export let settings: models.Setting[] = [];
    export let hasFolder = true;
    export let onPickFolder: () => void;
    export let onOpenRule: (settingId: string) => void;

    type Kind = "world" | "worldid" | "join" | "leave" | "shot" | "file" | "custom" | "err" | "sys";
    type Entry = {
        key: number;
        kind: Kind;
        time: string;
        log: models.NoticeLog;
        ruleTitle: string;
        channel: string;
        delivered: boolean;
        ok: boolean;
    };
    type Session = { head: (Entry & { wid: string }) | null; rows: Entry[] };

    // 通知先ごとのバッジ表示（Dispatcher の Label 末尾と対応）
    const CHANNELS: Record<string, { label: string; icon: string }> = {
        XSOverlay: { label: "XSOverlay", icon: "monitor" },
        "Discord Webhook": { label: "Discord", icon: "msg" },
        "Web Request": { label: "Web", icon: "send" },
        テキスト出力: { label: "ファイル", icon: "file" },
    };

    const KIND_ICON: Record<string, string> = {
        join: "userPlus",
        leave: "userMinus",
        shot: "camera",
        file: "file",
        custom: "zap",
        err: "alert",
        sys: "info",
    };

    function fmtTime(ts: string) {
        const m = (ts ?? "").match(/(\d{1,2}:\d{2}:\d{2})/);
        return m ? m[1] : "";
    }

    function splitTitle(title: string): [string, string] {
        const i = title.lastIndexOf(": ");
        return i >= 0 ? [title.slice(0, i), title.slice(i + 2)] : [title, ""];
    }

    function classify(log: models.NoticeLog, key: number): Entry {
        let kind: Kind;
        switch (log.settingId) {
            case undefined:
            case "":
                kind = log.isError ? "err" : "sys";
                break;
            case "core-world-join-name":
                kind = "world";
                break;
            case "core-world-join-id":
                kind = "worldid";
                break;
            case "core-user-join":
                kind = "join";
                break;
            case "core-user-left":
                kind = "leave";
                break;
            case "core-screenshot":
                kind = "shot";
                break;
            default:
                kind = log.text.includes("[Log Output]") ? "file" : "custom";
        }
        const [ruleTitle, channel] = log.settingId ? splitTitle(log.title) : [log.title, ""];
        const delivered = channel in CHANNELS;
        const ok = !delivered || log.text.includes("成功");
        return {
            key,
            kind,
            time: fmtTime(log.timestamp),
            log,
            ruleTitle,
            channel,
            delivered,
            ok,
        };
    }

    // ワールド入室ごとに区切る。ワールドIDは入室ログ（Entering Room）の直前に出るので次の見出しに付ける
    function buildSessions(logs: models.NoticeLog[]): Session[] {
        const sessions: Session[] = [{ head: null, rows: [] }];
        let pendingWid = "";
        logs.forEach((log, i) => {
            const e = classify(log, i);
            const cur = sessions[sessions.length - 1];
            if (e.kind === "worldid") {
                pendingWid = log.metaData;
            } else if (e.kind === "world") {
                sessions.push({ head: { ...e, wid: pendingWid }, rows: [] });
                pendingWid = "";
            } else {
                cur.rows.push(e);
            }
        });
        return sessions;
    }

    // 絞り込みは通知ルール単位。"" はすべて、SYSTEM はルールに紐づかないログ
    const ALL = "";
    const SYSTEM = "__system";
    let filter = ALL;
    let query = "";

    function matches(e: Entry) {
        if (!query) return true;
        const q = query.toLowerCase();
        return [e.log.metaData, e.log.text, e.log.title].join(" ").toLowerCase().includes(q);
    }

    const filterKey = (log: models.NoticeLog) => log.settingId || SYSTEM;

    $: sessions = buildSessions(noticeLogs);
    $: counts = noticeLogs.reduce(
        (c, log) => {
            const k = filterKey(log);
            c[k] = (c[k] ?? 0) + 1;
            return c;
        },
        {} as Record<string, number>,
    );
    // 通知ルール画面の並び順で、0件のルールも含めてすべて出す。削除済みのルールはログのタイトルから名前を取る
    $: filters = (() => {
        const list: { id: string; label: string }[] = settings.map((s) => ({
            id: s.id,
            label: s.title || "名前のないルール",
        }));
        noticeLogs.forEach((log) => {
            const k = filterKey(log);
            if (k !== SYSTEM && !list.some((f) => f.id === k)) list.push({ id: k, label: splitTitle(log.title)[0] });
        });
        list.push({ id: SYSTEM, label: "システム" });
        return list;
    })();
    // ワールド入室のルールは見出し側に表示されるので、見出しで絞り込む
    // ワールド名の帯を絞り込みバーのすぐ下に固定するため、バーの高さを使う
    let fbarHeight = 0;

    const headMatchesFilter = (head: Session["head"]) =>
        filter === ALL || filter === "core-world-join-name" || (filter === "core-world-join-id" && !!head?.wid);
    $: visible = sessions
        .map((s) => {
            let rows = filter === ALL ? s.rows : s.rows.filter((r) => filterKey(r.log) === filter);
            rows = rows.filter(matches);
            const headHit = !!s.head && headMatchesFilter(s.head) && matches(s.head);
            const show = filter === ALL && !query ? true : rows.length > 0 || headHit;
            return {
                head: s.head,
                headHit,
                allRows: s.rows,
                rows: rows.slice().reverse(),
                show,
            };
        })
        .filter((s) => s.show && (s.head || s.rows.length))
        .reverse();

    function clearFilters() {
        filter = ALL;
        query = "";
    }

    // ---- 操作 ----
    function isURL(text: string): boolean {
        try {
            const url = new URL(text);
            return url.protocol === "http:" || url.protocol === "https:";
        } catch {
            return false;
        }
    }

    function copy(text: string) {
        ClipboardSetText(text)
            .then(() => toast("コピーしました"))
            .catch((err) => console.error("コピーに失敗しました:", err));
    }

    async function openImage(path: string) {
        try {
            await OpenFile(path);
        } catch (e) {
            console.error("画像を開けませんでした:", e);
        }
    }

    async function showInExplorer(path: string) {
        try {
            await OpenFileInExplorer(path);
        } catch (error) {
            console.error("エクスプローラーで開けませんでした。ファイルを直接開きます:", error);
            await openImage(path);
        }
    }

    async function openOutputFolder(settingId: string) {
        const folderPath = await GetOutputFolderPath(settingId);
        if (folderPath) await OpenInExplorer(folderPath);
    }

    // サムネイルキャッシュ（ファイルパス → Promise<string>）
    const thumbnailPromises = new Map<string, Promise<string>>();
    function getThumbnail(filePath: string): Promise<string> {
        let promise = thumbnailPromises.get(filePath);
        if (!promise) {
            promise = GetImageThumbnail(filePath)
                .then((b64) => `data:image/jpeg;base64,${b64}`)
                .catch(() => "");
            thumbnailPromises.set(filePath, promise);
        }
        return promise;
    }

    const baseName = (p: string) => p.split(/[\\/]/).pop();
    const countKind = (rows: Entry[], kind: Kind) => rows.filter((r) => r.kind === kind).length;
</script>

{#if !hasFolder}
    <div class="onb">
        <div class="box">
            <div class="big"><Icon name="folder" size={30} /></div>
            <h3>VRChatのログフォルダを選んでください</h3>
            <p>いつもの場所にログが見つかりませんでした。VRChatを一度でも起動していれば、通常は次の場所にあります。</p>
            <code>C:\Users\（ユーザー名）\AppData\LocalLow\VRChat\VRChat</code>
            <button class="btn pri" on:click={onPickFolder}><Icon name="folder" />フォルダを選択</button>
            <p class="hint">あとから「アプリ設定」でも変更できます。</p>
        </div>
    </div>
{:else}
    <div class="fbar" bind:clientHeight={fbarHeight}>
        <div class="chips">
            <button class="chip" aria-pressed={filter === ALL} on:click={() => (filter = ALL)}>
                すべて<span class="n">{noticeLogs.length}</span>
            </button>
            {#each filters as f (f.id)}
                <button class="chip" aria-pressed={filter === f.id} on:click={() => (filter = f.id)}>
                    {f.label}<span class="n">{counts[f.id] ?? 0}</span>
                </button>
            {/each}
        </div>
        <label class="search">
            <Icon name="search" />
            <input type="search" placeholder="検索" bind:value={query} aria-label="ログを検索">
        </label>
    </div>

    <div class="feed">
        {#each visible as s (s.head ? s.head.key : -1)}
            <!-- ワールドごとに枠で囲み、上の帯でどのワールドかを示す。入室そのものは時刻どおり枠の一番下に置く -->
            <section class:world={!!s.head}>
                {#if s.head}
                    {@const head = s.head}
                    <header class="sess-h" style="top: {fbarHeight}px">
                        <Icon name="globe" size={14} />
                        <b class="wname">{head.log.metaData}</b>
                        <span class="stats">
                            入室 {countKind(s.allRows, "join")} · 退室 {countKind(s.allRows, "leave")}
                        </span>
                        <div class="acts show">
                            {#if head.wid}
                                <button
                                    class="ib"
                                    title="ワールドのページを開く"
                                    on:click={() => BrowserOpenURL(`https://vrchat.com/home/world/${head.wid}`)}
                                >
                                    <Icon name="ext" />
                                </button>
                            {/if}
                            <button class="ib" title="ワールド名をコピー" on:click={() => copy(head.log.metaData)}>
                                <Icon name="copy" />
                            </button>
                        </div>
                    </header>
                {/if}

                {#if s.rows.length}
                    <div class="rows">
                        {#each s.rows as e (e.key)}
                            {@const log = e.log}
                            <div class="row k-{e.kind}" class:failed={!e.ok} in:fly|local={{ y: -6, duration: 220 }}>
                                <time>{e.time}</time>
                                <span class="kind"><Icon name={KIND_ICON[e.kind]} size={14} /></span>
                                <div class="msg">
                                    {#if e.kind === "join"}
                                        <b>{log.metaData}</b>
                                        が入室しました
                                    {:else if e.kind === "leave"}
                                        <b>{log.metaData}</b>
                                        が退室しました
                                    {:else if e.kind === "shot"}
                                        スクリーンショットを撮影しました
                                        <span class="sub val">{baseName(log.metaData)}</span>
                                        {#await getThumbnail(log.metaData) then src}
                                            {#if src}
                                                <button
                                                    class="thumb"
                                                    title="クリックで画像を開く"
                                                    on:click={() => openImage(log.metaData)}
                                                >
                                                    <img {src} alt="スクリーンショット">
                                                </button>
                                            {/if}
                                        {/await}
                                    {:else if e.kind === "file"}
                                        <span class="rl">{e.ruleTitle}</span
                                        >「{log.metaData}」をファイルに書き出しました
                                    {:else if e.kind === "custom"}
                                        <span class="rl">{e.ruleTitle}</span><span class="val">{log.metaData}</span>
                                    {:else}
                                        {log.text}
                                    {/if}
                                    {#if !e.ok}
                                        <span class="sub err">
                                            {log.text}
                                            {#if log.settingId}
                                                <button class="link" on:click={() => onOpenRule(log.settingId)}>
                                                    ルールを確認
                                                </button>
                                            {/if}
                                        </span>
                                    {/if}
                                </div>
                                <div class="right">
                                    {#if e.delivered && e.kind !== "file"}
                                        <span class="dlv" class:ng={!e.ok} title={log.text}>
                                            <Icon name={CHANNELS[e.channel].icon} size={12} />
                                            {CHANNELS[e.channel].label}
                                            <Icon name={e.ok ? "check" : "x"} size={12} />
                                        </span>
                                    {/if}
                                    <div class="acts">
                                        {#if log.canCopy && log.metaData}
                                            {#if isURL(log.metaData)}
                                                <button
                                                    class="ib"
                                                    title="ブラウザで開く"
                                                    on:click={() => BrowserOpenURL(log.metaData)}
                                                >
                                                    <Icon name="ext" />
                                                </button>
                                            {/if}
                                            {#if e.kind === "shot"}
                                                <button
                                                    class="ib"
                                                    title="画像を開く"
                                                    on:click={() => openImage(log.metaData)}
                                                >
                                                    <Icon name="image" />
                                                </button>
                                                <button
                                                    class="ib"
                                                    title="フォルダで表示"
                                                    on:click={() => showInExplorer(log.metaData)}
                                                >
                                                    <Icon name="folder" />
                                                </button>
                                            {/if}
                                            {#if e.kind === "file"}
                                                <button
                                                    class="ib"
                                                    title="出力フォルダを開く"
                                                    on:click={() => openOutputFolder(log.settingId)}
                                                >
                                                    <Icon name="folder" />
                                                </button>
                                            {/if}
                                            <button class="ib" title="コピー" on:click={() => copy(log.metaData)}>
                                                <Icon name="copy" />
                                            </button>
                                        {/if}
                                    </div>
                                </div>
                            </div>
                        {/each}
                    </div>
                {/if}

                {#if s.head && s.headHit}
                    {@const head = s.head}
                    <div class="rows">
                        <div class="row k-world" class:failed={!head.ok}>
                            <time>{head.time}</time>
                            <span class="kind"><Icon name="globe" size={14} /></span>
                            <div class="msg">
                                <b>{head.log.metaData}</b>
                                に入室しました
                                {#if head.wid}
                                    <span class="sub val">{head.wid}</span>
                                {/if}
                                {#if !head.ok}
                                    <span class="sub err">{head.log.text}</span>
                                {/if}
                            </div>
                            <div class="right">
                                {#if head.delivered}
                                    <span class="dlv" class:ng={!head.ok} title={head.log.text}>
                                        <Icon name={CHANNELS[head.channel].icon} size={12} />
                                        {CHANNELS[head.channel].label}
                                        <Icon name={head.ok ? "check" : "x"} size={12} />
                                    </span>
                                {/if}
                            </div>
                        </div>
                    </div>
                {/if}
            </section>
        {:else}
            <div class="empty">
                {#if noticeLogs.length === 0}
                    <Icon name="activity" size={28} />
                    <span>VRChatで動きがあると、ここに表示されます</span>
                {:else}
                    <Icon name="search" size={28} />
                    <span>条件に合うログはありません</span>
                    <button class="btn sm" on:click={clearFilters}>絞り込みを解除</button>
                {/if}
            </div>
        {/each}
    </div>
{/if}

<style>
    .fbar {
        display: flex;
        gap: 8px;
        align-items: center;
        flex-wrap: wrap;
        padding: 12px 20px;
        position: sticky;
        top: 0;
        background: var(--bg);
        z-index: 5;
        border-bottom: 1px solid var(--line);
    }
    .chips {
        display: flex;
        gap: 6px;
        flex-wrap: wrap;
    }
    .chip {
        border: 1px solid var(--line-strong);
        background: transparent;
        color: var(--muted);
        border-radius: 999px;
        padding: 4px 11px;
        font-size: 12.5px;
        cursor: pointer;
        display: inline-flex;
        gap: 6px;
        align-items: center;
    }
    .chip:hover {
        color: var(--fg);
    }
    .chip[aria-pressed="true"] {
        background: var(--fg);
        color: var(--bg);
        border-color: var(--fg);
    }
    :global(:root[data-theme="red"]) .chip,
    :global(:root[data-theme="red"]) .search {
        border-radius: 2px;
    }
    :global(:root[data-theme="red"]) .chip[aria-pressed="true"] {
        background: var(--accent);
        border-color: var(--accent);
        color: #fff;
    }
    .n {
        font-family: var(--font-data);
        font-size: 11px;
        opacity: 0.65;
        font-variant-numeric: tabular-nums;
    }
    .search {
        margin-left: auto;
        display: flex;
        align-items: center;
        gap: 6px;
        background: var(--surface);
        border: 1px solid var(--line);
        border-radius: 999px;
        padding: 5px 12px;
        color: var(--faint);
        min-width: 0;
        flex: 0 1 240px;
    }
    .search:focus-within {
        border-color: var(--accent);
    }
    .search input {
        background: transparent;
        border: 0;
        outline: 0;
        min-width: 0;
        width: 100%;
        font-size: 12.5px;
        color: var(--fg);
    }
    .search input::placeholder {
        color: var(--faint);
    }

    .feed {
        padding: 14px 20px 48px;
        display: flex;
        flex-direction: column;
        gap: 20px;
        max-width: 1000px;
    }
    /* ワールドの枠 */
    .world {
        background: var(--surface);
        border: 1px solid var(--line);
        border-radius: var(--radius);
        padding: 0 6px 6px;
    }
    :global(:root[data-theme="red"]) .world {
        border-left: 3px solid var(--accent);
    }
    /* 枠の上の帯。スクロールしても絞り込みバーの下に残る */
    .sess-h {
        position: sticky;
        z-index: 4;
        display: flex;
        align-items: center;
        gap: 8px;
        margin: 0 -6px;
        padding: 6px 12px 6px 18px;
        background: var(--surface);
        border-bottom: 1px solid var(--line);
        border-radius: var(--radius) var(--radius) 0 0;
        color: var(--accent);
    }
    .wname {
        flex: 1 1 auto;
        min-width: 0;
        color: var(--fg);
        font-size: 13.5px;
        font-weight: 600;
        font-family: var(--font-display);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    .world .rows {
        margin-top: 4px;
    }
    .world .rows + .rows {
        margin-top: 0;
    }
    .k-world .kind {
        color: var(--accent);
        background: var(--accent-soft);
    }
    .stats {
        font-size: 12px;
        color: var(--muted);
        font-variant-numeric: tabular-nums;
        white-space: nowrap;
    }

    .rows {
        display: flex;
        flex-direction: column;
        margin-top: 4px;
    }
    .row {
        display: grid;
        grid-template-columns: 64px 26px minmax(0, 1fr) auto;
        gap: 10px;
        align-items: start;
        padding: 8px 12px;
        border-radius: var(--radius-s);
    }
    .row:hover {
        background: var(--hover);
    }
    .row time {
        font-family: var(--font-data);
        font-size: 12px;
        color: var(--faint);
        font-variant-numeric: tabular-nums;
        padding-top: 3px;
    }
    :global(:root[data-theme="red"]) .row time {
        color: color-mix(in srgb, var(--accent-ink) 70%, var(--faint));
    }
    .kind {
        width: 24px;
        height: 24px;
        border-radius: 6px;
        display: grid;
        place-items: center;
        color: var(--muted);
        background: var(--elev);
        clip-path: var(--cut);
    }
    .k-join .kind {
        color: var(--ok);
    }
    .k-shot .kind,
    .k-custom .kind,
    .k-file .kind {
        color: var(--accent);
    }
    .k-err,
    .row.failed {
        background: var(--err-soft);
    }
    .k-err .kind,
    .row.failed .kind {
        color: var(--err);
        background: transparent;
    }
    .k-sys {
        color: var(--faint);
        font-size: 12.5px;
    }
    .msg {
        min-width: 0;
        padding-top: 2px;
        overflow-wrap: anywhere;
    }
    .msg b {
        font-weight: 600;
    }
    .rl {
        font-size: 12px;
        color: var(--muted);
        margin-right: 8px;
    }
    .val {
        font-family: var(--font-data);
        font-size: 12.5px;
    }
    .sub {
        display: block;
        color: var(--muted);
        font-size: 12.5px;
        margin-top: 2px;
    }
    .sub.err {
        color: var(--fg);
        font-family: var(--font-ui);
    }
    .link {
        background: none;
        border: 0;
        padding: 0;
        margin-left: 6px;
        color: var(--accent-ink);
        text-decoration: underline;
        cursor: pointer;
        font-size: 12.5px;
    }
    .right {
        display: flex;
        align-items: center;
        gap: 6px;
        flex-wrap: wrap;
        justify-content: flex-end;
    }
    .dlv {
        display: inline-flex;
        align-items: center;
        gap: 4px;
        font-size: 11px;
        color: var(--muted);
        padding: 2px 8px;
        border: 1px solid var(--line);
        border-radius: 999px;
        white-space: nowrap;
    }
    .dlv :global(.ic:last-child) {
        color: var(--ok);
    }
    .dlv.ng,
    .dlv.ng :global(.ic:last-child) {
        color: var(--err);
        border-color: color-mix(in srgb, var(--err) 40%, transparent);
    }
    :global(:root[data-theme="red"]) .dlv {
        border-radius: 2px;
    }
    .acts {
        display: flex;
        gap: 2px;
        opacity: 0;
        transition: opacity 0.15s;
    }
    .acts.show,
    .row:hover .acts,
    .row:focus-within .acts {
        opacity: 1;
    }
    .thumb {
        margin-top: 8px;
        display: block;
        max-width: 320px;
        border-radius: var(--radius-s);
        border: 1px solid var(--line);
        overflow: hidden;
        cursor: pointer;
        padding: 0;
        background: none;
    }
    .thumb:hover {
        border-color: var(--accent);
    }
    .thumb img {
        display: block;
        max-width: 100%;
        max-height: 220px;
        width: auto;
        height: auto;
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

    .onb {
        display: grid;
        place-items: center;
        min-height: 100%;
        padding: 32px 20px;
    }
    .box {
        max-width: 460px;
        text-align: center;
        display: flex;
        flex-direction: column;
        gap: 12px;
        align-items: center;
    }
    .big {
        width: 64px;
        height: 64px;
        border-radius: var(--radius);
        background: var(--accent-soft);
        color: var(--accent);
        display: grid;
        place-items: center;
    }
    .onb h3 {
        margin: 6px 0 0;
        font-size: 18px;
        font-family: var(--font-display);
        font-weight: 600;
    }
    .onb p {
        margin: 0;
        color: var(--muted);
        font-size: 13px;
    }
    .onb code {
        font-family: var(--font-data);
        font-size: 12px;
        background: var(--surface);
        border: 1px solid var(--line);
        padding: 6px 10px;
        border-radius: var(--radius-s);
        overflow-wrap: anywhere;
    }
</style>
