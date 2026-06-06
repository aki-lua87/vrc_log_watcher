package fileutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// OpenInExplorer エクスプローラーでフォルダを開く
func OpenInExplorer(folderPath string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", folderPath)
	case "darwin":
		cmd = exec.Command("open", folderPath)
	case "linux":
		cmd = exec.Command("xdg-open", folderPath)
	default:
		return fmt.Errorf("サポートされていないOSです: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// OpenFileInExplorer エクスプローラーでファイルを選択した状態で開く
func OpenFileInExplorer(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが見つかりません: %s", filePath)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", filePath)
	case "darwin":
		cmd = exec.Command("open", "-R", filePath)
	case "linux":
		dir := filepath.Dir(filePath)
		cmd = exec.Command("xdg-open", dir)
	default:
		return fmt.Errorf("サポートされていないOSです: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// OpenFile ファイルをデフォルトアプリケーションで開く
func OpenFile(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("ファイルが見つかりません: %s", filePath)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		return fmt.Errorf("サポートされていないOSです: %s", runtime.GOOS)
	}

	return cmd.Start()
}
