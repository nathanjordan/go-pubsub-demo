package httppubsub

// Message is the JSON pubsub wire protocol.
type Message struct {
	Data string `json:"data"`
}
