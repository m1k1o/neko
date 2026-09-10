package event

const (
	SYSTEM_INIT       = "system/init"
	SYSTEM_ADMIN      = "system/admin"
	SYSTEM_SETTINGS   = "system/settings"
	SYSTEM_LOGS       = "system/logs"
	SYSTEM_DISCONNECT = "system/disconnect"
	SYSTEM_HEARTBEAT  = "system/heartbeat"
)

const (
	CLIENT_HEARTBEAT = "client/heartbeat"
)

const (
	SIGNAL_REQUEST   = "signal/request"
	SIGNAL_RESTART   = "signal/restart"
	SIGNAL_OFFER     = "signal/offer"
	SIGNAL_ANSWER    = "signal/answer"
	SIGNAL_PROVIDE   = "signal/provide"
	SIGNAL_CANDIDATE = "signal/candidate"
	SIGNAL_CLOSE     = "signal/close"
	SIGNAL_VIDEO     = "signal/video"
	SIGNAL_AUDIO     = "signal/audio"
)

const (
	SESSION_CREATED = "session/created"
	SESSION_DELETED = "session/deleted"
	SESSION_PROFILE = "session/profile"
	SESSION_STATE   = "session/state"
	SESSION_CURSORS = "session/cursors"
)

const (
	CONTROL_HOST    = "control/host"
	CONTROL_RELEASE = "control/release"
	CONTROL_REQUEST = "control/request"
	CONTROL_RENEW   = "control/renew"
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
	KEYBOARD_MAP       = "keyboard/map"
)

const (
	BROADCAST_STATUS = "broadcast/status"
)

const (
	SEND_UNICAST   = "send/unicast"
	SEND_BROADCAST = "send/broadcast"
)

const (
	FILE_CHOOSER_DIALOG_OPENED = "file_chooser_dialog/opened"
	FILE_CHOOSER_DIALOG_CLOSED = "file_chooser_dialog/closed"
)
