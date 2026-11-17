package worker // import "admin-console/internal/worker"
import (
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"log/slog"
	"sync"
	"time"
)

// Pool manages a set of background workers that process the background jobs.
type Pool struct {
	queue        chan model.Job
	breaker      *circuitBreaker
	jobTimeout   time.Duration
	shutdown     chan struct{}
	shutdownOnce sync.Once
	wg           sync.WaitGroup
}

// Push sends a list of jobs to the queue.
// Jobs pushed after Shutdown are discarded.
func (p *Pool) Push(jobs model.JobList) {
	for _, job := range jobs {
		select {
		case p.queue <- job:
		case <-p.shutdown:
			return
		}
	}
}

// Shutdown stops accepting new jobs and waits for all workers to finish their current jobs.
func (p *Pool) Shutdown() {
	p.shutdownOnce.Do(func() {
		close(p.shutdown)
	})
	p.wg.Wait()
}

// NewPool creates a pool of background workers.
func NewPool(store *storage.Storage, nbWorkers, queueSize, circuitBreakerThreshold int, jobTimeout time.Duration) *Pool {
	workerPool := &Pool{
		queue: make(chan model.Job, queueSize),
		breaker: &circuitBreaker{
			threshold: int64(circuitBreakerThreshold),
		},
		jobTimeout: jobTimeout,
		shutdown:   make(chan struct{}),
	}

	for i := range nbWorkers {
		workerPool.wg.Add(1)
		worker := &worker{
			id:         i,
			store:      store,
			breaker:    workerPool.breaker,
			jobTimeout: workerPool.jobTimeout,
		}

		go worker.Run(workerPool.queue, workerPool.shutdown, &workerPool.wg)
	}

	return workerPool
}

// The breaker drops jobs while open. After the cooldown, one job probes
// recovery; a successful probe closes the breaker and a failure reopens it.
type circuitBreaker struct {
	mu          sync.Mutex
	threshold   int64
	consecutive int64
	state       breakerState
	generation  uint64
	retryAt     time.Time
}

type breakerState uint8

const (
	breakerClosed breakerState = iota
	breakerOpen
	breakerHalfOpen
)

const breakerCooldown = 30 * time.Second

// allow returns a generation identifying the state in which a job was admitted.
func (b *circuitBreaker) allow() (uint64, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case breakerClosed:
		return b.generation, true
	case breakerOpen:
		if !time.Now().Before(b.retryAt) {
			b.state = breakerHalfOpen
			return b.generation, true
		}
	}
	return 0, false
}

func (b *circuitBreaker) recordSuccess(generation uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if generation != b.generation || b.state == breakerOpen {
		return
	}
	b.consecutive = 0
	if b.state == breakerHalfOpen {
		b.state = breakerClosed
		b.generation++
		slog.Info("Worker circuit breaker closed, resuming job processing")
	}
}

func (b *circuitBreaker) recordFailure(generation uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if generation != b.generation || b.state == breakerOpen {
		return
	}
	b.consecutive++
	if b.state == breakerHalfOpen || b.consecutive >= b.threshold {
		b.state = breakerOpen
		b.generation++
		b.retryAt = time.Now().Add(breakerCooldown)
		slog.Warn("Worker circuit breaker opened due to consecutive failures",
			slog.Int64("consecutive_failures", b.consecutive),
			slog.Int64("threshold", b.threshold),
			slog.Duration("cooldown", breakerCooldown),
		)
	}
}
