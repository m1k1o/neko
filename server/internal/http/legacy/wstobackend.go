package legacy

import (
	"encoding/json"
	"fmt"

	"github.com/demodesk/neko/internal/http/legacy/event"
	"github.com/demodesk/neko/internal/http/legacy/message"
)

func (h *LegacyHandler) wsToBackend(msg []byte) ([]byte, error) {
	header := message.Message{}
	err := json.Unmarshal(msg, &header)
	if err != nil {
		return nil, err
	}

	var response any
	switch header.Event {
	// Signal Events
	case event.SIGNAL_OFFER:
		request := &message.SignalOffer{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.SIGNAL_ANSWER:
		request := &message.SignalAnswer{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.SIGNAL_CANDIDATE:
		request := &message.SignalCandidate{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	// Control Events
	case event.CONTROL_RELEASE:
	case event.CONTROL_REQUEST:
	case event.CONTROL_GIVE:
		request := &message.Control{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.CONTROL_CLIPBOARD:
		request := &message.Clipboard{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.CONTROL_KEYBOARD:
		request := &message.Keyboard{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	// Chat Events
	case event.CHAT_MESSAGE:
		request := &message.ChatReceive{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.CHAT_EMOTE:
		request := &message.EmoteReceive{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	// File Transfer Events
	case event.FILETRANSFER_REFRESH:

	// Screen Events
	case event.SCREEN_RESOLUTION:
	case event.SCREEN_CONFIGURATIONS:
	case event.SCREEN_SET:
		request := &message.ScreenResolution{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	// Broadcast Events
	case event.BROADCAST_CREATE:
		request := &message.BroadcastCreate{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.BROADCAST_DESTROY:

	// Admin Events
	case event.ADMIN_LOCK:
		request := &message.AdminLock{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.ADMIN_UNLOCK:
		request := &message.AdminLock{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.ADMIN_CONTROL:
	case event.ADMIN_RELEASE:
	case event.ADMIN_GIVE:
		request := &message.Admin{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.ADMIN_BAN:
		request := &message.Admin{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.ADMIN_KICK:
		request := &message.Admin{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.ADMIN_MUTE:
		request := &message.Admin{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	case event.ADMIN_UNMUTE:
		request := &message.Admin{}
		err := json.Unmarshal(msg, request)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown event type: %s", header.Event)
	}

	return json.Marshal(request)
}
