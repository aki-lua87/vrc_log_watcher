package notify

import (
	"strings"

	"vrc_log_watcher/internal/models"
)

// applyTemplate テンプレート文字列に変数を展開する ({key} -> value)
func applyTemplate(s string, vars map[string]string) string {
	for key, value := range vars {
		s = strings.ReplaceAll(s, "{"+key+"}", value)
	}
	return s
}

// applyTemplateToSetting URL と ExtraFields の値にテンプレートを展開した Setting を返す
func applyTemplateToSetting(setting models.Setting, vars map[string]string) models.Setting {
	setting.URL = applyTemplate(setting.URL, vars)
	if setting.ExtraFields != nil {
		newFields := make(map[string]string, len(setting.ExtraFields))
		for k, v := range setting.ExtraFields {
			newFields[k] = applyTemplate(v, vars)
		}
		setting.ExtraFields = newFields
	}
	return setting
}
