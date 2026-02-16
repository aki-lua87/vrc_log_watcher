package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/metadata"
)

// DiscordClient Discord Webhook通知を管理
type DiscordClient struct {
	Logger applog.Logger
}

// Post テキストメッセージをDiscord Webhookに送信
func (c *DiscordClient) Post(eventString string, title string, webhookURL string) string {
	message := map[string]string{
		"content": title + ": " + eventString,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return err.Error()
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err.Error()
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()

	return "[Discord Webhook] 送信に成功しました: " + title + ": " + eventString
}

// PostWithImage 画像ファイルをDiscord Webhookに送信
func (c *DiscordClient) PostWithImage(imagePath string, title string, webhookURL string) string {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Sprintf("[Discord Webhook] 画像ファイルが見つかりません: %s", imagePath)
	}

	pngMetadata := metadata.ReadPNG(imagePath, c.Logger)

	content := ""
	if worldName, ok := pngMetadata["World Display Name"]; ok && worldName != "" {
		content = fmt.Sprintf("**%s**", worldName)
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] 画像ファイルの読み込みに失敗: %s", err.Error())
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(imagePath))
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] マルチパートフォームの作成に失敗: %s", err.Error())
	}

	if _, err = io.Copy(part, file); err != nil {
		return fmt.Sprintf("[Discord Webhook] ファイルのコピーに失敗: %s", err.Error())
	}

	_ = writer.WriteField("content", content)

	if err = writer.Close(); err != nil {
		return fmt.Sprintf("[Discord Webhook] マルチパートフォームのクローズに失敗: %s", err.Error())
	}

	req, err := http.NewRequest("POST", webhookURL, body)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] HTTPリクエストの作成に失敗: %s", err.Error())
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("[Discord Webhook] 送信に失敗: %s", err.Error())
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprintf("[Discord Webhook] HTTPエラー %d: %s - レスポンス: %s", resp.StatusCode, resp.Status, string(respBody))
	}

	return fmt.Sprintf("[Discord Webhook] 画像の送信に成功しました: %s", filepath.Base(imagePath))
}
