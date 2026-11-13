package api // import "admin-console/internal/api"
import (
	"admin-console/internal/storage"
	"admin-console/internal/worker"
	"net/http"
)

type handler struct {
	store *storage.Storage
	pool  *worker.Pool
}

// NewHandler returns a http.Handler that handles API calls.
// The returned handler expects the base path to be stripped from the request URL.
func NewHandler(store *storage.Storage, pool *worker.Pool) http.Handler {
	handler := &handler{store: store, pool: pool}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/captchaImage", handler.captchaImg)
	mux.HandleFunc("POST /api/login", handler.Login)

	return mux
}
