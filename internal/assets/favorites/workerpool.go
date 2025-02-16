package favorites

import (
	"context"
	"log/slog"
	"sync"

	"github.com/alesr/platform-go-challenge/internal/assets"
)

type storeFavoriteFunc func(ctx context.Context, params *FavoriteAssetParams) error

type favoriteTask struct {
	params        *FavoriteAssetParams
	storeFavorite storeFavoriteFunc
}

type workerPool struct {
	tasksCh       chan *FavoriteAssetParams
	logger        *slog.Logger
	jobs          int
	storeFavorite storeFavoriteFunc
	wg            sync.WaitGroup
	done          chan struct{}
}

func newWorkerPool(logger *slog.Logger, jobs int, storeFavorite storeFavoriteFunc) *workerPool {
	wp := &workerPool{
		tasksCh:       make(chan *FavoriteAssetParams),
		logger:        logger.WithGroup("worker-pool"),
		jobs:          jobs,
		storeFavorite: storeFavorite,
		done:          make(chan struct{}),
	}

	wp.wg.Add(jobs)
	for i := 0; i < jobs; i++ {
		go wp.worker()
	}
	return wp
}

func (wp *workerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.done:
			return
		case params, ok := <-wp.tasksCh:
			if !ok {
				return
			}

			wp.logger.Debug(
				"processing favorite asset",
				slog.String("asset_id", params.AssetID),
				slog.String("user_id", params.UserID),
			)

			ctx, cancel := context.WithTimeout(context.Background(), assets.BackgroundCtxTimeout)

			if err := wp.storeFavorite(ctx, params); err != nil {
				wp.logger.Error("failed to store favorite",
					"user_id", params.UserID,
					"asset_id", params.AssetID,
					"error", err,
				)
			}
			cancel()
		}
	}
}

func (wp *workerPool) submit(params *FavoriteAssetParams) {
	select {
	case <-wp.done:
		// if the pool is stopped, walk away
		return
	default:
		select {
		case <-wp.done: // double-checked locking pattern
			return
		case wp.tasksCh <- params:
		}
	}
}

func (wp *workerPool) stop() {
	close(wp.done)
	close(wp.tasksCh)
	wp.wg.Wait()
}
