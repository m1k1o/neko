package webrtc

import (
	"context"
	"net"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// how long to wait for a NAT 1-to-1 hostname to be resolved
const nat1to1LookupTimeout = 3 * time.Second

// nat1to1Resolver turns the configured NAT 1-to-1 entries into the IP
// addresses passed to pion. An entry is an external address, optionally
// followed by a slash and the local IP address that it maps to. The external
// address may be a hostname. Hostnames are resolved every time the addresses
// are requested, so that a changed DNS record, e.g. of a dynamic DNS name,
// applies to new peer connections without restarting neko.
type nat1to1Resolver struct {
	logger  zerolog.Logger
	entries []string
	lookup  func(ctx context.Context, network, host string) ([]netip.Addr, error)

	mu   sync.Mutex
	last map[string][]netip.Addr // last successful lookup of each hostname
}

func newNAT1To1Resolver(logger zerolog.Logger, entries []string) *nat1to1Resolver {
	return &nat1to1Resolver{
		logger:  logger,
		entries: entries,
		lookup:  net.DefaultResolver.LookupNetIP,
		last:    map[string][]netip.Addr{},
	}
}

// IPs returns the NAT 1-to-1 entries with every hostname replaced by all of
// its addresses. When a mapping to a local IP address is given, only the
// addresses of the same family are used. Entries with an IP address are
// passed through as they are, and pion reports those that are invalid.
func (r *nat1to1Resolver) IPs() []string {
	ips := []string{}
	add := func(ip string) {
		if !slices.Contains(ips, ip) {
			ips = append(ips, ip)
		}
	}

	for _, entry := range r.entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		external, local, hasLocal := strings.Cut(entry, "/")
		if _, err := netip.ParseAddr(external); err == nil {
			add(entry)
			continue
		}

		localAddr, localErr := netip.ParseAddr(local)
		for _, addr := range r.resolve(external) {
			if !hasLocal {
				add(addr.String())
				continue
			}

			// a local address can only be reached through an external address of its family
			if localErr == nil && addr.Is4() != localAddr.Unmap().Is4() {
				continue
			}

			add(addr.String() + "/" + local)
		}
	}

	return ips
}

// resolve looks up the addresses of the hostname. When the lookup fails, the
// addresses from the last successful lookup are returned, if there are any.
func (r *nat1to1Resolver) resolve(host string) []netip.Addr {
	ctx, cancel := context.WithTimeout(context.Background(), nat1to1LookupTimeout)
	defer cancel()

	addrs, err := r.lookup(ctx, "ip", host)

	r.mu.Lock()
	defer r.mu.Unlock()

	last, ok := r.last[host]
	if err != nil {
		if ok {
			r.logger.Warn().Err(err).Str("hostname", host).
				Strs("ips", addrsToStrings(last)).
				Msg("unable to resolve NAT 1-to-1 hostname, using the last resolved addresses")
		} else {
			r.logger.Warn().Err(err).Str("hostname", host).
				Msg("unable to resolve NAT 1-to-1 hostname, skipping it")
		}
		return last
	}

	// normalize the addresses, so that the order of the
	// records does not matter when looking for a change
	for i, addr := range addrs {
		addrs[i] = addr.Unmap()
	}
	slices.SortFunc(addrs, netip.Addr.Compare)
	addrs = slices.Compact(addrs)

	if !ok || !slices.Equal(addrs, last) {
		r.logger.Info().Str("hostname", host).
			Strs("ips", addrsToStrings(addrs)).
			Msg("resolved NAT 1-to-1 hostname")
	}

	r.last[host] = addrs
	return addrs
}

func addrsToStrings(addrs []netip.Addr) []string {
	strs := make([]string, len(addrs))
	for i, addr := range addrs {
		strs[i] = addr.String()
	}
	return strs
}
