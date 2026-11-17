package worker

import (
	"admin-console/internal/model"
	"context"
	"errors"
	"testing"
	"time"
)

func TestExecuteJobRecordsOutcome(t *testing.T) {
	for _, tt := range []struct {
		name        string
		dispatchErr error
		wantCount   int64
	}{
		{"success", nil, 0},
		{"failure", errors.New("job failed"), 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			breaker := &circuitBreaker{threshold: 5, consecutive: 1}
			w := &worker{breaker: breaker, jobTimeout: time.Second}
			job := model.Job{UserID: 42}
			w.executeJobWithDispatcher(job, func(ctx context.Context, got model.Job) error {
				if got != job {
					t.Errorf("job = %+v, want %+v", got, job)
				}
				if _, ok := ctx.Deadline(); !ok {
					t.Error("dispatcher did not receive a deadline")
				}
				return tt.dispatchErr
			})
			if got := breaker.consecutive; got != tt.wantCount {
				t.Errorf("consecutive failures = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

func TestExecuteJobTimeoutRecordsOneFailure(t *testing.T) {
	breaker := &circuitBreaker{threshold: 5}
	w := &worker{breaker: breaker, jobTimeout: 10 * time.Millisecond}
	w.executeJobWithDispatcher(model.Job{}, func(ctx context.Context, _ model.Job) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if got := breaker.consecutive; got != 1 {
		t.Fatalf("consecutive failures = %d, want 1", got)
	}
}

func TestExecuteJobWaitsForDispatcherAfterTimeout(t *testing.T) {
	breaker := &circuitBreaker{threshold: 5}
	w := &worker{breaker: breaker, jobTimeout: 10 * time.Millisecond}
	timedOut := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		w.executeJobWithDispatcher(model.Job{}, func(ctx context.Context, _ model.Job) error {
			<-ctx.Done()
			close(timedOut)
			<-release
			return nil
		})
	}()
	<-timedOut
	select {
	case <-done:
		t.Fatal("worker returned while dispatcher was still running")
	default:
	}
	close(release)
	<-done
	if got := breaker.consecutive; got != 1 {
		t.Fatalf("consecutive failures = %d, want 1", got)
	}
}
