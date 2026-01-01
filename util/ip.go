package util

import (
	"net"
	"strings"

	"github.com/labstack/echo/v4"
)

// GetRealIP extracts the real client IP address from the request.
// It handles proxy scenarios by checking X-Forwarded-For and X-Real-IP headers.
// If the application is configured to run behind a proxy (util.Proxy = true),
// it will prioritize the proxy headers. Otherwise, it falls back to RemoteAddr.
func GetRealIP(c echo.Context) string {
	// If behind a proxy, check proxy headers
	if Proxy {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// We want the leftmost (original client) IP
		if xff := c.Request().Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				clientIP := strings.TrimSpace(ips[0])
				if clientIP != "" {
					return clientIP
				}
			}
		}

		// X-Real-IP is typically set by nginx and similar proxies
		if xri := c.Request().Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	// Fall back to RemoteAddr
	// RemoteAddr format is "IP:port", so we need to extract just the IP
	ip := c.Request().RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}
