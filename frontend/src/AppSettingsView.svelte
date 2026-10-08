<script lang="ts">
    import { BrowserOpenURL } from "../wailsjs/runtime";
    import { OpenInExplorer } from "../wailsjs/go/main/App";
    import Icon from "./lib/Icon.svelte";
    import Logo from "./lib/Logo.svelte";
    import { THEMES } from "./lib/theme";
    import type { ThemeId } from "./lib/theme";

    export let logFolderPath = "";
    export let logFileName = "";
    export let theme: ThemeId = "light";
    export let onPickFolder: () => void;
    export let onChangeTheme: (theme: ThemeId) => void;

    const REPO_URL = "https://github.com/aki-lua87/vrc_log_watcher";

    async function openFolder() {
        try {
            await OpenInExplorer(logFolderPath);
        } catch (e) {
            console.error("フォルダを開けませんでした:", e);
        }
    }
</script>

<div class="pg">
    <section class="card">
        <h3><Icon name="folder" />VRChatのログ</h3>
        <div class="setrow">
            <div class="k">
                <b>ログフォルダ</b>
                <code>{logFolderPath || "未設定"}</code>
            </div>
            <div class="inline">
                {#if logFolderPath}
                    <button class="btn" on:click={openFolder}><Icon name="ext" />開く</button>
                {/if}
                <button class="btn" class:pri={!logFolderPath} on:click={onPickFolder}>{logFolderPath ? "変更" : "フォルダを選択"}</button>
            </div>
        </div>
        <div class="setrow">
            <div class="k">
                <b>監視中のファイル</b>
                <code>{logFileName || "見つかりません"}</code>
            </div>
            {#if logFileName}
                <span class="badge"><span class="dot" />監視中</span>
            {:else}
                <span class="badge"><span class="dot warn" />停止中</span>
            {/if}
        </div>
    </section>

    <section class="card">
        <h3><Icon name="palette" />カラーテーマ</h3>
        <p class="lead">見た目だけが変わります。通知の動きは同じです。</p>
        <div class="skins">
            {#each THEMES as t}
                <button class="skin" aria-pressed={theme === t.id} on:click={() => onChangeTheme(t.id)}>
                    <span class="prev" style="background:{t.preview[1]}">
                        <span style="background:{t.preview[0]}" />
                        <span class="pr">
                            <i style="background:{t.preview[2]};width:42%" />
                            <i style="background:{t.preview[3]};opacity:.55;width:82%" />
                            <i style="background:{t.preview[3]};opacity:.3;width:64%" />
                        </span>
                    </span>
                    <b>{t.name}</b>
                </button>
            {/each}
        </div>
    </section>

    <section class="card">
        <h3><Icon name="monitor" />XSOverlay</h3>
        <div class="setrow">
            <div class="k">
                <b>接続先</b>
                <span class="hint">localhost:42070（WebSocket）に通知を送り、1分ごとに接続を確認します。</span>
            </div>
        </div>
    </section>

    <section class="card">
        <h3><Icon name="info" />このアプリについて</h3>
        <div class="setrow">
            <div class="k about">
                <Logo size={36} />
                <div>
                    <b>VRC LogBridge</b>
                    <span class="hint">VRChatのログを読み取り、XSOverlay・Discord・Webなどへ届けます</span>
                </div>
            </div>
            <button class="btn" on:click={() => BrowserOpenURL(REPO_URL)}><Icon name="ext" />GitHub</button>
        </div>
    </section>
</div>

<style>
    .pg {
        padding: 24px 28px 64px;
        max-width: 860px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }
    .setrow {
        display: flex;
        gap: 16px;
        align-items: center;
        justify-content: space-between;
        flex-wrap: wrap;
        padding: 14px 0;
        border-top: 1px solid var(--line);
    }
    h3 + .setrow {
        border-top: 0;
        padding-top: 8px;
    }
    .k {
        min-width: 0;
        flex: 1 1 260px;
        display: flex;
        flex-direction: column;
        gap: 2px;
    }
    .k b {
        font-weight: 500;
    }
    .k code {
        font-family: var(--font-data);
        font-size: 12px;
        color: var(--muted);
        overflow-wrap: anywhere;
    }
    .about {
        flex-direction: row;
        align-items: center;
        gap: 12px;
    }
    .about div {
        display: flex;
        flex-direction: column;
    }
    .inline {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
    }
    .skins {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
        gap: 10px;
    }
    .skin {
        border: 1px solid var(--line);
        border-radius: var(--radius-s);
        padding: 8px;
        background: var(--bg);
        cursor: pointer;
        text-align: left;
        display: flex;
        flex-direction: column;
        gap: 8px;
        color: var(--fg);
        clip-path: var(--cut);
    }
    .skin:hover {
        border-color: var(--line-strong);
    }
    .skin[aria-pressed="true"] {
        border-color: var(--accent);
        box-shadow: inset 0 0 0 1px var(--accent);
    }
    .skin b {
        font-size: 13px;
        font-weight: 600;
        padding-inline: 2px;
    }
    .prev {
        height: 74px;
        border-radius: 4px;
        display: grid;
        grid-template-columns: 28% 1fr;
        overflow: hidden;
    }
    .pr {
        padding: 9px;
        display: flex;
        flex-direction: column;
        gap: 6px;
    }
    .prev i {
        height: 6px;
        border-radius: 3px;
        display: block;
    }
</style>
