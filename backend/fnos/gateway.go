package fnos

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type gatewayContextKey struct{}

type gatewayIdentity struct {
	UserID   string
	Username string
	IsAdmin  bool
}

// MarkGatewayConnection marks requests accepted by the fnOS Unix socket.
// Header values are trusted only when this marker is present.
func MarkGatewayConnection(ctx context.Context, _ net.Conn) context.Context {
	return context.WithValue(ctx, gatewayContextKey{}, true)
}

func identityFromRequest(r *http.Request) (gatewayIdentity, bool) {
	trusted, _ := r.Context().Value(gatewayContextKey{}).(bool)
	if !trusted {
		return gatewayIdentity{}, false
	}

	userID := strings.TrimSpace(r.Header.Get("X-Trim-Userid"))
	username := strings.TrimSpace(r.Header.Get("X-Trim-Username"))
	if userID == "" || username == "" {
		return gatewayIdentity{}, false
	}

	return gatewayIdentity{
		UserID:   userID,
		Username: username,
		IsAdmin:  strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Trim-Isadmin")), "true") || strings.TrimSpace(r.Header.Get("X-Trim-Isadmin")) == "1",
	}, true
}
