package favorites

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/logutil"
	"github.com/stretchr/testify/assert"
)

func TestWorkerPool(t *testing.T) {
	t.Parallel()

	concurrency := 2

	testCases := []struct {
		name            string
		givenNumOfTasks int
		expectErr       bool
	}{
		{
			name:            "process multiple tasks",
			givenNumOfTasks: 5,
			expectErr:       false,
		},
		{
			name:            "process tasks with errors",
			givenNumOfTasks: 5,
			expectErr:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var processedTasks atomic.Int32
			storeFunc := func(ctx context.Context, params *FavoriteAssetParams) error {
				processedTasks.Add(1)
				if tc.expectErr {
					return assert.AnError
				}
				return nil
			}

			wp := newWorkerPool(logutil.NewNoop(), concurrency, storeFunc)

			// send tasks
			for i := 0; i < tc.givenNumOfTasks; i++ {
				wp.submit(&FavoriteAssetParams{
					UserID:  fmt.Sprintf("user-%d", i),
					AssetID: fmt.Sprintf("asset-%d", i),
				})
			}

			// This test might fail if we add moar tasks
			// and the workers are stopped before all tasks are processed.
			// If you see this failing, sleep for a bit.
			// time.Sleep(100 * time.Millisecond)
			// Looking forward to try https://danp.net/posts/synctest-experiment/

			wp.stop()
			assert.Equal(t, int32(tc.givenNumOfTasks), processedTasks.Load())
		})
	}
}

func TestWorkerPool_StopBeforeSubmit(t *testing.T) {
	t.Parallel()

	storeFunc := func(ctx context.Context, params *FavoriteAssetParams) error {
		return nil
	}

	concurrency := 1
	wp := newWorkerPool(logutil.NewNoop(), concurrency, storeFunc)

	// stop the worker pool immediately
	wp.stop()

	var submitted bool
	assert.Eventually(t, func() bool {
		wp.submit(&FavoriteAssetParams{
			UserID:  "user-1",
			AssetID: "asset-1",
		})
		submitted = true
		return submitted
	}, time.Second, 100*time.Millisecond)

	// submission should not block after stop
	// since we use wg to wait for all tasks to be processed
	assert.True(t, submitted)
}

func TestWorkerPool_ContextCancel(t *testing.T) {
	t.Parallel()

	var ctxCanceled bool
	mockStoreFavorite := func(ctx context.Context, params *FavoriteAssetParams) error {
		if ctx != nil && ctx.Err() == nil {
			ctxCanceled = true
		}
		return nil
	}

	concurrency := 1
	wp := newWorkerPool(logutil.NewNoop(), concurrency, mockStoreFavorite)

	wp.submit(&FavoriteAssetParams{
		UserID:  "user-1",
		AssetID: "asset-1",
	})

	wp.stop()

	assert.True(t, ctxCanceled)
}
