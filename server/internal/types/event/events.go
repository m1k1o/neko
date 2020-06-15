package event

const (
	SYSTEM_DISCONNECT = "system/disconnect"
)

const (
	SIGNAL_ANSWER  = "signal/answer"
	SIGNAL_PROVIDE = "signal/provide"
)

const (
	MEMBER_LIST         = "member/list"
	MEMBER_CONNECTED    = "member/connected"
	MEMBER_DISCONNECTED = "member/disconnected"
)

const (
	CONTROL_LOCKED     = "control/locked"
	CONTROL_RELEASE    = "control/release"
	CONTROL_REQUEST    = "control/request"
	CONTROL_REQUESTING = "control/requesting"
	CONTROL_GIVE       = "control/give"
	CONTROL_CLIPBOARD  = "control/clipboard"
	CONTROL_KEYBOARD   = "control/keyboard"
)

const (
	CHAT_MESSAGE = "chat/message"
	CHAT_EMOTE   = "chat/emote"
)

const (
	SCREEN_CONFIGURATIONS = "screen/configurations"
	SCREEN_RESOLUTION     = "screen/resolution"
	SCREEN_SET            = "screen/set"
)

const (
	ADMIN_BAN     = "admin/ban"
	ADMIN_KICK    = "admin/kick"
	ADMIN_LOCK    = "admin/lock"
	ADMIN_MUTE    = "admin/mute"
	ADMIN_UNLOCK  = "admin/unlock"
	ADMIN_UNMUTE  = "admin/unmute"
	ADMIN_CONTROL = "admin/control"
	ADMIN_RELEASE = "admin/release"
	ADMIN_GIVE    = "admin/give"
)
