// Package proxy contains configuration primitives for Chromium's egress proxy.
// It does not start a proxy process; the runtime layer owns that lifecycle.
package proxy

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Protocol is the upstream proxy protocol supported by M1.
type Protocol string

const (
	ProtocolHTTPConnect Protocol = "http"
	ProtocolSOCKS5      Protocol = "socks5"
)

// Config describes an authenticated upstream proxy. Credentials are separate
// from Server so callers can source them from a secret and never include them
// in process arguments, profile files, or diagnostics.
type Config struct {
	Server   string
	Username string
	Password string
	Bypass   []string
}

// Endpoint is the parsed, credential-free upstream proxy address.
type Endpoint struct {
	Protocol Protocol
	Host     string
	Port     string
}

// Enabled reports whether an upstream proxy is configured.
func (c Config) Enabled() bool {
	return c.Server != ""
}

// Parse validates the configuration and returns a credential-free endpoint.
func (c Config) Parse() (Endpoint, error) {
	if !c.Enabled() {
		if c.Username != "" || c.Password != "" || len(c.Bypass) != 0 {
			return Endpoint{}, fmt.Errorf("proxy credentials or bypass rules require a proxy server")
		}
		return Endpoint{}, nil
	}
	if (c.Username == "") != (c.Password == "") {
		return Endpoint{}, fmt.Errorf("proxy username and password must be configured together")
	}

	u, err := url.Parse(c.Server)
	if err != nil {
		return Endpoint{}, fmt.Errorf("parse proxy server: %w", err)
	}
	if u.User != nil {
		return Endpoint{}, fmt.Errorf("proxy server must not embed credentials")
	}
	if u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return Endpoint{}, fmt.Errorf("proxy server must not contain a path, query, or fragment")
	}
	if u.Hostname() == "" || u.Port() == "" {
		return Endpoint{}, fmt.Errorf("proxy server must include a host and port")
	}
	if err := validateProxyPort(u.Port()); err != nil {
		return Endpoint{}, err
	}

	endpoint := Endpoint{Host: u.Hostname(), Port: u.Port()}
	switch Protocol(strings.ToLower(u.Scheme)) {
	case ProtocolHTTPConnect:
		endpoint.Protocol = ProtocolHTTPConnect
		if strings.Contains(c.Username, ":") {
			return Endpoint{}, fmt.Errorf("HTTP proxy username must not contain a colon")
		}
	case ProtocolSOCKS5:
		endpoint.Protocol = ProtocolSOCKS5
		if len(c.Username) > 255 || len(c.Password) > 255 {
			return Endpoint{}, fmt.Errorf("SOCKS5 username and password must not exceed 255 bytes")
		}
	default:
		return Endpoint{}, fmt.Errorf("unsupported proxy protocol %q", u.Scheme)
	}

	for _, rule := range c.Bypass {
		if strings.TrimSpace(rule) == "" || strings.ContainsAny(rule, ",\n\r") {
			return Endpoint{}, fmt.Errorf("invalid proxy bypass rule %q", rule)
		}
	}

	return endpoint, nil
}

func validateProxyPort(port string) error {
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		return fmt.Errorf("invalid proxy port %q", port)
	}
	return nil
}

// RedactedServer returns a safe value for logs and diagnostics.
func (e Endpoint) RedactedServer() string {
	if e.Protocol == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s", e.Protocol, net.JoinHostPort(e.Host, e.Port))
}

// ChromiumArguments returns arguments that point Chromium at a local proxy
// agent. Chromium never receives the upstream endpoint or its credentials.
func (c Config) ChromiumArguments(localAgent string) ([]string, error) {
	if !c.Enabled() {
		return nil, nil
	}
	if _, err := c.Parse(); err != nil {
		return nil, err
	}
	if _, err := ListenAddress(localAgent); err != nil {
		return nil, err
	}

	args := []string{"--proxy-server=http://" + localAgent}
	if len(c.Bypass) > 0 {
		args = append(args, "--proxy-bypass-list="+strings.Join(c.Bypass, ";"))
	}
	return args, nil
}
