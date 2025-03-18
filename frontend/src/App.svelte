<script lang="ts">
  import logo from "./assets/images/logo-universal.png";
  import { ClipboardSetText } from "../wailsjs/runtime";
  import { OpenFolderSelectWindow } from "../wailsjs/go/main/App.js";
  import { GetNewestFileName } from "../wailsjs/go/main/App.js";
  import { PingXSOverlay } from "../wailsjs/go/main/App.js";
  import { UpdateSetting } from "../wailsjs/go/main/App.js";
  import { LoadSetting } from "../wailsjs/go/main/App.js";
  import { LoadNoticeLog } from "../wailsjs/go/main/App.js";
  import { ReadFile } from "../wailsjs/go/main/App.js";

  import Header from "./Header.svelte";
  import Content from "./Content.svelte";
  import Tabs from "./Tabs.svelte";
  import Footer from "./Footer.svelte";

  import { main } from "../wailsjs/go/models";
  import { text } from "svelte/internal";

  let vrcLogFileName: string = "";
  let intervalId = 0;
  let saveData: main.SaveData;

  let contents: main.Setting[] = [];
  let selectedContent: main.Setting | null = null;
  let noticeLogs: main.NoticeLog[] = [];
  let idCount = 0;

  window.runtime.EventsOn("commonLogOutput", (noticeLog: main.NoticeLog) => {
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
      text: `${new Date().toLocaleTimeString()} Application start`,
      metaData: "",
      title: "[SYSTEM TS]",
    } as main.NoticeLog;
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
        canCopy: false
      } as main.NoticeLog;
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
    const newContent: main.Setting = {
      id: uuid(),
      title: `untitled ${idCount++}`,
      target: "",
      details: "",
      type: "Web Request",
      url: "",
      regexp: "",
      exclude: "",
    };
    contents = [...contents, newContent];
    // 選択を更新
    selectedContent = contents.find((content) => content.id === newContent.id);
    const noticeLog = {
      text: `${new Date().toLocaleTimeString()} 設定を追加しました: ${newContent.id} ${newContent.title}`,
      metaData: "",
      title: "[SYSTEM TS]",
    } as main.NoticeLog;
    noticeLogs = [...noticeLogs, noticeLog];
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  function selectContent(customEvent: CustomEvent<main.Setting>) {
    let selectContent = customEvent.detail;
    selectedContent = contents.find(
      (content) => content.id === selectContent.id,
    );
  }

  // CustomEvent<any>を使っているので、any型で受け取る
  async function updateContent(customEvent: CustomEvent<main.Setting>) {
    // CustomEvent<any> を Content型に変換
    let updateContent = customEvent.detail;
    contents = contents.map((content) =>
      content.id === updateContent.id ? updateContent : content,
    );
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  async function deleteContent(customEvent: CustomEvent<main.Setting>) {
    let deleteContent = customEvent.detail;
    contents = contents.filter((content) => content.id !== deleteContent.id);
    if (contents.length > 0) {
      selectedContent = contents[0];
    } else {
      selectedContent = null;
    }
    const noticeLog = {
      text: `${new Date().toLocaleTimeString()} 削除しました: ${deleteContent.id} ${deleteContent.title}`,
      metaData: "",
      title: "[SYSTEM TS]",
    } as main.NoticeLog;
    noticeLogs = [...noticeLogs, noticeLog];
    await UpdateSetting(contents).then((result) => console.log(result));
  }

  function sendLogEvent(customEvent: CustomEvent<main.NoticeLog>) {
    let event = customEvent.detail;
    noticeLogs = [...noticeLogs, event];
  }

  function clipboardData(customEvent: CustomEvent<string>) {
    let str = customEvent.detail;
    ClipboardSetText(str)
      .then(() => {
        console.log("success");
      })
      .catch((err) => {
        console.log("fail", err);
      });
  }
</script>

<main class="bg-dark-200 text-white min-h-screen">
  <div class="flex flex-col h-screen">
    <Header filename={vrcLogFileName} on:getLogFolderPath={getLogFolderPath} />
    <div class="flex flex-1 overflow-hidden p-2 gap-3">
      <Tabs
        {contents}
        on:selectContent={selectContent}
        on:addContent={addContent}
      />

      {#if selectedContent}
        <Content
          bind:content={selectedContent}
          on:updateContent={updateContent}
          on:deleteContent={deleteContent}
          on:logEvent={sendLogEvent}
        />
      {:else}
        <div class="flex-grow flex items-center justify-center bg-dark-100 rounded-lg shadow-card">
          <div class="text-center p-6">
            <h2 class="text-xl font-bold mb-2">設定が選択されていません</h2>
            <p class="text-gray-400 mb-4">左側のタブから設定を選択するか、新しい設定を追加してください</p>
            <button 
              class="bg-primary-600 hover:bg-primary-700 text-white py-2 px-4 rounded-lg transition-all duration-200 shadow-md"
              on:click={addContent}
            >
              新しい設定を追加
            </button>
          </div>
        </div>
      {/if}
    </div>
    <Footer {noticeLogs} on:clipboardData={clipboardData} />
  </div>
</main>
