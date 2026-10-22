package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsLocalRequest(t *testing.T) {
	for _, tc := range []struct {
		name, peer, trusted, realIP, forwarded string
		local                                  bool
	}{
		{name: "direct LAN", peer: "192.168.1.2:1234", local: true},
		{name: "direct loopback", peer: "127.0.0.1:1234", local: true},
		{name: "direct IPv6 loopback", peer: "[::1]:1234", local: true},
		{name: "direct IPv6 private", peer: "[fd00::1]:1234", local: true},
		{name: "direct IPv6 link local", peer: "[fe80::1]:1234", local: true},
		{name: "mapped IPv4", peer: "[::ffff:192.168.1.2]:1234", local: true},
		{name: "public", peer: "203.0.113.10:1234"},
		{name: "invalid peer", peer: "not-an-address"},
		{name: "spoof real IP", peer: "203.0.113.10:1234", realIP: "127.0.0.1"},
		{name: "spoof forwarded IP", peer: "203.0.113.10:1234", forwarded: "192.168.1.2"},
		{name: "unconfigured private proxy", peer: "172.19.0.2:1234", realIP: "203.0.113.10"},
		{name: "unconfigured proxy cannot claim local", peer: "172.19.0.2:1234", realIP: "192.168.1.2"},
		{name: "trusted proxy local client", peer: "172.19.0.2:1234", trusted: "172.19.0.2/32", realIP: "192.168.1.2", local: true},
		{name: "trusted proxy public client", peer: "172.19.0.2:1234", trusted: "172.19.0.2", realIP: "203.0.113.10"},
		{name: "trusted proxy missing client", peer: "172.19.0.2:1234", trusted: "172.19.0.2"},
		{name: "trust is restricted to configured address", peer: "172.19.0.3:1234", trusted: "172.19.0.2", realIP: "127.0.0.1"},
		{name: "trusted IPv6 proxy", peer: "[fd00::2]:1234", trusted: "fd00::2/128", realIP: "fd01::1", local: true},
		{name: "mapped proxy address", peer: "[::ffff:172.19.0.2]:1234", trusted: "::ffff:172.19.0.2/128", realIP: "192.168.1.2", local: true},
		{name: "CIDR contains peer", peer: "172.19.0.3:1234", trusted: "172.19.0.2/30", realIP: "192.168.1.2", local: true},
		{name: "CIDR excludes peer", peer: "172.19.0.4:1234", trusted: "172.19.0.2/30", realIP: "192.168.1.2"},
		{name: "allowlist", peer: "172.19.0.2:1234", trusted: "127.0.0.1, 172.19.0.2/32", realIP: "192.168.1.2", local: true},
		{name: "invalid trust entry", peer: "172.19.0.2:1234", trusted: "172.19.0.2,typo", realIP: "192.168.1.2"},
		{name: "invalid config fails closed for direct local", peer: "127.0.0.1:1234", trusted: "typo"},
		{name: "empty trust entry fails closed", peer: "172.19.0.2:1234", trusted: "172.19.0.2,", realIP: "192.168.1.2"},
		{name: "real IP is authoritative", peer: "172.19.0.2:1234", trusted: "172.19.0.2", realIP: "203.0.113.10", forwarded: "127.0.0.1"},
		{name: "malformed real IP cannot fall back", peer: "172.19.0.2:1234", trusted: "172.19.0.2", realIP: "unknown", forwarded: "127.0.0.1"},
		{name: "real IP list rejected", peer: "172.19.0.2:1234", trusted: "172.19.0.2", realIP: "127.0.0.1, 203.0.113.10"},
		{name: "invalid real IP port rejected", peer: "172.19.0.2:1234", trusted: "172.19.0.2", realIP: "127.0.0.1:invalid"},
		{name: "forwarded local client", peer: "172.19.0.2:1234", trusted: "172.19.0.2", forwarded: "192.168.1.2", local: true},
		{name: "forged leftmost address", peer: "172.19.0.2:1234", trusted: "172.19.0.2", forwarded: "127.0.0.1, 203.0.113.10"},
		{name: "stop at untrusted proxy", peer: "172.19.0.2:1234", trusted: "172.19.0.2", forwarded: "127.0.0.1, 203.0.113.11, 172.19.0.2"},
		{name: "multiple trusted proxies", peer: "172.19.0.2:1234", trusted: "172.19.0.2,172.19.0.3", forwarded: "192.168.1.2, 172.19.0.3", local: true},
		{name: "invalid chain", peer: "172.19.0.2:1234", trusted: "172.19.0.2", forwarded: "unknown, 192.168.1.2"},
		{name: "empty chain hop", peer: "172.19.0.2:1234", trusted: "172.19.0.2", forwarded: ", 192.168.1.2"},
		{name: "too many hops", peer: "172.19.0.2:1234", trusted: "172.19.0.2", forwarded: strings.Repeat("172.19.0.2,", 32) + "127.0.0.1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TRUSTED_PROXY_CIDRS", tc.trusted)
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tc.peer
			if tc.realIP != "" {
				r.Header.Set("X-Real-IP", tc.realIP)
			}
			if tc.forwarded != "" {
				r.Header.Set("X-Forwarded-For", tc.forwarded)
			}
			if got := IsLocalRequest(r); got != tc.local {
				t.Fatalf("IsLocalRequest() = %v, want %v", got, tc.local)
			}
		})
	}
}

func TestIsLocalRequestRejectsAmbiguousHeaders(t *testing.T) {
	for _, tc := range []struct {
		name, trusted string
		headers       http.Header
	}{
		{"duplicate real IP", "172.19.0.2", http.Header{"X-Real-Ip": {"127.0.0.1", "203.0.113.10"}}},
		{"empty real IP", "172.19.0.2", http.Header{"X-Real-Ip": {""}, "X-Forwarded-For": {"127.0.0.1"}}},
		{"empty forwarded IP", "", http.Header{"X-Forwarded-For": {""}}},
		{"unsupported forwarded header", "", http.Header{"Forwarded": {"for=127.0.0.1"}}},
		{"oversized chain", "172.19.0.2", http.Header{"X-Forwarded-For": {strings.Repeat(" ", 4096) + "127.0.0.1"}}},
		{"multiple chain lines include untrusted peer", "172.19.0.2", http.Header{"X-Forwarded-For": {"127.0.0.1", "203.0.113.10"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TRUSTED_PROXY_CIDRS", tc.trusted)
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = "172.19.0.2:1234"
			r.Header = tc.headers
			if IsLocalRequest(r) {
				t.Fatal("ambiguous forwarding metadata authorized local sign-in")
			}
		})
	}
}
