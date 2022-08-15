package event

import "time"

type Event struct {
	ID         string
	RuleID     string
	Webhooks   []string
	Msg        interface{}
	CreateTime time.Time
}

func NewEvent(ruleID string, webhooks []string, msg interface{}) *Event {
	// TODO: create event ID
	return &Event{
		ID:         "",
		RuleID:     ruleID,
		Webhooks:   webhooks,
		Msg:        msg,
		CreateTime: time.Now(),
	}
}
