package webrtc

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"

	"github.com/m1k1o/neko/server/internal/config"
	"github.com/m1k1o/neko/server/pkg/types/codec"
)

// fakeDNS answers lookups from a map, which can be changed between lookups
type fakeDNS map[string][]string

func (dns fakeDNS) lookup(ctx context.Context, network, host string) ([]netip.Addr, error) {
	if network != "ip" {
		return nil, errors.New("unexpected network " + network)
	}

	records, ok := dns[host]
	if !ok {
		return nil, errors.New("no such host")
	}

	addrs := []netip.Addr{}
	for _, record := range records {
		addrs = append(addrs, netip.MustParseAddr(record))
	}
	return addrs, nil
}

func newTestNAT1To1Resolver(dns fakeDNS, entries ...string) *nat1to1Resolver {
	resolver := newNAT1To1Resolver(zerolog.Nop(), entries)
	resolver.lookup = dns.lookup
	return resolver
}

func TestNAT1To1ResolverIPs(t *testing.T) {
	dns := fakeDNS{
		"neko.example.com":  {"203.0.113.20", "2001:db8::20", "203.0.113.10"},
		"other.example.com": {"203.0.113.10", "::ffff:203.0.113.30"},
	}

	tests := []struct {
		name     string
		entries  []string
		expected []string
	}{
		{
			name:     "no entries",
			entries:  nil,
			expected: []string{},
		},
		{
			name:     "addresses are kept",
			entries:  []string{"203.0.113.1", " 2001:db8::1 ", "203.0.113.2/10.0.0.2", ""},
			expected: []string{"203.0.113.1", "2001:db8::1", "203.0.113.2/10.0.0.2"},
		},
		{
			name:     "hostname is replaced by all of its addresses",
			entries:  []string{"neko.example.com"},
			expected: []string{"203.0.113.10", "203.0.113.20", "2001:db8::20"},
		},
		{
			name:     "hostname mapped to a local address uses the same family",
			entries:  []string{"neko.example.com/10.0.0.2", "neko.example.com/fd00::2"},
			expected: []string{"203.0.113.10/10.0.0.2", "203.0.113.20/10.0.0.2", "2001:db8::20/fd00::2"},
		},
		{
			name:     "duplicate addresses are removed",
			entries:  []string{"203.0.113.10", "neko.example.com", "other.example.com"},
			expected: []string{"203.0.113.10", "203.0.113.20", "2001:db8::20", "203.0.113.30"},
		},
		{
			name:     "unresolvable hostname is skipped",
			entries:  []string{"missing.example.com", "203.0.113.1"},
			expected: []string{"203.0.113.1"},
		},
		{
			name:     "invalid local address is left for pion to report",
			entries:  []string{"other.example.com/invalid"},
			expected: []string{"203.0.113.10/invalid", "203.0.113.30/invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := newTestNAT1To1Resolver(dns, tt.entries...)
			if ips := resolver.IPs(); !reflect.DeepEqual(ips, tt.expected) {
				t.Errorf("IPs() = %q, want %q", ips, tt.expected)
			}
		})
	}
}

func TestNAT1To1ResolverFollowsDNSChanges(t *testing.T) {
	dns := fakeDNS{"neko.example.com": {"203.0.113.10"}}
	resolver := newTestNAT1To1Resolver(dns, "neko.example.com")

	if ips := resolver.IPs(); !reflect.DeepEqual(ips, []string{"203.0.113.10"}) {
		t.Fatalf("IPs() = %q, want the initial address", ips)
	}

	// the dynamic DNS record changes
	dns["neko.example.com"] = []string{"203.0.113.11"}
	if ips := resolver.IPs(); !reflect.DeepEqual(ips, []string{"203.0.113.11"}) {
		t.Fatalf("IPs() = %q, want the changed address", ips)
	}

	// the DNS server is unreachable, the last addresses are used
	delete(dns, "neko.example.com")
	if ips := resolver.IPs(); !reflect.DeepEqual(ips, []string{"203.0.113.11"}) {
		t.Fatalf("IPs() = %q, want the last resolved address", ips)
	}

	// the record is back with another address
	dns["neko.example.com"] = []string{"203.0.113.12"}
	if ips := resolver.IPs(); !reflect.DeepEqual(ips, []string{"203.0.113.12"}) {
		t.Fatalf("IPs() = %q, want the new address", ips)
	}
}

func TestNAT1To1ResolverLookupTimeout(t *testing.T) {
	resolver := newNAT1To1Resolver(zerolog.Nop(), []string{"neko.example.com"})
	resolver.lookup = func(ctx context.Context, network, host string) ([]netip.Addr, error) {
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("lookup context has no deadline")
		}
		if timeout := time.Until(deadline); timeout <= 0 || timeout > nat1to1LookupTimeout {
			t.Errorf("lookup timeout = %v, want at most %v", timeout, nat1to1LookupTimeout)
		}
		return nil, context.DeadlineExceeded
	}

	if ips := resolver.IPs(); len(ips) != 0 {
		t.Errorf("IPs() = %q, want none", ips)
	}
}

// the host candidates of a new peer connection
// advertise all addresses of the configured hostname
func TestPeerConnectionNAT1To1Hostname(t *testing.T) {
	if !hasIPv4Interface(t) {
		t.Skip("no non-loopback IPv4 interface to gather host candidates from")
	}

	dns := fakeDNS{"neko.example.com": {"203.0.113.10", "203.0.113.11"}}

	manager := &WebRTCManagerCtx{
		logger: zerolog.Nop(),
		config: &config.WebRTC{
			EphemeralMin: 59000,
			EphemeralMax: 59100,
		},
		nat1to1: newTestNAT1To1Resolver(dns, "neko.example.com"),
	}

	gatherHostIPs := func() map[string]bool {
		t.Helper()

		connection, _, err := manager.newPeerConnection(zerolog.Nop(), []codec.RTPCodec{codec.Opus(), codec.VP8()})
		if err != nil {
			t.Fatalf("creating peer connection: %v", err)
		}
		defer connection.Close()

		if _, err := connection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
			t.Fatalf("adding transceiver: %v", err)
		}

		offer, err := connection.CreateOffer(nil)
		if err != nil {
			t.Fatalf("creating offer: %v", err)
		}

		gatherComplete := webrtc.GatheringCompletePromise(connection)
		if err := connection.SetLocalDescription(offer); err != nil {
			t.Fatalf("setting local description: %v", err)
		}

		select {
		case <-gatherComplete:
		case <-time.After(10 * time.Second):
			t.Fatal("timed out gathering candidates")
		}

		ips := map[string]bool{}
		for line := range strings.SplitSeq(connection.LocalDescription().SDP, "\n") {
			attr, ok := strings.CutPrefix(strings.TrimSpace(line), "a=candidate:")
			if !ok {
				continue
			}

			candidate, err := ice.UnmarshalCandidate(attr)
			if err != nil {
				t.Fatalf("parsing candidate %q: %v", attr, err)
			}

			if candidate.Type() == ice.CandidateTypeHost && candidate.NetworkType().IsIPv4() {
				ips[candidate.Address()] = true
			}
		}
		return ips
	}

	expected := map[string]bool{"203.0.113.10": true, "203.0.113.11": true}
	if ips := gatherHostIPs(); !reflect.DeepEqual(ips, expected) {
		t.Fatalf("IPv4 host candidates = %v, want %v", ips, expected)
	}

	// a new peer connection uses the changed DNS record
	dns["neko.example.com"] = []string{"203.0.113.12"}

	expected = map[string]bool{"203.0.113.12": true}
	if ips := gatherHostIPs(); !reflect.DeepEqual(ips, expected) {
		t.Fatalf("IPv4 host candidates after DNS change = %v, want %v", ips, expected)
	}
}

func hasIPv4Interface(t *testing.T) bool {
	t.Helper()

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Fatalf("listing interface addresses: %v", err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return true
		}
	}
	return false
}
