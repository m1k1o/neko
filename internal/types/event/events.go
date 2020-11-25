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
	SCREEN_CONFIGURATIONS = "screen/configurations"
	SCREEN_RESOLUTION     = "screen/resolution"
	SCREEN_SET            = "screen/set"
)

const (
	BORADCAST_STATUS  = "broadcast/status"
	BORADCAST_CREATE  = "broadcast/create"
	BORADCAST_DESTROY = "broadcast/destroy"
)

const (
	ADMIN_KICK    = "admin/kick"
	ADMIN_CONTROL = "admin/control"
	ADMIN_RELEASE = "admin/release"
	ADMIN_GIVE    = "admin/give"
)
