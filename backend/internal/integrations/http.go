package integrations

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"time"
)

// Resolver is injected only for deterministic policy tests, never configured by provider YAML.
type Resolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}
type networkDialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

var forbiddenNetworks = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"), netip.MustParsePrefix("127.0.0.0/8"), netip.MustParsePrefix("169.254.0.0/16"), netip.MustParsePrefix("172.16.0.0/12"), netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"), netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("198.18.0.0/15"), netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"), netip.MustParsePrefix("224.0.0.0/4"), netip.MustParsePrefix("240.0.0.0/4"), netip.MustParsePrefix("::/96"), netip.MustParsePrefix("64:ff9b::/96"), netip.MustParsePrefix("64:ff9b:1::/48"), netip.MustParsePrefix("100::/64"), netip.MustParsePrefix("2001::/23"), netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2002::/16"), netip.MustParsePrefix("fc00::/7"), netip.MustParsePrefix("fe80::/10"),
}

func publicAddress(a netip.Addr) bool {
	a = a.Unmap()
	if !a.IsValid() || !a.IsGlobalUnicast() {
		return false
	}
	for _, prefix := range forbiddenNetworks {
		if prefix.Contains(a) {
			return false
		}
	}
	return true
}
func safeDial(resolver Resolver, dialer networkDialer) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, e := net.SplitHostPort(address)
		if e != nil {
			return nil, errors.New("invalid provider address")
		}
		if port != "443" {
			return nil, errors.New("provider connections require HTTPS port 443")
		}
		addresses, e := resolver.LookupNetIP(ctx, "ip", host)
		if e != nil || len(addresses) == 0 {
			return nil, errors.New("provider DNS lookup failed")
		}
		for _, ip := range addresses {
			if !publicAddress(ip) {
				return nil, errors.New("provider DNS includes a restricted address")
			}
		}
		var last error
		for _, ip := range addresses {
			connection, e := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if e == nil {
				return connection, nil
			}
			last = e
		}
		return nil, last
	}
}

// NewHTTPClient pins each connection to a validated public DNS answer. It ignores proxy environment
// variables and refuses redirects so a provider move requires explicit destination review.
func NewHTTPClient() *http.Client {
	transport := &http.Transport{Proxy: nil, DialContext: safeDial(net.DefaultResolver, &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}), TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 15 * time.Second, MaxIdleConns: 8, MaxIdleConnsPerHost: 2, MaxConnsPerHost: 4, IdleConnTimeout: 30 * time.Second}
	return &http.Client{Transport: transport, Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}
