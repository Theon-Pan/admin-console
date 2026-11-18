package api // import "admin-console/internal/api"
import (
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"admin-console/internal/worker"
	"fmt"
	"log/slog"
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
	middleware := newMiddleware(store)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/captchaImage", handler.captchaImg)
	mux.HandleFunc("POST /api/login", handler.Login)

	return middleware.withCORSHeaders(middleware.validateJWTokenAuth(mux))
}

// EnqueueJob put a request to the pool to do the job according to the type of the request.
func (h *handler) EnqueueJob(jobRequest any) {
	var job model.Job
	switch request := jobRequest.(type) {
	case *model.LoginInfoCreationRequest:
		if request == nil {
			slog.Warn("Cannot enqueue login info job: nil request")
			return
		}

		job.Type = model.JobTypeRecordLogin
		job.UserID = request.UserID
	case *model.MockJobRequest:
		if request == nil {
			slog.Warn("Cannot enqueue mock job: nil request")
			return
		}

		job.Type = model.JobTypeMock
	default:
		slog.Warn("Cannot enqueue job: unsupported request type",
			slog.String("type", fmt.Sprintf("%T", jobRequest)))
		return
	}
	job.JobArg = &jobRequest
	h.pool.Push(model.JobList{job})
}
