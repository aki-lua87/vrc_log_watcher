package notify

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/metadata"
	"vrc_log_watcher/internal/models"
)

// HTTPClient HTTP通知の送信を管理
type HTTPClient struct {
	Logger applog.Logger
}

// Post テキストデータをHTTPリクエストで送信
func (c *HTTPClient) Post(eventString string, setting models.Setting) string {
	if setting.URL == "" {
		return "URLが空です"
	}
	if !strings.HasPrefix(setting.URL, "http") {
		return "URLが無効です"
	}

	data := make(map[string]interface{})

	messageKey := setting.MessageKey
	if messageKey == "" {
		messageKey = "message"
	}
	data[messageKey] = eventString

	if setting.ExtraFields != nil {
		for key, value := range setting.ExtraFields {
			data[key] = value
		}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		errorMsg := fmt.Sprintf("JSONのマーシャルに失敗しました: %s", err.Error())
		c.Logger.LogError(err, "JSONマーシャル")
		return errorMsg
	}

	res, err := http.Post(setting.URL, "application/json", bytes.NewBuffer(dataJSON))
	if err != nil {
		c.Logger.LogError(err, "HTTPリクエスト")
		return err.Error()
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.Logger.LogError(err, "レスポンス読み取り")
		return err.Error()
	}

	log.Default().Println(string(body))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("[Web Request] HTTPエラー %d: %s - レスポンス: %s", res.StatusCode, res.Status, string(body))
		c.Logger.LogError(fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status), "HTTPレスポンスエラー")
		return errorMsg
	}

	return "[Web Request] 送信に成功しました: " + setting.Title + ": " + eventString
}

// PostWithImage 画像ファイルをBase64エンコードしてHTTPリクエストで送信
func (c *HTTPClient) PostWithImage(imagePath string, setting models.Setting) string {
	if setting.URL == "" {
		return "URLが空です"
	}
	if !strings.HasPrefix(setting.URL, "http") {
		return "URLが無効です"
	}

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		errorMsg := fmt.Sprintf("[Web Request] 画像ファイルが見つかりません: %s", imagePath)
		c.Logger.LogError(err, "画像ファイル確認")
		return errorMsg
	}

	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		errorMsg := fmt.Sprintf("[Web Request] 画像ファイルの読み込みに失敗: %s", err.Error())
		c.Logger.LogError(err, "画像ファイル読み込み")
		return errorMsg
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	pngMetadata := metadata.ReadPNG(imagePath, c.Logger)

	data := make(map[string]interface{})

	messageKey := setting.MessageKey
	if messageKey == "" {
		messageKey = "message"
	}
	data[messageKey] = imagePath
	data["image"] = base64Image
	data["filename"] = filepath.Base(imagePath)
	data["mimetype"] = "image/png"

	if worldID, ok := pngMetadata["World ID"]; ok {
		data["world_id"] = worldID
	}
	if worldName, ok := pngMetadata["World Display Name"]; ok {
		data["world_name"] = worldName
	}

	if setting.ExtraFields != nil {
		for key, value := range setting.ExtraFields {
			data[key] = value
		}
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		errorMsg := fmt.Sprintf("JSONのマーシャルに失敗しました: %s", err.Error())
		c.Logger.LogError(err, "JSONマーシャル")
		return errorMsg
	}

	res, err := http.Post(setting.URL, "application/json", bytes.NewBuffer(dataJSON))
	if err != nil {
		c.Logger.LogError(err, "HTTPリクエスト")
		return err.Error()
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		c.Logger.LogError(err, "レスポンス読み取り")
		return err.Error()
	}

	log.Default().Println(string(body))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		errorMsg := fmt.Sprintf("[Web Request] HTTPエラー %d: %s - レスポンス: %s", res.StatusCode, res.Status, string(body))
		c.Logger.LogError(fmt.Errorf("HTTP %d: %s", res.StatusCode, res.Status), "HTTPレスポンスエラー")
		return errorMsg
	}

	return fmt.Sprintf("[Web Request] 画像の送信に成功しました: %s", filepath.Base(imagePath))
}
