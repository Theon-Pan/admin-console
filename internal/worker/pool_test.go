package worker

import (
	"admin-console/internal/model"
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func TestPushAfterShutdownDiscardsJobs(t *testing.T) {
	pool := NewPool(nil, 2, 1, 10, 60*time.Second)
	pool.Shutdown()

	done := make(chan struct{})
	go func() {
		defer close(done)
		pool.Push(model.JobList{
			{UserID: 0},
			{UserID: 1},
		})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Push blocked after Shutdown instead of discard job and return immediately.")
	}
}

func TestShutdownUnblocksPendingPush(t *testing.T) {
	pool := NewPool(nil, 0, 1, 10, 60*time.Second)

	pushed := make(chan struct{})
	go func() {
		defer close(pushed)
		pool.Push(model.JobList{{UserID: 1}})
	}()

	// Give Push time to block on the unbuffered queue before shutting down.
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		defer close(done)
		pool.Shutdown()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown deadlocked while a Push was pending.")
	}

	select {
	case <-pushed:
	case <-time.After(5 * time.Second):
		t.Fatal("Push remained blocked after Shutdown")
	}

}

func TestCircuitBreakerAvailable(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const threshold = 5
		const jobDuration = 2 * time.Millisecond
		pool := NewPool(nil, threshold, 10, threshold, time.Second)
		defer pool.Shutdown()

		failure := any(&model.MockJobRequest{SleepTime: jobDuration})
		jobs := make(model.JobList, threshold)
		for i := range jobs {
			jobs[i] = model.Job{UserID: 1, Type: model.JobTypeMock, JobArg: &failure}
		}
		pool.Push(jobs)
		time.Sleep(jobDuration)
		synctest.Wait()
		if !pool.IsCircuitBreakerOpened() {
			t.Fatal("breaker did not open after reaching the failure threshold")
		}

		time.Sleep(breakerCooldown)
		if !pool.IsCircuitBreakerOpened() {
			t.Fatal("breaker should remain open until a successful probe")
		}

		success := any(&model.MockJobRequest{SleepTime: jobDuration, ReturnSuccess: true})
		pool.Push(model.JobList{{UserID: 1, Type: model.JobTypeMock, JobArg: &success}})
		time.Sleep(jobDuration)
		synctest.Wait()
		if pool.IsCircuitBreakerOpened() {
			t.Fatal("breaker did not close after a successful probe")
		}
	})
}

func TestCircuitBreakerRecovery(t *testing.T) {
	b := &circuitBreaker{threshold: 2}
	first, ok := b.allow()
	if !ok {
		t.Fatal("closed breaker rejected a job")
	}
	stale, ok := b.allow()
	if !ok {
		t.Fatal("closed breaker rejected a concurrent job")
	}
	second, ok := b.allow()
	if !ok {
		t.Fatal("closed breaker rejected a third job")
	}
	b.recordFailure(second)
	if _, ok := b.allow(); !ok {
		t.Fatal("breaker opened before reaching threshold")
	}
	b.recordFailure(first)
	if _, ok := b.allow(); ok {
		t.Fatal("open breaker accepted a job before cooldown")
	}
	b.recordSuccess(stale)
	if _, ok := b.allow(); ok {
		t.Fatal("stale success closed the breaker")
	}

	b.mu.Lock()
	b.retryAt = time.Now().Add(-time.Second)
	b.mu.Unlock()
	probe, ok := b.allow()
	if !ok {
		t.Fatal("breaker rejected recovery probe")
	}
	if _, ok := b.allow(); ok {
		t.Fatal("breaker accepted a second concurrent probe")
	}
	b.recordFailure(stale)
	b.recordSuccess(stale)
	if _, ok := b.allow(); ok {
		t.Fatal("stale result changed the probe state")
	}
	b.recordFailure(probe)
	if _, ok := b.allow(); ok {
		t.Fatal("failed probe did not restart cooldown")
	}

	b.mu.Lock()
	b.retryAt = time.Now().Add(-time.Second)
	b.mu.Unlock()
	probe, ok = b.allow()
	if !ok {
		t.Fatal("breaker rejected second probe")
	}
	b.recordSuccess(probe)
	if _, ok := b.allow(); !ok {
		t.Fatal("successful probe did not close breaker")
	}
	if b.consecutive != 0 {
		t.Fatalf("failure count = %d, want 0", b.consecutive)
	}
}

func TestCircuitBreakerSingleConcurrentProbe(t *testing.T) {
	b := &circuitBreaker{threshold: 1}
	generation, _ := b.allow()
	b.recordFailure(generation)
	b.mu.Lock()
	b.retryAt = time.Now().Add(-time.Second)
	b.mu.Unlock()

	const attempts = 64
	results := make(chan bool, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, ok := b.allow()
			results <- ok
		}()
	}
	wg.Wait()
	close(results)
	admitted := 0
	for ok := range results {
		if ok {
			admitted++
		}
	}
	if admitted != 1 {
		t.Fatalf("admitted %d probes, want 1", admitted)
	}
}

func TestCircuitBreakerConcurrentResults(t *testing.T) {
	b := &circuitBreaker{threshold: 10000}
	const jobs = 100
	var wg sync.WaitGroup
	for i := 0; i < jobs; i++ {
		generation, ok := b.allow()
		if !ok {
			t.Fatal("closed breaker rejected a job")
		}
		wg.Add(1)
		go func(fail bool, generation uint64) {
			defer wg.Done()
			if fail {
				b.recordFailure(generation)
			} else {
				b.recordSuccess(generation)
			}
		}(i%2 == 0, generation)
	}
	wg.Wait()
	if b.state != breakerClosed || b.consecutive < 0 || b.consecutive > jobs/2 {
		t.Fatalf("unexpected breaker state %v, failures %d", b.state, b.consecutive)
	}
}

func TestWorkerSkipsOpenBreakerAndProbesRecovery(t *testing.T) {
	b := &circuitBreaker{threshold: 1}
	w := &worker{breaker: b, jobTimeout: time.Second}
	calls := 0
	dispatch := func(context.Context, model.Job) error {
		calls++
		if calls == 1 {
			return errors.New("database unavailable")
		}
		return nil
	}
	w.executeJobWithDispatcher(model.Job{}, dispatch)
	w.executeJobWithDispatcher(model.Job{}, dispatch)
	if calls != 1 {
		t.Fatalf("dispatcher called %d times while breaker open, want 1", calls)
	}
	b.mu.Lock()
	b.retryAt = time.Now().Add(-time.Second)
	b.mu.Unlock()
	w.executeJobWithDispatcher(model.Job{}, dispatch)
	w.executeJobWithDispatcher(model.Job{}, dispatch)
	if calls != 3 {
		t.Fatalf("dispatcher called %d times after recovery, want 3", calls)
	}
}
