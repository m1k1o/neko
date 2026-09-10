package proxy

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAgentForwardsThroughAuthenticatedSOCKS5Proxy(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "through socks5")
	}))
	defer origin.Close()

	socksListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer socksListener.Close()
	handshake := make(chan error, 1)
	go serveOneSOCKS5Connection(socksListener, "socks-user", "socks-password", handshake)

	agent, err := NewAgent(Config{
		Server:   "socks5://" + socksListener.Addr().String(),
		Username: "socks-user",
		Password: "socks-password",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.transport.CloseIdleConnections()
	local := httptest.NewServer(agent)
	defer local.Close()
	localURL, _ := url.Parse(local.URL)
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(localURL)}}
	response, err := client.Get(origin.URL)
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "through socks5" {
		t.Fatalf("unexpected SOCKS5 response %q", body)
	}
	select {
	case err := <-handshake:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for SOCKS5 authentication")
	}
}

func TestAgentForwardsThroughAuthenticatedHTTPProxy(t *testing.T) {
	const username = "proxy-user"
	const password = "proxy-password"

	echoListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer echoListener.Close()
	go func() {
		for {
			conn, err := echoListener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				_, _ = io.Copy(conn, conn)
			}()
		}
	}()

	var requests atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUsername, gotPassword, ok := parseProxyBasicAuth(r.Header.Get("Proxy-Authorization"))
		if !ok || gotUsername != username || gotPassword != password {
			http.Error(w, "proxy authentication required", http.StatusProxyAuthRequired)
			return
		}
		requests.Add(1)
		if r.Method != http.MethodConnect {
			w.Header().Set("X-Upstream-Proxy", "authenticated")
			_, _ = io.WriteString(w, r.URL.String())
			return
		}

		destination, err := net.DialTimeout("tcp", r.Host, time.Second)
		if err != nil {
			http.Error(w, "dial target", http.StatusBadGateway)
			return
		}
		hijacker := w.(http.Hijacker)
		client, buffered, err := hijacker.Hijack()
		if err != nil {
			destination.Close()
			return
		}
		defer client.Close()
		defer destination.Close()
		_, _ = buffered.WriteString("HTTP/1.1 200 Connection Established\r\n\r\n")
		_ = buffered.Flush()
		go func() {
			_, _ = io.Copy(destination, buffered.Reader)
			destination.Close()
		}()
		_, _ = io.Copy(client, destination)
	}))
	defer upstream.Close()

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatal(err)
	}
	agent, err := NewAgent(Config{
		Server:   "http://" + upstreamURL.Host,
		Username: username,
		Password: password,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.transport.CloseIdleConnections()
	local := httptest.NewServer(agent)
	defer local.Close()

	localURL, err := url.Parse(local.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(localURL)}}
	response, err := client.Get("http://origin.example.test/resource")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if response.Header.Get("X-Upstream-Proxy") != "authenticated" {
		t.Fatalf("request did not pass through authenticated upstream proxy")
	}
	if string(body) != "http://origin.example.test/resource" {
		t.Fatalf("unexpected upstream request URL %q", body)
	}

	proxyConn, err := net.DialTimeout("tcp", localURL.Host, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer proxyConn.Close()
	_ = proxyConn.SetDeadline(time.Now().Add(3 * time.Second))
	_, _ = fmt.Fprintf(proxyConn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", echoListener.Addr(), echoListener.Addr())
	reader := bufio.NewReader(proxyConn)
	connectResponse, err := http.ReadResponse(reader, &http.Request{Method: http.MethodConnect})
	if err != nil {
		t.Fatal(err)
	}
	if connectResponse.StatusCode != http.StatusOK {
		t.Fatalf("CONNECT status = %s", connectResponse.Status)
	}
	if _, err := io.WriteString(proxyConn, "tunnel payload\n"); err != nil {
		t.Fatal(err)
	}
	payload, err := reader.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if payload != "tunnel payload\n" {
		t.Fatalf("unexpected tunneled payload %q", payload)
	}
	if requests.Load() != 2 {
		t.Fatalf("authenticated upstream requests = %d, want 2", requests.Load())
	}
}

func TestListenAddressRejectsNonLoopback(t *testing.T) {
	for _, address := range []string{"127.0.0.1:18080", "localhost:18080", "[::1]:18080"} {
		if _, err := ListenAddress(address); err != nil {
			t.Errorf("ListenAddress(%q) returned %v", address, err)
		}
	}
	if _, err := ListenAddress("0.0.0.0:18080"); err == nil {
		t.Fatal("non-loopback listen address unexpectedly succeeded")
	}
}

func parseProxyBasicAuth(header string) (string, string, bool) {
	request := &http.Request{Header: http.Header{"Proxy-Authorization": []string{header}}}
	authorization := request.Header.Get("Proxy-Authorization")
	if !strings.HasPrefix(authorization, "Basic ") {
		return "", "", false
	}
	request.Header.Set("Authorization", authorization)
	return request.BasicAuth()
}

func serveOneSOCKS5Connection(listener net.Listener, username, password string, result chan<- error) {
	conn, err := listener.Accept()
	if err != nil {
		result <- err
		return
	}
	defer conn.Close()
	reader := bufio.NewReader(conn)

	header := make([]byte, 2)
	if _, err := io.ReadFull(reader, header); err != nil {
		result <- err
		return
	}
	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(reader, methods); err != nil {
		result <- err
		return
	}
	if header[0] != 5 || !strings.ContainsRune(string(methods), rune(2)) {
		result <- fmt.Errorf("SOCKS5 client did not offer username/password authentication")
		return
	}
	if _, err := conn.Write([]byte{5, 2}); err != nil {
		result <- err
		return
	}

	if _, err := io.ReadFull(reader, header); err != nil {
		result <- err
		return
	}
	user := make([]byte, int(header[1]))
	if _, err := io.ReadFull(reader, user); err != nil {
		result <- err
		return
	}
	length, err := reader.ReadByte()
	if err != nil {
		result <- err
		return
	}
	pass := make([]byte, int(length))
	if _, err := io.ReadFull(reader, pass); err != nil {
		result <- err
		return
	}
	if header[0] != 1 || string(user) != username || string(pass) != password {
		_, _ = conn.Write([]byte{1, 1})
		result <- fmt.Errorf("SOCKS5 credentials did not match")
		return
	}
	if _, err := conn.Write([]byte{1, 0}); err != nil {
		result <- err
		return
	}

	target, err := readSOCKS5Target(reader)
	if err != nil {
		result <- err
		return
	}
	destination, err := net.DialTimeout("tcp", target, time.Second)
	if err != nil {
		result <- err
		return
	}
	defer destination.Close()
	if _, err := conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}); err != nil {
		result <- err
		return
	}
	result <- nil
	go func() {
		_, _ = io.Copy(destination, reader)
		destination.Close()
	}()
	_, _ = io.Copy(conn, destination)
}

func readSOCKS5Target(reader *bufio.Reader) (string, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(reader, header); err != nil {
		return "", err
	}
	if header[0] != 5 || header[1] != 1 {
		return "", fmt.Errorf("unexpected SOCKS5 connect request")
	}
	var host string
	switch header[3] {
	case 1:
		address := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(reader, address); err != nil {
			return "", err
		}
		host = net.IP(address).String()
	case 3:
		length, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		address := make([]byte, int(length))
		if _, err := io.ReadFull(reader, address); err != nil {
			return "", err
		}
		host = string(address)
	case 4:
		address := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(reader, address); err != nil {
			return "", err
		}
		host = net.IP(address).String()
	default:
		return "", fmt.Errorf("unsupported SOCKS5 address type %d", header[3])
	}
	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(reader, portBytes); err != nil {
		return "", err
	}
	return net.JoinHostPort(host, fmt.Sprint(binary.BigEndian.Uint16(portBytes))), nil
}
