package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/fileutil"
	"vrc_log_watcher/internal/logwatcher"
	"vrc_log_watcher/internal/models"
	"vrc_log_watcher/internal/notify"
	"vrc_log_watcher/internal/settings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	fileLogger *applog.FileLogger
	watcher    *logwatcher.Watcher
	dispatcher *notify.Dispatcher
	SaveData   models.SaveData
	NoticeLog  models.NoticeLog
}

func NewApp() *App {
	return &App{}
}

// applog.Logger インターフェースの実装
func (a *App) Log(msg string) {
	logMsg := "[DEBUG] [LOG] OutputLog:" + msg
	log.Default().Println(logMsg)
	if a.fileLogger != nil {
		a.fileLogger.WriteLog(logMsg)
	}
}

func (a *App) LogError(err error, context string) {
	if err == nil {
		return
	}
	errorMsg := fmt.Sprintf("[ERROR] %s: %v", context, err)
	log.Default().Println(errorMsg)
	a.SendNoticeLog(errorMsg, "", "[ERROR]", false)
	if a.fileLogger != nil {
		a.fileLogger.WriteLog(errorMsg)
	}
}

// --- Wails ライフサイクル ---

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.LogInfo(ctx, "アプリケーション起動")

	a.fileLogger = applog.NewFileLogger()
	a.watcher = logwatcher.NewWatcher(a)
	a.dispatcher = notify.NewDispatcher(a)
}

func (a *App) shutdown(ctx context.Context) {
	if a.fileLogger != nil {
		a.fileLogger.Close()
	}
}

// --- フロントエンド連携: コンソールログ ---

// OutputConsoleLog フロントエンドから呼ばれるコンソールログ出力（後方互換）
func (a *App) OutputConsoleLog(logstring string) {
	a.Log(logstring)
}

// OutputErrorLog エラーログ出力（後方互換）
func (a *App) OutputErrorLog(err error, context string) {
	a.LogError(err, context)
}

// --- フロントエンド連携: 通知ログ ---

func (a *App) SendNoticeLog(text string, metaData string, title string, canCopy bool) {
	a.SendNoticeLogWithID(text, metaData, title, "", canCopy)
}

func (a *App) SendNoticeLogWithID(text string, metaData string, title string, settingID string, canCopy bool) {
	logTemplate := models.NoticeLog{
		Text:          text,
		MetaData:      metaData,
		Title:         title,
		SettingID:     settingID,
		CanCopy:       canCopy,
		Timestamp:     a.watcher.LastLogTime,
		IsSystem:      strings.Contains(title, "SYSTEM"),
		IsError:       strings.Contains(title, "ERROR") || strings.Contains(title, "WARNING"),
		ActionSuccess: !(strings.Contains(title, "ERROR") || strings.Contains(title, "WARNING")),
	}
	runtime.EventsEmit(a.ctx, "commonLogOutput", logTemplate)

	if a.fileLogger != nil {
		a.fileLogger.WriteLog(fmt.Sprintf("[%s] %s", title, text))
	}
}

// --- フロントエンド連携: ダミー関数（Wailsバインディング用） ---

func (a *App) LoadNoticeLog() models.NoticeLog {
	return a.NoticeLog
}

// --- フロントエンド連携: 設定 ---

func (a *App) GetLastLogTime() string {
	return a.watcher.LastLogTime
}

func (a *App) LoadSetting() models.SaveData {
	a.SendNoticeLog("", "setting.json 読み込み", "[SYSTEM]", false)
	a.SaveData = settings.Load(a)
	a.SendNoticeLog("", "setting.json 読み込みに成功しました", "[SYSTEM]", false)
	return a.SaveData
}

func (a *App) UpdateSetting(ss []models.Setting) {
	a.SaveData.Settings = ss
	settings.Save(a.SaveData, a)
}

// --- フロントエンド連携: フォルダ選択 ---

func (a *App) OpenFolderSelectWindow() string {
	log.Default().Println("[DEBUG] [LOG] OpenFolderSelectWindow")
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "ログファイルフォルダを選択",
	})
	if err != nil {
		a.SendNoticeLog("", "フォルダ選択エラー: "+err.Error(), "[ERROR]", false)
	}

	log.Default().Println("[DEBUG] [LOG] Target Path:" + path)
	a.SaveData.LogPath = path

	jsonData, err := json.Marshal(a.SaveData)
	if err != nil {
		log.Default().Println(err)
		log.Fatal(err)
	}
	if err := os.WriteFile("setting.json", jsonData, 0644); err != nil {
		log.Default().Println(err)
		log.Fatal(err)
	}

	return path
}

// --- フロントエンド連携: ログファイル監視 ---

func (a *App) GetNewestFileName(path string) string {
	if path == "" {
		a.SendNoticeLog("ログフォルダが指定されていません。「フォルダを指定」ボタンをクリックしてVRChatのログフォルダを選択してください。", "", "[WARNING]", false)
		return ""
	}

	name, isNew := a.watcher.GetNewestFileName(path)
	if name == "" {
		return ""
	}

	if isNew {
		a.ReadFile()
		a.SendNoticeLog("ログファイル名: "+name, "", "[SYSTEM]", false)
	}

	return name
}

func (a *App) ResetOffset() {
	a.watcher.ResetOffset()
}

func (a *App) ReadFile() {
	results := a.watcher.ReadFile(a.SaveData.LogPath, a.SaveData.Settings)
	for _, result := range results {
		dispatchResult := a.dispatcher.Send(result.Text, result.Setting, result.IsScreenshot)
		if dispatchResult.Label != "" {
			a.SendNoticeLogWithID(dispatchResult.Message, result.Text, dispatchResult.Label, result.Setting.ID, true)
		}
	}
}

// --- フロントエンド連携: XSOverlay ---

func (a *App) PingXSOverlay() {
	a.dispatcher.XSOverlay.Ping()
}

// --- フロントエンド連携: ファイル操作 ---

func (a *App) GetOutputFolderPath(settingID string) string {
	return notify.GetOutputFolderPath(settingID)
}

func (a *App) OpenInExplorer(folderPath string) error {
	return fileutil.OpenInExplorer(folderPath)
}

func (a *App) OpenFileInExplorer(filePath string) error {
	return fileutil.OpenFileInExplorer(filePath)
}

func (a *App) OpenFile(filePath string) error {
	return fileutil.OpenFile(filePath)
}
