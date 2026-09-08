package proxy

import (
	"reflect"
	"testing"
)

func TestConfigParse(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		want         Endpoint
		wantErr      bool
		wantRedacted string
	}{
		{
			name:         "HTTP CONNECT with credentials",
			config:       Config{Server: "http://proxy.example.test:8080", Username: "user", Password: "secret"},
			want:         Endpoint{Protocol: ProtocolHTTPConnect, Host: "proxy.example.test", Port: "8080"},
			wantRedacted: "http://proxy.example.test:8080",
		},
		{
			name:         "SOCKS5 with credentials",
			config:       Config{Server: "socks5://[2001:db8::1]:1080", Username: "user", Password: "secret"},
			want:         Endpoint{Protocol: ProtocolSOCKS5, Host: "2001:db8::1", Port: "1080"},
			wantRedacted: "socks5://[2001:db8::1]:1080",
		},
		{name: "credentials without server", config: Config{Username: "user", Password: "secret"}, wantErr: true},
		{name: "partial credentials", config: Config{Server: "http://proxy.example.test:8080", Username: "user"}, wantErr: true},
		{name: "credentials in URL", config: Config{Server: "http://user:secret@proxy.example.test:8080"}, wantErr: true},
		{name: "missing port", config: Config{Server: "http://proxy.example.test"}, wantErr: true},
		{name: "unsupported protocol", config: Config{Server: "https://proxy.example.test:443"}, wantErr: true},
		{name: "invalid bypass", config: Config{Server: "http://proxy.example.test:8080", Bypass: []string{"localhost,example.test"}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.Parse()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse() = %#v, want %#v", got, tt.want)
			}
			if got.RedactedServer() != tt.wantRedacted {
				t.Fatalf("RedactedServer() = %q, want %q", got.RedactedServer(), tt.wantRedacted)
			}
		})
	}
}

func TestChromiumArguments(t *testing.T) {
	config := Config{
		Server:   "socks5://proxy.example.test:1080",
		Username: "user",
		Password: "secret",
		Bypass:   []string{"localhost", "127.0.0.1", "*.internal"},
	}

	got, err := config.ChromiumArguments("127.0.0.1:18080")
	if err != nil {
		t.Fatalf("ChromiumArguments() error = %v", err)
	}
	want := []string{
		"--proxy-server=http://127.0.0.1:18080",
		"--proxy-bypass-list=localhost;127.0.0.1;*.internal",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChromiumArguments() = %#v, want %#v", got, want)
	}
}
