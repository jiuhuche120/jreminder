package hook

type Webhook interface {
	// ID return hookId
	ID() string
	// Call the webhook
	Call(msg interface{}) error
}
