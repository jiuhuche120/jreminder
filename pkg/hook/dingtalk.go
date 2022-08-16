package hook

import (
	"fmt"

	"github.com/jiuhuche120/jhttp"
)

var _ Webhook = (*DingTalk)(nil)

type DingTalk struct {
	hookID  string
	webhook string
}

func NewDingTalk(hookID, webhook string) *DingTalk {
	return &DingTalk{
		hookID:  hookID,
		webhook: webhook,
	}
}
func (d *DingTalk) ID() string {
	return d.hookID
}

func (d *DingTalk) Call(msg interface{}) error {
	client := jhttp.NewClient(
		jhttp.AddHeader("Content-Type", "application/json"),
	)
	resp, err := client.Post(d.webhook, msg)
	if err != nil {
		return err
	}
	if !resp.IsSuccess() {
		body, err := resp.Body()
		if err != nil {
			return err
		}
		return fmt.Errorf("error call webhook: %v", string(body))
	}
	return nil
}
