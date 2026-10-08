<script lang="ts">
    import { onDestroy } from "svelte";
    import { GetWatcherDebugInfo } from "../../wailsjs/go/main/App";
    import Icon from "../lib/Icon.svelte";

    export let fileName = "";
    export let hasFolder = false;
    export let onSetupFolder: () => void;

    let open = false;
    let debug = {
        lastReadLine: "",
        lastOffset: 0,
        fileSize: 0,
        linesRead: 0,
        lastReadAt: "",
        lastNewLinesAt: "",
        isRunning: false,
        skipCount: 0,
        consecutiveNoProgress: 0,
        refreshCount: 0,
    };

    async function fetchDebug() {
        try {
            debug = await GetWatcherDebugInfo();
        } catch (e) {
            console.error("監視状態の取得に失敗:", e);
        }
    }
    fetchDebug();
    const timer = setInterval(fetchDebug, 3000);
    onDestroy(() => clearInterval(timer));

    // 停滞が続いているときは注意表示にする
    $: stalled = debug.consecutiveNoProgress > 10;
    $: state = !hasFolder ? "nofolder" : !fileName ? "nofile" : stalled ? "stalled" : "ok";

    const fmt = (n: number) => n.toLocaleString("ja-JP");
</script>

<svelte:window on:keydown={(e) => e.key === "Escape" && (open = false)} />

<div class="wrap">
    {#if state === "nofolder"}
        <button class="status" on:click={onSetupFolder}>
            <span class="dot warn" /><b>ログフォルダ未設定</b>
        </button>
    {:else}
        <button class="status" on:click={() => (open = !open)} aria-expanded={open}>
            {#if state === "ok"}
                <span class="dot live" /><b>監視中</b>
            {:else if state === "stalled"}
                <span class="dot warn" /><b>読み取りが止まっています</b>
            {:else}
                <span class="dot warn" /><b>ログファイルが見つかりません</b>
            {/if}
            {#if fileName}<span class="file">{fileName}</span>{/if}
            <Icon name="chev" size={13} />
        </button>
    {/if}

    {#if open}
        <div class="pop" role="dialog" aria-label="監視の状態">
            <h4>監視の状態 <button class="ib" on:click={() => (open = false)} title="閉じる"><Icon name="x" /></button></h4>
            <dl>
                <dt>ログファイル</dt>
                <dd>{fileName || "---"}</dd>
                <dt>最後に新しい行を読んだ</dt>
                <dd>{debug.lastNewLinesAt || "---"}</dd>
                <dt>最後に確認した</dt>
                <dd>{debug.lastReadAt || "---"}</dd>
                <dt>読み取り位置</dt>
                <dd>{fmt(debug.lastOffset)} / {fmt(debug.fileSize)} バイト</dd>
                <dt>前回読んだ行数</dt>
                <dd>{debug.linesRead}</dd>
                <dt>読み取り中</dt>
                <dd class:bad={debug.isRunning}>{debug.isRunning ? "はい" : "いいえ"}</dd>
                <dt>スキップ</dt>
                <dd class:bad={debug.skipCount > 0}>{debug.skipCount}</dd>
                <dt>停滞の回数</dt>
                <dd class:warn={debug.consecutiveNoProgress > 0} class:bad={stalled}>{debug.consecutiveNoProgress}</dd>
                <dt>再読み込み</dt>
                <dd>{debug.refreshCount} 回</dd>
                <dt>最後に読んだ行</dt>
                <dd class="line">{debug.lastReadLine || "（まだ読んでいません）"}</dd>
            </dl>
            <p class="hint">うまく通知されないときに確認する情報です。</p>
        </div>
    {/if}
</div>

<style>
    .wrap {
        margin-left: auto;
        position: relative;
        min-width: 0;
        max-width: 100%;
    }
    .status {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 6px 12px;
        border: 1px solid var(--line);
        border-radius: 999px;
        background: var(--surface);
        color: var(--muted);
        font-size: 12px;
        cursor: pointer;
        max-width: 100%;
        min-width: 0;
    }
    :global(:root[data-theme="red"]) .status {
        border-radius: 2px;
    }
    .status:hover {
        border-color: var(--line-strong);
    }
    .status b {
        color: var(--fg);
        font-weight: 500;
        white-space: nowrap;
    }
    .file {
        font-family: var(--font-data);
        font-size: 11.5px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 280px;
    }
    .live {
        animation: pulse 2s ease-out infinite;
    }
    @keyframes pulse {
        0% {
            box-shadow: 0 0 0 0 color-mix(in srgb, var(--ok) 55%, transparent);
        }
        100% {
            box-shadow: 0 0 0 7px transparent;
        }
    }
    .pop {
        position: absolute;
        right: 0;
        top: calc(100% + 6px);
        width: 400px;
        max-width: calc(100vw - 32px);
        background: var(--elev);
        border: 1px solid var(--line-strong);
        border-radius: var(--radius);
        padding: 16px;
        z-index: 30;
        box-shadow: var(--shadow);
    }
    .pop h4 {
        margin: 0 0 10px;
        font-size: 13px;
        font-weight: 600;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    dl {
        display: grid;
        grid-template-columns: auto minmax(0, 1fr);
        gap: 7px 16px;
        margin: 0;
        font-size: 12px;
    }
    dt {
        color: var(--muted);
    }
    dd {
        margin: 0;
        font-family: var(--font-data);
        font-variant-numeric: tabular-nums;
        overflow-wrap: anywhere;
    }
    dd.line {
        max-height: 80px;
        overflow: auto;
    }
    dd.warn {
        color: var(--warn);
    }
    dd.bad {
        color: var(--err);
    }
    .pop .hint {
        margin-top: 12px;
    }
</style>
