package utils

import (
	"net"
	"net/http"
	"strings"
)

func isLocalRequest(r *http.Request) bool {
	host := r.RemoteAddr
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		host = xri
	} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// take the first IP in the list
		parts := strings.Split(xff, ",")
		host = strings.TrimSpace(parts[0])
	}
	// Strip port if present
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	// Loopback
	if ip.IsLoopback() {
		return true
	}
	// IPv4 private ranges
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 10 { // 10.0.0.0/8
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 { // 172.16.0.0/12
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 { // 192.168.0.0/16
			return true
		}
	}
	// IPv6 unique local addresses (fc00::/7) or link-local (fe80::/10)
	if ip.To16() != nil && ip.To4() == nil {
		// fc00::/7 -> first 7 bits are 1111110x, we can check prefix fc00::/7 approximately
		if strings.HasPrefix(ip.String(), "fc") || strings.HasPrefix(ip.String(), "fd") {
			return true
		}
		if strings.HasPrefix(ip.String(), "fe80:") {
			return true
		}
	}
	return false
}
