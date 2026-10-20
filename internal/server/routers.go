package server // import "admin-console/internal/server"
import (
	"admin-console/internal/config"
	"admin-console/internal/storage"
	"admin-console/internal/worker"
	"net/http"
)

func newRouter(store *storage.Storage, pool *worker.Pool) http.Handler {
	readinessProbe := newReadinessProbe(store)

	// Application routes served under the base path.
	appMux := http.NewServeMux()

	appMux.HandleFunc("GET /healthcheck", readinessProbe)

	var appHandler http.Handler = appMux

	// Root Router: health probes at root, app routes under base path.
	rootMux := http.NewServeMux()

	basePath := config.Opts.BasePath()
	if basePath != "" {
		rootMux.Handle(basePath+"/", http.StripPrefix(basePath, appHandler))
	} else {
		rootMux.Handle("/", appHandler)
	}

	return rootMux

}
