package request // import "admin-console/internal/request"
import (
	"net/http"
	"net/netip"
)

// ContextKey  represents a context key.
type ContextKey int

// List of context keys.
const (
	ClientIPContextKey ContextKey = iota
)

// ClientIP returns the IP from the request context or the direct connection's
// address. It returns an empty string when neither provides a valid IP.
func ClientIP(r *http.Request) string {
	if ip, err := netip.ParseAddr(getContextStringValue(r, ClientIPContextKey)); err == nil {
		return ip.WithZone("").String()
	}
	if peer, err := netip.ParseAddrPort(r.RemoteAddr); err == nil {
		return peer.Addr().WithZone("").String()
	}
	return ""
}

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := r.Context().Value(key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}
