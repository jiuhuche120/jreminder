package types

type Msg struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
	At      At     `json:"at"`
}
type Text struct {
	Content string `json:"content"`
}
type At struct {
	AtMobiles []string `json:"atMobiles"`
	AtUserIds []string `json:"atUserIds"`
	IsAtAll   bool     `json:"isAtAll"`
}
