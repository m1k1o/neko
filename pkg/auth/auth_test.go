package auth

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/demodesk/neko/internal/config"
	"github.com/demodesk/neko/internal/session"
	"github.com/demodesk/neko/pkg/types"
)

var i = 0
var sessionManager = session.New(&config.Session{})

func rWithSession(profile types.MemberProfile) (*http.Request, types.Session, error) {
	i++
	r := &http.Request{}
	session, _, err := sessionManager.Create(fmt.Sprintf("id-%d", i), profile)
	ctx := SetSession(r, session)
	r = r.WithContext(ctx)
	return r, session, err
}

func TestSessionCtx(t *testing.T) {
	r, session, err := rWithSession(types.MemberProfile{})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	sess, ok := GetSession(r)
	if !ok {
		t.Errorf("session not found")
		return
	}

	if !reflect.DeepEqual(sess, session) {
		t.Errorf("sessions not equal")
		return
	}
}

func TestAdminsOnly(t *testing.T) {
	r1, _, err := rWithSession(types.MemberProfile{IsAdmin: false})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	r2, _, err := rWithSession(types.MemberProfile{IsAdmin: true})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	tests := []struct {
		name    string
		r       *http.Request
		wantErr bool
	}{
		{
			name:    "is not admin",
			r:       r1,
			wantErr: true,
		},
		{
			name:    "is admin",
			r:       r2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := AdminsOnly(nil, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("AdminsOnly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestHostsOnly(t *testing.T) {
	r1, _, err := rWithSession(types.MemberProfile{CanHost: true})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	r2, session, err := rWithSession(types.MemberProfile{CanHost: true})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	// r2 is host
	session.SetAsHost()

	r3, _, err := rWithSession(types.MemberProfile{CanHost: false})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	tests := []struct {
		name    string
		r       *http.Request
		wantErr bool
	}{
		{
			name:    "is not hosting",
			r:       r1,
			wantErr: true,
		},
		{
			name:    "is hosting",
			r:       r2,
			wantErr: false,
		},
		{
			name:    "cannot host",
			r:       r3,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HostsOnly(nil, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("HostsOnly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCanWatchOnly(t *testing.T) {
	r1, _, err := rWithSession(types.MemberProfile{CanWatch: false})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	r2, _, err := rWithSession(types.MemberProfile{CanWatch: true})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	tests := []struct {
		name    string
		r       *http.Request
		wantErr bool
	}{
		{
			name:    "can not watch",
			r:       r1,
			wantErr: true,
		},
		{
			name:    "can watch",
			r:       r2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CanWatchOnly(nil, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanWatchOnly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCanHostOnly(t *testing.T) {
	r1, _, err := rWithSession(types.MemberProfile{CanHost: false})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	r2, _, err := rWithSession(types.MemberProfile{CanHost: true})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	tests := []struct {
		name        string
		r           *http.Request
		wantErr     bool
		privateMode bool
	}{
		{
			name:    "can not host",
			r:       r1,
			wantErr: true,
		},
		{
			name:    "can host",
			r:       r2,
			wantErr: false,
		},
		{
			name:        "private mode enabled: can not host",
			r:           r1,
			wantErr:     true,
			privateMode: true,
		},
		{
			name:        "private mode enabled: can host",
			r:           r2,
			wantErr:     true,
			privateMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, _ := GetSession(tt.r)
			sessionManager.UpdateSettingsFunc(session, func(s *types.Settings) bool {
				s.PrivateMode = tt.privateMode
				return true
			})

			_, err := CanHostOnly(nil, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanHostOnly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestCanAccessClipboardOnly(t *testing.T) {
	r1, _, err := rWithSession(types.MemberProfile{CanAccessClipboard: false})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	r2, _, err := rWithSession(types.MemberProfile{CanAccessClipboard: true})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	tests := []struct {
		name    string
		r       *http.Request
		wantErr bool
	}{
		{
			name:    "can not access clipboard",
			r:       r1,
			wantErr: true,
		},
		{
			name:    "can access clipboard",
			r:       r2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CanAccessClipboardOnly(nil, tt.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanAccessClipboardOnly() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestPluginsGenericOnly(t *testing.T) {
	r1, _, err := rWithSession(types.MemberProfile{
		Plugins: map[string]any{
			"foo.bar": 1,
		},
	})
	if err != nil {
		t.Errorf("could not create session %s", err.Error())
		return
	}

	t.Run("test if exists", func(t *testing.T) {
		key := "foo.bar"
		val := 1
		wantErr := false

		handler := PluginsGenericOnly(key, val)
		_, err := handler(nil, r1)
		if (err != nil) != wantErr {
			t.Errorf("PluginsGenericOnly(%q, %v) error = %v, wantErr %v", key, val, err, wantErr)
			return
		}
	})

	t.Run("test when gets different value", func(t *testing.T) {
		key := "foo.bar"
		val := 2
		wantErr := true

		handler := PluginsGenericOnly(key, val)
		_, err := handler(nil, r1)
		if (err != nil) != wantErr {
			t.Errorf("PluginsGenericOnly(%q, %v) error = %v, wantErr %v", key, val, err, wantErr)
			return
		}
	})

	t.Run("test when gets different type", func(t *testing.T) {
		key := "foo.bar"
		val := "1"
		wantErr := true

		handler := PluginsGenericOnly(key, val)
		_, err := handler(nil, r1)
		if (err != nil) != wantErr {
			t.Errorf("PluginsGenericOnly(%q, %v) error = %v, wantErr %v", key, val, err, wantErr)
			return
		}
	})

	t.Run("test if does not exists", func(t *testing.T) {
		key := "foo.bar_not_extist"
		val := 1
		wantErr := true

		handler := PluginsGenericOnly(key, val)
		_, err := handler(nil, r1)
		if (err != nil) != wantErr {
			t.Errorf("PluginsGenericOnly(%q, %v) error = %v, wantErr %v", key, val, err, wantErr)
			return
		}
	})

	t.Run("test if session does not exists", func(t *testing.T) {
		key := "foo.bar_not_extist"
		val := 1
		wantErr := true

		handler := PluginsGenericOnly(key, val)
		_, err := handler(nil, &http.Request{})
		if (err != nil) != wantErr {
			t.Errorf("PluginsGenericOnly(%q, %v) error = %v, wantErr %v", key, val, err, wantErr)
			return
		}
	})
}
