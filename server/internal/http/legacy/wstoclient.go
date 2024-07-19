package legacy

import (
	"encoding/json"
	"fmt"

	"github.com/demodesk/neko/internal/http/legacy/event"
	"github.com/demodesk/neko/internal/http/legacy/message"
)

func (h *LegacyHandler) wsToClient(msg []byte) ([]byte, error) {
	header := message.Message{}
	err := json.Unmarshal(msg, &header)
	if err != nil {
		return nil, err
	}

	var payload any
	switch header.Event {
	// System Events
	case:
		payload = &message.SystemMessage{
			Event: event.SYSTEM_DISCONNECT,
		}
	case:
		payload = &message.SystemMessage{
			Event: event.SYSTEM_ERROR,
		}
	case:
		payload = &message.SystemInit{
			Event: event.SYSTEM_INIT,
		}

	// Member Events
	case:
		payload = &message.MembersList{
			Event: event.MEMBER_LIST,
		}
	case:
		payload = &message.Member{
			Event: event.MEMBER_CONNECTED,
		}
	case:
		payload = &message.MemberDisconnected{
			Event: event.MEMBER_DISCONNECTED,
		}

	// Signal Events
	case:
		payload = &message.SignalOffer{
			Event: event.SIGNAL_OFFER,
		}
	case:
		payload = &message.SignalAnswer{
			Event: event.SIGNAL_ANSWER,
		}
	case:
		payload = &message.SignalCandidate{
			Event: event.SIGNAL_CANDIDATE,
		}
	case:
		payload = &message.SignalProvide{
			Event: event.SIGNAL_PROVIDE,
		}

	// Control Events
	case:
		payload = &message.Clipboard{
			Event: event.CONTROL_CLIPBOARD,
		}
	case:
		payload = &message.Control{
			Event: event.CONTROL_REQUEST,
		}
	case:
		payload = &message.Control{
			Event: event.CONTROL_REQUESTING,
		}
	case:
		payload = &message.ControlTarget{
			Event: event.CONTROL_GIVE,
		} // message.AdminTarget
	case:
		payload = &message.Control{
			Event: event.CONTROL_RELEASE,
		}
	case:
		payload = &message.Control{
			Event: event.CONTROL_LOCKED,
		}

	// Chat Events
	case:
		payload = &message.ChatSend{
			Event: event.CHAT_MESSAGE,
		}
	case:
		payload = &message.EmoteSend{
			Event: event.CHAT_EMOTE,
		}

	// File Transfer Events
	case:
		payload = &message.FileTransferList{
			Event: event.FILETRANSFER_LIST,
		}

	// Screen Events
	case:
		payload = &message.ScreenResolution{
			Event: event.SCREEN_RESOLUTION,
		}
	case:
		payload = &message.ScreenConfigurations{
			Event: event.SCREEN_CONFIGURATIONS,
		}

	// Broadcast Events
	case:
		payload = &message.BroadcastStatus{
			Event: event.BROADCAST_STATUS,
		}

	// Admin Events
	case:
		payload = &message.AdminLock{
			Event: event.ADMIN_UNLOCK,
		}
	case:
		payload = &message.AdminLock{
			Event: event.ADMIN_LOCK,
		}
	case:
		payload = &message.AdminTarget{
			Event: event.ADMIN_CONTROL,
		} // message.Admin
	case:
		payload = &message.AdminTarget{
			Event: event.ADMIN_RELEASE,
		} // message.Admin
	case:
		payload = &message.AdminTarget{
			Event: event.ADMIN_MUTE,
		}
	case:
		payload = &message.AdminTarget{
			Event: event.ADMIN_UNMUTE,
		}
	case:
		payload = &message.AdminTarget{
			Event: event.ADMIN_KICK,
		}
	case:
		payload = &message.AdminTarget{
			Event: event.ADMIN_BAN,
		}
	default:
		return nil, fmt.Errorf("unknown event type: %s", header.Event)
	}

	return json.Marshal(payload)
}
