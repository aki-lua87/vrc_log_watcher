package notify

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

const baseFolder = "OutputText"

// OutputTextFile イベント文字列をテキストファイルに出力
func OutputTextFile(eventString string, settingID string, title string) string {
	t := time.Now()

	if _, err := os.Stat(baseFolder); os.IsNotExist(err) {
		os.Mkdir(baseFolder, 0777)
	}

	folderName := filepath.Join(baseFolder, settingID)
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		os.Mkdir(folderName, 0777)
	}

	fileName := t.Format("150405.000")
	fileName = strings.Replace(fileName, ".", "", 1)
	filePath := filepath.Join(folderName, fileName+".txt")

	file, err := os.Create(filePath)
	if err != nil {
		return err.Error()
	}
	defer file.Close()

	if _, err = file.WriteString(eventString); err != nil {
		return err.Error()
	}

	return "[Log Output] 出力に成功しました " + title + ": " + filePath
}

// GetOutputFolderPath テキストファイル出力先のフォルダパスを取得
func GetOutputFolderPath(settingID string) string {
	folderName := filepath.Join(baseFolder, settingID)
	absPath, err := filepath.Abs(folderName)
	if err != nil {
		return ""
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return ""
	}

	return absPath
}
