// Package accesscontrol provides middleware for controlling access based on IP addresses.
package accesscontrol

import (
	"net"
	"net/http"
	"strings"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
)

// CidrAccessMiddleware is a middleware that checks if the client IP is in a given CIDR block.
func CidrAccessMiddleware(app *app.App, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trustedSubnet := app.TrustedSubnet

		if trustedSubnet == "" {
			http.Error(w, "no trusted subnet", http.StatusForbidden)
			return
		}

		clientIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if clientIP == "" {
			http.Error(w, "no trusted subnet", http.StatusForbidden)
			return
		}

		ip := net.ParseIP(clientIP)
		if ip == nil {
			http.Error(w, "no trusted subnet", http.StatusForbidden)
			return
		}

		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			http.Error(w, "no trusted subnet", http.StatusForbidden)
			return

		}

		if !ipNet.Contains(ip) {
			http.Error(w, "no trusted subnet", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
