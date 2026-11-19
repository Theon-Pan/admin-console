package api // import "admin-console/internal/api"
import (
	"admin-console/internal/http/request"
	"admin-console/internal/http/response"
	"admin-console/internal/storage"
	"admin-console/internal/utils"
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const bearerPrefix = "Bearer "

// Paths (matched as suffixes of the request path) that skip token authentication.
var anonymousPaths = []string{
	"/login",
	"/captchaImage",
}

type middleware struct {
	store *storage.Storage
}

func newMiddleware(store *storage.Storage) *middleware {
	return &middleware{store: store}
}

func (m *middleware) withCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Auth-Token, Authorization, Content-Type, Accept")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "3600")
			response.NoContent(w, r)
		}
		next.ServeHTTP(w, r)
	})
}

// handle validates the bearer token and ensures the login session exists in storage.
func (m *middleware) validateJWTokenAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAnonymousPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), bearerPrefix)
		tokenString = strings.TrimSpace(tokenString)
		if !ok || tokenString == "" {
			response.JSONUnauthorized(w, r)
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil || claims.UUID == "" {
			response.JSONUnauthorized(w, r)
			return
		}

		loginUser, err := m.store.RetrieveLoginUser(r.Context(), claims.UUID)
		if err != nil || loginUser == nil {
			response.JSONUnauthorized(w, r)
			return
		}
		slog.Info("Authentication passed",
			slog.String("url", r.RequestURI))

		ctx := r.Context()
		ctx = context.WithValue(ctx, request.LoginUserContextKey, loginUser)
		ctx = context.WithValue(ctx, request.IsAuthenticatedContextKey, true)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isAnonymousPath(path string) bool {
	for _, p := range anonymousPaths {
		if strings.HasSuffix(path, p) {
			return true
		}
	}
	return false
}
