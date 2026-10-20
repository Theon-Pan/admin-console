package worker // import "admin-console/internal/worker"
import (
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"log/slog"
	"sync"
)

type worker struct {
	id    int
	store *storage.Storage
}

// Run processes background jobs from the channel until the pool is shut down.
func (w *worker) Run(c <-chan model.Job, shutdown <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	slog.Debug("Worker started",
		slog.Int("worker_id", w.id),
	)

	for {
		var job model.Job
		select {
		case <-shutdown:
			return
		case job = <-c:
		}

		slog.Debug("Job received by worker",
			slog.Int("worker_id", w.id),
			slog.Int64("user_id", job.UserID),
		)

		// @todo Add the concrete logic of the worker
	}
}
