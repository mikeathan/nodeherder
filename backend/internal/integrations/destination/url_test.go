package destination

// Checks provider destination URL validation and origin normalization.

import (
	"net/url"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, raw := range []string{"https://api.example", "https://api.example:443/v1", "https://[2001:db8::1]/v1"} {
		if _, err := Parse(raw); err != nil {
			t.Fatalf("rejected valid base URL: %v", err)
		}
	}
	for _, raw := range []string{"http://api.example", "https://api.example:8443", "https://user:secret@api.example", "https://api.example?key=secret", "https://api.example/#fragment", ":invalid", "https:///missing", "https://" + strings.Repeat("a", 2049)} {
		if _, err := Parse(raw); err == nil {
			t.Fatal("accepted invalid destination")
		}
	}
}

func TestOrigin(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"https://API.Example:443/path", "https://api.example"},
		{"https://api.example:8443/path", "https://api.example:8443"},
		{"https://[2001:db8::1]:443/path", "https://[2001:db8::1]"},
		{"https://[2001:db8::1]:8443/path", "https://[2001:db8::1]:8443"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			u, err := url.Parse(tc.raw)
			if err != nil {
				t.Fatal(err)
			}
			if got := Origin(u); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
