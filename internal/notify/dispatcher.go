package notify

import (
	"vrc_log_watcher/internal/applog"
	"vrc_log_watcher/internal/models"
)

// Dispatcher 各通知チャネルへのルーティングを管理
type Dispatcher struct {
	HTTP      *HTTPClient
	XSOverlay *XSOverlayClient
	Discord   *DiscordClient
}

// NewDispatcher 新しいDispatcherを作成
func NewDispatcher(logger applog.Logger) *Dispatcher {
	return &Dispatcher{
		HTTP:      &HTTPClient{Logger: logger},
		XSOverlay: &XSOverlayClient{},
		Discord:   &DiscordClient{Logger: logger},
	}
}

// DispatchResult 通知送信結果
type DispatchResult struct {
	Message string
	Label   string
}

// Send 設定に基づいて適切な通知チャネルに送信
func (d *Dispatcher) Send(eventString string, setting models.Setting, isScreenshot bool) DispatchResult {
	switch setting.Type {
	case "WebRequest":
		var message string
		if isScreenshot {
			message = d.HTTP.PostWithImage(eventString, setting)
		} else {
			message = d.HTTP.Post(eventString, setting)
		}
		return DispatchResult{Message: message, Label: setting.Title + ": Web Request"}

	case "SendXSOverlay":
		message := d.XSOverlay.Post(eventString, setting.Title)
		return DispatchResult{Message: message, Label: setting.Title + ": XSOverlay"}

	case "SendDiscordWebHook":
		var message string
		if isScreenshot {
			message = d.Discord.PostWithImage(eventString, setting.Title, setting.URL)
		} else {
			message = d.Discord.Post(eventString, setting.Title, setting.URL)
		}
		return DispatchResult{Message: message, Label: setting.Title + ": Discord Webhook"}

	case "OutputTextFile":
		message := OutputTextFile(eventString, setting.ID, setting.Title)
		return DispatchResult{Message: message, Label: setting.Title + ": テキスト出力"}

	case "LogOnly":
		message := "[ログ出力] " + setting.Title + ": " + eventString
		return DispatchResult{Message: message, Label: setting.Title + ": ログ出力"}

	case "Disable":
		return DispatchResult{}

	default:
		return DispatchResult{}
	}
}
