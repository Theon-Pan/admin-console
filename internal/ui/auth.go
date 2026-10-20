package ui // import "admin-console/internal/ui"
import (
	"admin-console/internal/config"
	"admin-console/internal/model"
	"net/http"
	"time"
)

const sessionCookieName = "AdminConsoleSessionID"

// setSessionCookie writes the session cookie to the response with the
// security attributes used by admin-console (HttpOnly, SameSite=Lax, Secure
// when HTTPS).
func SetSessionCookie(w http.ResponseWriter, session *model.WebSession, secret string) {
	path := config.Opts.BasePath()
	if path == "" {
		path = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID + "." + secret,
		Path:     path,
		Secure:   config.Opts.HTTPS(),
		HttpOnly: true,
		Expires:  time.Now().Add(config.Opts.CleanupRemoveSessionsInterval()),
		SameSite: http.SameSiteLaxMode,
	})
}
