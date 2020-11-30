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
	CONTROL_LOCKED     = "control/locked" // TODO: Remove.
	CONTROL_REQUESTING = "control/requesting" // TODO: Remove.
	CONTROL_GIVE       = "control/give" // TODO: Remove.
	CONTROL_CLIPBOARD  = "control/clipboard" // TODO: Remove.
	CONTROL_KEYBOARD   = "control/keyboard" // TODO: Remove.
	CONTROL_UPDATED = "control/updated" // TODO: New.
	CONTROL_RELEASE = "control/release"
	CONTROL_REQUEST = "control/request"
	CONTROL_MOVE    = "control/move" // TODO: New. (fallback)
	CONTROL_SCROLL  = "control/scroll" // TODO: New. (fallback)
	CONTROL_KEYDOWN = "control/keydown" // TODO: New. (fallback)
	CONTROL_KEYUP   = "control/keyup" // TODO: New. (fallback)
)

const (
	SCREEN_CONFIGURATIONS = "screen/configurations" // TODO: Remove.
	SCREEN_RESOLUTION     = "screen/resolution" // TODO: Remove.

	SCREEN_UPDATED = "screen/updated" // TODO: New.
	SCREEN_SET     = "screen/set"
)

const (
	CLIPBOARD_UPDATED = "clipboard/updated" // TODO: New.
	CLIPBOARD_SET     = "clipboard/set" // TODO: New.
)

const (
	KEYBOARD_MODIFIERS = "keyboard/modifiers" // TODO: New.
	KEYBOARD_LAYOUT    = "keyboard/layout" // TODO: New.
)

const (
	BORADCAST_STATUS  = "broadcast/status" // TODO: Remove.
	BORADCAST_CREATE  = "broadcast/create" // TODO: Remove.
	BORADCAST_DESTROY = "broadcast/destroy" // TODO: Remove.
)

const (
	ADMIN_CONTROL = "admin/control" // TODO: Remove.
	ADMIN_RELEASE = "admin/release" // TODO: Remove.
	ADMIN_GIVE    = "admin/give" // TODO: Remove.
)
