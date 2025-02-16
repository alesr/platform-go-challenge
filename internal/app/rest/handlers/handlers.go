package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/alesr/resterr"
	"github.com/oklog/ulid/v2"
)

var (
	// Enumerate all possible errors returned by the handlers.
	// If this grows too large, consider moving it to errors.go

	ErrDescriptionMaxLen           = errors.New("description is too long")
	ErrFavoriteIDRequired          = errors.New("favorite id is required")
	ErrInvalidFavoriteAssetPayload = errors.New("invalid favorite asset request payload")
	ErrInvalidFavoriteID           = errors.New("invalid favorite id")
	ErrInvalidPageMaxResults       = errors.New("invalid page max results")
	ErrInvalidPageSize             = errors.New("invalid page size")
	ErrInvalidPageToken            = errors.New("invalid page token")
	ErrInvalidUserID               = errors.New("invalid user id")
	ErrUserIDRequired              = errors.New("user id is required")
)

type usersService interface {
	ListUsers(ctx context.Context) []users.User
}

type assetsService interface {
	ListAssets(ctx context.Context, params *assets.ListAssetsParams) ([]assets.Asseter, string, error)
}

type favoritesService interface {
	FavoriteAsset(ctx context.Context, params *favorites.FavoriteAssetParams) error
	FetchUserFavorites(ctx context.Context, userID string) ([]favorites.FavoriteAsset, error)
	UpdateFavorite(ctx context.Context, userID, favoriteID string, params *favorites.UpdateFavoriteParams) (*favorites.FavoriteAsset, error)
	DeleteFavorite(ctx context.Context, favoriteID, userID string) error
}

type errorHandler interface {
	Handle(ctx context.Context, w resterr.Writer, err error)
}

type Handler struct {
	logger       *slog.Logger
	errHandler   errorHandler
	usersSvc     usersService
	assetsSvc    assetsService
	favoritesSvc favoritesService
}

func New(
	logger *slog.Logger,
	errHandler errorHandler,
	usersSvc usersService,
	assetsSvc assetsService,
	favoritesSvc favoritesService,
) *Handler {
	return &Handler{
		logger:       logger.WithGroup("rest-handlers"),
		errHandler:   errHandler,
		usersSvc:     usersSvc,
		assetsSvc:    assetsSvc,
		favoritesSvc: favoritesSvc,
	}
}

func (h *Handler) Shutdown(ctx context.Context) error {
	// Shutdown services that need graceful shutdown.
	// In our case, the favorites service.
	if shutdownable, ok := h.favoritesSvc.(interface{ Shutdown(context.Context) error }); ok {
		if err := shutdownable.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown favorites service: %w", err)
		}
	}
	return nil
}

// Helper methods

func validateID(id string) error {
	if _, err := ulid.Parse(id); err != nil {
		return fmt.Errorf("could not validate id '%s': %w", id, err)
	}
	return nil
}
