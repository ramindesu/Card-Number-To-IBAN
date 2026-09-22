package queue

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoundedExecutorRejectsWorkWhenQueueIsFull(t *testing.T) {
	t.Parallel()

	executor, err := NewBoundedExecutor(Config{
		Capacity:          1,
		Workers:           1,
		RequestsPerSecond: 100,
		Burst:             1,
	})
	require.NoError(t, err)
	t.Cleanup(executor.Close)

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	go func() {
		defer waitGroup.Done()
		_ = executor.Execute(context.Background(), func(context.Context) error {
			close(firstStarted)
			<-releaseFirst
			return nil
		})
	}()
	<-firstStarted

	go func() {
		defer waitGroup.Done()
		_ = executor.Execute(context.Background(), func(context.Context) error {
			return nil
		})
	}()

	require.Eventually(t, func() bool {
		return len(executor.jobs) == 1
	}, time.Second, time.Millisecond)

	err = executor.Execute(context.Background(), func(context.Context) error {
		return nil
	})
	assert.ErrorIs(t, err, ErrFull)

	close(releaseFirst)
	waitGroup.Wait()
}

func TestTokenBucketHonorsContextWhileWaiting(t *testing.T) {
	t.Parallel()

	bucket, err := newTokenBucket(1, 1)
	require.NoError(t, err)
	t.Cleanup(bucket.Close)
	require.NoError(t, bucket.Wait(context.Background()))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err = bucket.Wait(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestInlineExecutorReturnsTaskError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("task failed")
	err := (InlineExecutor{}).Execute(context.Background(), func(context.Context) error {
		return expectedError
	})

	assert.ErrorIs(t, err, expectedError)
}
