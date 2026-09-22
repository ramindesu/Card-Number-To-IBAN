package queue

import (
	"context"
	"errors"
	"sync"
	"time"
)

type tokenBucket struct {
	tokens    chan struct{}
	stop      chan struct{}
	closeOnce sync.Once
}

func newTokenBucket(requestsPerSecond, burst int) (*tokenBucket, error) {
	if requestsPerSecond <= 0 {
		return nil, errors.New("rate limit requests per second must be positive")
	}
	if burst <= 0 {
		return nil, errors.New("rate limit burst must be positive")
	}

	refillInterval := time.Second / time.Duration(requestsPerSecond)
	if refillInterval <= 0 {
		return nil, errors.New("rate limit requests per second is too large")
	}

	bucket := &tokenBucket{
		tokens: make(chan struct{}, burst),
		stop:   make(chan struct{}),
	}
	for range burst {
		bucket.tokens <- struct{}{}
	}

	go bucket.refill(refillInterval)
	return bucket, nil
}

func (b *tokenBucket) Wait(ctx context.Context) error {
	select {
	case <-b.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-b.stop:
		return ErrClosed
	}
}

func (b *tokenBucket) Close() {
	b.closeOnce.Do(func() {
		close(b.stop)
	})
}

func (b *tokenBucket) refill(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case b.tokens <- struct{}{}:
			default:
			}
		case <-b.stop:
			return
		}
	}
}
