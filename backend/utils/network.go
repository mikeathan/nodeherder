package utils

import (
	"net/http"
	"net/netip"
	"os"
	"strings"
)

// IsLocalRequest reports whether the client can use the existing local sign-in
// flow. Forwarded requests require an explicitly trusted immediate proxy; a
// proxy's own private address never substitutes for a missing client address.
func IsLocalRequest(r *http.Request) bool {
	peer, ok := requestIP(r.RemoteAddr)
	if !ok {
		return false
	}
	trusted, ok := trustedProxyPrefixes(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if !ok {
		return false
	}
	if !containsIP(trusted, peer) {
		// Fail closed for unconfigured private proxies too: merely ignoring
		// headers would mistake their private peer address for a local client.
		if len(r.Header.Values("X-Real-IP")) > 0 || len(r.Header.Values("X-Forwarded-For")) > 0 || len(r.Header.Values("Forwarded")) > 0 {
			return false
		}
		return isLocalIP(peer)
	}

	if realIP := r.Header.Values("X-Real-IP"); len(realIP) > 0 {
		if len(realIP) != 1 {
			return false
		}
		client, valid := requestIP(realIP[0])
		return valid && isLocalIP(client)
	}

	chain := strings.Join(r.Header.Values("X-Forwarded-For"), ",")
	if len(chain) == 0 || len(chain) > 4096 {
		return false
	}
	hops := strings.Split(chain, ",")
	if len(hops) > 32 {
		return false
	}
	addresses := make([]netip.Addr, len(hops))
	for i, hop := range hops {
		addresses[i], ok = requestIP(hop)
		if !ok {
			return false
		}
	}
	// A caller can prepend arbitrary values. Only skip proxies we explicitly
	// trust, walking from the socket peer toward the original client.
	for i := len(addresses) - 1; i >= 0; i-- {
		if i == 0 || !containsIP(trusted, addresses[i]) {
			return isLocalIP(addresses[i])
		}
	}
	return false
}

func requestIP(raw string) (netip.Addr, bool) {
	raw = strings.TrimSpace(raw)
	ip, err := netip.ParseAddr(raw)
	if err != nil {
		endpoint, endpointErr := netip.ParseAddrPort(raw)
		if endpointErr != nil {
			return netip.Addr{}, false
		}
		ip = endpoint.Addr()
	}
	if ip.Zone() != "" {
		return netip.Addr{}, false
	}
	return ip.Unmap(), true
}

func trustedProxyPrefixes(raw string) ([]netip.Prefix, bool) {
	if strings.TrimSpace(raw) == "" {
		return nil, true
	}
	entries := strings.Split(raw, ",")
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		prefix, err := netip.ParsePrefix(entry)
		if err != nil {
			ip, parseErr := netip.ParseAddr(entry)
			if parseErr != nil || ip.Zone() != "" {
				return nil, false
			}
			ip = ip.Unmap()
			prefix = netip.PrefixFrom(ip, ip.BitLen())
		}
		if prefix.Addr().Is4In6() {
			if prefix.Bits() < 96 {
				return nil, false
			}
			prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes, true
}

func containsIP(prefixes []netip.Prefix, ip netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(ip) {
			return true
		}
	}
	return false
}

func isLocalIP(ip netip.Addr) bool {
	return ip.IsLoopback() || ip.IsPrivate() || (ip.Is6() && ip.IsLinkLocalUnicast())
}
