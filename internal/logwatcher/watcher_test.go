package logwatcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"vrc_log_watcher/internal/models"
)

type nopLogger struct{}

func (nopLogger) Log(string)             {}
func (nopLogger) LogInfo(string)         {}
func (nopLogger) LogError(error, string) {}

var worldSettings = []models.Setting{
	{ID: "core-world-join-name", Title: "ワールド入室 (名前)", Type: "LogOnly", RegExp: `\[Behaviour\] Entering Room: (.+)`},
}

func writeLog(t *testing.T, dir, name, content string, modTime time.Time) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
}

func texts(results []MatchResult) []string {
	var out []string
	for _, r := range results {
		out = append(out, r.Text)
	}
	return out
}

// 起動時は既存のログを読み飛ばし、監視中に切り替わった新しいファイルは先頭から読む
func TestReadFileSkipsOnStartupButReadsSwitchedFileFromStart(t *testing.T) {
	dir := t.TempDir()
	base := time.Now().Add(-time.Hour)
	writeLog(t, dir, "output_log_1.txt", "2026.11.08 21:00:00 Debug      -  [Behaviour] Entering Room: 前回のワールド\n", base)

	w := NewWatcher(nopLogger{})

	// 起動時: 既存のワールド入室は出さない
	if name, isNew := w.GetNewestFileName(dir); name != "output_log_1.txt" || !isNew {
		t.Fatalf("最初のファイルを選んでいない: %s %v", name, isNew)
	}
	if got := w.ReadFile(dir, worldSettings); len(got) != 0 {
		t.Fatalf("起動時に既存ログを読んではいけない: %v", texts(got))
	}

	// 監視中に追記された行は読む
	f, err := os.OpenFile(filepath.Join(dir, "output_log_1.txt"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("2026.11.08 21:10:00 Debug      -  [Behaviour] Entering Room: 追記されたワールド\n")
	f.Close()
	if got := texts(w.ReadFile(dir, worldSettings)); len(got) != 1 || got[0] != "追記されたワールド" {
		t.Fatalf("追記された行を読んでいない: %v", got)
	}

	// VRChat を起動し直して新しいファイルができた: 気づくまでの間に出た行も読む
	writeLog(t, dir, "output_log_2.txt", "2026.11.08 22:00:00 Debug      -  [Behaviour] Entering Room: 新しいワールド\n", time.Now().Add(time.Minute))
	if name, isNew := w.GetNewestFileName(dir); name != "output_log_2.txt" || !isNew {
		t.Fatalf("新しいファイルに切り替わっていない: %s %v", name, isNew)
	}
	if got := texts(w.ReadFile(dir, worldSettings)); len(got) != 1 || got[0] != "新しいワールド" {
		t.Fatalf("切り替え後のファイルを先頭から読んでいない: %v", got)
	}
	// 2回目は同じ行を読まない
	if got := w.ReadFile(dir, worldSettings); len(got) != 0 {
		t.Fatalf("同じ行を2回読んでいる: %v", texts(got))
	}
}

// ログフォルダを変えたときは、起動時と同じく既存ログを読み飛ばす
func TestFolderChangeSkipsExistingLines(t *testing.T) {
	oldDir, newDir := t.TempDir(), t.TempDir()
	base := time.Now().Add(-time.Hour)
	writeLog(t, oldDir, "output_log_1.txt", "", base)
	writeLog(t, newDir, "output_log_2.txt", "2026.11.08 22:00:00 Debug      -  [Behaviour] Entering Room: 古いワールド\n", base)

	w := NewWatcher(nopLogger{})
	w.GetNewestFileName(oldDir)
	w.ReadFile(oldDir, worldSettings)
	if name, isNew := w.GetNewestFileName(newDir); name != "output_log_2.txt" || !isNew {
		t.Fatalf("新しいフォルダのファイルに切り替わっていない: %s %v", name, isNew)
	}
	if got := w.ReadFile(newDir, worldSettings); len(got) != 0 {
		t.Fatalf("フォルダを変えたときに既存ログを読んではいけない: %v", texts(got))
	}
}

// ResetOffset の後は起動時と同じく読み飛ばす
func TestResetOffsetSkipsExistingLines(t *testing.T) {
	dir := t.TempDir()
	base := time.Now().Add(-time.Hour)
	writeLog(t, dir, "output_log_1.txt", "", base)

	w := NewWatcher(nopLogger{})
	w.GetNewestFileName(dir)
	writeLog(t, dir, "output_log_2.txt", "2026.11.08 22:00:00 Debug      -  [Behaviour] Entering Room: A\n", base.Add(time.Minute))
	w.GetNewestFileName(dir)
	w.ResetOffset()

	if got := w.ReadFile(dir, worldSettings); len(got) != 0 {
		t.Fatalf("ResetOffset の後は既存ログを読み飛ばす: %v", texts(got))
	}
}
