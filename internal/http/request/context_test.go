package request

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		contextIP  string
		remoteAddr string
		want       string
	}{
		{"context takes precedence", "192.0.2.1", "127.0.0.1:1234", "192.0.2.1"},
		{"direct IPv4", "", "127.0.0.1:1234", "127.0.0.1"},
		{"direct IPv6 without zone", "", "[fe80::1%eth0]:1234", "fe80::1"},
		{"invalid context IP", "invalid", "[::1]:1234", "::1"},
		{"unknown IP ignores forwarding header", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/api/login", nil)
			r.RemoteAddr = tt.remoteAddr
			r.Header.Set("X-Forwarded-For", "192.0.2.99")
			if tt.contextIP != "" {
				r = r.WithContext(context.WithValue(r.Context(), ClientIPContextKey, tt.contextIP))
			}
			if got := ClientIP(r); got != tt.want {
				t.Errorf("ClientIP = %q, want %q", got, tt.want)
			}
		})
	}
}
