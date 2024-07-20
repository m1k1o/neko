package legacy

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v3"

	oldEvent "github.com/demodesk/neko/internal/http/legacy/event"
	oldMessage "github.com/demodesk/neko/internal/http/legacy/message"
	oldTypes "github.com/demodesk/neko/internal/http/legacy/types"

	"github.com/demodesk/neko/internal/plugins/chat"
	"github.com/demodesk/neko/internal/plugins/filetransfer"
	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/types/event"
	"github.com/demodesk/neko/pkg/types/message"
)

func sessionDataToMember(id string, session message.SessionData) (*oldTypes.Member, error) {
	settings := chat.Settings{
		CanSend:    true, // defaults to true
		CanReceive: true, // defaults to true
	}

	err := session.Profile.Plugins.Unmarshal(chat.PluginName, &settings)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return nil, fmt.Errorf("unable to unmarshal %s plugin settings from global settings: %w", chat.PluginName, err)
	}

	return &oldTypes.Member{
		ID:    id,
		Name:  session.Profile.Name,
		Admin: session.Profile.IsAdmin,
		Muted: !settings.CanSend,
	}, nil
}

func sendControlHost(request message.ControlHost, send func(payload any) error) error {
	if request.HasHost {
		if request.ID == request.HostID {
			return send(&oldMessage.Control{
				Event: oldEvent.CONTROL_LOCKED,
				ID:    request.HostID,
			})
		}

		return send(&oldMessage.ControlTarget{
			Event:  oldEvent.CONTROL_GIVE,
			ID:     request.HostID,
			Target: request.ID,
		})
	}

	if request.ID != "" {
		return send(&oldMessage.Control{
			Event: oldEvent.CONTROL_RELEASE,
			ID:    request.ID,
		})
	}

	return nil
}

func (s *session) wsToClient(msg []byte, sendMsg func([]byte) error) error {
	data := types.WebSocketMessage{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return err
	}

	send := func(payload any) error {
		msg, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		return sendMsg(msg)
	}

	switch data.Event {
	// System Events
	case event.SYSTEM_DISCONNECT:
		request := &message.SystemDisconnect{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.SystemMessage{
			Event:   oldEvent.SYSTEM_DISCONNECT,
			Message: request.Message,
		})

	// case:
	// 	send(&oldMessage.SystemMessage{
	// 		Event: oldEvent.SYSTEM_ERROR,
	// 	})
	case event.SYSTEM_INIT:
		request := &message.SystemInit{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		//
		// MembersList
		//

		membersList := []*oldTypes.Member{}
		s.sessions = map[string]*oldTypes.Member{}
		for id, session := range request.Sessions {
			if !session.State.IsConnected {
				continue
			}
			member, err := sessionDataToMember(id, session)
			if err != nil {
				return err
			}
			membersList = append(membersList, member)
			s.sessions[id] = member
		}

		err = send(&oldMessage.MembersList{
			Event:   oldEvent.MEMBER_LIST,
			Members: membersList,
		})
		if err != nil {
			return err
		}

		//
		// ScreenSize
		//

		err = send(&oldMessage.ScreenResolution{
			Event:  oldEvent.SCREEN_RESOLUTION,
			Width:  request.ScreenSize.Width,
			Height: request.ScreenSize.Height,
			Rate:   request.ScreenSize.Rate,
		})
		if err != nil {
			return err
		}

		// actually its already set when we create the session
		s.id = request.SessionId

		//
		// ControlHost
		//

		err = sendControlHost(request.ControlHost, send)
		if err != nil {
			return err
		}

		//
		// FileTransfer
		//

		filetransferSettings := filetransfer.Settings{
			Enabled: true, // defaults to true
		}

		err = request.Settings.Plugins.Unmarshal(filetransfer.PluginName, &filetransferSettings)
		if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
			return fmt.Errorf("unable to unmarshal %s plugin settings from global settings: %w", filetransfer.PluginName, err)
		}

		//
		// Locks
		//
		locks := map[string]string{}
		if request.Settings.LockedLogins {
			locks["login"] = "" // TODO: We don't know who locked the login.
		}
		if request.Settings.LockedControls {
			locks["control"] = "" // TODO: We don't know who locked the control.
		}
		if !filetransferSettings.Enabled {
			locks["filetransfer"] = "" // TODO: We don't know who locked the file transfer.
		}

		return send(&oldMessage.SystemInit{
			Event:           oldEvent.SYSTEM_INIT,
			ImplicitHosting: request.Settings.ImplicitHosting,
			Locks:           locks,
			FileTransfer:    true, // TODO: We don't know if file transfer is enabled, we would need to check the global config somehow.
		})

	case event.SYSTEM_ADMIN:
		request := &message.SystemAdmin{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		//
		// ScreenSizesList
		//

		rates := map[string][]int16{}
		for _, size := range request.ScreenSizesList {
			key := fmt.Sprintf("%dx%d", size.Width, size.Height)
			rates[key] = append(rates[key], size.Rate)
		}

		usedScreenSizes := map[string]struct{}{}
		screenSizesList := map[int]oldTypes.ScreenConfiguration{}
		for i, size := range request.ScreenSizesList {
			key := fmt.Sprintf("%dx%d", size.Width, size.Height)
			if _, ok := usedScreenSizes[key]; ok {
				continue
			}

			ratesMap := map[int]int16{}
			for i, rate := range rates[key] {
				ratesMap[i] = rate
			}

			screenSizesList[i] = oldTypes.ScreenConfiguration{
				Width:  size.Width,
				Height: size.Height,
				Rates:  ratesMap,
			}
		}

		err = send(&oldMessage.ScreenConfigurations{
			Event:          oldEvent.SCREEN_CONFIGURATIONS,
			Configurations: screenSizesList,
		})
		if err != nil {
			return err
		}

		//
		// BroadcastStatus
		//

		return send(&oldMessage.BroadcastStatus{
			Event:    oldEvent.BROADCAST_STATUS,
			URL:      request.BroadcastStatus.URL,
			IsActive: request.BroadcastStatus.IsActive,
		})

	// Member Events

	case event.SESSION_CREATED:
		request := &message.SessionData{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		member, err := sessionDataToMember(request.ID, *request)
		if err != nil {
			return err
		}

		// only save session - will be notified on connect
		s.sessions[request.ID] = member

		return nil

	case event.SESSION_DELETED:
		request := &message.SessionID{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		// only continue if session is in the list - should have been already removed
		if _, ok := s.sessions[request.ID]; !ok {
			return nil
		}

		delete(s.sessions, request.ID)

		return send(&oldMessage.MemberDisconnected{
			Event: oldEvent.MEMBER_DISCONNECTED,
			ID:    request.ID,
		})

	case event.SESSION_STATE:
		request := &message.SessionState{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		member, ok := s.sessions[request.ID]
		if !ok {
			return nil
		}

		if request.IsConnected && member != nil {
			s.sessions[request.ID] = nil

			// oldEvent.MEMBER_CONNECTED if not sent already
			return send(&oldMessage.Member{
				Event:  oldEvent.MEMBER_CONNECTED,
				Member: member,
			})
		}

		if !request.IsConnected {
			delete(s.sessions, request.ID)

			// oldEvent.MEMBER_DISCONNECTED if nor sent already
			return send(&oldMessage.MemberDisconnected{
				Event: oldEvent.MEMBER_DISCONNECTED,
				ID:    request.ID,
			})
		}

		return nil

	// Signal Events
	case event.SIGNAL_OFFER:
		request := &message.SignalDescription{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.SignalOffer{
			Event: oldEvent.SIGNAL_OFFER,
			SDP:   request.SDP,
		})

	case event.SIGNAL_ANSWER:
		request := &message.SignalDescription{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.SignalAnswer{
			Event:       oldEvent.SIGNAL_ANSWER,
			DisplayName: s.profile.Name, // DisplayName
			SDP:         request.SDP,
		})

	case event.SIGNAL_CANDIDATE:
		request := &message.SignalCandidate{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		json, err := json.Marshal(request.ICECandidateInit)
		if err != nil {
			return err
		}

		return send(&oldMessage.SignalCandidate{
			Event: oldEvent.SIGNAL_CANDIDATE,
			Data:  string(json),
		})

	case event.SIGNAL_PROVIDE:
		request := &message.SignalProvide{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		iceServers := []webrtc.ICEServer{}
		for _, ice := range request.ICEServers {
			iceServers = append(iceServers, webrtc.ICEServer{
				URLs:           ice.URLs,
				Username:       ice.Username,
				Credential:     ice.Credential,
				CredentialType: webrtc.ICECredentialTypePassword,
			})
		}

		return send(&oldMessage.SignalProvide{
			Event: oldEvent.SIGNAL_PROVIDE,
			ID:    s.id, // SessionId
			SDP:   request.SDP,
			Lite:  len(iceServers) == 0, // if no ICE servers are provided, it's a lite offer
			ICE:   iceServers,
		})

	// Control Events
	case event.CLIPBOARD_UPDATED:
		request := &message.ClipboardData{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.Clipboard{
			Event: oldEvent.CONTROL_CLIPBOARD,
			Text:  request.Text,
		})

	case event.CONTROL_HOST:
		request := &message.ControlHost{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return sendControlHost(*request, send)

	case event.CONTROL_REQUEST:
		request := &message.SessionID{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		if s.id == request.ID {
			// if i am the one that is requesting, send CONTROL_REQUEST to me
			return send(&oldMessage.Control{
				Event: oldEvent.CONTROL_REQUEST,
				ID:    request.ID,
			})
		} else {
			// if not, let me know someone else is requesting
			return send(&oldMessage.Control{
				Event: oldEvent.CONTROL_REQUESTING,
				ID:    request.ID,
			})
		}

	// Chat Events
	case chat.CHAT_MESSAGE:
		request := &chat.Message{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.ChatSend{
			Event:   oldEvent.CHAT_MESSAGE,
			ID:      request.ID,
			Content: request.Content.Text,
		})

	// TODO: emotes.
	//case:
	//	send(&oldMessage.EmoteSend{
	//		Event: oldEvent.CHAT_EMOTE,
	//	})

	// File Transfer Events
	case filetransfer.FILETRANSFER_UPDATE:
		request := &filetransfer.Message{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		files := []oldTypes.FileListItem{}
		for _, file := range request.Files {
			var itemType string
			switch file.Type {
			case filetransfer.ItemTypeFile:
				itemType = "file"
			case filetransfer.ItemTypeDir:
				itemType = "dir"
			}
			files = append(files, oldTypes.FileListItem{
				Filename: file.Name,
				Type:     itemType,
				Size:     file.Size,
			})
		}

		return send(&oldMessage.FileTransferList{
			Event: oldEvent.FILETRANSFER_LIST,
			Cwd:   request.RootDir,
			Files: files,
		})

	// Screen Events
	case event.SCREEN_UPDATED:
		request := &message.ScreenSizeUpdate{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.ScreenResolution{
			Event:  oldEvent.SCREEN_RESOLUTION,
			ID:     request.ID,
			Width:  request.ScreenSize.Width,
			Height: request.ScreenSize.Height,
			Rate:   request.ScreenSize.Rate,
		})

	// Broadcast Events
	case event.BROADCAST_STATUS:
		request := &message.BroadcastStatus{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return send(&oldMessage.BroadcastStatus{
			Event:    oldEvent.BROADCAST_STATUS,
			URL:      request.URL,
			IsActive: request.IsActive,
		})

		/*
			// Admin Events
			case:
				send(&oldMessage.AdminLock{
					Event: oldEvent.ADMIN_UNLOCK,
				})
			case:
				send(&oldMessage.AdminLock{
					Event: oldEvent.ADMIN_LOCK,
				})
			case:
				send(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_CONTROL,
				} // )oldMessage.Admin
			case:
				send(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_RELEASE,
				} // )oldMessage.Admin
			case:
				send(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_MUTE,
				})
			case:
				send(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_UNMUTE,
				})
			case:
				send(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_KICK,
				})
			case:
				send(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_BAN,
				})
		*/

	case event.SYSTEM_HEARTBEAT:
		return nil

	default:
		return fmt.Errorf("unknown event type: %s", data.Event)
	}
}
