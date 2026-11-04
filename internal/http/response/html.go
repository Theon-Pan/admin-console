package response // import "admin-console/internal/response"
import (
	"admin-console/internal/http/request"
	"log/slog"
	"net/http"
)

// HTMLNotAcceptable sends a not acceptable error to the client.
func HTMLNotAcceptable(w http.ResponseWriter, r *http.Request) {
	slog.Warn(http.StatusText(http.StatusNotAcceptable),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
			slog.String("accept_encoding", r.Header.Get("Accept-Encoding")),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusNotAcceptable),
		),
	)

	NewBuilder(w, r).
		WithStatus(http.StatusNotAcceptable).
		WithHeader("Content-Type", "text/html; charset=utf-8").
		WithHeader("Cache-Control", "no-cache, max-age=0, must-revalidate, no-store").
		WithBodyAsString("Unknown Accept-Encoding").
		Write()
}
