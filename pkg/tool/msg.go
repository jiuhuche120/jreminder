package tool

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

func NewGithubMsg(pulls []*PullRequest, text string) Msg {
	var msg Msg
	msg.MsgType = "text"
	for i := 0; i < len(pulls); i++ {
		msg.Text.Content += "🔗:" + pulls[i].HtmlUrl + " " + text + " @" + pulls[i].DingTalk + "\n"
		msg.At.AtMobiles = append(msg.At.AtMobiles, pulls[i].DingTalk)
	}
	msg.At.IsAtAll = false
	return msg
}

func NewTeambitionMsg(tasks []*SubTask, text string) Msg {
	var msg Msg
	msg.MsgType = "text"
	for _, task := range tasks {
		msg.Text.Content += "🔗:" + task.Url + " " + text + " @" + task.DingTalk + "\n"
		msg.At.AtMobiles = append(msg.At.AtMobiles, task.DingTalk)
	}
	msg.At.IsAtAll = false
	return msg
}
