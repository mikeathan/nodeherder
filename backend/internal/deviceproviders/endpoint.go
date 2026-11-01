package deviceproviders

// This file validates the network target supplied during enrollment. Only a
// numeric private unicast literal is accepted, so no DNS resolution or redirect
// can turn an authenticated configuration request into a network probe.
import (
	"errors"
	"net"
	"net/netip"
	"strconv"
)

// Endpoint is a validated private unicast address and TCP port. The address is a
// numeric literal, so its NetworkAddress does not require DNS resolution.
type Endpoint struct {
	address netip.Addr
	port    uint16
}

// NewEndpoint accepts RFC 1918 IPv4 or RFC 4193 IPv6 addresses and ports 1–65535.
// IPv4-mapped IPv6 is normalized to IPv4. Hostnames and scoped addresses are rejected.
// Validation errors do not contain the supplied address.
func NewEndpoint(address string, port int) (Endpoint, error) {
	if port < MinPort || port > MaxPort {
		return Endpoint{}, errors.New("device endpoint port must be between 1 and 65535")
	}
	ip, err := netip.ParseAddr(address)
	if err != nil || ip.Zone() != "" {
		return Endpoint{}, errors.New("device endpoint must be a numeric private IP address")
	}
	ip = ip.Unmap()
	if !ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsMulticast() || ip.IsUnspecified() {
		return Endpoint{}, errors.New("device endpoint must be a numeric private IP address")
	}
	return Endpoint{address: ip, port: uint16(port)}, nil
}

// NetworkAddress returns the normalized host:port passed to a TCP dialer. The
// zero Endpoint is invalid and returns an empty string.
func (e Endpoint) NetworkAddress() string {
	if !e.address.IsValid() || e.port == 0 {
		return ""
	}
	return net.JoinHostPort(e.address.String(), strconv.Itoa(int(e.port)))
}
