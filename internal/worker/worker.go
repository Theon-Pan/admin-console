package worker // import "admin-console/internal/worker"
import (
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"context"
	"log/slog"
	"sync"
	"time"
)

type worker struct {
	id         int
	store      *storage.Storage
	breaker    *circuitBreaker
	jobTimeout time.Duration
}

// Run processes background jobs from the channel until the pool is shut down.
func (w *worker) Run(c <-chan model.Job, shutdown <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	slog.Debug("Worker started",
		slog.Int("worker_id", w.id),
		slog.Duration("job_timeout", w.jobTimeout),
	)

	for {
		var job model.Job
		select {
		case <-shutdown:
			return
		case job = <-c:
		}

		w.executeJob(job)
	}

}

func (w *worker) executeJob(job model.Job) {
	w.executeJobWithDispatcher(job, w.jobDispatcher)
}

func (w *worker) executeJobWithDispatcher(job model.Job, dispatch func(context.Context, model.Job) error) {
	var generation uint64
	if w.breaker != nil {
		var allowed bool
		generation, allowed = w.breaker.allow()
		if !allowed {
			slog.Warn("Skipping job: worker circuit breaker is open or probing",
				slog.Int("worker_id", w.id),
			)
			return
		}
	}

	slog.Debug("Job received by worker",
		slog.Int("worker_id", w.id),
		slog.Int64("user_id", job.UserID),
	)
	ctx, cancel := context.WithTimeout(context.Background(), w.jobTimeout)
	defer cancel()

	err := dispatch(ctx, job)
	if ctx.Err() != nil {
		slog.Warn("Job exceeded worker timeout",
			slog.Int("worker_id", w.id),
			slog.Duration("timeout", w.jobTimeout),
			slog.Any("error", err),
		)
		if w.breaker != nil {
			w.breaker.recordFailure(generation)
		}
		return
	}

	if err != nil {
		slog.Error("Error when running worker",
			slog.Int("worker_id", w.id),
			slog.Any("error", err),
		)
		if w.breaker != nil {
			w.breaker.recordFailure(generation)
		}
		return
	}

	if w.breaker != nil {
		w.breaker.recordSuccess(generation)
	}
}

func (w *worker) jobDispatcher(ctx context.Context, job model.Job) error {
	return ctx.Err()
}
