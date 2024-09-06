package legacy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	oldTypes "github.com/demodesk/neko/internal/http/legacy/types"

	"github.com/demodesk/neko/internal/api"
	"github.com/demodesk/neko/pkg/types"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var (
	ErrWebsocketSend  = fmt.Errorf("failed to send message to websocket")
	ErrBackendRespone = fmt.Errorf("error response from backend")
)

type memberStruct struct {
	member    *oldTypes.Member
	connected bool
	sent      bool
}

type session struct {
	logger     zerolog.Logger
	serverAddr string

	id     string
	token  string
	name   string
	client *http.Client

	lastHostID         string
	lockedControls     bool
	lockedLogins       bool
	lockedFileTransfer bool
	sessions           map[string]*memberStruct

	connClient  *websocket.Conn
	connBackend *websocket.Conn
}

func newSession(logger zerolog.Logger, serverAddr string) *session {
	return &session{
		logger:     logger,
		serverAddr: serverAddr,
		client:     http.DefaultClient,
		sessions:   make(map[string]*memberStruct),
	}
}

func (s *session) apiReq(method, path string, request, response any) error {
	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, "http://"+s.serverAddr+path, bytes.NewReader(body))
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
		return fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	if res.Body == nil {
		return nil
	}

	if response == nil {
		io.Copy(io.Discard, res.Body)
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

func (s *session) create(username, password string) (string, error) {
	data := api.SessionDataPayload{}

	err := s.apiReq(http.MethodPost, "/api/login", api.SessionLoginPayload{
		Username: username,
		Password: password,
	}, &data)
	if err != nil {
		return "", err
	}

	s.id = data.ID
	s.token = data.Token
	s.name = data.Profile.Name

	// if Cookie auth, the token will be empty
	if s.token == "" {
		return "", fmt.Errorf("token not found - make sure you are not using Cookie auth on the server")
	}

	return data.Token, nil
}

func (s *session) destroy() {
	defer s.client.CloseIdleConnections()

	// logout session
	err := s.apiReq(http.MethodPost, "/api/logout", nil, nil)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to logout")
	}
}
