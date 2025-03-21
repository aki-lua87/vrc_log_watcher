package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx            context.Context
	xsWS           *websocket.Conn
	targetFileName string
	SaveData       SaveData
	NoticeLog      NoticeLog
	appLogFile     *os.File
}

type NoticeLog struct {
	Text     string `json:"text"`
	MetaData string `json:"metaData"`
	Title    string `json:"title"`
	CanCopy  bool   `json:"canCopy"`
}

type SaveData struct {
	LogPath  string    `json:"path"`
	Settings []Setting `json:"settings"`
}

type HttpRequestModel struct {
	Message     string `json:"message"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Regexp      string `json:"regexp"`
}

type XSOApiObject struct {
	Sender   string `json:"sender"`
	Target   string `json:"target"`
	Command  string `json:"command"`
	JsonData string `json:"jsonData"`
	RawData  string `json:"rawData"`
	// Timeout       int     `json:"timeout"`
	// Volume        float32 `json:"volume"`
	// AudioPath     string  `json:"audioPath"`
	// UseBase64Icon bool    `json:"useBase64Icon"`
	// Icon          string  `json:"icon"`
	// Opacity       float32 `json:"opacity"`
}

type XSONotificationObject struct {
	Type    int     `json:"type"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Timeout float32 `json:"timeout"`
	Height  float32 `json:"height"`
	// Height        float32 `json:"height"`
	// SourceApp     string  `json:"sourceApp"`
	// Timeout       int     `json:"timeout"`
	// Volume        float32 `json:"volume"`
	// AudioPath     string  `json:"audioPath"`
	// UseBase64Icon bool    `json:"useBase64Icon"`
	// Icon          string  `json:"icon"`
	// Opacity       float32 `json:"opacity"`
}

type LogOutputModel struct {
}

type Setting struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Details string `json:"details"`
	Target  string `json:"target"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	RegExp  string `json:"regexp"`
	Exclude string `json:"exclude"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.xsWS = nil
	runtime.LogInfo(ctx, "Application Startup called!")

	// アプリケーションログファイルの初期化
	a.initAppLogFile()
}

// アプリケーションログファイルの初期化
func (a *App) initAppLogFile() {
	// ログディレクトリの作成
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

	// 現在の日時を取得してファイル名に使用
	currentTime := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logDir, "app_log_"+currentTime+".txt")

	// ログファイルを作成
	file, err := os.Create(logFilePath)
	if err != nil {
		log.Printf("[ERROR] ログファイルの作成に失敗しました: %v", err)
		return
	}

	a.appLogFile = file

	// アプリケーション起動ログを書き込む
	startupMsg := fmt.Sprintf("[%s] アプリケーションを起動しました\n", time.Now().Format("2006/01/02 15:04:05"))
	a.appLogFile.WriteString(startupMsg)
}

// アプリケーション終了時にログファイルを閉じる
func (a *App) shutdown(ctx context.Context) {
	if a.appLogFile != nil {
		a.appLogFile.WriteString(fmt.Sprintf("[%s] アプリケーションを終了しました\n", time.Now().Format("2006/01/02 15:04:05")))
		a.appLogFile.Close()
	}
}

func (a *App) OutputConsoleLog(logstring string) {
	logMsg := "[DEBUG] [LOG] OutputLog:" + logstring
	log.Default().Println(logMsg)

	// ログファイルにも書き込む
	if a.appLogFile != nil {
		a.appLogFile.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006/01/02 15:04:05"), logMsg))
	}
}

// エラーログを出力する関数
func (a *App) OutputErrorLog(err error, context string) {
	if err == nil {
		return
	}

	errorMsg := fmt.Sprintf("[ERROR] %s: %v", context, err)
	log.Default().Println(errorMsg)

	// エラーログをアプリケーションログに追加
	a.SendNoticeLog(errorMsg, "", "[ERROR]", false)

	// ログファイルにも書き込む
	if a.appLogFile != nil {
		a.appLogFile.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006/01/02 15:04:05"), errorMsg))
	}
}

func (a *App) SendNoticeLog(text string, metaData string, title string, canCopy bool) {
	var logTemplate NoticeLog
	logTemplate.Text = text
	logTemplate.MetaData = metaData
	logTemplate.Title = title
	logTemplate.CanCopy = canCopy
	runtime.EventsEmit(a.ctx, "commonLogOutput", logTemplate)

	// ログファイルにも書き込む
	if a.appLogFile != nil {
		a.appLogFile.WriteString(fmt.Sprintf("[%s] [%s] %s\n", time.Now().Format("2006/01/02 15:04:05"), title, text))
	}
}

// dummy function NOTE: on load ts for models
func (a *App) LoadNoticeLog() NoticeLog {
	return a.NoticeLog
}

func (a *App) LoadSetting() SaveData {
	log.Default().Println("[DEBUG] [LOG] Load Setting")
	logTemplateText := ""
	logTemplateMetaData := ""
	logTemplateTitle := "[SYSTEM GO]"

	logTemplateTitle = "setting.json Load"
	a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
	// 設定ファイルの読み込み
	file, err := os.ReadFile("setting.json")
	if err != nil {
		logTemplateTitle = "setting.json Load Error" + err.Error()
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
		a.UpdateSetting([]Setting{})
		return a.SaveData
	}
	// JSONをStructに変換
	var saveData SaveData
	err = json.Unmarshal(file, &saveData)
	if err != nil {
		logTemplateText = "setting.json Unmarshal Error" + err.Error()
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
	}
	log.Default().Println(saveData)
	logTemplateText = "setting.json Loaded Successfully"
	a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
	a.SaveData = saveData
	return saveData
}

func (a *App) UpdateSetting(ss []Setting) {
	log.Default().Println("[DEBUG] [LOG] UpdateSetting:", len(ss))
	logTemplateText := ""
	logTemplateMetaData := ""
	logTemplateTitle := "[SYSTEM GO]"
	a.SaveData.Settings = ss
	// StructをJSONに変換
	jsonData, err := json.Marshal(a.SaveData)
	if err != nil {
		logTemplateText = "setting.json Marshal Error" + err.Error()
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
	}
	// JSONをファイルに書き込む
	err = os.WriteFile("setting.json", jsonData, 0644)
	if err != nil {
		logTemplateText = "setting.json Write Error" + err.Error()
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
	}
}

func (a *App) OpenFolderSelectWindow() string {
	log.Default().Println("[DEBUG] [LOG] OpenFolderSelectWindow")
	logTemplateText := ""
	logTemplateMetaData := ""
	logTemplateTitle := "[SYSTEM GO]"
	// フォルダ選択ダイアログを開く
	// 選択されたフォルダのパスを返す
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select LogFile Folder",
	})
	if err != nil {
		logTemplateText = "OpenFolderSelectWindow Error" + err.Error()
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
	}
	log.Default().Println("[DEBUG] [LOG] Target Path:" + path)
	// JSONに保存
	// saveData := SaveData{LogPath: path}
	a.SaveData.LogPath = path
	// StructをJSONに変換
	jsonData, err := json.Marshal(a.SaveData)
	if err != nil {
		log.Default().Println(err)
		log.Fatal(err)
	}
	// JSONをファイルに書き込む
	err = os.WriteFile("setting.json", jsonData, 0644)
	if err != nil {
		log.Default().Println(err)
		log.Fatal(err)
	}
	a.SaveData.LogPath = path
	return path
}

// フォルダ内の最新のtxtファイルを探索し、そのファイル名を返す
func (a *App) GetNewestFileName(path string) string {
	logTemplateText := ""
	logTemplateMetaData := ""
	logTemplateTitle := "[SYSTEM GO]"

	// ログフォルダが指定されていない場合
	if path == "" {
		logTemplateText = "ログフォルダが指定されていません。「フォルダを指定」ボタンをクリックしてVRChatのログフォルダを選択してください。"
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, "[WARNING]", false)
		return ""
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		errorMsg := fmt.Sprintf("ログフォルダの読み取りに失敗しました: %s", err.Error())
		a.SendNoticeLog(errorMsg, logTemplateMetaData, "[ERROR]", false)
		a.OutputErrorLog(err, "ログフォルダの読み取り")
		return ""
	}

	var newestFile os.DirEntry
	var newestTime time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				a.OutputErrorLog(err, "ファイル情報の取得")
				continue
			}
			// 拡張子が.txtのファイルのみを対象とする
			if filepath.Ext(entry.Name()) != ".txt" {
				continue
			}
			if info.IsDir() {
				continue
			}
			if info.ModTime().After(newestTime) {
				newestFile = entry
				newestTime = info.ModTime()
			}
		}
	}

	if newestFile == nil {
		logTemplateText = "ログフォルダ内にログファイル(.txt)が見つかりません。VRChatのログフォルダを正しく指定しているか確認してください。"
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, "[WARNING]", false)
		return ""
	}

	if newestFile.Name() == a.targetFileName {
		return a.targetFileName // Viewへ反映(別に更新しなくてもいいけど)
	}

	if newestFile != nil {
		log.Default().Println("[DEBUG] [LOG] 監視対象を変更します")
		a.targetFileName = newestFile.Name()
		a.ResetOffset() // オフセット削除
		a.ReadFile()    // 初回内容読み取り
		logTemplateText = "Reading log file name:" + a.targetFileName
		a.SendNoticeLog(logTemplateText, logTemplateMetaData, logTemplateTitle, false)
		return newestFile.Name() // Viewへ反映
	}

	return ""
}

var lastOffset int64
var isWatchFileRunning bool

func (a *App) ResetOffset() {
	lastOffset = 0
}

func (a *App) ReadFile() {
	if a.targetFileName == "" {
		log.Default().Println("[DEBUG] [LOG] No FileName")
		return
	}
	if a.SaveData.LogPath == "" {
		log.Default().Println("[DEBUG] [LOG] No LogPath")
		return
	}
	if isWatchFileRunning {
		log.Default().Println("[DEBUG] [LOG] Watching now")
		return
	}

	isWatchFileRunning = true
	path := a.SaveData.LogPath + "\\" + a.targetFileName
	file, err := os.Open(path)
	if err != nil {
		a.OutputErrorLog(err, "ログファイルのオープン")
		isWatchFileRunning = false
		return
	}
	defer file.Close()

	_, err = file.Seek(lastOffset, 0)
	if err != nil {
		a.OutputErrorLog(err, "ファイルシーク")
		isWatchFileRunning = false
		return
	}

	// スキャナーのバッファサイズを増やす（デフォルトは64KB）
	scanner := bufio.NewScanner(file)
	// バッファを確保
	const maxCapacity = 1000000 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		a.evaluateLine(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		a.OutputErrorLog(err, "ファイルスキャン")
		isWatchFileRunning = false
		return
	}

	lastOffset, err = file.Seek(0, io.SeekCurrent)
	if err != nil {
		a.OutputErrorLog(err, "ファイルオフセット取得")
		isWatchFileRunning = false
		return
	}

	isWatchFileRunning = false
}

// 行の評価
func (a *App) evaluateLine(line string) {
	if lastOffset == 0 {
		return
	}
	// a.SaveData.Settings をループさせる
	for _, setting := range a.SaveData.Settings {
		if setting.RegExp != "" {
			pattern := regexp.MustCompile(setting.RegExp)
			matches := pattern.FindStringSubmatch(line)
			text := ""
			if len(matches) > 1 {
				text = strings.Join(matches[1:], "")
			}
			if text != "" {
				// オフセットが0の場合は初回読み込みと判断してスキップ
				if lastOffset == 0 {
					continue
				}
				// 除外文字と一致する場合はスキップ
				if setting.Exclude == text {
					continue
				}
				a.OutputConsoleLog(setting.Title + " : " + text)
				// setting.Type によって処理を分岐
				if setting.Type == "WebRequest" {
					message := a.postHttpRequest(text, setting.Title, setting.URL, setting.RegExp, setting.Details)
					a.SendNoticeLog(message, text, setting.Title, true)
				} else if setting.Type == "SendXSOverlay" {
					message := a.postXSOverlay(text, setting.Title)
					a.SendNoticeLog(message, text, setting.Title, true)
				} else if setting.Type == "SendDiscordWebHook" {
					message := postDiscordWebhook(text, setting.Title, setting.URL)
					a.SendNoticeLog(message, text, setting.Title, true)
				} else if setting.Type == "OutputTextFile" {
					message := outputTextFile(text, setting.Title)
					a.SendNoticeLog(message, text, setting.Title, true)
				} else if setting.Type == "Disable" {
					a.SendNoticeLog("[何もしない] "+setting.Title+": "+text, text, setting.Title, true)
				}
			}
		}
	}
}

func (a *App) postHttpRequest(eventString string, title string, url string, regx string, desc string) string {
	if url == "" {
		return "URL is empty"
	}
	// url形式じゃない場合の処理
	if !strings.HasPrefix(url, "http") {
		return "URL is invalid"
	}

	data := new(HttpRequestModel)
	data.Message = eventString
	data.Title = title
	data.Description = desc
	data.Regexp = regx

	data_json, err := json.Marshal(data)
	if err != nil {
		errorMsg := fmt.Sprintf("JSONのマーシャルに失敗しました: %s", err.Error())
		a.OutputErrorLog(err, "JSONマーシャル")
		return errorMsg
	}

	res, err := http.Post(url, "application/json", bytes.NewBuffer(data_json))
	if err != nil {
		errorMsg := err.Error()
		a.OutputErrorLog(err, "HTTPリクエスト")
		return errorMsg
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		errorMsg := err.Error()
		a.OutputErrorLog(err, "レスポンス読み取り")
		return errorMsg
	}

	log.Default().Println(string(body))
	return "[Web Request] Sent Successfully: " + title + ": " + eventString
}

func (a *App) PingXSOverlay() {
	if a.xsWS != nil {
		log.Default().Println("[DEBUG] [LOG] a.xsWS != nil")
		rwerrr1 := a.xsWS.WriteMessage(2, []byte{})
		if rwerrr1 != nil {
			log.Default().Println("[DEBUG] [LOG] a.xsWS.WriteMessage Error:" + rwerrr1.Error())
			a.xsWS.Close()
			a.xsWS = nil
		}
	} else {
		log.Default().Println("[DEBUG] [LOG] a.xsWS == nil")
	}
}

func (a *App) postXSOverlay(eventString string, title string) string {
	// XSOverlayへ通知の送信
	// https://xsoverlay.vercel.app/Developer/API/websockets/apicommands
	notification := new(XSONotificationObject)
	notification.Type = 1
	notification.Title = title
	notification.Content = eventString
	notification.Timeout = 1.2
	notification.Height = 100.0
	notification_json, _ := json.Marshal(notification)
	notification_json_str := string(notification_json)
	url := "localhost"
	port := "42070"

	apiObject := new(XSOApiObject)
	apiObject.Sender = "VRCLogWatcher"
	apiObject.Target = "XSOverlay"
	apiObject.Command = "SendNotification"
	apiObject.JsonData = notification_json_str
	if a.xsWS == nil {
		log.Default().Println("[DEBUG] [LOG] a.xsWS != nil")
		ws, _, err := websocket.DefaultDialer.Dial("ws://"+url+":"+port+"/?client=VRCLogWatcher", nil)
		if err != nil {
			// log.Fatal(err)
			a.xsWS = nil
			return err.Error()
		}
		a.xsWS = ws
	}
	// defer ws.Close()
	err := a.xsWS.WriteJSON(apiObject)
	if err != nil {
		// log.Fatal(err)
		a.xsWS.Close()
		// retry
		rws, _, errr1 := websocket.DefaultDialer.Dial("ws://"+url+":"+port+"/?client=VRCLogWatcher", nil)
		if errr1 != nil {
			// log.Fatal(err)
			a.xsWS = nil
			return errr1.Error()
		}
		a.xsWS = rws
		errr2 := a.xsWS.WriteJSON(apiObject)
		if errr2 != nil {
			a.xsWS.Close()
			a.xsWS = nil
		}
		return err.Error()
	}
	return "[XS Overlay] Notification Sent Successfully: " + title + ": " + eventString
}

func postDiscordWebhook(eventString string, title string, webhookURL string) string {
	// メッセージの内容を定義
	message := map[string]string{
		"content": title + ": " + eventString,
	}
	// メッセージをJSON形式に変換
	jsonData, err := json.Marshal(message)
	if err != nil {
		return err.Error()
	}
	// HTTP POSTリクエストを作成
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err.Error()
	}
	// Content-Typeを設定
	req.Header.Set("Content-Type", "application/json")
	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	return "[Discord Webhook] Sent Successfully: " + title + ": " + eventString
}

func outputTextFile(eventString string, title string) string {
	t := time.Now()
	// フォルダが存在しない場合は作成 フォルダ名は YYYYMMDD
	folderName := t.Format("20060102") + "_" + title
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		os.Mkdir(folderName, 0777)
	}
	// ファイル名は HHmmssSSS.txt
	fileName := t.Format("150405.000")
	// ファイル名から.を削除
	fileName = strings.Replace(fileName, ".", "", 1)
	// ファイルパスはこのアプリケーションの実行フォルダ内の YYYYMMDD/HHmmssSSS.txt
	filePath := folderName + "/" + fileName + ".txt"
	// ファイルを作成
	file, err := os.Create(filePath)
	if err != nil {
		return err.Error()
	}
	defer file.Close()
	// ファイルに書き込み
	_, err = file.WriteString(eventString)
	if err != nil {
		return err.Error()
	}
	return "[Log Output] Successfully " + title + ":" + filePath
}
