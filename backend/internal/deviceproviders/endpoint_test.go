package deviceproviders

import (
	"strconv"
	"testing"
)

func TestNewEndpoint(t *testing.T) {
	tests := []struct {
		address string
		port    int
		want    string
	}{
		{"10.0.0.1", 6053, "10.0.0.1:6053"},
		{"172.31.255.255", 1, "172.31.255.255:1"},
		{"192.168.50.160", 65535, "192.168.50.160:65535"},
		{"fd12::1", 6053, "[fd12::1]:6053"},
		{"::ffff:192.168.1.2", 6053, "192.168.1.2:6053"},
		{"localhost", 6053, ""},
		{"192.168.1.2:6053", 6053, ""},
		{"fe80::1%en0", 6053, ""},
		{"127.0.0.1", 6053, ""},
		{"::1", 6053, ""},
		{"169.254.1.1", 6053, ""},
		{"172.32.0.1", 6053, ""},
		{"8.8.8.8", 6053, ""},
		{"224.0.0.1", 6053, ""},
		{"::", 6053, ""},
		{"192.168.1.2", 0, ""},
		{"192.168.1.2", 65536, ""},
	}
	for _, tt := range tests {
		t.Run(tt.address+"/"+strconv.Itoa(tt.port), func(t *testing.T) {
			endpoint, err := NewEndpoint(tt.address, tt.port)
			if tt.want == "" {
				if err == nil {
					t.Fatal("invalid endpoint accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := endpoint.NetworkAddress(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
	if got := (Endpoint{}).NetworkAddress(); got != "" {
		t.Fatalf("zero endpoint: %q", got)
	}
}
