package proxy

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	xproxy "golang.org/x/net/proxy"
)

const dialTimeout = 15 * time.Second

// Agent is a credential-hiding HTTP proxy for Chromium. It forwards requests
// through one authenticated HTTP CONNECT or SOCKS5 upstream proxy.
type Agent struct {
	transport  *http.Transport
	dialTunnel func(context.Context, string) (net.Conn, error)
}

// NewAgent constructs an Agent without opening a listener.
func NewAgent(config Config) (*Agent, error) {
	endpoint, err := config.Parse()
	if err != nil {
		return nil, err
	}
	if endpoint.Protocol == "" {
		return nil, errors.New("proxy server is required")
	}

	agent := &Agent{}
	switch endpoint.Protocol {
	case ProtocolHTTPConnect:
		proxyURL := &url.URL{
			Scheme: string(endpoint.Protocol),
			Host:   net.JoinHostPort(endpoint.Host, endpoint.Port),
		}
		if config.Username != "" {
			proxyURL.User = url.UserPassword(config.Username, config.Password)
		}
		agent.transport = &http.Transport{
			Proxy:                 http.ProxyURL(proxyURL),
			ProxyConnectHeader:    basicAuthHeader(config.Username, config.Password),
			DialContext:           (&net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second}).DialContext,
			ForceAttemptHTTP2:     false,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: time.Second,
		}
		agent.dialTunnel = func(ctx context.Context, target string) (net.Conn, error) {
			return dialHTTPConnect(ctx, endpoint, config.Username, config.Password, target)
		}

	case ProtocolSOCKS5:
		var auth *xproxy.Auth
		if config.Username != "" {
			auth = &xproxy.Auth{User: config.Username, Password: config.Password}
		}
		dialer, err := xproxy.SOCKS5("tcp", net.JoinHostPort(endpoint.Host, endpoint.Port), auth, &net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second})
		if err != nil {
			return nil, fmt.Errorf("create SOCKS5 dialer: %w", err)
		}
		contextDialer, ok := dialer.(xproxy.ContextDialer)
		if !ok {
			return nil, errors.New("SOCKS5 dialer does not support contexts")
		}
		agent.transport = &http.Transport{
			DialContext:           contextDialer.DialContext,
			ForceAttemptHTTP2:     false,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: time.Second,
		}
		agent.dialTunnel = func(ctx context.Context, target string) (net.Conn, error) {
			return contextDialer.DialContext(ctx, "tcp", target)
		}
	}

	return agent, nil
}

func (a *Agent) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		a.serveTunnel(w, r)
		return
	}
	a.serveHTTP(w, r)
}

func (a *Agent) serveHTTP(w http.ResponseWriter, r *http.Request) {
	request := r.Clone(r.Context())
	request.RequestURI = ""
	request.Header = r.Header.Clone()
	removeHopHeaders(request.Header)

	response, err := a.transport.RoundTrip(request)
	if err != nil {
		http.Error(w, "upstream proxy request failed", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	removeHopHeaders(response.Header)
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func (a *Agent) serveTunnel(w http.ResponseWriter, r *http.Request) {
	if err := validateTarget(r.Host); err != nil {
		http.Error(w, "invalid CONNECT target", http.StatusBadRequest)
		return
	}

	upstream, err := a.dialTunnel(r.Context(), r.Host)
	if err != nil {
		http.Error(w, "upstream proxy connection failed", http.StatusBadGateway)
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		upstream.Close()
		http.Error(w, "connection hijacking is unavailable", http.StatusInternalServerError)
		return
	}
	client, buffered, err := hijacker.Hijack()
	if err != nil {
		upstream.Close()
		return
	}
	if _, err := buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n"); err != nil {
		client.Close()
		upstream.Close()
		return
	}
	if err := buffered.Flush(); err != nil {
		client.Close()
		upstream.Close()
		return
	}

	tunnel(client, buffered.Reader, upstream)
}

func dialHTTPConnect(ctx context.Context, endpoint Endpoint, username, password, target string) (net.Conn, error) {
	conn, err := (&net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second}).DialContext(ctx, "tcp", net.JoinHostPort(endpoint.Host, endpoint.Port))
	if err != nil {
		return nil, &CheckError{Reason: HealthUpstreamUnreachable, Err: err}
	}

	request := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: target},
		Host:   target,
		Header: make(http.Header),
	}
	for key, values := range basicAuthHeader(username, password) {
		request.Header[key] = values
	}
	if err := request.Write(conn); err != nil {
		conn.Close()
		return nil, &CheckError{Reason: HealthCheckFailed, Err: err}
	}

	reader := bufio.NewReader(conn)
	response, err := http.ReadResponse(reader, request)
	if err != nil {
		conn.Close()
		return nil, &CheckError{Reason: HealthCheckFailed, Err: err}
	}
	if response.StatusCode != http.StatusOK {
		if response.Body != nil {
			response.Body.Close()
		}
		conn.Close()
		reason := HealthTargetRejected
		if response.StatusCode == http.StatusProxyAuthRequired {
			reason = HealthAuthenticationFailed
		}
		return nil, &CheckError{Reason: reason, Err: fmt.Errorf("upstream HTTP proxy returned %s", response.Status)}
	}
	return &bufferedConn{Conn: conn, reader: reader}, nil
}

func basicAuthHeader(username, password string) http.Header {
	header := make(http.Header)
	if username != "" {
		credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		header.Set("Proxy-Authorization", "Basic "+credentials)
	}
	return header
}

func validateTarget(target string) error {
	host, port, err := net.SplitHostPort(target)
	if err != nil || host == "" || port == "" {
		return errors.New("target must contain a host and port")
	}
	if err := validateProxyPort(port); err != nil {
		return errors.New("target contains an invalid port")
	}
	return nil
}

// TargetAddress validates an explicit proxy health-check destination.
func TargetAddress(target string) (string, error) {
	if err := validateTarget(target); err != nil {
		return "", fmt.Errorf("invalid proxy health-check target %q: %w", target, err)
	}
	return target, nil
}

func removeHopHeaders(header http.Header) {
	for _, connection := range header.Values("Connection") {
		for _, key := range strings.Split(connection, ",") {
			header.Del(strings.TrimSpace(key))
		}
	}
	for _, key := range []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"} {
		header.Del(key)
	}
}

func tunnel(client net.Conn, clientReader io.Reader, upstream net.Conn) {
	done := make(chan struct{}, 2)
	copyOneWay := func(destination net.Conn, source io.Reader) {
		_, _ = io.Copy(destination, source)
		if halfCloser, ok := destination.(interface{ CloseWrite() error }); ok {
			_ = halfCloser.CloseWrite()
		}
		done <- struct{}{}
	}
	go copyOneWay(upstream, clientReader)
	go copyOneWay(client, upstream)
	<-done
	<-done
	client.Close()
	upstream.Close()
}

type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

func (c *bufferedConn) CloseWrite() error {
	if halfCloser, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return halfCloser.CloseWrite()
	}
	return nil
}

// ListenAddress validates that the agent cannot be exposed beyond loopback.
func ListenAddress(address string) (string, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("invalid proxy agent listen address %q: %w", address, err)
	}
	if err := validateProxyPort(port); err != nil {
		return "", fmt.Errorf("invalid proxy agent listen port %q", port)
	}
	if host != "localhost" {
		ip := net.ParseIP(strings.Trim(host, "[]"))
		if ip == nil || !ip.IsLoopback() {
			return "", errors.New("proxy agent must listen on a loopback address")
		}
	}
	return address, nil
}
