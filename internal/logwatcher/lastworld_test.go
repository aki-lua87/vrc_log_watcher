package logwatcher

import (
	"testing"

	"vrc_log_watcher/internal/models"
)

func worldSettings(nameType string) []models.Setting {
	return []models.Setting{
		{ID: worldNameSettingID, Title: "ワールド入室 (名前)", Type: nameType, RegExp: `\[Behaviour\] Entering Room: (.+)`},
		{ID: worldIDSettingID, Title: "ワールド入室 (ID)", Type: "LogOnly", RegExp: `Joining[g]*\s+(wrld_[a-f0-9-]+)`},
	}
}

func TestLastWorldTracker(t *testing.T) {
	lines := []string{
		"2026.11.08 22:09:20 Debug      -  [Behaviour] Entering Room: ワールド間対戦リバーシ",
		"2026.11.08 22:09:20 Debug      -  [Behaviour] Joining wrld_ffb19b3e-8891-426a-a100-54e578509743:18506~private(usr_x)~region(jp)",
		"2026.11.08 22:09:20 Debug      -  [Behaviour] Joining or Creating Room: ワールド間対戦リバーシ",
		"2026.11.08 22:20:00 Debug      -  [Behaviour] OnPlayerJoinComplete aki_lua87",
		"2026.11.08 22:29:30 Debug      -  [Behaviour] Entering Room: トークン撮影ワールド",
		"2026.11.08 22:29:30 Debug      -  [Behaviour] Joining wrld_c9f8358e-659b-40d2-af26-f9344ed10f98:31727~private(usr_x)~region(jp)",
		"2026.11.08 22:29:40 Debug      -  [Behaviour] OnPlayerJoinComplete aki_lua87",
	}

	tracker := newLastWorldTracker(worldSettings("LogOnly"))
	for _, line := range lines {
		tracker.feed(line, line[:19])
	}
	got := tracker.results()

	if len(got) != 2 {
		t.Fatalf("結果は名前とIDの2件のはず: %d 件", len(got))
	}
	if got[0].Setting.ID != worldNameSettingID || got[0].Text != "トークン撮影ワールド" || got[0].LogTime != "2026.11.08 22:29:30" {
		t.Errorf("名前が最後の入室になっていない: %+v", got[0])
	}
	if got[1].Setting.ID != worldIDSettingID || got[1].Text != "wrld_c9f8358e-659b-40d2-af26-f9344ed10f98" {
		t.Errorf("IDが最後の入室になっていない: %+v", got[1])
	}
	for _, r := range got {
		if !r.Initial {
			t.Errorf("Initial が立っていない: %+v", r)
		}
	}
}

func TestLastWorldTrackerDropsOldIDOnNewEntry(t *testing.T) {
	tracker := newLastWorldTracker(worldSettings("LogOnly"))
	tracker.feed("2026.11.08 22:09:20 Debug      -  [Behaviour] Entering Room: A", "")
	tracker.feed("2026.11.08 22:09:20 Debug      -  [Behaviour] Joining wrld_aaaa", "")
	// 入室の途中で終わったログ（IDの行がまだない）
	tracker.feed("2026.11.08 22:29:30 Debug      -  [Behaviour] Entering Room: B", "")

	got := tracker.results()
	if len(got) != 1 || got[0].Text != "B" {
		t.Fatalf("前のワールドのIDを引き継いではいけない: %+v", got)
	}
}

func TestLastWorldTrackerSkipsDisabled(t *testing.T) {
	tracker := newLastWorldTracker(worldSettings("Disable"))
	tracker.feed("2026.11.08 22:29:30 Debug      -  [Behaviour] Entering Room: B", "")
	tracker.feed("2026.11.08 22:29:30 Debug      -  [Behaviour] Joining wrld_bbbb", "")

	got := tracker.results()
	if len(got) != 1 || got[0].Setting.ID != worldIDSettingID {
		t.Fatalf("無効にしたルールは拾わない: %+v", got)
	}
}
