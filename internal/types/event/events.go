package event

const (
	SYSTEM_CONNECT    = "system/connect" // TODO: New.
	SYSTEM_DISCONNECT = "system/disconnect"
)

const (
	SIGNAL_REQUEST = "signal/request" // TODO: New.
	SIGNAL_ANSWER  = "signal/answer"
	SIGNAL_PROVIDE = "signal/provide"
)

const (
	MEMBER_LIST         = "member/list" // TODO: Remove.
	MEMBER_CONNECTED    = "member/connected"
	MEMBER_UPDATED      = "member/updated" // TODO: New.
	MEMBER_DISCONNECTED = "member/disconnected"
)

const (
	CONTROL_HOST    = "control/host"
	CONTROL_RELEASE = "control/release"
	CONTROL_REQUEST = "control/request"
	CONTROL_MOVE    = "control/move" // TODO: New. (fallback)
	CONTROL_SCROLL  = "control/scroll" // TODO: New. (fallback)
	CONTROL_KEYDOWN = "control/keydown" // TODO: New. (fallback)
	CONTROL_KEYUP   = "control/keyup" // TODO: New. (fallback)
)

const (
	SCREEN_UPDATED = "screen/updated"
	SCREEN_SET     = "screen/set"
)

const (
	CLIPBOARD_UPDATED = "clipboard/updated"
	CLIPBOARD_SET     = "clipboard/set"
)

const (
	KEYBOARD_MODIFIERS = "keyboard/modifiers"
	KEYBOARD_LAYOUT    = "keyboard/layout"
)

const (
	BORADCAST_STATUS = "broadcast/status"
)
