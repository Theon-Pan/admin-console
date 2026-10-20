package request // import "admin-console/internal/http/request"
import "net/http"

// CookieValue returns the named cookie value, or an empty string if the cookie is missing.
func CookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}
