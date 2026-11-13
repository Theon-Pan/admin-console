package response // import "admin-console/internal/response"
import (
	"admin-console/internal/http/request"
	"log/slog"
	"net/http"
)

func LogInfo(r *http.Request, statusCode int) {
	slog.Info(http.StatusText(statusCode),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", statusCode),
		),
	)
}

func LogWarn(r *http.Request, statusCode int, err error) {
	slog.Warn(http.StatusText(statusCode),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", statusCode),
		),
	)
}

func LogError(r *http.Request, statusCode int, err error) {
	slog.Error(http.StatusText(statusCode),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", statusCode),
		),
	)
}
