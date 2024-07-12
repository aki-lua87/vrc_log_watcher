package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.xsWS = nil
	runtime.LogInfo(ctx, "Application Startup called!")
}

func (a *App) OutputLog(logstring string) {
	log.Default().Println("[DEBUG] [LOG] OutputLog:" + logstring)
}

func (a *App) LoadSetting() SaveData {
	log.Default().Println("[DEBUG] [LOG] Load Setting")
	runtime.EventsEmit(a.ctx, "commonLogOutput", "setting.json Load")
	// 設定ファイルの読み込み
	file, err := os.ReadFile("setting.json")
	if err != nil {
		runtime.EventsEmit(a.ctx, "commonLogOutput", "Setting file read error:"+err.Error())
		a.UpdateSetting([]Setting{})
		return a.SaveData
	}
	// JSONをStructに変換
	var saveData SaveData
	err = json.Unmarshal(file, &saveData)
	if err != nil {
		runtime.EventsEmit(a.ctx, "commonLogOutput", "ERRPR:"+err.Error())
	}
	log.Default().Println(saveData)
	runtime.EventsEmit(a.ctx, "commonLogOutput", "Target log folder:"+saveData.LogPath)
	// a.SaveData.LogPath = saveData.LogPath
	a.SaveData = saveData
	return saveData
}

func (a *App) UpdateSetting(ss []Setting) {
	log.Default().Println("[DEBUG] [LOG] UpdateSetting:", len(ss))
	a.SaveData.Settings = ss
	// StructをJSONに変換
	jsonData, err := json.Marshal(a.SaveData)
	if err != nil {
		runtime.EventsEmit(a.ctx, "commonLogOutput", "ERRPR:"+err.Error())
	}
	// runtime.EventsEmit(a.ctx, "commonLogOutput", string(jsonData))
	// JSONをファイルに書き込む
	err = os.WriteFile("setting.json", jsonData, 0644)
	if err != nil {
		runtime.EventsEmit(a.ctx, "commonLogOutput", "ERRPR:"+err.Error())
	}
	// runtime.EventsEmit(a.ctx, "commonLogOutput", "Setting Updated Successfully")
}

func (a *App) OpenFolderSelectWindow() string {
	log.Default().Println("[DEBUG] [LOG] OpenFolderSelectWindow")
	// フォルダ選択ダイアログを開く
	// 選択されたフォルダのパスを返す
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select LogFile Folder",
	})
	if err != nil {
		runtime.EventsEmit(a.ctx, "commonLogOutput", "ERRPR:"+err.Error())
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
	// log.Default().Println("[DEBUG] [LOG] GetNewestFileName")
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	var newestFile os.DirEntry
	var newestTime time.Time
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				runtime.EventsEmit(a.ctx, "commonLogOutput", "ERROR:"+err.Error())
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
		runtime.EventsEmit(a.ctx, "commonLogOutput", "監視対象のファイルが見つかりませんでした")
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
		runtime.EventsEmit(a.ctx, "commonLogOutput", "Reading log file name:"+a.targetFileName)
		return newestFile.Name() // Viewへ反映
	}
	runtime.EventsEmit(a.ctx, "commonLogOutput", "監視対象のファイルが見つかりませんでした")
	return ""
}

var lastOffset int64
var isWatchFileRunning bool

func (a *App) ResetOffset() {
	// runtime.EventsEmit(a.ctx, "commonLogOutput", "Reset And Read New File")
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
		log.Default().Println("[ERROR] [LOG] ", err.Error())
		isWatchFileRunning = false
		return
	}
	defer file.Close()
	_, err = file.Seek(lastOffset, 0)
	if err != nil {
		log.Default().Println("[ERROR] [LOG] ", err.Error())
		isWatchFileRunning = false
		return
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		a.evaluateLine(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Default().Println("[ERROR] [LOG] ", err.Error())
		isWatchFileRunning = false
		return
	}
	lastOffset, err = file.Seek(0, io.SeekCurrent)
	if err != nil {
		log.Default().Println("[ERROR] [LOG] ", err.Error())
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
				a.OutputLog(setting.Title + " : " + text)
				// setting.Type によって処理を分岐
				if setting.Type == "WebRequest" {
					message := a.postHttpRequest(text, setting.Title, setting.URL, setting.RegExp, setting.Details)
					runtime.EventsEmit(a.ctx, "commonLogOutput", message)
				} else if setting.Type == "SendXSOverlay" {
					message := a.postXSOverlay(text, setting.Title)
					runtime.EventsEmit(a.ctx, "commonLogOutput", message)
				} else if setting.Type == "SendDiscordWebHook" {
					message := postDiscordWebhook(text, setting.Title, setting.URL)
					runtime.EventsEmit(a.ctx, "commonLogOutput", message)
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
	data_json, _ := json.Marshal(data)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(data_json))
	if err != nil {
		return err.Error()
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err.Error()
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
