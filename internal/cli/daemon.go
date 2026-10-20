package cli // import "admin-console/internal/cli"
import (
	"admin-console/internal/config"
	"admin-console/internal/server"
	"admin-console/internal/storage"
	"admin-console/internal/worker"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func startDaemon(store *storage.Storage) {
	slog.Debug("Starting daemon...")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)

	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGHUP)

	pool := worker.NewPool(store, config.Opts.WorkerPoolSize())

	var httpServers []*http.Server
	var certReloadFn func()
	httpServers, certReloadFn = server.StartWebServer(store, pool)

	for {
		select {
		case <-stop:
			slog.Debug("Shutting down the process")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if len(httpServers) > 0 {
				slog.Debug("Shutting down HTTP servers...")
				for _, srv := range httpServers {
					if srv != nil {
						if err := srv.Shutdown(ctx); err != nil {
							slog.Error("HTTP server shutdown error", slog.Any("error", err), slog.String("addr", srv.Addr))
						}
					}
				}
				slog.Debug("All HTTP servers shut down.")
			} else {
				slog.Debug("No HTTP servers to shut down.")
			}

			slog.Debug("Shutting down worker pool...")
			pool.Shutdown()
			slog.Debug("Worker pool shut down.")

			slog.Debug("Process gracefully stopped")
			return

		case <-reload:
			slog.Info("Received SIGHUP, reloading TLS certificates")
			if certReloadFn != nil {
				certReloadFn()
			}
		}
	}
}
