package request // import "admin-console/internal/request"
import (
	"admin-console/internal/model"
	"fmt"
	"net/http"
	"net/netip"
)

// ContextKey  represents a context key.
type ContextKey int

// List of context keys.
const (
	ClientIPContextKey ContextKey = iota
	LoginUserContextKey
	IsAuthenticatedContextKey
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

func IsAuthenticated(r *http.Request) bool {
	if isAuthenticated, ok := getContextAnyValue(r, IsAuthenticatedContextKey).(bool); ok {
		return isAuthenticated
	}
	return false
}

func GetLoginUser(r *http.Request) (*model.LoginUser, error) {
	if loginUser, ok := getContextAnyValue(r, LoginUserContextKey).(*model.LoginUser); ok {
		if loginUser == nil {
			return nil, fmt.Errorf("Login user is nil")
		}
		return loginUser, nil
	}
	return nil, fmt.Errorf("No login user")
}

func getContextStringValue(r *http.Request, key ContextKey) string {
	if v := getContextAnyValue(r, key); v != nil {
		if value, valid := v.(string); valid {
			return value
		}
	}
	return ""
}

func getContextAnyValue(r *http.Request, key ContextKey) any {
	return r.Context().Value(key)
}
