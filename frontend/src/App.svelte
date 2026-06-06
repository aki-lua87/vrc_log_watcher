<script lang="ts">
  import { fly } from "svelte/transition";
  import { OpenFolderSelectWindow } from "../wailsjs/go/main/App.js";
  import { GetNewestFileName } from "../wailsjs/go/main/App.js";
  import { PingXSOverlay } from "../wailsjs/go/main/App.js";
  import { UpdateSetting } from "../wailsjs/go/main/App.js";
  import { LoadSetting } from "../wailsjs/go/main/App.js";
  import { LoadNoticeLog } from "../wailsjs/go/main/App.js";
  import { ReadFile } from "../wailsjs/go/main/App.js";

  import MainView from "./MainView.svelte";
  import SettingsView from "./SettingsView.svelte";
  import HelpView from "./HelpView.svelte";

  import { models } from "../wailsjs/go/models";

  // 画面状態管理
  let currentView: "main" | "settings" | "help" = "main";

  let vrcLogFileName: string = "";
  let intervalId = 0;
  let saveData: models.SaveData;

  let contents: models.Setting[] = [];
  let noticeLogs: models.NoticeLog[] = [];
  let idCount = 0;

  window.runtime.EventsOn("commonLogOutput", (noticeLog: models.NoticeLog) => {
    noticeLogs = [...noticeLogs, noticeLog];
  });

  // 使ってない？
  // window.runtime.EventsOn("pushHttpEvent", (eventString) => {
  //   logs = [
  //     ...logs,
  //     `${new Date().toLocaleTimeString()} POST HTTP REQUEST: ${eventString}`,
  //   ];
  // });

  init();
  async function init() {
    await LoadNoticeLog(); // 構造体のロードのためのダミーコール
    const firstLog = {
      text: `${new Date().toLocaleTimeString()} アプリケーション起動`,
      metaData: "",
      title: "[SYSTEM]",
    } as models.NoticeLog;
    noticeLogs = [...noticeLogs, firstLog];
    await LoadSetting().then((result) => (saveData = result));
    contents = saveData.settings;
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
    if (saveData.path == undefined || saveData.path == "") {
      // ログフォルダが指定されていない場合は警告を表示
      const noticeLog = {
        text: `ログフォルダが指定されていません。「フォルダを指定」ボタンをクリックしてVRChatのログフォルダを選択してください。`,
        metaData: "",
        title: "[WARNING]",
        canCopy: false,
      } as models.NoticeLog;
      noticeLogs = [...noticeLogs, noticeLog];
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
      title: `無題 ${idCount++}`,
      target: "",
      details: "",
      isCore: false,
      type: "LogOnly",
      url: "",
      regexp: "",
      exclude: "",
      messageKey: "",
      extraFields: {},
      simpleMode: false,
      simplePattern: "",
      simpleBlocks: [],
    });
    contents = [...contents, newContent];
    const noticeLog = {
      text: `${new Date().toLocaleTimeString()} 設定を追加しました: ${newContent.id} ${newContent.title}`,
      metaData: "",
      title: "[SYSTEM]",
      canCopy: false,
      timestamp: "",
      isSystem: true,
      isError: false,
      actionSuccess: true,
    } as models.NoticeLog;
    noticeLogs = [...noticeLogs, noticeLog];
    await UpdateSetting(contents).then((result) => console.log(result));
    
    // 設定画面を開く
    currentView = "settings";
  }

  async function updateContent(setting: models.Setting) {
    contents = contents.map((content) =>
      content.id === setting.id ? setting : content,
    );
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  async function deleteContent(setting: models.Setting) {
    contents = contents.filter((content) => content.id !== setting.id);
    const noticeLog = {
      text: `${new Date().toLocaleTimeString()} 削除しました: ${setting.id} ${setting.title}`,
      metaData: "",
      title: "[SYSTEM]",
      canCopy: false,
      timestamp: "",
      isSystem: true,
      isError: false,
      actionSuccess: true,
    } as models.NoticeLog;
    noticeLogs = [...noticeLogs, noticeLog];
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  async function handleReorderContents(newContents: models.Setting[]) {
    contents = newContents;
    await UpdateSetting(contents).then((result) => console.log(result));
    const noticeLog = {
      text: `${new Date().toLocaleTimeString()} 設定の順序を変更しました`,
      metaData: "",
      title: "[SYSTEM]",
      canCopy: false,
      timestamp: "",
      isSystem: true,
      isError: false,
      actionSuccess: true,
    } as models.NoticeLog;
    noticeLogs = [...noticeLogs, noticeLog];
  }

  // 画面切り替え
  function openSettings() {
    currentView = "settings";
  }

  function closeSettings() {
    currentView = "main";
  }

  function openHelp() {
    currentView = "help";
  }

  function closeHelp() {
    currentView = "main";
  }
</script>

<main class="bg-dark-200 text-white min-h-screen relative">
  <MainView
    {noticeLogs}
    {vrcLogFileName}
    logFolderPath={saveData?.path || ""}
    onOpenSettings={openSettings}
    onGetLogFolder={getLogFolderPath}
    onOpenHelp={openHelp}
  />
  
  {#if currentView === "help"}
    <div
      class="fixed inset-0 z-50 flex items-end"
      on:click={closeHelp}
      on:keydown={(e) => e.key === 'Escape' && closeHelp()}
      role="button"
      tabindex="0"
    >
      <div
        class="w-full bg-dark-100 rounded-t-2xl shadow-2xl h-[80vh] overflow-hidden"
        on:click|stopPropagation
        on:keydown={(e) => e.key === 'Escape' && closeHelp()}
        role="dialog"
        tabindex="-1"
        transition:fly={{ y: 1000, duration: 400 }}
      >
        <HelpView onBack={closeHelp} />
      </div>
    </div>
  {/if}

  {#if currentView === "settings"}
    <div 
      class="fixed inset-0 z-50 flex items-end" 
      on:click={closeSettings}
      on:keydown={(e) => e.key === 'Escape' && closeSettings()}
      role="button"
      tabindex="0"
    >
      <div 
        class="w-full bg-dark-100 rounded-t-2xl shadow-2xl h-[80vh] overflow-hidden"
        on:click|stopPropagation
        on:keydown={(e) => e.key === 'Escape' && closeSettings()}
        role="dialog"
        tabindex="-1"
        transition:fly={{ y: 1000, duration: 400 }}
      >
        <SettingsView
          settings={contents}
          onBack={closeSettings}
          onUpdate={updateContent}
          onDelete={deleteContent}
          onAdd={addContent}
          onReorder={handleReorderContents}
        />
      </div>
    </div>
  {/if}
</main>
