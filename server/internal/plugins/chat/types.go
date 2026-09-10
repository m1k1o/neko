package chat

import appchat "github.com/m1k1o/neko/server/internal/application/chat"

const PluginName = "chat"

const (
	CHAT_INIT    = "chat/init"
	CHAT_MESSAGE = "chat/message"
)

type Init struct {
	Enabled bool `json:"enabled"`
}

type Content = appchat.Content
type Message = appchat.Message
