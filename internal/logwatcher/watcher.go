package logwatcher

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/models"
)

// Watcher VRChatログファイルの監視・解析を管理
type Watcher struct {
	TargetFileName string
	LastLogTime    string
	lastOffset     int64
	isRunning      bool
	Logger         applog.Logger
}

// NewWatcher 新しいWatcherを作成
func NewWatcher(logger applog.Logger) *Watcher {
	return &Watcher{Logger: logger}
}

// ResetOffset 読み取りオフセットをリセット
func (w *Watcher) ResetOffset() {
	w.lastOffset = 0
}

// IsFirstRead 初回読み込みかどうか
func (w *Watcher) IsFirstRead() bool {
	return w.lastOffset == 0
}

// MatchResult 行の評価結果
type MatchResult struct {
	Setting      models.Setting
	Text         string
	IsScreenshot bool
}

// GetNewestFileName フォルダ内の最新のtxtファイルを探索
func (w *Watcher) GetNewestFileName(logPath string) (string, bool) {
	if logPath == "" {
		w.Logger.Log("ログフォルダが指定されていません。「フォルダを指定」ボタンをクリックしてVRChatのログフォルダを選択してください。")
		return "", false
	}

	entries, err := os.ReadDir(logPath)
	if err != nil {
		errorMsg := fmt.Sprintf("ログフォルダの読み取りに失敗しました: %s", err.Error())
		w.Logger.LogError(err, "ログフォルダの読み取り")
		w.Logger.Log(errorMsg)
		return "", false
	}

	var newestFile os.DirEntry
	var newestTime time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".txt" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			w.Logger.LogError(err, "ファイル情報の取得")
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

	if newestFile == nil {
		w.Logger.Log("ログフォルダ内にログファイル(.txt)が見つかりません。VRChatのログフォルダを正しく指定しているか確認してください。")
		return "", false
	}

	if newestFile.Name() == w.TargetFileName {
		return w.TargetFileName, false
	}

	log.Default().Println("[DEBUG] [LOG] 監視対象を変更します")
	w.TargetFileName = newestFile.Name()
	w.ResetOffset()
	return newestFile.Name(), true
}

// ReadFile ログファイルを読み取り、各行を評価して結果を返す
func (w *Watcher) ReadFile(logPath string, settings []models.Setting) []MatchResult {
	if w.TargetFileName == "" {
		log.Default().Println("[DEBUG] [LOG] No FileName")
		return nil
	}
	if logPath == "" {
		log.Default().Println("[DEBUG] [LOG] No LogPath")
		return nil
	}
	if w.isRunning {
		log.Default().Println("[DEBUG] [LOG] Watching now")
		return nil
	}

	w.isRunning = true
	defer func() { w.isRunning = false }()

	path := logPath + "\\" + w.TargetFileName
	file, err := os.Open(path)
	if err != nil {
		w.Logger.LogError(err, "ログファイルのオープン")
		return nil
	}
	defer file.Close()

	if _, err = file.Seek(w.lastOffset, 0); err != nil {
		w.Logger.LogError(err, "ファイルシーク")
		return nil
	}

	scanner := bufio.NewScanner(file)
	const maxCapacity = 1000000
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var results []MatchResult
	for scanner.Scan() {
		line := scanner.Text()
		w.extractTimeFromVRCLog(line)

		if w.lastOffset == 0 {
			continue
		}

		matches := w.evaluateLine(line, settings)
		results = append(results, matches...)
	}

	if err := scanner.Err(); err != nil {
		w.Logger.LogError(err, "ファイルスキャン")
		return results
	}

	w.lastOffset, err = file.Seek(0, io.SeekCurrent)
	if err != nil {
		w.Logger.LogError(err, "ファイルオフセット取得")
	}

	return results
}

// extractTimeFromVRCLog VRCログから時刻を抽出
func (w *Watcher) extractTimeFromVRCLog(line string) {
	timePattern := regexp.MustCompile(`^(\d{4}\.\d{2}\.\d{2} \d{2}:\d{2}:\d{2})`)
	matches := timePattern.FindStringSubmatch(line)
	if len(matches) > 1 {
		w.LastLogTime = matches[1]
	}
}

// evaluateLine 1行を全設定パターンと照合し、マッチ結果を返す
func (w *Watcher) evaluateLine(line string, settings []models.Setting) []MatchResult {
	var results []MatchResult

	for _, setting := range settings {
		if setting.RegExp == "" {
			continue
		}

		pattern := regexp.MustCompile(setting.RegExp)
		matches := pattern.FindStringSubmatch(line)
		text := ""
		if len(matches) > 1 {
			text = strings.Join(matches[1:], "")
		}

		if text == "" {
			continue
		}

		if setting.Exclude == text {
			continue
		}

		w.Logger.Log(setting.Title + " : " + text)

		results = append(results, MatchResult{
			Setting:      setting,
			Text:         text,
			IsScreenshot: setting.ID == "core-screenshot",
		})
	}

	return results
}
