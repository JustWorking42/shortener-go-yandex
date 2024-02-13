// Package accesscontrol provides middleware for controlling access based on IP addresses.
package accesscontrol

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/JustWorking42/shortener-go-yandex/internal/app"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CidrAccessMiddleware is a middleware that checks if the client IP is in a given CIDR block.
func CidrAccessMiddleware(app *app.App, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		trustedSubnet := app.TrustedSubnet

		if trustedSubnet == "" {
			sendError(w)
			return
		}

		clientIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		if clientIP == "" {
			sendError(w)
			return
		}

		ip := net.ParseIP(clientIP)
		if ip == nil {
			sendError(w)
			return
		}

		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			sendError(w)
			return

		}

		if !ipNet.Contains(ip) {
			sendError(w)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func CidrAccessInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, app *app.App) (interface{}, error) {
	trustedSubnet := app.TrustedSubnet

	if trustedSubnet == "" {
		return nil, status.Error(codes.PermissionDenied, "no trusted subnet")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "no metadata")
	}
	clientIP := md.Get("x-real-ip")[0]
	if clientIP == "" {
		return nil, status.Error(codes.PermissionDenied, "no client IP")
	}
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return nil, status.Error(codes.PermissionDenied, "invalid client IP")
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid trusted subnet")
	}

	if !ipNet.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, "client IP not in trusted subnet")
	}

	return handler(ctx, req)
}

func sendError(w http.ResponseWriter) {
	http.Error(w, "no trusted subnet", http.StatusForbidden)
}
