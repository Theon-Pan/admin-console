package request // import "admin-console/internal/request"
import "net/http"

// ContextKey  represents a context key.
type ContextKey int

// List of context keys.
const (
	ClientIPContextKey ContextKey = iota
)

// ClientIP returns the client IP address stored in the request context.
func ClientIP(r *http.Request) string {
	return getContextStringValue(r, ClientIPContextKey)
}

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}
