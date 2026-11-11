package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/types"
)

// loadWebRTCFromEnv mirrors the production config flow: env prefix setup,
// flag registration via Init, then value population via Set.
func loadWebRTCFromEnv(t *testing.T, env map[string]string) *WebRTC {
	t.Helper()

	viper.Reset()
	viper.SetEnvPrefix("NEKO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	for k, v := range env {
		t.Setenv(k, v)
	}

	s := &WebRTC{}
	cmd := &cobra.Command{}
	if err := s.Init(cmd); err != nil {
		t.Fatal(err)
	}
	s.Set()
	return s
}

func TestICEServersSecretEnv(t *testing.T) {
	s := loadWebRTCFromEnv(t, map[string]string{
		"NEKO_WEBRTC_ICESERVERS_FRONTEND": `[
			{"urls": ["turn:rest.example.com:3478"], "secret": "north", "ttl": 3600},
			{"urls": "stun:stun.example.com:3478"}
		]`,
		"NEKO_WEBRTC_ICESERVERS_BACKEND": `[
			{"urls": ["turn:rest.example.com:3478"], "secret": "south"}
		]`,
	})

	expectedFrontend := []types.ICEServer{
		{URLs: []string{"turn:rest.example.com:3478"}, Secret: "north", TTL: 3600},
		{URLs: []string{"stun:stun.example.com:3478"}},
	}
	if !reflect.DeepEqual(s.ICEServersFrontend, expectedFrontend) {
		t.Errorf("frontend ICE servers:\n got %+v\nwant %+v", s.ICEServersFrontend, expectedFrontend)
	}

	expectedBackend := []types.ICEServer{
		{URLs: []string{"turn:rest.example.com:3478"}, Secret: "south"},
	}
	if !reflect.DeepEqual(s.ICEServersBackend, expectedBackend) {
		t.Errorf("backend ICE servers:\n got %+v\nwant %+v", s.ICEServersBackend, expectedBackend)
	}
}
