package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"tailscale.com/client/local"
	"tailscale.com/client/tailscale/apitype"
	"tailscale.com/tailcfg"

	"github.com/jaxxstorm/landlord/internal/config"
	"github.com/jaxxstorm/landlord/internal/logger"
)

var (
	ErrTailscaleCapabilityDenied = errors.New("tailscale capability denied")
	ErrTailscaleIdentityMissing  = errors.New("tailscale identity unavailable")
)

type tailscaleIdentityContextKey struct{}

// TailscaleIdentity contains request-scoped caller identity from Tailscale.
type TailscaleIdentity struct {
	LoginName  string
	NodeName   string
	RemoteAddr string
}

// TailscaleAuthorizer authorizes a remote caller for a capability.
type TailscaleAuthorizer interface {
	Authorize(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error)
}

type tailscaleCapabilityAuthorizer struct {
	client tailscaleWhoIsAPI
}

type tailscaleWhoIsAPI interface {
	WhoIs(ctx context.Context, remoteAddr string) (*apitype.WhoIsResponse, error)
}

// WithTailscaleIdentity stores the caller identity in request context.
func WithTailscaleIdentity(ctx context.Context, identity *TailscaleIdentity) context.Context {
	return context.WithValue(ctx, tailscaleIdentityContextKey{}, identity)
}

// TailscaleIdentityFromContext returns the caller identity stored by auth middleware.
func TailscaleIdentityFromContext(ctx context.Context) (*TailscaleIdentity, bool) {
	identity, ok := ctx.Value(tailscaleIdentityContextKey{}).(*TailscaleIdentity)
	return identity, ok
}

func newTailscaleCapabilityAuthorizer(client *local.Client) TailscaleAuthorizer {
	return &tailscaleCapabilityAuthorizer{client: client}
}

func (a *tailscaleCapabilityAuthorizer) Authorize(ctx context.Context, remoteAddr, capability string) (*TailscaleIdentity, error) {
	who, err := a.client.WhoIs(ctx, remoteAddr)
	if err != nil {
		if errors.Is(err, local.ErrPeerNotFound) {
			return nil, ErrTailscaleIdentityMissing
		}
		return nil, fmt.Errorf("tailscale whois failed: %w", err)
	}

	identity := &TailscaleIdentity{
		LoginName:  tailscaleLoginName(who),
		NodeName:   tailscaleNodeName(who),
		RemoteAddr: remoteAddr,
	}

	if !who.CapMap.HasCapability(tailcfg.PeerCapability(capability)) {
		return identity, ErrTailscaleCapabilityDenied
	}

	return identity, nil
}

func (s *Server) tailscaleAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			policy, ok := s.matchProtectedEndpoint(r.Method, r.URL.Path)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = chimiddleware.GetReqID(r.Context())
			}

			reqLogger := logger.FromContext(r.Context()).With(
				zap.String("auth_mechanism", "tailscale"),
				zap.String("auth_capability", policy.Capability),
			)

			if s.tailscaleAuthorizer == nil {
				reqLogger.Error("tailscale authorization unavailable",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				s.writeErrorResponse(w, http.StatusServiceUnavailable, "Authorization unavailable", []string{"tailscale authorizer is not initialized"}, requestID)
				return
			}

			identity, err := s.tailscaleAuthorizer.Authorize(r.Context(), r.RemoteAddr, policy.Capability)
			if err != nil {
				status, message, level := classifyTailscaleAuthError(err)
				fields := []zap.Field{
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("remote_addr", r.RemoteAddr),
					zap.String("decision", "deny"),
					zap.String("reason", message),
				}
				if identity != nil {
					fields = append(fields,
						zap.String("tailscale_login", identity.LoginName),
						zap.String("tailscale_node", identity.NodeName),
					)
				}

				switch level {
				case zap.WarnLevel:
					reqLogger.Warn("tailscale authorization denied", fields...)
				default:
					fields = append(fields, zap.Error(err))
					reqLogger.Error("tailscale authorization failed", fields...)
				}

				s.writeErrorResponse(w, status, message, []string{err.Error()}, requestID)
				return
			}

			reqLogger.Info("tailscale authorization allowed",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("decision", "allow"),
				zap.String("tailscale_login", identity.LoginName),
				zap.String("tailscale_node", identity.NodeName),
			)

			next.ServeHTTP(w, r.WithContext(WithTailscaleIdentity(r.Context(), identity)))
		})
	}
}

func classifyTailscaleAuthError(err error) (int, string, zapcore.Level) {
	switch {
	case errors.Is(err, ErrTailscaleCapabilityDenied):
		return http.StatusForbidden, "Forbidden", zap.WarnLevel
	case errors.Is(err, ErrTailscaleIdentityMissing):
		return http.StatusForbidden, "Forbidden", zap.WarnLevel
	default:
		return http.StatusServiceUnavailable, "Authorization unavailable", zap.ErrorLevel
	}
}

func tailscaleLoginName(who *apitype.WhoIsResponse) string {
	if who == nil || who.UserProfile == nil {
		return ""
	}
	return who.UserProfile.LoginName
}

func tailscaleNodeName(who *apitype.WhoIsResponse) string {
	if who == nil || who.Node == nil {
		return ""
	}
	return who.Node.ComputedName
}

func (s *Server) matchProtectedEndpoint(method, path string) (config.TailscaleProtectedEndpoint, bool) {
	if !s.httpConfig.TailscaleAuth.Enabled {
		return config.TailscaleProtectedEndpoint{}, false
	}

	for _, endpoint := range s.httpConfig.TailscaleAuth.ProtectedEndpoints {
		if endpoint.Method != method {
			continue
		}
		if tailscalePathMatch(endpoint.Path, path) {
			return endpoint, true
		}
	}

	return config.TailscaleProtectedEndpoint{}, false
}

func tailscalePathMatch(pattern, path string) bool {
	if pattern == path {
		return true
	}

	patternParts := splitPath(pattern)
	pathParts := splitPath(path)
	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		part := patternParts[i]
		switch {
		case part == "*":
			continue
		case strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}"):
			continue
		case part != pathParts[i]:
			return false
		}
	}

	return true
}

func splitPath(path string) []string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}
