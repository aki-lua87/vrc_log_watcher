package models

// NoticeLog フロントエンドに送信するログ情報
type NoticeLog struct {
	Text          string `json:"text"`
	MetaData      string `json:"metaData"`
	Title         string `json:"title"`
	SettingID     string `json:"settingId"`
	CanCopy       bool   `json:"canCopy"`
	Timestamp     string `json:"timestamp"`
	IsSystem      bool   `json:"isSystem"`
	IsError       bool   `json:"isError"`
	ActionSuccess bool   `json:"actionSuccess"`
}

// SaveData アプリケーション設定の保存データ
type SaveData struct {
	LogPath  string    `json:"path"`
	Settings []Setting `json:"settings"`
}

// Setting 個別の監視・通知ルール設定
type Setting struct {
	ID            string            `json:"id"`
	Title         string            `json:"title"`
	Details       string            `json:"details"`
	IsCore        bool              `json:"isCore"`
	Target        string            `json:"target"`
	Type          string            `json:"type"`
	URL           string            `json:"url"`
	RegExp        string            `json:"regexp"`
	Exclude       string            `json:"exclude"`
	MessageKey    string            `json:"messageKey"`
	ExtraFields   map[string]string `json:"extraFields"`
	SimpleMode    bool              `json:"simpleMode"`
	SimplePattern string            `json:"simplePattern"`
	SimpleBlocks  []SimpleBlock     `json:"simpleBlocks"`
}

// SimpleBlock 簡易入力のブロック定義
type SimpleBlock struct {
	Type  string `json:"type"`
	Start string `json:"start"`
	End   string `json:"end"`
}

// XSOApiObject XSOverlay API通信用オブジェクト
type XSOApiObject struct {
	Sender   string `json:"sender"`
	Target   string `json:"target"`
	Command  string `json:"command"`
	JsonData string `json:"jsonData"`
	RawData  string `json:"rawData"`
}

// XSONotificationObject XSOverlay通知データ
type XSONotificationObject struct {
	Type    int     `json:"type"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Timeout float32 `json:"timeout"`
	Height  float32 `json:"height"`
}
