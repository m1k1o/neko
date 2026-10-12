package event

const (
	SYSTEM_INIT       = "system/init"
	SYSTEM_DISCONNECT = "system/disconnect"
	SYSTEM_ERROR      = "system/error"
)

const (
	CLIENT_HEARTBEAT = "client/heartbeat"
)

const (
	SIGNAL_OFFER     = "signal/offer"
	SIGNAL_ANSWER    = "signal/answer"
	SIGNAL_PROVIDE   = "signal/provide"
	SIGNAL_CANDIDATE = "signal/candidate"
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
	CONTROL_MOVE       = "control/move"
	CONTROL_SCROLL     = "control/scroll"
	CONTROL_BUTTONDOWN = "control/buttondown"
	CONTROL_BUTTONUP   = "control/buttonup"
	CONTROL_KEYDOWN    = "control/keydown"
	CONTROL_KEYUP      = "control/keyup"
)

const (
	CHAT_MESSAGE = "chat/message"
	CHAT_EMOTE   = "chat/emote"
)

const (
	FILETRANSFER_LIST    = "filetransfer/list"
	FILETRANSFER_REFRESH = "filetransfer/refresh"
)

const (
	SCREEN_CONFIGURATIONS = "screen/configurations"
	SCREEN_RESOLUTION     = "screen/resolution"
	SCREEN_SET            = "screen/set"
)

const (
	BROADCAST_STATUS  = "broadcast/status"
	BROADCAST_CREATE  = "broadcast/create"
	BROADCAST_DESTROY = "broadcast/destroy"
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
