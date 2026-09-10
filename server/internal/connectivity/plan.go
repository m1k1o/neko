// Package connectivity defines deployment-neutral network plans for media
// workers. It deliberately has no dependency on Pion, Docker, or a tunnel
// implementation so configuration can be validated before a worker starts.
package connectivity

import (
	"fmt"
	"net/netip"
)

// Mode describes how a media worker is exposed to WebRTC clients.
type Mode string

const (
	// ModeDirect exposes the worker's MUX ports directly from the host.
	ModeDirect Mode = "direct"
	// ModeFRP exposes the worker through a TCP and UDP tunnel on one public port.
	ModeFRP Mode = "frp"
)

// MediaPortPlan is the public MUX contract for a room worker. The same port
// number can safely be used for TCP and UDP because they are distinct
// transports.
type MediaPortPlan struct {
	Mode       Mode
	UDPMuxPort int
	TCPMuxPort int
	NAT1To1IP  string
}

// Validate ensures a plan can be represented by Neko's TCP/UDP MUX settings.
// It intentionally does not probe reachability: client-side NAT and firewall
// policies are runtime conditions, not worker startup failures.
func (p MediaPortPlan) Validate() error {
	if p.Mode == "" {
		return fmt.Errorf("connectivity mode is required")
	}
	if p.Mode != ModeDirect && p.Mode != ModeFRP {
		return fmt.Errorf("unsupported connectivity mode %q", p.Mode)
	}
	if err := validatePort("udp mux", p.UDPMuxPort); err != nil {
		return err
	}
	if err := validatePort("tcp mux", p.TCPMuxPort); err != nil {
		return err
	}
	if p.UDPMuxPort == 0 && p.TCPMuxPort == 0 {
		return fmt.Errorf("at least one media mux port is required")
	}

	if p.NAT1To1IP != "" {
		if _, err := netip.ParseAddr(p.NAT1To1IP); err != nil {
			return fmt.Errorf("invalid nat 1:1 IP %q: %w", p.NAT1To1IP, err)
		}
	}

	if p.Mode == ModeFRP {
		if p.UDPMuxPort == 0 || p.TCPMuxPort == 0 {
			return fmt.Errorf("frp mode requires both UDP and TCP mux ports")
		}
		if p.UDPMuxPort != p.TCPMuxPort {
			return fmt.Errorf("frp mode requires matching UDP and TCP mux ports")
		}
		if p.NAT1To1IP == "" {
			return fmt.Errorf("frp mode requires the tunnel public NAT 1:1 IP")
		}
	}

	return nil
}

func validatePort(name string, port int) error {
	if port < 0 || port > 65535 {
		return fmt.Errorf("%s port %d is outside the valid range", name, port)
	}
	return nil
}
