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

type session struct {
	url     string
	id      string
	token   string
	profile types.MemberProfile
	client  *http.Client

	lastHostID string
	sessions   map[string]*oldTypes.Member

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

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", res.StatusCode, body)
	}

	return json.NewDecoder(res.Body).Decode(response)
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
