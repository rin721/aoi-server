package utils

import (
	"net"
	"net/netip"
	"strings"

	"github.com/rei0721/go-scaffold/pkg/web"
)

func ClientIPRealIP(c web.Context) string {
	for _, candidate := range []string{
		forwardedForHeader(c.GetHeader("Forwarded")),
		firstForwardedIP(c.GetHeader("X-Forwarded-For")),
		c.GetHeader("X-Real-IP"),
		c.GetHeader("CF-Connecting-IP"),
		c.GetHeader("True-Client-IP"),
		c.ClientIP(),
		remoteAddrIP(c),
	} {
		if ip := normalizeIP(candidate); ip != "" {
			return ip
		}
	}
	return ""
}

func firstForwardedIP(value string) string {
	for _, part := range strings.Split(value, ",") {
		if normalizeIP(part) != "" {
			return part
		}
	}
	return ""
}

func forwardedForHeader(value string) string {
	for _, element := range strings.Split(value, ",") {
		for _, part := range strings.Split(element, ";") {
			key, val, ok := strings.Cut(strings.TrimSpace(part), "=")
			if !ok || !strings.EqualFold(key, "for") {
				continue
			}
			if normalizeIP(val) != "" {
				return val
			}
		}
	}
	return ""
}

func remoteAddrIP(c web.Context) string {
	req := c.Request()
	if req == nil {
		return ""
	}
	return req.RemoteAddr
}

func normalizeIP(value string) string {
	value = strings.Trim(strings.TrimSpace(value), `"`)
	if value == "" || strings.EqualFold(value, "unknown") {
		return ""
	}
	if ip, err := netip.ParseAddr(value); err == nil {
		return ip.String()
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		if ip, err := netip.ParseAddr(strings.Trim(host, "[]")); err == nil {
			return ip.String()
		}
	}
	if ip, err := netip.ParseAddr(strings.Trim(value, "[]")); err == nil {
		return ip.String()
	}
	return ""
}
