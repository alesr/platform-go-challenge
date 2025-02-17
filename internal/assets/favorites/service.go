package favorites

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/users"
)

var (
	// Enumerate service errors

	ErrFavoriteAssetNotFound = errors.New("favorite asset not found")
	ErrInvalidAssetID        = errors.New("invalid asset id")
)

type Repository interface {
	StoreFavoriteAsset(ctx context.Context, params *FavoriteAssetParams) error
	GetUserFavorites(ctx context.Context, userID string) ([]FavoriteAsset, error)
	UpdateFavorite(ctx context.Context, favID, userID string, params *UpdateFavoriteParams) (*FavoriteAsset, error)
	DeleteFavorite(ctx context.Context, favoriteID, userID string) error
}

type usersService interface {
	FetchUser(ctx context.Context, id string) (*users.User, error)
}

// Service provides asset favorite service for managing user favorites.
type Service struct {
	logger     *slog.Logger
	repository Repository
	usersSvc   usersService
	workerPool *workerPool
}

const (
	workerpoolJobs = 10
)

// NewService creates a new asset favorite service.
func NewService(logger *slog.Logger, repo Repository, usersSvc usersService) *Service {
	return &Service{
		logger:     logger.WithGroup("assets-service"),
		repository: repo,
		usersSvc:   usersSvc,
		workerPool: newWorkerPool(logger, workerpoolJobs, repo.StoreFavoriteAsset),
	}
}

// FavoriteAsset marks an asset as favorite for a user.
func (s *Service) FavoriteAsset(ctx context.Context, params *FavoriteAssetParams) error {
	// In a real-case scenario, peharps we could get the user ID from the context after
	// some auth mechanism. This would help us  decoupling assets from users service.
	if _, err := s.usersSvc.FetchUser(ctx, params.UserID); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("could not fetch user id '%s': %w", params.UserID, err)
	}

	// I'm assuming the user doesn't need to get a confirmation that the asset was marked as favorite.
	// The goroutines will eventually die due to context timeout.
	// But we might consider using a worker pool here to avoid creating too many of them in case of a high volume of requests.
	// I'm a bit concerned about being a feature-creep here just for showing off.
	// Also, "Premature optimization is the root of all evil."
	s.workerPool.submit(params)
	return nil
}

// FetchUserFavorites fetches the user's favorite assets.
func (s *Service) FetchUserFavorites(ctx context.Context, userID string) ([]FavoriteAsset, error) {
	// We could live without this check and just return an empty slice if we can't find any favorites for this user.
	// But only the big picture of the system and business requirements would tell us the appropriate approach here.
	if _, err := s.usersSvc.FetchUser(ctx, userID); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("could not fetch user: %w", err)
	}

	favorites, err := s.repository.GetUserFavorites(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get user favorites: %w", err)
	}
	return favorites, nil
}

// UpdateFavorite updates a user's favorite asset.
// The favorite must belong to the user that created it which will be checked by the repository.
func (s *Service) UpdateFavorite(ctx context.Context, favID, userID string, params *UpdateFavoriteParams) (*FavoriteAsset, error) {
	// Detach context to prevent cancellation while writing data.
	ctx, cancel := context.WithTimeout(context.Background(), assets.BackgroundCtxTimeout)
	defer cancel()

	favAsset, err := s.repository.UpdateFavorite(ctx, favID, userID, params)
	if err != nil {
		return nil, fmt.Errorf("could not update favorite: %w", err)
	}
	return favAsset, nil
}

// DeleteFavorite deletes a user's favorite asset.
// The favorite must belong to the user that created it which will be checked by the repository.
func (s *Service) DeleteFavorite(ctx context.Context, favoriteID, userID string) error {
	if _, err := s.usersSvc.FetchUser(ctx, userID); err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("could not fetch user: %w", err)
	}
	if err := s.repository.DeleteFavorite(ctx, favoriteID, userID); err != nil {
		return fmt.Errorf("could not delete favorite: %w", err)
	}
	return nil
}

func (s *Service) handleAsyncFavorite(params *FavoriteAssetParams) {
	// Create a context detached from the original to avoid cancellation due to the request being closed.
	// Yet, use a timeout to prevent goroutines from running indefinitely (pgx will handle it)
	ctx, cancel := context.WithTimeout(context.Background(), assets.BackgroundCtxTimeout)
	defer cancel()

	if err := s.repository.StoreFavoriteAsset(ctx, params); err != nil {
		if errors.Is(err, assets.ErrAssetNotFound) {
			s.logger.Error("Asset not found when trying to favorite",
				slog.String("user_id", params.UserID),
				slog.String("asset_id", params.AssetID),
				slog.String("error", err.Error()),
			)
			return
		}
		s.logger.Error("Failed to store favorite asset",
			"error", err,
			"user_id", params.UserID,
			"asset_id", params.AssetID,
		)
		// NOTE: I should stop writing tests and collect metrics to generate
		// alerts in case we see too many errors here =]
	}
}

func (s *Service) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down favorites service")

	doneCh := make(chan struct{})

	go func() {
		s.workerPool.stop()
		close(doneCh)
	}()

	// wait for shutdown to complete or context to timeout
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		return nil
	}
}
