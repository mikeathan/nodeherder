package httpclient

// Constructs the restricted HTTP client used for provider requests.

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// NewHTTPClient pins each connection to a validated public DNS answer. It ignores proxy environment
// variables and refuses redirects so a provider move requires explicit destination review.
func NewHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy: nil,
		DialContext: safeDial(net.DefaultResolver, &net.Dialer{
			Timeout: 10 * time.Second, KeepAlive: 30 * time.Second,
		}),
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		MaxIdleConns:          8,
		MaxIdleConnsPerHost:   2,
		MaxConnsPerHost:       4,
		IdleConnTimeout:       30 * time.Second,
	}
	return &http.Client{
		Transport:     transport,
		Timeout:       15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}
