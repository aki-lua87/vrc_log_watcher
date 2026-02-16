package notify

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"vrc_log_watcher/internal/models"
)

// XSOverlayClient XSOverlayとのWebSocket通信を管理
type XSOverlayClient struct {
	ws *websocket.Conn
	mu sync.Mutex
}

// Ping XSOverlayへの接続をチェック
func (c *XSOverlayClient) Ping() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws != nil {
		log.Default().Println("[DEBUG] [LOG] a.xsWS != nil")
		if err := c.ws.WriteMessage(2, []byte{}); err != nil {
			log.Default().Println("[DEBUG] [LOG] a.xsWS.WriteMessage Error:" + err.Error())
			c.ws.Close()
			c.ws = nil
		}
	} else {
		log.Default().Println("[DEBUG] [LOG] a.xsWS == nil")
	}
}

// Post XSOverlayへ通知を送信
func (c *XSOverlayClient) Post(eventString string, title string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	notification := &models.XSONotificationObject{
		Type:    1,
		Title:   title,
		Content: eventString,
		Timeout: 1.2,
		Height:  100.0,
	}
	notificationJSON, _ := json.Marshal(notification)

	url := "localhost"
	port := "42070"

	apiObject := &models.XSOApiObject{
		Sender:   "VRCLogWatcher",
		Target:   "XSOverlay",
		Command:  "SendNotification",
		JsonData: string(notificationJSON),
	}

	if c.ws == nil {
		log.Default().Println("[DEBUG] [LOG] a.xsWS != nil")
		ws, _, err := websocket.DefaultDialer.Dial("ws://"+url+":"+port+"/?client=VRCLogWatcher", nil)
		if err != nil {
			c.ws = nil
			return err.Error()
		}
		c.ws = ws
	}

	if err := c.ws.WriteJSON(apiObject); err != nil {
		c.ws.Close()
		// リトライ
		rws, _, errr1 := websocket.DefaultDialer.Dial("ws://"+url+":"+port+"/?client=VRCLogWatcher", nil)
		if errr1 != nil {
			c.ws = nil
			return errr1.Error()
		}
		c.ws = rws
		if errr2 := c.ws.WriteJSON(apiObject); errr2 != nil {
			c.ws.Close()
			c.ws = nil
		}
		return err.Error()
	}

	return "[XS Overlay] 通知の送信に成功しました: " + title + ": " + eventString
}
