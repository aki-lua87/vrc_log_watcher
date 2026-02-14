package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	rt "runtime"
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
	lastLogTime    string
}

type NoticeLog struct {
	Text          string `json:"text"`
	MetaData      string `json:"metaData"`
	Title         string `json:"title"`
	SettingID     string `json:"settingId"` // 設定のID
	CanCopy       bool   `json:"canCopy"`
	Timestamp     string `json:"timestamp"`
	IsSystem      bool   `json:"isSystem"`      // システムログかどうか
	IsError       bool   `json:"isError"`       // エラーログかどうか
	ActionSuccess bool   `json:"actionSuccess"` // アクション成功/失敗
}

type SaveData struct {
	LogPath  string    `json:"path"`
	Settings []Setting `json:"settings"`
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

type Setting struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Details       string            `json:"details"`
	IsCore        bool              `json:"isCore"` // 基本機能フラグ
	Target        string            `json:"target"`
	Type          string            `json:"type"`
	URL           string            `json:"url"`
	RegExp        string            `json:"regexp"`
	Exclude       string            `json:"exclude"`
	MessageKey    string            `json:"messageKey"`    // メッセージのキー名
	ExtraFields   map[string]string `json:"extraFields"`   // 追加フィールド
	SimpleMode    bool              `json:"simpleMode"`    // 簡易入力モード有効/無効
	SimplePattern string            `json:"simplePattern"` // 簡易パターンタイプ
	SimpleBlocks  []SimpleBlock     `json:"simpleBlocks"`  // 簡易入力のブロック定義
}

type SimpleBlock struct {
	Type  string `json:"type"`  // "between", "contains", "after", etc.
	Start string `json:"start"` // 開始文字列
	End   string `json:"end"`   // 終了文字列（必要な場合）
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.xsWS = nil
	runtime.LogInfo(ctx, "アプリケーション起動")

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
	a.SendNoticeLogWithID(text, metaData, title, "", canCopy)
}

func (a *App) SendNoticeLogWithID(text string, metaData string, title string, settingID string, canCopy bool) {
	var logTemplate NoticeLog
	logTemplate.Text = text
	logTemplate.MetaData = metaData
	logTemplate.Title = title
	logTemplate.SettingID = settingID
	logTemplate.CanCopy = canCopy
	logTemplate.Timestamp = a.lastLogTime
	logTemplate.IsSystem = strings.Contains(title, "SYSTEM")
	logTemplate.IsError = strings.Contains(title, "ERROR") || strings.Contains(title, "WARNING")
	logTemplate.ActionSuccess = !logTemplate.IsError
	runtime.EventsEmit(a.ctx, "commonLogOutput", logTemplate)

	// ログファイルにも書き込む
	if a.appLogFile != nil {
		a.appLogFile.WriteString(fmt.Sprintf("[%s] [%s] %s\n", time.Now().Format("2006/01/02 15:04:05"), title, text))
	}
}

// VRCイベントの基本機能プリセットを初期化
func (a *App) initCoreSettings() []Setting {
	coreSettings := []Setting{
		{
			ID:      "core-world-join-name",
			Title:   "ワールド入室 (名前)",
			Details: "ワールド入室時のワールド名を取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[Behaviour\] Entering Room: (.+)`,
			Exclude: "",
		},
		{
			ID:      "core-world-join-id",
			Title:   "ワールド入室 (ID)",
			Details: "ワールド入室時のワールドIDを取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `Joining[g]*\s+(wrld_[a-f0-9-]+)`,
			Exclude: "",
		},
		{
			ID:      "core-user-join",
			Title:   "ユーザー入室",
			Details: "ユーザー入室時のユーザー名を取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[Behaviour\] OnPlayerJoinComplete (.+)`,
			Exclude: "",
		},
		{
			ID:      "core-user-left",
			Title:   "ユーザー退室",
			Details: "ユーザー退室時のユーザー名を取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[Behaviour\] OnPlayerLeft (.+)`,
			Exclude: "",
		},
		{
			ID:      "core-screenshot",
			Title:   "スクリーンショット",
			Details: "写真撮影時のファイルパスを取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[VRC Camera\] Took screenshot to: (.+)`,
			Exclude: "",
		},
	}
	return coreSettings
}

// dummy function NOTE: on load ts for models
func (a *App) LoadNoticeLog() NoticeLog {
	return a.NoticeLog
}

// 最新のVRCログ時刻を取得
func (a *App) GetLastLogTime() string {
	return a.lastLogTime
}

// ログヘルパー関数
func (a *App) sendSystemLog(text, title string) {
	a.SendNoticeLog(text, "", title, false)
}

func (a *App) sendSystemError(text, title string) {
	a.SendNoticeLog(text, "", title, false)
}

func (a *App) LoadSetting() SaveData {
	log.Default().Println("[DEBUG] [LOG] Load Setting")

	a.sendSystemLog("", "setting.json 読み込み")
	// 設定ファイルの読み込み
	file, err := os.ReadFile("setting.json")
	if err != nil {
		a.sendSystemError("", "setting.json 読み込みエラー: "+err.Error())
		// 初回起動時は基本機能をセットアップ
		coreSettings := a.initCoreSettings()
		a.UpdateSetting(coreSettings)
		a.SaveData.Settings = coreSettings
		return a.SaveData
	}
	// JSONをStructに変換
	var saveData SaveData
	err = json.Unmarshal(file, &saveData)
	if err != nil {
		a.sendSystemError("", "setting.json パースエラー: "+err.Error())
	}

	// 設定が空の場合は基本機能を追加
	if len(saveData.Settings) == 0 {
		coreSettings := a.initCoreSettings()
		saveData.Settings = coreSettings
		a.UpdateSetting(coreSettings)
	} else {
		// 既存設定に基本機能が存在しない場合は追加、存在する場合はタイトルとDetailsを更新
		coreSettings := a.initCoreSettings()
		coreSettingsMap := make(map[string]Setting)
		for _, coreSetting := range coreSettings {
			coreSettingsMap[coreSetting.ID] = coreSetting
		}

		// 既存の基本機能のタイトルとDetailsを更新
		needsUpdate := false
		for i, setting := range saveData.Settings {
			if coreSetting, exists := coreSettingsMap[setting.ID]; exists {
				// 基本機能のタイトルまたはDetailsが古い場合は更新
				if setting.Title != coreSetting.Title || setting.Details != coreSetting.Details {
					saveData.Settings[i].Title = coreSetting.Title
					saveData.Settings[i].Details = coreSetting.Details
					needsUpdate = true
				}
				// 既存IDから削除（追加済み）
				delete(coreSettingsMap, setting.ID)
			}
		}

		// 存在しない基本機能を追加
		for _, coreSetting := range coreSettings {
			if _, exists := coreSettingsMap[coreSetting.ID]; exists {
				saveData.Settings = append([]Setting{coreSetting}, saveData.Settings...)
				needsUpdate = true
			}
		}

		// 更新があった場合は保存
		if needsUpdate {
			a.UpdateSetting(saveData.Settings)
		}
	}

	log.Default().Println(saveData)
	a.sendSystemLog("", "setting.json 読み込みに成功しました")
	a.SaveData = saveData
	return saveData
}

func (a *App) UpdateSetting(ss []Setting) {
	log.Default().Println("[DEBUG] [LOG] UpdateSetting:", len(ss))
	a.SaveData.Settings = ss
	// StructをJSONに変換
	jsonData, err := json.Marshal(a.SaveData)
	if err != nil {
		a.sendSystemError("", "setting.json 変換エラー: "+err.Error())
	}
	// JSONをファイルに書き込む
	err = os.WriteFile("setting.json", jsonData, 0644)
	if err != nil {
		a.sendSystemError("", "setting.json 書き込みエラー: "+err.Error())
	}
}

func (a *App) OpenFolderSelectWindow() string {
	log.Default().Println("[DEBUG] [LOG] OpenFolderSelectWindow")
	// フォルダ選択ダイアログを開く
	// 選択されたフォルダのパスを返す
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "ログファイルフォルダを選択",
	})
	if err != nil {
		a.sendSystemError("", "フォルダ選択エラー: "+err.Error())
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

	// ログフォルダが指定されていない場合
	if path == "" {
		a.sendSystemLog("ログフォルダが指定されていません。「フォルダを指定」ボタンをクリックしてVRChatのログフォルダを選択してください。", "[WARNING]")
		return ""
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		errorMsg := fmt.Sprintf("ログフォルダの読み取りに失敗しました: %s", err.Error())
		a.sendSystemError(errorMsg, "[ERROR]")
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
		a.sendSystemLog("ログフォルダ内にログファイル(.txt)が見つかりません。VRChatのログフォルダを正しく指定しているか確認してください。", "[WARNING]")
		return ""
	}

	if newestFile.Name() == a.targetFileName {
		return a.targetFileName // Viewへ反映(別に更新しなくてもいいけど)
	}

	log.Default().Println("[DEBUG] [LOG] 監視対象を変更します")
	a.targetFileName = newestFile.Name()
	a.ResetOffset() // オフセット削除
	a.ReadFile()    // 初回内容読み取り
	a.sendSystemLog("ログファイル名: "+a.targetFileName, "[SYSTEM]")
	return newestFile.Name() // Viewへ反映
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

// VRCログから時刻を抽出
func (a *App) extractTimeFromVRCLog(line string) {
	// VRCログの時刻形式: "2025.06.23 10:53:50"
	timePattern := regexp.MustCompile(`^(\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2})`)
	matches := timePattern.FindStringSubmatch(line)
	if len(matches) > 1 {
		a.lastLogTime = matches[1]
	}
}

// 行の評価
func (a *App) evaluateLine(line string) {
	// VRCログから時刻を抽出
	a.extractTimeFromVRCLog(line)

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

				// スクリーンショットの場合は画像を送信
				isScreenshot := setting.ID == "core-screenshot"

				// setting.Type によって処理を分岐
				if setting.Type == "WebRequest" {
					var message string
					if isScreenshot {
						message = a.postHttpRequestWithImage(text, setting)
					} else {
						message = a.postHttpRequest(text, setting)
					}
					a.SendNoticeLogWithID(message, text, setting.Title, setting.ID, true)
				} else if setting.Type == "SendXSOverlay" {
					message := a.postXSOverlay(text, setting.Title)
					a.SendNoticeLogWithID(message, text, setting.Title, setting.ID, true)
				} else if setting.Type == "SendDiscordWebHook" {
					var message string
					if isScreenshot {
						message = a.postDiscordWebhookWithImage(text, setting.Title, setting.URL)
					} else {
						message = postDiscordWebhook(text, setting.Title, setting.URL)
					}
					a.SendNoticeLogWithID(message, text, setting.Title, setting.ID, true)
				} else if setting.Type == "OutputTextFile" {
					message := outputTextFile(text, setting.ID, setting.Title)
					a.SendNoticeLogWithID(message, text, setting.Title, setting.ID, true)
				} else if setting.Type == "Disable" {
					a.SendNoticeLogWithID("[何もしない] "+setting.Title+": "+text, text, setting.Title, setting.ID, true)
				}
			}
		}
	}
}

func (a *App) postHttpRequest(eventString string, setting Setting) string {
	if setting.URL == "" {
		return "URLが空です"
	}
	// url形式じゃない場合の処理
	if !strings.HasPrefix(setting.URL, "http") {
		return "URLが無効です"
	}

	// 動的にJSONペイロードを構築
	data := make(map[string]interface{})

	// メッセージキーが指定されていない場合はデフォルトで"message"を使用
	messageKey := setting.MessageKey
	if messageKey == "" {
		messageKey = "message"
	}
	data[messageKey] = eventString

	// 追加フィールドを設定
	if setting.ExtraFields != nil {
		for key, value := range setting.ExtraFields {
			data[key] = value
		}
	}

	data_json, err := json.Marshal(data)
	if err != nil {
		errorMsg := fmt.Sprintf("JSONのマーシャルに失敗しました: %s", err.Error())
		a.OutputErrorLog(err, "JSONマーシャル")
		return errorMsg
	}

	res, err := http.Post(setting.URL, "application/json", bytes.NewBuffer(data_json))
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

	// HTTPステータスコードをチェック
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("[Web Request] HTTPエラー %d: %s - レスポンス: %s", res.StatusCode, res.Status, string(body))
		a.OutputErrorLog(fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status), "HTTPレスポンスエラー")
		return errorMsg
	}

	return "[Web Request] 送信に成功しました: " + setting.Title + ": " + eventString
}

// PNGファイルからVRChatのメタデータを読み取る
func (a *App) readPNGMetadata(filePath string) map[string]string {
	metadata := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] 画像ファイルを開けませんでした: %v", err))
		return metadata
	}
	defer file.Close()

	// PNGシグネチャを確認（8バイト）
	signature := make([]byte, 8)
	if _, err := file.Read(signature); err != nil {
		a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] PNGシグネチャの読み取りに失敗: %v", err))
		return metadata
	}

	// PNGシグネチャの検証
	expectedSignature := []byte{137, 80, 78, 71, 13, 10, 26, 10}
	if !bytes.Equal(signature, expectedSignature) {
		a.OutputConsoleLog("[PNG Metadata] 有効なPNGファイルではありません")
		return metadata
	}

	a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] PNGファイルの解析を開始: %s", filePath))

	// チャンクを読み取る
	chunkCount := 0
	for {
		// チャンク長（4バイト）
		lengthBytes := make([]byte, 4)
		if _, err := file.Read(lengthBytes); err != nil {
			break // ファイル終端
		}
		length := binary.BigEndian.Uint32(lengthBytes)

		// チャンクタイプ（4バイト）
		chunkType := make([]byte, 4)
		if _, err := file.Read(chunkType); err != nil {
			break
		}

		chunkTypeStr := string(chunkType)
		chunkCount++

		// チャンクデータ
		chunkData := make([]byte, length)
		if _, err := file.Read(chunkData); err != nil {
			break
		}

		// CRC（4バイト）をスキップ
		crc := make([]byte, 4)
		if _, err := file.Read(crc); err != nil {
			break
		}

		// tEXtチャンクの場合、メタデータを抽出
		if chunkTypeStr == "tEXt" {
			// tEXtチャンクの形式: keyword\0text
			nullIndex := bytes.IndexByte(chunkData, 0)
			if nullIndex > 0 && nullIndex < len(chunkData)-1 {
				key := string(chunkData[:nullIndex])
				value := string(chunkData[nullIndex+1:])
				metadata[key] = value
				a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] tEXt: %s = %s", key, value))
			}
		}

		// iTXtチャンク（国際化テキスト）の場合も抽出
		if chunkTypeStr == "iTXt" {
			// iTXtチャンクの形式: keyword\0compression_flag\0compression_method\0language_tag\0translated_keyword\0text
			if len(chunkData) == 0 {
				continue
			}

			// keywordを抽出
			nullIndex := bytes.IndexByte(chunkData, 0)
			if nullIndex <= 0 || nullIndex >= len(chunkData)-1 {
				a.OutputConsoleLog("[PNG Metadata] iTXt: 不正なフォーマット（keyword）")
				continue
			}

			key := string(chunkData[:nullIndex])
			pos := nullIndex + 1

			// compression flag（1バイト）
			if pos >= len(chunkData) {
				a.OutputConsoleLog("[PNG Metadata] iTXt: 不正なフォーマット（compression flag）")
				continue
			}
			compressionFlag := chunkData[pos]
			pos++

			// compression method（1バイト）
			if pos >= len(chunkData) {
				a.OutputConsoleLog("[PNG Metadata] iTXt: 不正なフォーマット（compression method）")
				continue
			}
			pos++ // compression methodをスキップ

			// language tagを探す（NULLで終端）
			nullIndex2 := bytes.IndexByte(chunkData[pos:], 0)
			if nullIndex2 < 0 {
				a.OutputConsoleLog("[PNG Metadata] iTXt: 不正なフォーマット（language tag）")
				continue
			}
			languageTag := string(chunkData[pos : pos+nullIndex2])
			pos += nullIndex2 + 1

			// translated keywordを探す（NULLで終端）
			if pos >= len(chunkData) {
				a.OutputConsoleLog("[PNG Metadata] iTXt: 不正なフォーマット（translated keyword）")
				continue
			}
			nullIndex3 := bytes.IndexByte(chunkData[pos:], 0)
			if nullIndex3 < 0 {
				a.OutputConsoleLog("[PNG Metadata] iTXt: 不正なフォーマット（translated keyword終端）")
				continue
			}
			translatedKeyword := string(chunkData[pos : pos+nullIndex3])
			pos += nullIndex3 + 1

			// 残りがテキストデータ
			if pos < len(chunkData) {
				value := string(chunkData[pos:])
				metadata[key] = value
				a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] iTXt: key=%s, lang=%s, translated=%s, value=%s, compressed=%d",
					key, languageTag, translatedKeyword, value, compressionFlag))
			}
		}

		// zTXtチャンク（圧縮テキスト）の場合も抽出
		if chunkTypeStr == "zTXt" {
			// zTXtチャンクの形式: keyword\0compression_method\0compressed_text
			nullIndex := bytes.IndexByte(chunkData, 0)
			if nullIndex > 0 && nullIndex+2 < len(chunkData) {
				key := string(chunkData[:nullIndex])
				// compression methodをスキップして圧縮データを取得
				// 簡易実装のため、zTXtは現時点ではスキップ
				a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] zTXt チャンクを検出（スキップ）: %s", key))
			}
		}

		// IENDチャンクに到達したら終了
		if chunkTypeStr == "IEND" {
			a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] IENDチャンクに到達。合計 %d チャンクを処理", chunkCount))
			break
		}
	}

	a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] 抽出されたメタデータ: %d 件", len(metadata)))
	for k, v := range metadata {
		a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata]   %s: %s", k, v))
	}

	// XMPメタデータからVRChatの情報を抽出
	if xmpData, ok := metadata["XML:com.adobe.xmp"]; ok {
		// vrc:WorldDisplayNameを抽出
		if startIdx := strings.Index(xmpData, "<vrc:WorldDisplayName>"); startIdx >= 0 {
			startIdx += len("<vrc:WorldDisplayName>")
			if endIdx := strings.Index(xmpData[startIdx:], "</vrc:WorldDisplayName>"); endIdx >= 0 {
				worldName := xmpData[startIdx : startIdx+endIdx]
				metadata["World Display Name"] = worldName
				a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] XMPから抽出: World Display Name = %s", worldName))
			}
		}
		// vrc:WorldIDを抽出
		if startIdx := strings.Index(xmpData, "<vrc:WorldID>"); startIdx >= 0 {
			startIdx += len("<vrc:WorldID>")
			if endIdx := strings.Index(xmpData[startIdx:], "</vrc:WorldID>"); endIdx >= 0 {
				worldID := xmpData[startIdx : startIdx+endIdx]
				metadata["World ID"] = worldID
				a.OutputConsoleLog(fmt.Sprintf("[PNG Metadata] XMPから抽出: World ID = %s", worldID))
			}
		}
	}

	return metadata
}

// Webリクエストで画像ファイルを送信（Base64エンコード）
func (a *App) postHttpRequestWithImage(imagePath string, setting Setting) string {
	if setting.URL == "" {
		return "URLが空です"
	}
	// url形式じゃない場合の処理
	if !strings.HasPrefix(setting.URL, "http") {
		return "URLが無効です"
	}

	// ファイルが存在するか確認
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		errorMsg := fmt.Sprintf("[Web Request] 画像ファイルが見つかりません: %s", imagePath)
		a.OutputErrorLog(err, "画像ファイル確認")
		return errorMsg
	}

	// 画像ファイルを読み込む
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		errorMsg := fmt.Sprintf("[Web Request] 画像ファイルの読み込みに失敗: %s", err.Error())
		a.OutputErrorLog(err, "画像ファイル読み込み")
		return errorMsg
	}

	// Base64エンコード
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// PNGメタデータを読み取る
	metadata := a.readPNGMetadata(imagePath)

	// 動的にJSONペイロードを構築
	data := make(map[string]interface{})

	// メッセージキーが指定されていない場合はデフォルトで"message"を使用
	messageKey := setting.MessageKey
	if messageKey == "" {
		messageKey = "message"
	}
	data[messageKey] = imagePath // ファイルパスも含める
	data["image"] = base64Image  // Base64エンコードした画像データ
	data["filename"] = filepath.Base(imagePath)
	data["mimetype"] = "image/png" // VRChatのスクリーンショットは通常PNG形式

	// VRChatメタデータを追加
	if worldID, ok := metadata["World ID"]; ok {
		data["world_id"] = worldID
	}
	if worldName, ok := metadata["World Display Name"]; ok {
		data["world_name"] = worldName
	}

	// 追加フィールドを設定
	if setting.ExtraFields != nil {
		for key, value := range setting.ExtraFields {
			data[key] = value
		}
	}

	data_json, err := json.Marshal(data)
	if err != nil {
		errorMsg := fmt.Sprintf("JSONのマーシャルに失敗しました: %s", err.Error())
		a.OutputErrorLog(err, "JSONマーシャル")
		return errorMsg
	}

	res, err := http.Post(setting.URL, "application/json", bytes.NewBuffer(data_json))
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

	// HTTPステータスコードをチェック
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("[Web Request] HTTPエラー %d: %s - レスポンス: %s", res.StatusCode, res.Status, string(body))
		a.OutputErrorLog(fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status), "HTTPレスポンスエラー")
		return errorMsg
	}

	return fmt.Sprintf("[Web Request] 画像の送信に成功しました: %s", filepath.Base(imagePath))
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
	return "[XS Overlay] 通知の送信に成功しました: " + title + ": " + eventString
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
	return "[Discord Webhook] 送信に成功しました: " + title + ": " + eventString
}

// Discord Webhookに画像ファイルを送信
func (a *App) postDiscordWebhookWithImage(imagePath string, title string, webhookURL string) string {
	// ファイルが存在するか確認
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Sprintf("[Discord Webhook] 画像ファイルが見つかりません: %s", imagePath)
	}

	// PNGメタデータを読み取る
	metadata := a.readPNGMetadata(imagePath)

	// メッセージコンテンツを構築（ワールド名がある場合のみ）
	content := ""
	if worldName, ok := metadata["World Display Name"]; ok && worldName != "" {
		content = fmt.Sprintf("**%s**", worldName)
	}

	// 画像ファイルを開く
	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] 画像ファイルの読み込みに失敗: %s", err.Error())
	}
	defer file.Close()

	// マルチパートフォームを作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// ファイルパートを追加
	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] マルチパートフォームの作成に失敗: %s", err.Error())
	}

	// ファイル内容をコピー
	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] ファイルのコピーに失敗: %s", err.Error())
	}

	// メッセージコンテンツを追加（World Display Name含む）
	_ = writer.WriteField("content", content)

	// マルチパートフォームを閉じる
	err = writer.Close()
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] マルチパートフォームのクローズに失敗: %s", err.Error())
	}

	// HTTP POSTリクエストを作成
	req, err := http.NewRequest("POST", webhookURL, body)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] HTTPリクエストの作成に失敗: %s", err.Error())
	}

	// Content-Typeを設定（マルチパート形式）
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// HTTPリクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] 送信に失敗: %s", err.Error())
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	respBody, _ := io.ReadAll(resp.Body)

	// HTTPステータスコードをチェック
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprintf("[Discord Webhook] HTTPエラー %d: %s - レスポンス: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return fmt.Sprintf("[Discord Webhook] 画像の送信に成功しました: %s", filepath.Base(imagePath))
}

func outputTextFile(eventString string, settingID string, title string) string {
	t := time.Now()
	// 基底フォルダを作成
	baseFolder := "OutputText"
	if _, err := os.Stat(baseFolder); os.IsNotExist(err) {
		os.Mkdir(baseFolder, 0777)
	}
	// 設定IDごとのフォルダを作成
	folderName := filepath.Join(baseFolder, settingID)
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		os.Mkdir(folderName, 0777)
	}
	// ファイル名は HHmmssSSS.txt
	fileName := t.Format("150405.000")
	// ファイル名から.を削除
	fileName = strings.Replace(fileName, ".", "", 1)
	// ファイルパス
	filePath := filepath.Join(folderName, fileName+".txt")
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
	return "[Log Output] 出力に成功しました " + title + ": " + filePath
}

// テキストファイル出力先のフォルダパスを取得
func (a *App) GetOutputFolderPath(settingID string) string {
	// 設定IDごとのフォルダパス
	folderName := filepath.Join("OutputText", settingID)
	// 絶対パスを取得
	absPath, err := filepath.Abs(folderName)
	if err != nil {
		return ""
	}
	// フォルダが存在するか確認
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return ""
	}
	return absPath
}

// エクスプローラーでフォルダを開く
func (a *App) OpenInExplorer(folderPath string) error {
	var cmd *exec.Cmd

	// OSに応じてコマンドを変更
	switch rt.GOOS {
	case "windows":
		cmd = exec.Command("explorer", folderPath)
	case "darwin": // macOS
		cmd = exec.Command("open", folderPath)
	case "linux":
		cmd = exec.Command("xdg-open", folderPath)
	default:
		return fmt.Errorf("サポートされていないOSです: %s", rt.GOOS)
	}

	return cmd.Start()
}

// エクスプローラーでファイルを選択した状態で開く
func (a *App) OpenFileInExplorer(filePath string) error {
	var cmd *exec.Cmd

	// ファイルが存在するか確認
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// ファイルが存在しない場合はエラーを返す
		return fmt.Errorf("ファイルが見つかりません: %s", filePath)
	}

	// OSに応じてコマンドを変更
	switch rt.GOOS {
	case "windows":
		// Windowsの場合は /select オプションでファイルを選択した状態で開く
		cmd = exec.Command("explorer", "/select,", filePath)
	case "darwin": // macOS
		// macOSの場合は open -R でファイルを選択した状態で開く
		cmd = exec.Command("open", "-R", filePath)
	case "linux":
		// Linuxの場合はファイルマネージャーによって異なるため、ファイルを直接開く
		// xdg-openでファイルのディレクトリを開く
		dir := filepath.Dir(filePath)
		cmd = exec.Command("xdg-open", dir)
	default:
		return fmt.Errorf("サポートされていないOSです: %s", rt.GOOS)
	}

	return cmd.Start()
}

// ファイルをデフォルトアプリケーションで開く
func (a *App) OpenFile(filePath string) error {
	var cmd *exec.Cmd

	// ファイルが存在するか確認
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが見つかりません: %s", filePath)
	}

	// OSに応じてコマンドを変更
	switch rt.GOOS {
	case "windows":
		// Windowsの場合は cmd /c start でファイルを開く
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin": // macOS
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		return fmt.Errorf("サポートされていないOSです: %s", rt.GOOS)
	}

	return cmd.Start()
}
