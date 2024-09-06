package chat

import "time"

const PluginName = "chat"

const (
	CHAT_INIT    = "chat/init"
	CHAT_MESSAGE = "chat/message"
)

type Init struct {
	Enabled bool `json:"enabled"`
}

type Content struct {
	Text string `json:"text"`
}

type Message struct {
	ID      string    `json:"id"`
	Created time.Time `json:"created"`
	Content Content   `json:"content"`
}
