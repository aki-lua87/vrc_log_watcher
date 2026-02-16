package settings

import (
	"encoding/json"
	"log"
	"os"

	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/models"
)

const settingFile = "setting.json"

// InitCoreSettings VRCイベントの基本機能プリセットを返す
func InitCoreSettings() []models.Setting {
	return []models.Setting{
		{
			ID:      "core-world-join-name",
			Title:   "ワールド入室 (名前)",
			Details: "ワールド入室時のワールド名を取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[Behaviour\] Entering Room: (.+)`,
			Exclude: "",
		},
		{
			ID:      "core-world-join-id",
			Title:   "ワールド入室 (ID)",
			Details: "ワールド入室時のワールドIDを取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `Joining[g]*\s+(wrld_[a-f0-9-]+)`,
			Exclude: "",
		},
		{
			ID:      "core-user-join",
			Title:   "ユーザー入室",
			Details: "ユーザー入室時のユーザー名を取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[Behaviour\] OnPlayerJoinComplete (.+)`,
			Exclude: "",
		},
		{
			ID:      "core-user-left",
			Title:   "ユーザー退室",
			Details: "ユーザー退室時のユーザー名を取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[Behaviour\] OnPlayerLeft ([^\s]+)`,
			Exclude: "",
		},
		{
			ID:      "core-screenshot",
			Title:   "スクリーンショット",
			Details: "写真撮影時のファイルパスを取得",
			IsCore:  true,
			Type:    "Disable",
			RegExp:  `\[VRC Camera\] Took screenshot to: (.+)`,
			Exclude: "",
		},
	}
}

// Load 設定ファイルを読み込む
func Load(logger applog.Logger) models.SaveData {
	log.Default().Println("[DEBUG] [LOG] Load Setting")
	logger.Log("setting.json 読み込み")

	file, err := os.ReadFile(settingFile)
	if err != nil {
		logger.LogError(err, "setting.json 読み込みエラー")
		coreSettings := InitCoreSettings()
		saveData := models.SaveData{Settings: coreSettings}
		Save(saveData, logger)
		return saveData
	}

	var saveData models.SaveData
	if err := json.Unmarshal(file, &saveData); err != nil {
		logger.LogError(err, "setting.json パースエラー")
	}

	if len(saveData.Settings) == 0 {
		coreSettings := InitCoreSettings()
		saveData.Settings = coreSettings
		Save(saveData, logger)
	} else {
		saveData = mergeCoreSettings(saveData, logger)
	}

	log.Default().Println(saveData)
	logger.Log("setting.json 読み込みに成功しました")
	return saveData
}

// Save 設定をファイルに保存する
func Save(saveData models.SaveData, logger applog.Logger) {
	log.Default().Println("[DEBUG] [LOG] UpdateSetting:", len(saveData.Settings))

	jsonData, err := json.Marshal(saveData)
	if err != nil {
		logger.LogError(err, "setting.json 変換エラー")
		return
	}

	if err := os.WriteFile(settingFile, jsonData, 0644); err != nil {
		logger.LogError(err, "setting.json 書き込みエラー")
	}
}

// mergeCoreSettings 既存設定に基本機能の定義を反映させる
func mergeCoreSettings(saveData models.SaveData, logger applog.Logger) models.SaveData {
	coreSettings := InitCoreSettings()
	coreSettingsMap := make(map[string]models.Setting)
	for _, cs := range coreSettings {
		coreSettingsMap[cs.ID] = cs
	}

	needsUpdate := false
	for i, setting := range saveData.Settings {
		coreSetting, exists := coreSettingsMap[setting.ID]
		if !exists {
			continue
		}

		simpleBlocksChanged := false
		if len(setting.SimpleBlocks) != len(coreSetting.SimpleBlocks) {
			simpleBlocksChanged = true
		} else {
			for j := range setting.SimpleBlocks {
				if setting.SimpleBlocks[j] != coreSetting.SimpleBlocks[j] {
					simpleBlocksChanged = true
					break
				}
			}
		}

		if setting.Title != coreSetting.Title ||
			setting.Details != coreSetting.Details ||
			setting.RegExp != coreSetting.RegExp ||
			setting.Exclude != coreSetting.Exclude ||
			setting.SimpleMode != coreSetting.SimpleMode ||
			simpleBlocksChanged {
			saveData.Settings[i].Title = coreSetting.Title
			saveData.Settings[i].Details = coreSetting.Details
			saveData.Settings[i].RegExp = coreSetting.RegExp
			saveData.Settings[i].Exclude = coreSetting.Exclude
			saveData.Settings[i].SimpleMode = coreSetting.SimpleMode
			saveData.Settings[i].SimpleBlocks = coreSetting.SimpleBlocks
			needsUpdate = true
		}

		delete(coreSettingsMap, setting.ID)
	}

	// 存在しない基本機能を追加
	for _, coreSetting := range coreSettings {
		if _, exists := coreSettingsMap[coreSetting.ID]; exists {
			saveData.Settings = append([]models.Setting{coreSetting}, saveData.Settings...)
			needsUpdate = true
		}
	}

	if needsUpdate {
		Save(saveData, logger)
	}

	return saveData
}
