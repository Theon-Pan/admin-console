package ui // import "admin-console/internal/ui"
import (
	"net/http"
	"strings"
)

// isStaticAssetRoute checks if the request path corresponds to a static
// asset route that does not require a web session.
func isStaticAssetRoute(r *http.Request) bool {
	path := r.URL.Path
	switch path {
	case "/favicon.ico", "/robots.txt":
		return true
	}

	return strings.HasPrefix(path, "/stylesheets") ||
		strings.HasPrefix(path, "/js/") ||
		strings.HasPrefix(path, "/icon/") ||
		strings.HasPrefix(path, "/feed-icon/")
}
