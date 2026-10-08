<script lang="ts">
    import Icon from "./lib/Icon.svelte";
    import { DESTINATIONS, TEMPLATE_VARS } from "./lib/destinations";

    const destDetails: Record<string, string> = {
        SendXSOverlay: "XSOverlay（VR内のオーバーレイ）へ通知を送ります。localhost:42070 の WebSocket を使います。",
        SendDiscordWebHook: "Discord の Webhook へメッセージを投稿します。スクリーンショットのときは画像ファイルを添付して送ります。",
        LogOnly: "このアプリのログ画面に表示するだけで、外には送りません。動作確認にも便利です。",
        OutputTextFile: "OutputText/{ルールID}/ フォルダに、日時の名前を付けた .txt ファイルを書き出します。",
        WebRequest: "指定したURLへ JSON を POST します。URLと追加の項目に変数を使えます。",
        Disable: "ルールを一時的に止めます。条件に合っても何もしません。",
    };

    type Section = "start" | "vars" | "screenshot" | "dest";
    let active: Section = "start";
    const SECTIONS: [Section, string][] = [
        ["start", "はじめかた"],
        ["dest", "通知先"],
        ["vars", "使える変数"],
        ["screenshot", "スクリーンショット"],
    ];
</script>

<div class="help">
    <nav class="hnav" aria-label="ヘルプの項目">
        {#each SECTIONS as [id, label]}
            <button class="hitem" aria-current={active === id} on:click={() => (active = id)}>{label}</button>
        {/each}
    </nav>

    <div class="body">
        {#if active === "start"}
            <section class="card">
                <h3><Icon name="zap" />はじめかた</h3>
                <ol class="steps">
                    <li>
                        <b>ログフォルダを確認する</b>
                        <span>ふつうは自動で見つかります。見つからないときは「アプリ設定」から選びます。</span>
                    </li>
                    <li>
                        <b>通知先を選ぶ</b>
                        <span>「通知ルール」で、入室やスクリーンショットをどこに送るか決めます。XSOverlay ならVRの中で通知が見られます。</span>
                    </li>
                    <li>
                        <b>VRChatを遊ぶ</b>
                        <span>このアプリを開いたままにしておくと、ログ画面にイベントが流れてきます。</span>
                    </li>
                </ol>
            </section>
            <section class="card">
                <h3><Icon name="alert" />うまく通知されないとき</h3>
                <p class="lead flat">
                    画面右上の「監視中」を押すと、ログの読み取り状況を確認できます。停滞の回数が増え続けている場合は、ログフォルダの場所とVRChatのログ出力設定を確認してください。
                </p>
            </section>
        {:else if active === "dest"}
            <section class="card">
                <h3><Icon name="bell" />通知先</h3>
                <p class="lead">ルールの条件に合ったときに、どこへ送るかを選べます。</p>
                <div class="dl">
                    {#each DESTINATIONS as d}
                        <div class="di">
                            <span class="dico"><Icon name={d.icon} /></span>
                            <div>
                                <b>{d.label}</b>
                                <p>{destDetails[d.type]}</p>
                            </div>
                        </div>
                    {/each}
                </div>
            </section>
        {:else if active === "vars"}
            <section class="card">
                <h3><Icon name="scan" />使える変数</h3>
                <p class="lead">
                    Webリクエストの URL と「追加で送る項目」の値に <code>{"{変数名}"}</code> の形で書くと、送るときに実際の値に置き換わります。
                </p>
                <div class="tw">
                    <table>
                        <thead><tr><th>変数</th><th>内容</th><th>例</th></tr></thead>
                        <tbody>
                            {#each TEMPLATE_VARS as v}
                                <tr><td class="vn">{v.name}</td><td>{v.desc}</td><td class="ex">{v.example}</td></tr>
                            {/each}
                        </tbody>
                    </table>
                </div>
                <span class="lb sp">使用例</span>
                <pre class="code"><span class="c">// URL</span>
https://example.com/hook?src=<em>{"{log_file}"}</em>

<span class="c">// 追加で送る項目</span>
"timestamp": "<em>{"{time}"}</em>"
"label": "[<em>{"{title}"}</em>] <em>{"{matched_text}"}</em>"</pre>
                <div class="note sp">
                    <Icon name="info" size={14} />変数が置き換わるのは値だけです。項目のキー名には使えません。
                </div>
            </section>
        {:else}
            <section class="card">
                <h3><Icon name="camera" />スクリーンショット</h3>
                <p class="lead">標準ルール「スクリーンショット」には、写真ならではの機能があります。</p>
                <div class="dl">
                    <div class="di">
                        <span class="dico"><Icon name="globe" /></span>
                        <div>
                            <b>ワールド情報の読み取り</b>
                            <p>写真のPNGファイルから、撮影したワールドのIDと名前を自動で読み取ります。</p>
                        </div>
                    </div>
                    <div class="di">
                        <span class="dico"><Icon name="msg" /></span>
                        <div>
                            <b>Discord</b>
                            <p>画像ファイルを添付して投稿します。投稿文にはワールド名を太字で入れます。</p>
                        </div>
                    </div>
                    <div class="di">
                        <span class="dico"><Icon name="send" /></span>
                        <div>
                            <b>Webリクエスト</b>
                            <p>画像を Base64 にして JSON で送ります。次の項目が自動で付きます。</p>
                            <div class="tags">
                                {#each ["image", "filename", "mimetype", "world_id", "world_name"] as f}<code>{f}</code>{/each}
                            </div>
                        </div>
                    </div>
                </div>
            </section>
        {/if}
    </div>
</div>

<style>
    .help {
        display: grid;
        grid-template-columns: 200px minmax(0, 1fr);
        min-height: 100%;
    }
    .hnav {
        border-right: 1px solid var(--line);
        padding: 18px 10px;
        display: flex;
        flex-direction: column;
        gap: 2px;
    }
    .hitem {
        text-align: left;
        padding: 8px 12px;
        border-radius: var(--radius-s);
        border: 0;
        background: none;
        color: var(--muted);
        cursor: pointer;
        font-size: 13.5px;
    }
    .hitem:hover {
        background: var(--hover);
        color: var(--fg);
    }
    .hitem[aria-current="true"] {
        background: var(--accent-soft);
        color: var(--accent-ink);
        font-weight: 500;
    }
    .body {
        padding: 24px 28px 64px;
        max-width: 860px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }
    .lead.flat {
        margin: 0;
    }
    .lead code,
    .tags code {
        font-family: var(--font-data);
        font-size: 12px;
        background: var(--elev);
        padding: 1px 6px;
        border-radius: 4px;
        color: var(--accent-ink);
    }
    .steps {
        margin: 8px 0 0;
        padding-left: 1.4em;
        display: flex;
        flex-direction: column;
        gap: 12px;
        list-style: decimal;
    }
    .steps li b {
        display: block;
        font-weight: 600;
    }
    .steps li span {
        color: var(--muted);
        font-size: 13px;
    }
    .dl {
        display: flex;
        flex-direction: column;
    }
    .di {
        display: flex;
        gap: 12px;
        padding: 12px 0;
        border-top: 1px solid var(--line);
    }
    .di:first-child {
        border-top: 0;
        padding-top: 4px;
    }
    .dico {
        width: 32px;
        height: 32px;
        border-radius: var(--radius-s);
        background: var(--accent-soft);
        color: var(--accent);
        display: grid;
        place-items: center;
        flex: none;
        clip-path: var(--cut);
    }
    .di b {
        font-weight: 600;
    }
    .di p {
        margin: 2px 0 0;
        color: var(--muted);
        font-size: 13px;
    }
    .tags {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        margin-top: 8px;
    }
    .tw {
        overflow-x: auto;
    }
    table {
        width: 100%;
        border-collapse: collapse;
        font-size: 13px;
    }
    th {
        text-align: left;
        font-size: 11.5px;
        color: var(--faint);
        font-weight: 600;
        padding: 0 12px 8px 0;
    }
    td {
        padding: 9px 12px 9px 0;
        border-top: 1px solid var(--line);
        vertical-align: top;
    }
    td.vn {
        font-family: var(--font-data);
        font-size: 12.5px;
        color: var(--accent-ink);
        white-space: nowrap;
    }
    td.ex {
        font-family: var(--font-data);
        font-size: 12px;
        color: var(--muted);
        overflow-wrap: anywhere;
    }
    .sp {
        display: flex;
        margin-top: 18px;
    }
    .code {
        margin: 8px 0 0;
        padding: 14px;
        background: var(--bg);
        border: 1px solid var(--line);
        border-radius: var(--radius-s);
        font-family: var(--font-data);
        font-size: 12.5px;
        line-height: 1.7;
        overflow-x: auto;
        white-space: pre;
    }
    .code .c {
        color: var(--faint);
    }
    .code em {
        font-style: normal;
        color: var(--accent-ink);
    }
</style>
