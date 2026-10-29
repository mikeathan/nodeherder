package httpclient

// Checks provider HTTP client behavior and destination restrictions.

import (
	"net/http"
	"net/url"
	"testing"
)

func TestClientKeepsRestrictedTransportPolicy(t *testing.T) {
	client := NewHTTPClient()
	defer client.CloseIdleConnections()
	transport, ok := client.Transport.(*http.Transport)
	if !ok || transport.Proxy != nil || transport.DialContext == nil || client.Timeout == 0 || transport.MaxConnsPerHost == 0 {
		t.Fatal("unbounded/unrestricted provider client")
	}
	if err := client.CheckRedirect(&http.Request{URL: &url.URL{Scheme: "https", Host: "other.test"}}, nil); err != http.ErrUseLastResponse {
		t.Fatal("redirect allowed", err)
	}
}
