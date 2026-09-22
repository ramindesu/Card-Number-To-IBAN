package queue

import (
	"context"
	"errors"
	"sync"
)

var (
	// ErrFull indicates that the bounded queue cannot accept more work.
	ErrFull = errors.New("card-to-IBAN queue is full")
	// ErrClosed indicates that the queue is shutting down.
	ErrClosed = errors.New("card-to-IBAN queue is closed")
)

// Task is one rate-limited unit of work.
type Task func(context.Context) error

// Executor admits a task to the queue and waits for its result.
type Executor interface {
	Execute(context.Context, Task) error
}

// Config controls bounded capacity, concurrency, and provider request rate.
type Config struct {
	Capacity          int
	Workers           int
	RequestsPerSecond int
	Burst             int
}

type job struct {
	ctx    context.Context
	task   Task
	result chan error
}

// BoundedExecutor is an in-process worker queue with token-bucket rate limiting.
type BoundedExecutor struct {
	jobs    chan job
	limiter *tokenBucket
	workers sync.WaitGroup
	mutex   sync.RWMutex
	closed  bool
}

// NewBoundedExecutor starts a bounded worker pool.
func NewBoundedExecutor(config Config) (*BoundedExecutor, error) {
	if config.Capacity <= 0 {
		return nil, errors.New("queue capacity must be positive")
	}
	if config.Workers <= 0 {
		return nil, errors.New("queue worker count must be positive")
	}

	limiter, err := newTokenBucket(config.RequestsPerSecond, config.Burst)
	if err != nil {
		return nil, err
	}

	executor := &BoundedExecutor{
		jobs:    make(chan job, config.Capacity),
		limiter: limiter,
	}

	executor.workers.Add(config.Workers)
	for range config.Workers {
		go executor.work()
	}

	return executor, nil
}

// Execute queues a task without blocking admission when capacity is exhausted.
func (e *BoundedExecutor) Execute(ctx context.Context, task Task) error {
	if ctx == nil {
		return errors.New("queue context is required")
	}
	if task == nil {
		return errors.New("queue task is required")
	}

	queuedJob := job{
		ctx:    ctx,
		task:   task,
		result: make(chan error, 1),
	}

	e.mutex.RLock()
	if e.closed {
		e.mutex.RUnlock()
		return ErrClosed
	}

	select {
	case e.jobs <- queuedJob:
		e.mutex.RUnlock()
	case <-ctx.Done():
		e.mutex.RUnlock()
		return ctx.Err()
	default:
		e.mutex.RUnlock()
		return ErrFull
	}

	select {
	case err := <-queuedJob.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close stops admission and waits for workers to exit.
func (e *BoundedExecutor) Close() {
	e.mutex.Lock()
	if e.closed {
		e.mutex.Unlock()
		return
	}
	e.closed = true
	close(e.jobs)
	e.mutex.Unlock()

	e.limiter.Close()
	e.workers.Wait()
}

func (e *BoundedExecutor) work() {
	defer e.workers.Done()

	for queuedJob := range e.jobs {
		if err := queuedJob.ctx.Err(); err != nil {
			queuedJob.result <- err
			continue
		}

		if err := e.limiter.Wait(queuedJob.ctx); err != nil {
			queuedJob.result <- err
			continue
		}

		queuedJob.result <- queuedJob.task(queuedJob.ctx)
	}
}

// InlineExecutor is useful for tests and deployments that explicitly disable
// queueing while keeping the service contract unchanged.
type InlineExecutor struct{}

func (InlineExecutor) Execute(ctx context.Context, task Task) error {
	return task(ctx)
}

var _ Executor = (*BoundedExecutor)(nil)
var _ Executor = InlineExecutor{}
