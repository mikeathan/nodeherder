// Package destination owns the URL rules used to bind provider requests to credentials.
// Network address and DNS enforcement remains the transport's responsibility.
package destination

// Validates provider destination URLs and canonicalizes origins used for credential binding.

import (
	"net"
	"net/url"
	"strings"

	model "node-herder/models/integrations"
)

// Parse validates a provider base URL without DNS resolution or network I/O. The URL is
// caller-owned; errors do not include the potentially private input string.
func Parse(raw string) (*url.URL, error) {
	blocked := model.Fail("destination_blocked", "Use a public HTTPS provider on port 443.")
	if len(raw) > 2048 {
		return nil, blocked
	}
	destination, err := url.Parse(raw)
	if err != nil {
		return nil, blocked
	}
	if destination.Scheme != "https" || destination.Hostname() == "" {
		return nil, blocked
	}
	if destination.User != nil || destination.RawQuery != "" || destination.Fragment != "" {
		return nil, blocked
	}
	if port := destination.Port(); port != "" && port != "443" {
		return nil, blocked
	}
	return destination, nil
}

// Origin canonicalizes the URL's scheme, host and non-default port for credential binding.
// It does not validate the URL; callers must validate untrusted destinations before using it.
func Origin(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	if port != "" && port != "443" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return strings.ToLower(u.Scheme) + "://" + host
}
