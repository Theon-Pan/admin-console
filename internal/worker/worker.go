package worker // import "admin-console/internal/worker"
import (
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"context"
	"fmt"
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
	switch job.Type {
	case model.JobRecordLogin:
		if job.JobArg == nil {
			return fmt.Errorf("Error when executing job of recording login info, no arg specified.")
		}

		if loginInfoRequest, ok := (*job.JobArg).(*model.LoginInfoCreationRequest); ok {
			if _, err := w.store.InsertLoginInfoContext(ctx, loginInfoRequest); err != nil {
				return err
			}
		}
	case model.JobMock:
		sleepTime := 2 * time.Second
		if job.JobArg == nil {
			slog.Warn("No arg, just sleep in default time, and return successfully.",
				slog.Int("sleep_time", 2),
			)
			time.Sleep(sleepTime)
			return nil
		}
		if mockArg, ok := (*job.JobArg).(*model.MockJobArg); ok {
			sleepTime = mockArg.SleepTime
			time.Sleep(sleepTime)
			if mockArg.ReturnSuccess {
				return nil
			} else {
				return fmt.Errorf("Mock an error return from Mock Job.")
			}
		}

	}

	return nil
}
