package api // import "admin-console/internal/api"
import (
	"admin-console/internal/http/request"
	"admin-console/internal/http/response"
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"admin-console/internal/utils"
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const loginUserContextKey contextKey = "loginUser"

const bearerPrefix = "Bearer "

// Paths (matched as suffixes of the request path) that skip token authentication.
var anonymousPaths = []string{
	"/login",
	"/captchaImage",
}

type authMiddleware struct {
	store *storage.Storage
}

func newAuthMiddleware(store *storage.Storage) *authMiddleware {
	return &authMiddleware{store: store}
}

// handle validates the bearer token and ensures the login session exists in storage.
func (m *authMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAnonymousPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), bearerPrefix)
		tokenString = strings.TrimSpace(tokenString)
		if !ok || tokenString == "" {
			unauthorized(w, r, "Missing authentication token", nil)
			return
		}

		claims, err := utils.ValidateToken(tokenString)
		if err != nil || claims.UUID == "" {
			unauthorized(w, r, "Invalid authentication token", err)
			return
		}

		loginUser, err := m.store.RetrieveLoginUser(r.Context(), claims.UUID)
		if err != nil || loginUser == nil {
			unauthorized(w, r, "Login session expired, please login again", err)
			return
		}

		ctx := context.WithValue(r.Context(), loginUserContextKey, loginUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoginUserFromContext returns the authenticated login user set by the auth middleware.
func LoginUserFromContext(ctx context.Context) *model.LoginUser {
	loginUser, _ := ctx.Value(loginUserContextKey).(*model.LoginUser)
	return loginUser
}

func isAnonymousPath(path string) bool {
	for _, p := range anonymousPaths {
		if strings.HasSuffix(path, p) {
			return true
		}
	}
	return false
}

func unauthorized(w http.ResponseWriter, r *http.Request, message string, err error) {
	slog.Warn("Authentication failed",
		slog.String("reason", message),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.String("method", r.Method),
		slog.String("uri", r.RequestURI),
	)
	// TODO: Use i18n to get error message.
	response.JSON(w, r, response.NewErrorResultWithCodeAndMessage(response.Unauthorized, message))
}
