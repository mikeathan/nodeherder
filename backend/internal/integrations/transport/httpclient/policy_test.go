package httpclient

import (
	"context"
	"net"
	"net/netip"
	"testing"
)

type fixedResolver struct{ ips []netip.Addr }

func (r fixedResolver) LookupNetIP(context.Context, string, string) ([]netip.Addr, error) {
	return r.ips, nil
}

type recordingDialer struct {
	called  bool
	address string
}

func (d *recordingDialer) DialContext(_ context.Context, _, a string) (net.Conn, error) {
	d.called = true
	d.address = a
	return nil, nil
}
func TestSafeDial(t *testing.T) {
	for _, s := range []string{"127.0.0.1", "10.0.0.1", "100.100.100.200", "169.254.169.254", "192.168.1.1", "::1", "::ffff:127.0.0.1", "fe80::1", "fd00::1", "64:ff9b::a00:1", "2001:db8::1"} {
		t.Run(s, func(t *testing.T) {
			d := &recordingDialer{}
			_, e := safeDial(fixedResolver{[]netip.Addr{netip.MustParseAddr(s)}}, d)(context.Background(), "tcp", "provider.test:443")
			if e == nil || d.called {
				t.Fatal("restricted integration admitted")
			}
		})
	}
	d := &recordingDialer{}
	_, e := safeDial(fixedResolver{[]netip.Addr{netip.MustParseAddr("1.1.1.1")}}, d)(context.Background(), "tcp", "provider.test:443")
	if e != nil || d.address != "1.1.1.1:443" {
		t.Fatal(e, d.address)
	}
	d.called = false
	_, e = safeDial(fixedResolver{[]netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("10.0.0.1")}}, d)(context.Background(), "tcp", "provider.test:443")
	if e == nil || d.called {
		t.Fatal("mixed DNS bypass")
	}
}
