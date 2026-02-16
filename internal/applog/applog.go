package applog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger ロギング用インターフェース（内部パッケージで使用）
type Logger interface {
	Log(msg string)
	LogError(err error, context string)
}

// FileLogger アプリケーションログファイルへの書き込みを管理
type FileLogger struct {
	file *os.File
}

// NewFileLogger ログファイルを初期化して返す
func NewFileLogger() *FileLogger {
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

	currentTime := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logDir, "app_log_"+currentTime+".txt")

	file, err := os.Create(logFilePath)
	if err != nil {
		log.Printf("[ERROR] ログファイルの作成に失敗しました: %v", err)
		return &FileLogger{}
	}

	startupMsg := fmt.Sprintf("[%s] アプリケーションを起動しました\n", time.Now().Format("2006/01/02 15:04:05"))
	file.WriteString(startupMsg)

	return &FileLogger{file: file}
}

// WriteLog ログメッセージをファイルに書き込む
func (l *FileLogger) WriteLog(msg string) {
	if l.file != nil {
		l.file.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006/01/02 15:04:05"), msg))
	}
}

// Close ログファイルをクローズ
func (l *FileLogger) Close() {
	if l.file != nil {
		l.file.WriteString(fmt.Sprintf("[%s] アプリケーションを終了しました\n", time.Now().Format("2006/01/02 15:04:05")))
		l.file.Close()
	}
}
