package logwatcher

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/models"
)

// DebugInfo デバッグ用の監視状態情報
type DebugInfo struct {
	LastReadLine          string `json:"lastReadLine"`
	LastOffset            int64  `json:"lastOffset"`
	FileSize              int64  `json:"fileSize"`
	LinesRead             int    `json:"linesRead"`
	LastReadAt            string `json:"lastReadAt"`
	LastNewLinesAt        string `json:"lastNewLinesAt"`
	IsRunning             bool   `json:"isRunning"`
	SkipCount             int    `json:"skipCount"`
	ConsecutiveNoProgress int    `json:"consecutiveNoProgress"`
	RefreshCount          int    `json:"refreshCount"`
}

// Watcher VRChatログファイルの監視・解析を管理
type Watcher struct {
	TargetFileName string
	LastLogTime    string
	lastOffset     int64
	isRunning      bool
	mu             sync.Mutex
	Logger         applog.Logger
	Debug          DebugInfo

	lastNewLinesAt        time.Time
	consecutiveNoProgress int
	refreshCount          int
}

// NewWatcher 新しいWatcherを作成
func NewWatcher(logger applog.Logger) *Watcher {
	return &Watcher{Logger: logger}
}

// ResetOffset 読み取りオフセットをリセット
func (w *Watcher) ResetOffset() {
	w.lastOffset = 0
	w.consecutiveNoProgress = 0
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

	w.mu.Lock()
	defer w.mu.Unlock()

	if newestFile.Name() == w.TargetFileName {
		return w.TargetFileName, false
	}

	log.Default().Println("[DEBUG] [LOG] 監視対象を変更します")
	w.TargetFileName = newestFile.Name()
	w.lastOffset = 0
	w.consecutiveNoProgress = 0
	w.lastNewLinesAt = time.Time{}
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

	w.mu.Lock()
	if w.isRunning {
		w.Debug.SkipCount++
		w.mu.Unlock()
		return nil
	}
	w.isRunning = true
	currentOffset := w.lastOffset
	targetFile := w.TargetFileName
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.isRunning = false
		w.mu.Unlock()
	}()

	path := filepath.Join(logPath, targetFile)
	file, err := os.Open(path)
	if err != nil {
		w.Logger.LogError(err, "ログファイルのオープン")
		return nil
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		w.Logger.LogError(err, "ファイル情報取得")
		return nil
	}
	fileSize := stat.Size()

	if currentOffset > fileSize {
		w.Logger.Log(fmt.Sprintf("オフセット(%d)がファイルサイズ(%d)を超えています。先頭から再読み込みします。", currentOffset, fileSize))
		currentOffset = 0
	}

	if fileSize <= currentOffset {
		w.updateDebugInfo("", 0, fileSize, currentOffset)
		return nil
	}

	if _, err = file.Seek(currentOffset, 0); err != nil {
		w.Logger.LogError(err, "ファイルシーク")
		return nil
	}

	isFirstRead := currentOffset == 0

	// 初回読み込み: 既存ログを読み飛ばしてオフセットをファイル末尾に設定
	if isFirstRead {
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1000000), 1000000)
		var lastLine string
		linesRead := 0
		for scanner.Scan() {
			line := scanner.Text()
			lastLine = line
			linesRead++
			w.extractTimeFromVRCLog(line)
		}
		if err := scanner.Err(); err != nil {
			w.Logger.LogError(err, "初期読み込みスキャン")
		}
		w.mu.Lock()
		w.lastOffset = fileSize
		w.lastNewLinesAt = time.Now()
		w.consecutiveNoProgress = 0
		w.mu.Unlock()
		w.updateDebugInfo(lastLine, linesRead, fileSize, fileSize)
		return nil
	}

	// 増分読み取り: lastOffset 以降の新しいデータを直接読み取る
	newData, err := io.ReadAll(file)
	if err != nil {
		w.Logger.LogError(err, "ファイル読み取り")
		return nil
	}

	if len(newData) == 0 {
		w.handleNoProgress(logPath, fileSize, currentOffset)
		w.updateDebugInfo("", 0, fileSize, currentOffset)
		return nil
	}

	// 完全な行（\n で終わる）のみ処理し、書き込み途中の不完全行は次回に持ち越す
	lastNewline := bytes.LastIndexByte(newData, '\n')
	if lastNewline == -1 {
		w.handleNoProgress(logPath, fileSize, currentOffset)
		w.updateDebugInfo("", 0, fileSize, currentOffset)
		return nil
	}

	completeData := string(newData[:lastNewline+1])
	newOffset := currentOffset + int64(lastNewline+1)

	lines := strings.Split(completeData, "\n")
	var results []MatchResult
	linesRead := 0
	var lastLine string

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		lastLine = line
		linesRead++
		w.extractTimeFromVRCLog(line)
		matches := w.evaluateLine(line, settings)
		results = append(results, matches...)
	}

	w.mu.Lock()
	if w.TargetFileName == targetFile {
		w.lastOffset = newOffset
	}
	if linesRead > 0 {
		w.lastNewLinesAt = time.Now()
		w.consecutiveNoProgress = 0
	}
	w.mu.Unlock()

	w.updateDebugInfo(lastLine, linesRead, fileSize, newOffset)
	return results
}

// handleNoProgress ファイルが成長しているのに行が読めない場合の停滞検知・リカバリ
func (w *Watcher) handleNoProgress(logPath string, fileSize int64, currentOffset int64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if fileSize <= currentOffset {
		w.consecutiveNoProgress = 0
		return
	}

	w.consecutiveNoProgress++

	// 120回連続（約2分）進捗なし & ファイルが1KB以上成長している場合リフレッシュ
	if w.consecutiveNoProgress >= 120 && fileSize > currentOffset+1024 {
		w.refreshCount++
		gap := fileSize - currentOffset
		w.Logger.Log(fmt.Sprintf(
			"ログの読み取りが停滞しています。オフセットをリフレッシュします。(未読: %dバイト, 回数: %d)",
			gap, w.refreshCount,
		))
		w.performRefresh(logPath, fileSize, currentOffset)
	}
}

// performRefresh 停滞時のオフセットリカバリ（mu ロック済み前提）
func (w *Watcher) performRefresh(logPath string, fileSize int64, currentOffset int64) {
	log.Default().Printf("[DEBUG] [LOG] Refresh: offset=%d, fileSize=%d, gap=%d", currentOffset, fileSize, fileSize-currentOffset)

	if currentOffset > fileSize {
		w.lastOffset = fileSize
		w.consecutiveNoProgress = 0
		return
	}

	if w.TargetFileName == "" || logPath == "" {
		return
	}

	// ファイルを開いて現在のオフセット付近から次の行境界を探す
	path := filepath.Join(logPath, w.TargetFileName)
	file, err := os.Open(path)
	if err != nil {
		w.lastOffset = currentOffset
		w.consecutiveNoProgress = 0
		return
	}
	defer file.Close()

	if _, err := file.Seek(currentOffset, 0); err != nil {
		w.lastOffset = currentOffset
		w.consecutiveNoProgress = 0
		return
	}

	buf := make([]byte, 4096)
	n, _ := file.Read(buf)
	if n > 0 {
		idx := bytes.IndexByte(buf[:n], '\n')
		if idx >= 0 {
			w.lastOffset = currentOffset + int64(idx) + 1
			w.Logger.Log(fmt.Sprintf("リフレッシュ: 次の行境界を発見しました。(新オフセット: %d)", w.lastOffset))
		} else {
			w.lastOffset = fileSize
			w.Logger.Log("リフレッシュ: 行境界が見つからないためファイル末尾にスキップします。")
		}
	} else {
		w.lastOffset = fileSize
	}
	w.consecutiveNoProgress = 0
}

func (w *Watcher) updateDebugInfo(lastLine string, linesRead int, fileSize int64, offset int64) {
	w.Debug.FileSize = fileSize
	w.Debug.LastOffset = offset
	w.Debug.IsRunning = false
	if linesRead > 0 {
		w.Debug.LastReadLine = lastLine
		w.Debug.LinesRead = linesRead
		w.Debug.LastReadAt = time.Now().Format("15:04:05")
	}
	w.mu.Lock()
	w.Debug.LastNewLinesAt = w.lastNewLinesAt.Format("15:04:05")
	w.Debug.ConsecutiveNoProgress = w.consecutiveNoProgress
	w.Debug.RefreshCount = w.refreshCount
	w.mu.Unlock()
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

		pattern, err := regexp.Compile(setting.RegExp)
		if err != nil {
			w.Logger.LogError(err, "正規表現コンパイルエラー: "+setting.Title)
			continue
		}
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
