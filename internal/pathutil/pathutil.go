package pathutil

import (
	"os"
	"path/filepath"
)

// ExeDir は実行ファイルが置かれているディレクトリの絶対パスを返す。
// 取得に失敗した場合はカレントディレクトリ "." にフォールバックする。
func ExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}
