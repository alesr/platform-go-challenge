package assets

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

var (
	// Enumerate service errors

	ErrAssetNotFound = errors.New("asset not found")
)

// Repository defines the interface for asset storage operations.
// Exported so we can guarantee that the postgres implementation implements this interface.
type Repository interface {
	StoreAsset(ctx context.Context, asset Asseter) error
	ListAssets(ctx context.Context, params *ListAssetsParams) ([]Asseter, string, error)
}

// Service provides asset management operations including listing, storing, and managing user favorites.
type Service struct {
	logger     *slog.Logger
	repository Repository
}

// NewService instantiates a new assets service.
func NewService(logger *slog.Logger, repo Repository) *Service {
	return &Service{
		logger:     logger.WithGroup("assets-service"),
		repository: repo,
	}
}

// StoreAssets receives a list of assets and stores them in the repository.
// We only use this method for for populating the database with test data.
func (s *Service) StoreAssets(_ context.Context, assets []Asseter) error {
	// detach so we prevent writing interruption if context is canceled
	ctx, cancel := context.WithTimeout(context.Background(), BackgroundCtxTimeout)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	for _, asset := range assets {
		eg.Go(func() error {
			if err := s.repository.StoreAsset(ctx, asset); err != nil {
				return fmt.Errorf("could not store asset: %w", err)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("error while storing assets: %w", err)
	}
	return nil
}

// ListAssets returns a paginated list of assets.
func (s *Service) ListAssets(ctx context.Context, params *ListAssetsParams) ([]Asseter, string, error) {
	assets, nextPageToken, err := s.repository.ListAssets(ctx, params)
	if err != nil {
		return nil, "", fmt.Errorf("could not list assets: %w", err)
	}
	return assets, nextPageToken, nil
}
