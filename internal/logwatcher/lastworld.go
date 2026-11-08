package logwatcher

import (
	"regexp"
	"strings"

	"vrc_log_watcher/internal/models"
)

const (
	worldNameSettingID = "core-world-join-name"
	worldIDSettingID   = "core-world-join-id"
)

// lastWorldTracker 既存ログを読み飛ばす間に、最後に入室したワールドの名前とIDを覚えておく
type lastWorldTracker struct {
	nameSetting, idSetting *models.Setting
	namePattern, idPattern *regexp.Regexp
	name, id               *MatchResult
}

func newLastWorldTracker(settings []models.Setting) *lastWorldTracker {
	t := &lastWorldTracker{}
	for i := range settings {
		s := &settings[i]
		if s.Type == "Disable" || s.RegExp == "" {
			continue
		}
		pattern, err := regexp.Compile(s.RegExp)
		if err != nil {
			continue
		}
		switch s.ID {
		case worldNameSettingID:
			t.nameSetting, t.namePattern = s, pattern
		case worldIDSettingID:
			t.idSetting, t.idPattern = s, pattern
		}
	}
	return t
}

func (t *lastWorldTracker) feed(line string, logTime string) {
	// ワールド入室の行はどちらも [Behaviour] なので、それ以外は正規表現にかけずに飛ばす
	if !strings.Contains(line, "[Behaviour]") {
		return
	}
	if m := t.match(t.namePattern, t.nameSetting, line, logTime); m != nil {
		// VRChat は「Entering Room: 名前」→「Joining wrld_...」の順に出すので、新しい入室で前のIDは捨てる
		t.name, t.id = m, nil
		return
	}
	if m := t.match(t.idPattern, t.idSetting, line, logTime); m != nil {
		t.id = m
	}
}

func (t *lastWorldTracker) match(pattern *regexp.Regexp, setting *models.Setting, line string, logTime string) *MatchResult {
	if pattern == nil {
		return nil
	}
	matches := pattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}
	text := strings.Join(matches[1:], "")
	if text == "" || text == setting.Exclude {
		return nil
	}
	return &MatchResult{Setting: *setting, Text: text, LogTime: logTime, Initial: true}
}

// results ログに出た順（名前 → ID）で返す
func (t *lastWorldTracker) results() []MatchResult {
	var results []MatchResult
	if t.name != nil {
		results = append(results, *t.name)
	}
	if t.id != nil {
		results = append(results, *t.id)
	}
	return results
}
