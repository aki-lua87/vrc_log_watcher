<script lang="ts">
  import { fly } from "svelte/transition";
  import { OpenFolderSelectWindow } from "../wailsjs/go/main/App.js";
  import { GetNewestFileName } from "../wailsjs/go/main/App.js";
  import { PingXSOverlay } from "../wailsjs/go/main/App.js";
  import { UpdateSetting } from "../wailsjs/go/main/App.js";
  import { UpdateTheme } from "../wailsjs/go/main/App.js";
  import { LoadSetting } from "../wailsjs/go/main/App.js";
  import { LoadNoticeLog } from "../wailsjs/go/main/App.js";
  import { ReadFile } from "../wailsjs/go/main/App.js";

  import LogView from "./LogView.svelte";
  import RulesView from "./RulesView.svelte";
  import AppSettingsView from "./AppSettingsView.svelte";
  import HelpView from "./HelpView.svelte";
  import StatusPill from "./components/StatusPill.svelte";
  import Icon from "./lib/Icon.svelte";
  import Logo from "./lib/Logo.svelte";
  import { toastMessage } from "./lib/toast";
  import { applyTheme, normalizeTheme } from "./lib/theme";
  import type { ThemeId } from "./lib/theme";

  import { models } from "../wailsjs/go/models";

  // 画面状態管理
  type View = "log" | "rules" | "settings" | "help";
  let currentView: View = "log";
  let focusRuleId = "";

  let vrcLogFileName: string = "";
  let intervalId = 0;
  let saveData: models.SaveData;
  let loaded = false;
  let theme: ThemeId = "light";

  let contents: models.Setting[] = [];
  let noticeLogs: models.NoticeLog[] = [];
  let idCount = 0;

  applyTheme(theme);

  window.runtime.EventsOn("commonLogOutput", (noticeLog: models.NoticeLog) => {
    noticeLogs = [...noticeLogs, noticeLog];
  });

  function systemLog(text: string) {
    const noticeLog = {
      text,
      metaData: "",
      title: "[SYSTEM]",
      canCopy: false,
      timestamp: new Date().toLocaleTimeString(),
      isSystem: true,
      isError: false,
      actionSuccess: true,
    } as models.NoticeLog;
    noticeLogs = [...noticeLogs, noticeLog];
  }

  init();
  async function init() {
    await LoadNoticeLog(); // 構造体のロードのためのダミーコール
    systemLog("アプリケーションを起動しました");
    await LoadSetting().then((result) => (saveData = result));
    contents = saveData.settings;
    theme = normalizeTheme(saveData.theme);
    applyTheme(theme);
    loaded = true;
    if (intervalId != 0) {
      clearInterval(intervalId);
    }
    await getLogFiles();
    intervalId = setInterval(getLogFiles, 1 * (60 / 4) * 1000);
    setInterval(watchFile, 1 * 1 * 1000);
    setInterval(PingXSOverlay, 1 * 60 * 1000);
  }

  async function getLogFolderPath() {
    await OpenFolderSelectWindow().then((result) => (saveData.path = result));
    console.log(saveData.path);
    await getLogFiles();
  }

  async function getLogFiles() {
    // ログフォルダ未設定の間はログ画面に案内を表示する
    if (saveData.path == undefined || saveData.path == "") {
      vrcLogFileName = "";
      return;
    }
    // ログフォルダ内のファイルを取得する
    await GetNewestFileName(saveData.path).then(
      (result) => (vrcLogFileName = result),
    );
  }

  async function watchFile() {
    // ここをSetIntervalで無理やり監視する
    ReadFile().then((result) => {}); // awaitいらない？
  }

  async function addContent() {
    // uuid作成
    const uuid = () =>
      Math.floor((1 + Math.random()) * 0x10000)
        .toString(16)
        .substring(1);
    const newContent = models.Setting.createFrom({
      id: uuid(),
      title: `新しいルール ${++idCount}`,
      target: "",
      details: "",
      isCore: false,
      type: "LogOnly",
      url: "",
      regexp: "",
      exclude: "",
      messageKey: "",
      extraFields: {},
      simpleMode: true,
      simplePattern: "",
      simpleBlocks: [{ type: "between", start: "", end: "" }],
    });
    contents = [...contents, newContent];
    systemLog(`ルールを追加しました: ${newContent.title}`);
    await UpdateSetting(contents).then((result) => console.log(result));
    focusRuleId = newContent.id;
  }

  async function updateContent(setting: models.Setting) {
    contents = contents.map((content) =>
      content.id === setting.id ? setting : content,
    );
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  async function deleteContent(setting: models.Setting) {
    contents = contents.filter((content) => content.id !== setting.id);
    systemLog(`ルールを削除しました: ${setting.title}`);
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  async function handleReorderContents(newContents: models.Setting[]) {
    contents = newContents;
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  async function changeTheme(next: ThemeId) {
    theme = next;
    applyTheme(theme);
    saveData.theme = theme;
    await UpdateTheme(theme);
  }

  // 画面切り替え
  function go(view: View) {
    currentView = view;
  }

  function openRule(settingId: string) {
    focusRuleId = settingId;
    currentView = "rules";
  }

  const NAV: { view: View; icon: string; label: string }[] = [
    { view: "log", icon: "activity", label: "ログ" },
    { view: "rules", icon: "bell", label: "通知ルール" },
    { view: "settings", icon: "sliders", label: "アプリ設定" },
    { view: "help", icon: "help", label: "ヘルプ" },
  ];
  const TITLES: Record<View, string> = {
    log: "ログ",
    rules: "通知ルール",
    settings: "アプリ設定",
    help: "ヘルプ",
  };

  $: hasFolder = !loaded || !!saveData?.path;
</script>

<div class="win">
  <aside class="side">
    <div class="brand">
      <Logo size={32} />
      <div>
        <b>VRC LogBridge</b>
        <small>VRChatログ通知</small>
      </div>
    </div>
    {#each NAV as item}
      <button
        class="nav"
        aria-current={currentView === item.view ? "page" : undefined}
        on:click={() => go(item.view)}
      >
        <Icon name={item.icon} />
        <span>{item.label}</span>
        {#if item.view === "rules"}<span class="count">{contents.length}</span>{/if}
      </button>
    {/each}
  </aside>

  <section class="main">
    <header class="top">
      <h2>{TITLES[currentView]}</h2>
      <StatusPill
        fileName={vrcLogFileName}
        {hasFolder}
        onSetupFolder={getLogFolderPath}
      />
    </header>
    <div class="view">
      {#if currentView === "log"}
        <LogView
          {noticeLogs}
          settings={contents}
          {hasFolder}
          onPickFolder={getLogFolderPath}
          onOpenRule={openRule}
        />
      {:else if currentView === "rules"}
        <RulesView
          settings={contents}
          bind:focusId={focusRuleId}
          onUpdate={updateContent}
          onDelete={deleteContent}
          onAdd={addContent}
          onReorder={handleReorderContents}
        />
      {:else if currentView === "settings"}
        <AppSettingsView
          logFolderPath={saveData?.path || ""}
          logFileName={vrcLogFileName}
          {theme}
          onPickFolder={getLogFolderPath}
          onChangeTheme={changeTheme}
        />
      {:else}
        <HelpView />
      {/if}
    </div>
  </section>
</div>

{#if $toastMessage}
  <div class="toast" role="status" transition:fly={{ y: 16, duration: 180 }}>
    <Icon name="check" />{$toastMessage}
  </div>
{/if}

<style>
  .win {
    height: 100%;
    display: grid;
    grid-template-columns: 220px minmax(0, 1fr);
    background: var(--bg);
  }
  .side {
    background: var(--surface);
    border-right: 1px solid var(--line);
    display: flex;
    flex-direction: column;
    padding: 16px 12px;
    gap: 3px;
    min-height: 0;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 2px 8px 18px;
  }
  .brand b {
    font-family: var(--font-display);
    font-size: 14.5px;
    font-weight: 600;
    display: block;
    line-height: 1.2;
    white-space: nowrap;
  }
  .brand small {
    color: var(--faint);
    font-size: 11px;
  }
  .nav {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px;
    border-radius: var(--radius-s);
    color: var(--muted);
    background: none;
    border: 0;
    cursor: pointer;
    width: 100%;
    text-align: left;
    font-size: 13.5px;
  }
  .nav:hover {
    background: var(--hover);
    color: var(--fg);
  }
  .nav[aria-current="page"] {
    background: var(--accent-soft);
    color: var(--accent-ink);
    font-weight: 500;
  }
  :global(:root[data-theme="red"]) .nav[aria-current="page"] {
    background: linear-gradient(90deg, var(--accent-soft), transparent);
    box-shadow: inset 3px 0 0 var(--accent);
    color: var(--fg);
  }
  .count {
    margin-left: auto;
    font-size: 11px;
    font-family: var(--font-data);
    color: var(--faint);
  }
  .main {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }
  .top {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 20px;
    border-bottom: 1px solid var(--line);
    flex-wrap: wrap;
    min-height: 58px;
  }
  :global(:root[data-theme="red"]) .top {
    background: repeating-linear-gradient(-45deg, transparent 0 11px, rgba(232, 17, 45, 0.045) 11px 22px);
    box-shadow: inset 0 -1px 0 var(--line-strong);
  }
  .top h2 {
    font-family: var(--font-display);
    font-size: 16px;
    margin: 0;
    font-weight: 600;
  }
  :global(:root[data-theme="red"]) .top h2,
  :global(:root[data-theme="red"]) .brand b {
    letter-spacing: 0.06em;
  }
  .view {
    flex: 1;
    overflow: auto;
    min-height: 0;
  }
  .toast {
    position: fixed;
    left: 50%;
    bottom: 28px;
    transform: translateX(-50%);
    background: var(--fg);
    color: var(--bg);
    padding: 9px 16px;
    border-radius: 999px;
    font-size: 13px;
    z-index: 99;
    display: flex;
    gap: 7px;
    align-items: center;
    max-width: calc(100% - 32px);
    pointer-events: none;
  }
</style>
