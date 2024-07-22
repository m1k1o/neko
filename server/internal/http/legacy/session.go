package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	oldTypes "github.com/demodesk/neko/internal/http/legacy/types"

	"github.com/demodesk/neko/internal/api"
	"github.com/demodesk/neko/pkg/types"
	"github.com/gorilla/websocket"
)

var (
	ErrWebsocketSend  = fmt.Errorf("failed to send message to websocket")
	ErrBackendRespone = fmt.Errorf("error response from backend")
)

type session struct {
	url     string
	id      string
	token   string
	profile types.MemberProfile
	client  *http.Client

	lastHostID         string
	lockedControls     bool
	lockedLogins       bool
	lockedFileTransfer bool
	sessions           map[string]*oldTypes.Member

	connClient  *websocket.Conn
	connBackend *websocket.Conn
}

func newSession(url string) *session {
	return &session{
		url:      url,
		client:   http.DefaultClient,
		sessions: make(map[string]*oldTypes.Member),
	}
}

func (s *session) apiReq(method, path string, request, response any) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, s.url+path, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		// try to unmarsal as json error message
		var apiErr struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return fmt.Errorf("%w: %s", ErrBackendRespone, apiErr.Message)
		}
		// return raw body if failed to unmarshal
		return fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
	}

	if res.Body == nil {
		return nil
	}

	return json.NewDecoder(res.Body).Decode(response)
}

// send message to client (in old format)
func (s *session) toClient(payload any) error {
	msg, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	err = s.connClient.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrWebsocketSend, err)
	}

	return nil
}

// send message to backend (in new format)
func (s *session) toBackend(event string, payload any) error {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg, err := json.Marshal(&types.WebSocketMessage{
		Event:   event,
		Payload: rawPayload,
	})
	if err != nil {
		return err
	}

	err = s.connBackend.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrWebsocketSend, err)
	}

	return nil
}

func (s *session) create(password string) (string, error) {
	data := api.SessionDataPayload{}

	// pefrom login with arbitrary username that will be changed later
	err := s.apiReq(http.MethodPost, "/api/login", api.SessionLoginPayload{
		Username: "admin",
		Password: password,
	}, &data)
	if err != nil {
		return "", err
	}

	s.id = data.ID
	s.token = data.Token
	s.profile = data.Profile

	if s.token == "" {
		return "", fmt.Errorf("token not found")
	}

	return data.Token, nil
}

func (s *session) destroy() {
	defer s.client.CloseIdleConnections()

	// logout session
	err := s.apiReq(http.MethodPost, "/api/logout", nil, nil)
	if err != nil {
		log.Println("failed to logout session:", err)
	}
}
