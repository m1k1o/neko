package legacy

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v3"
	"github.com/spf13/viper"

	oldEvent "github.com/m1k1o/neko/server/internal/http/legacy/event"
	oldMessage "github.com/m1k1o/neko/server/internal/http/legacy/message"
	oldTypes "github.com/m1k1o/neko/server/internal/http/legacy/types"

	"github.com/m1k1o/neko/server/internal/plugins/chat"
	"github.com/m1k1o/neko/server/internal/plugins/filetransfer"
	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/types/event"
	"github.com/m1k1o/neko/server/pkg/types/message"
)

func profileToMember(id string, profile types.MemberProfile) (*oldTypes.Member, error) {
	settings := chat.Settings{
		CanSend:    true, // defaults to true
		CanReceive: true, // defaults to true
	}

	err := profile.Plugins.Unmarshal(chat.PluginName, &settings)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return nil, fmt.Errorf("unable to unmarshal %s plugin settings from global settings: %w", chat.PluginName, err)
	}

	return &oldTypes.Member{
		ID:    id,
		Name:  profile.Name,
		Admin: profile.IsAdmin,
		Muted: !settings.CanSend,
	}, nil
}

func screenConfigurations(screenSizes []types.ScreenSize) map[int]oldTypes.ScreenConfiguration {
	rates := map[string]map[int16]struct{}{}
	for _, size := range screenSizes {
		key := fmt.Sprintf("%dx%d", size.Width, size.Height)
		if _, ok := rates[key]; !ok {
			rates[key] = map[int16]struct{}{}
		}
		rates[key][size.Rate] = struct{}{}
	}

	usedScreenSizes := map[string]struct{}{}
	screenSizesList := map[int]oldTypes.ScreenConfiguration{}
	for i, size := range screenSizes {
		key := fmt.Sprintf("%dx%d", size.Width, size.Height)
		if _, ok := usedScreenSizes[key]; ok {
			continue
		}
		usedScreenSizes[key] = struct{}{}

		c, ratesMap := 0, map[int]int16{}
		for rate := range rates[key] {
			ratesMap[c] = rate
			c++
		}

		screenSizesList[i] = oldTypes.ScreenConfiguration{
			Width:  size.Width,
			Height: size.Height,
			Rates:  ratesMap,
		}
	}

	return screenSizesList
}

func (s *session) settingsToLocks(settings types.Settings) (map[string]string, error) {
	//
	// FileTransfer
	//

	filetransferSettings := filetransfer.Settings{
		Enabled: true, // defaults to true
	}

	err := settings.Plugins.Unmarshal(filetransfer.PluginName, &filetransferSettings)
	if err != nil && !errors.Is(err, types.ErrPluginSettingsNotFound) {
		return nil, fmt.Errorf("unable to unmarshal %s plugin settings from global settings: %w", filetransfer.PluginName, err)
	}

	//
	// Locks
	//

	locks := map[string]string{}
	if settings.LockedLogins {
		locks["login"] = "" // TODO: We don't know who locked the login.
		s.lockedLogins = true
	}
	if settings.LockedControls {
		locks["control"] = "" // TODO: We don't know who locked the control.
		s.lockedControls = true
	}
	if !filetransferSettings.Enabled {
		locks["file_transfer"] = "" // TODO: We don't know who locked the file transfer.
		s.lockedFileTransfer = true
	}

	return locks, nil
}

func (s *session) sendControlHost(request message.ControlHost) error {
	lastHostID := s.lastHostID

	if request.HasHost {
		s.lastHostID = request.ID

		if request.ID == request.HostID {
			if request.ID == lastHostID || lastHostID == "" {
				return s.toClient(&oldMessage.Control{
					Event: oldEvent.CONTROL_LOCKED,
					ID:    request.HostID,
				})
			} else {
				return s.toClient(&oldMessage.AdminTarget{
					Event:  oldEvent.ADMIN_CONTROL,
					ID:     request.ID,
					Target: lastHostID,
				})
			}
		} else if request.ID == "" {
			return s.toClient(&oldMessage.Control{
				Event: oldEvent.CONTROL_LOCKED,
				ID:    request.HostID,
			})
		} else {
			return s.toClient(&oldMessage.ControlTarget{
				Event:  oldEvent.CONTROL_GIVE,
				ID:     request.ID,
				Target: request.HostID,
			})
		}
	}

	if request.ID != "" {
		s.lastHostID = ""

		if request.ID == lastHostID {
			return s.toClient(&oldMessage.Control{
				Event: oldEvent.CONTROL_RELEASE,
				ID:    request.ID,
			})
		} else {
			return s.toClient(&oldMessage.Control{
				Event: oldEvent.ADMIN_RELEASE,
				ID:    request.ID,
			})
		}
	}

	return nil
}

func (s *session) wsToClient(msg []byte) error {
	data := types.WebSocketMessage{}
	err := json.Unmarshal(msg, &data)
	if err != nil {
		return err
	}

	switch data.Event {
	// System Events
	case event.SYSTEM_DISCONNECT:
		request := &message.SystemDisconnect{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		// if profile change is the reason for disconnect, it's a kick event
		if request.Message == "profile changed" {
			request.Message = "kicked"
		}

		return s.toClient(&oldMessage.SystemMessage{
			Event:   oldEvent.SYSTEM_DISCONNECT,
			Message: request.Message,
		})

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
		s.sessions = map[string]*memberStruct{}
		for id, session := range request.Sessions {
			if !session.State.IsConnected {
				continue
			}
			member, err := profileToMember(id, session.Profile)
			if err != nil {
				return err
			}
			membersList = append(membersList, member)
			s.sessions[id] = &memberStruct{
				sent:      member.Name != "",
				connected: true,
				member:    member,
			}
		}

		err = s.toClient(&oldMessage.MembersList{
			Event:   oldEvent.MEMBER_LIST,
			Members: membersList,
		})
		if err != nil {
			return err
		}

		//
		// ScreenSize
		//

		err = s.toClient(&oldMessage.ScreenResolution{
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

		err = s.sendControlHost(request.ControlHost)
		if err != nil {
			return err
		}

		locks, err := s.settingsToLocks(request.Settings)
		if err != nil {
			return err
		}

		return s.toClient(&oldMessage.SystemInit{
			Event:           oldEvent.SYSTEM_INIT,
			ImplicitHosting: request.Settings.ImplicitHosting,
			Locks:           locks,
			// TODO: hack - we don't know if file transfer is enabled, we would need to check the global config.
			FileTransfer:      viper.GetBool("filetransfer.enabled") || (viper.GetBool("legacy") && viper.GetBool("file_transfer_enabled")),
			HeartbeatInterval: request.Settings.HeartbeatInterval,
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

		err = s.toClient(&oldMessage.ScreenConfigurations{
			Event:          oldEvent.SCREEN_CONFIGURATIONS,
			Configurations: screenConfigurations(request.ScreenSizesList),
		})
		if err != nil {
			return err
		}

		//
		// BroadcastStatus
		//

		return s.toClient(&oldMessage.BroadcastStatus{
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

		member, err := profileToMember(request.ID, request.Profile)
		if err != nil {
			return err
		}

		// only save session - will be notified on connect
		s.sessions[request.ID] = &memberStruct{
			member: member,
		}

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

		return s.toClient(&oldMessage.MemberDisconnected{
			Event: oldEvent.MEMBER_DISCONNECTED,
			ID:    request.ID,
		})

	case event.SESSION_PROFILE:
		request := &message.MemberProfile{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		// session profile is expected to change when updating a name after connecting
		m, ok := s.sessions[request.ID]
		if !ok || m == nil {
			return nil
		}

		// update member profile
		member, err := profileToMember(request.ID, request.MemberProfile)
		if err != nil {
			return err
		}

		mutedChanged := m.member.Muted != member.Muted
		m.member = member

		if m.connected && !m.sent && member.Name != "" {
			m.sent = true

			// oldEvent.MEMBER_CONNECTED if not sent already
			err = s.toClient(&oldMessage.Member{
				Event:  oldEvent.MEMBER_CONNECTED,
				Member: member,
			})
			if err != nil {
				return err
			}
		}

		if mutedChanged && member.Muted {
			return s.toClient(&oldMessage.AdminTarget{
				Event:  oldEvent.ADMIN_MUTE,
				ID:     "", // TODO: We don't know who (un)muted the user.
				Target: request.ID,
			})
		} else if mutedChanged && !member.Muted {
			return s.toClient(&oldMessage.AdminTarget{
				Event:  oldEvent.ADMIN_UNMUTE,
				ID:     "", // TODO: We don't know who (un)muted the user.
				Target: request.ID,
			})
		}

		return nil

	case event.SESSION_STATE:
		request := &message.SessionState{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		m, ok := s.sessions[request.ID]
		if !ok {
			return nil
		}

		if request.IsConnected {
			m.connected = true

			if m.member.Muted {
				err = s.toClient(&oldMessage.AdminTarget{
					Event:  oldEvent.ADMIN_MUTE,
					ID:     "", // TODO: We don't know who (un)muted the user.
					Target: request.ID,
				})
				if err != nil {
					return err
				}
			}

			if !m.sent && m.member.Name != "" {
				m.sent = true

				// oldEvent.MEMBER_CONNECTED if not sent already
				return s.toClient(&oldMessage.Member{
					Event:  oldEvent.MEMBER_CONNECTED,
					Member: m.member,
				})
			}
		}

		if !request.IsConnected {
			delete(s.sessions, request.ID)

			// oldEvent.MEMBER_DISCONNECTED if nor sent already
			return s.toClient(&oldMessage.MemberDisconnected{
				Event: oldEvent.MEMBER_DISCONNECTED,
				ID:    request.ID,
			})
		}

		return nil

	// Signal Events
	case event.SIGNAL_VIDEO, event.SIGNAL_AUDIO:
		// ignore video and audio signals, because they are not part of the legacy protocol
		return nil

	case event.SIGNAL_OFFER:
		request := &message.SignalDescription{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return s.toClient(&oldMessage.SignalOffer{
			Event: oldEvent.SIGNAL_OFFER,
			SDP:   request.SDP,
		})

	case event.SIGNAL_ANSWER:
		request := &message.SignalDescription{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return s.toClient(&oldMessage.SignalAnswer{
			Event:       oldEvent.SIGNAL_ANSWER,
			DisplayName: s.name,
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

		return s.toClient(&oldMessage.SignalCandidate{
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

		return s.toClient(&oldMessage.SignalProvide{
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

		return s.toClient(&oldMessage.Clipboard{
			Event: oldEvent.CONTROL_CLIPBOARD,
			Text:  request.Text,
		})

	case event.CONTROL_HOST:
		request := &message.ControlHost{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return s.sendControlHost(*request)

	case event.CONTROL_REQUEST:
		request := &message.SessionID{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		if s.id == request.ID {
			// if i am the one that is requesting, send CONTROL_REQUEST to me
			return s.toClient(&oldMessage.Control{
				Event: oldEvent.CONTROL_REQUEST,
				ID:    request.ID,
			})
		} else {
			// if not, let me know someone else is requesting
			return s.toClient(&oldMessage.Control{
				Event: oldEvent.CONTROL_REQUESTING,
				ID:    request.ID,
			})
		}

	// Chat Events
	case chat.CHAT_INIT:
		request := &chat.Init{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		history := []oldMessage.ChatSend{}
		for _, msg := range request.History {
			history = append(history, oldMessage.ChatSend{
				Event:   oldEvent.CHAT_MESSAGE,
				ID:      msg.ID,
				Content: msg.Content.Text,
				Created: msg.Created,
			})
		}

		return s.toClient(&oldMessage.ChatInit{
			Event:   oldEvent.CHAT_INIT,
			Enabled: request.Enabled,
			History: history,
		})

	case chat.CHAT_MESSAGE:
		request := &chat.Message{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		return s.toClient(&oldMessage.ChatSend{
			Event:   oldEvent.CHAT_MESSAGE,
			ID:      request.ID,
			Content: request.Content.Text,
			Created: request.Created,
		})

	case event.SEND_BROADCAST:
		request := &message.SendBroadcast{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		if request.Subject == "emote" {
			return s.toClient(&oldMessage.EmoteSend{
				Event: oldEvent.CHAT_EMOTE,
				ID:    request.Sender,
				Emote: request.Body.(string),
			})
		}

		return nil

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

		return s.toClient(&oldMessage.FileTransferList{
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

		return s.toClient(&oldMessage.ScreenResolution{
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

		return s.toClient(&oldMessage.BroadcastStatus{
			Event:    oldEvent.BROADCAST_STATUS,
			URL:      request.URL,
			IsActive: request.IsActive,
		})

	// Admin Events
	case event.SYSTEM_SETTINGS:
		request := &message.SystemSettingsUpdate{}
		err := json.Unmarshal(data.Payload, request)
		if err != nil {
			return err
		}

		if s.lockedControls != request.LockedControls {
			s.lockedControls = request.LockedControls

			if request.LockedControls {
				err = s.toClient(&oldMessage.AdminLock{
					Event:    oldEvent.ADMIN_LOCK,
					Resource: "control",
					ID:       request.ID,
				})
			} else {
				err = s.toClient(&oldMessage.AdminLock{
					Event:    oldEvent.ADMIN_UNLOCK,
					Resource: "control",
					ID:       request.ID,
				})
			}

			if err != nil {
				return err
			}
		}

		if s.lockedLogins != request.LockedLogins {
			s.lockedLogins = request.LockedLogins

			if request.LockedLogins {
				err = s.toClient(&oldMessage.AdminLock{
					Event:    oldEvent.ADMIN_LOCK,
					Resource: "login",
					ID:       request.ID,
				})
			} else {
				err = s.toClient(&oldMessage.AdminLock{
					Event:    oldEvent.ADMIN_UNLOCK,
					Resource: "login",
					ID:       request.ID,
				})
			}

			if err != nil {
				return err
			}
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

		if s.lockedFileTransfer != !filetransferSettings.Enabled {
			s.lockedFileTransfer = !filetransferSettings.Enabled

			if !filetransferSettings.Enabled {
				err = s.toClient(&oldMessage.AdminLock{
					Event:    oldEvent.ADMIN_LOCK,
					Resource: "file_transfer",
					ID:       request.ID,
				})
			} else {
				err = s.toClient(&oldMessage.AdminLock{
					Event:    oldEvent.ADMIN_UNLOCK,
					Resource: "file_transfer",
					ID:       request.ID,
				})
			}

			if err != nil {
				return err
			}
		}

		return nil

		/*
			case:
				s.toClient(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_BAN,
				})
			case:
				s.toClient(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_KICK,
				})
			case:
				s.toClient(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_MUTE,
				})
			case:
				s.toClient(&oldMessage.AdminTarget{
					Event: oldEvent.ADMIN_UNMUTE,
				})
		*/

	case event.SYSTEM_HEARTBEAT:
		return nil

	default:
		return fmt.Errorf("unknown event type: %s", data.Event)
	}
}
