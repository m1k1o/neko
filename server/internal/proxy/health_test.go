package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthMonitorReportsRedactedStatus(t *testing.T) {
	agent := &Agent{}
	agent.dialTunnel = func(context.Context, string) (net.Conn, error) {
		client, server := net.Pipe()
		go server.Close()
		return client, nil
	}
	endpoint := Endpoint{Protocol: ProtocolHTTPConnect, Host: "proxy.example.test", Port: "8080"}
	monitor, err := NewHealthMonitor(agent, endpoint, "target.example.test:443")
	if err != nil {
		t.Fatal(err)
	}
	if err := monitor.Check(context.Background()); err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	monitor.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("healthy status code = %d", response.Code)
	}
	var snapshot HealthSnapshot
	if err := json.Unmarshal(response.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Status != "healthy" || snapshot.Upstream != "http://proxy.example.test:8080" || snapshot.Reason != "" {
		t.Fatalf("unexpected healthy snapshot: %#v", snapshot)
	}
	if strings.Contains(response.Body.String(), "secret") {
		t.Fatal("health response contains proxy credentials")
	}

	agent.dialTunnel = func(context.Context, string) (net.Conn, error) {
		return nil, &CheckError{Reason: HealthAuthenticationFailed, Err: errors.New("secret upstream response")}
	}
	if err := monitor.Check(context.Background()); err == nil {
		t.Fatal("failed health check unexpectedly succeeded")
	}
	response = httptest.NewRecorder()
	monitor.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("unhealthy status code = %d", response.Code)
	}
	if strings.Contains(response.Body.String(), "secret") {
		t.Fatal("unhealthy response contains raw error details")
	}
	if err := json.Unmarshal(response.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Reason != HealthAuthenticationFailed {
		t.Fatalf("health reason = %q", snapshot.Reason)
	}
}

func TestHealthMonitorRejectsInvalidRequests(t *testing.T) {
	monitor, err := NewHealthMonitor(&Agent{}, Endpoint{Protocol: ProtocolSOCKS5}, "target.example.test:443")
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	monitor.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/healthz", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status code = %d", response.Code)
	}
	response = httptest.NewRecorder()
	monitor.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/unknown", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("unknown path status code = %d", response.Code)
	}
}

func TestClassifySOCKS5HealthErrors(t *testing.T) {
	tests := []struct {
		err  error
		want HealthReason
	}{
		{err: errors.New("username/password authentication failed"), want: HealthAuthenticationFailed},
		{err: errors.New("no acceptable authentication methods"), want: HealthAuthenticationFailed},
		{err: errors.New("unknown error host unreachable"), want: HealthTargetRejected},
		{err: &net.OpError{Op: "dial", Err: errors.New("connection refused")}, want: HealthUpstreamUnreachable},
		{err: errors.New("unexpected protocol response"), want: HealthCheckFailed},
	}
	for _, test := range tests {
		if got := classifyCheckError(test.err); got != test.want {
			t.Errorf("classifyCheckError(%q) = %q, want %q", test.err, got, test.want)
		}
	}
}

func TestTargetAddressValidation(t *testing.T) {
	if _, err := TargetAddress("example.test:443"); err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"", "example.test", "example.test:0", "example.test:65536"} {
		if _, err := TargetAddress(target); err == nil {
			t.Errorf("invalid target %q unexpectedly succeeded", target)
		}
	}
}
