package event

const (
	SYSTEM_INIT       = "system/init"
	SYSTEM_ADMIN      = "system/admin"
	SYSTEM_DISCONNECT = "system/disconnect"
)

const (
	SIGNAL_REQUEST = "signal/request"
	SIGNAL_ANSWER  = "signal/answer"
	SIGNAL_PROVIDE = "signal/provide"
)

const (
	MEMBER_CREATED      = "member/created"
	MEMBER_DELETED      = "member/deleted"
	MEMBER_PROFILE      = "member/profile"
	MEMBER_STATE        = "member/state"
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
	CURSOR_IMAGE = "cursor/image"
)

const (
	BORADCAST_STATUS = "broadcast/status"
)