package ui // import "admin-console/internal/ui"
import (
	"admin-console/internal/http/request"
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"net/http"
	"strings"
)

type webSessionMiddleware struct {
	basePath string
	store    *storage.Storage
}

func NewWebSessionMiddleware(basePath string, store *storage.Storage) *webSessionMiddleware {
	return &webSessionMiddleware{
		basePath: basePath,
		store:    store,
	}
}

func (m *webSessionMiddleware) handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isStaticAssetRoute(r) {
			next.ServeHTTP(w, r)
			return
		}

		// session, err := m.loadWebSessionFromCookie(r)
		// if err != nil {
		// 	response.
		// }

	})
}

func (m *webSessionMiddleware) loadWebSessionFromCookie(r *http.Request) (*model.WebSession, error) {
	cookieValue := request.CookieValue(r, sessionCookieName)
	if cookieValue == "" {
		return nil, nil
	}

	sessionID, secret, ok := strings.Cut(cookieValue, ".")
	if !ok || sessionID == "" || secret == "" {
		return nil, nil
	}

	session, err := m.store.WebSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	if session == nil || !session.VerifySecret(secret) {
		return nil, nil
	}

	return session, nil
}
