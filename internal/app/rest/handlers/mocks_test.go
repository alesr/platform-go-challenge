package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/alesr/resterr"
)

// User service

var _ usersService = &usersSvcMock{}

type usersSvcMock struct {
	listUsersFunc func(ctx context.Context) []users.User
}

func (m *usersSvcMock) ListUsers(ctx context.Context) []users.User { return m.listUsersFunc(ctx) }

// Assets service

var _ assetsService = &assetsSvcMock{}

type assetsSvcMock struct {
	listAssetsFunc func(ctx context.Context, params *assets.ListAssetsParams) ([]assets.Asseter, string, error)
}

func (m *assetsSvcMock) ListAssets(ctx context.Context, params *assets.ListAssetsParams) ([]assets.Asseter, string, error) {
	return m.listAssetsFunc(ctx, params)
}

// Favorites service

var _ favoritesService = &favoritesSvcMock{}

type favoritesSvcMock struct {
	favoriteAssetFunc      func(ctx context.Context, params *favorites.FavoriteAssetParams) error
	fetchUserFavoritesFunc func(ctx context.Context, userID string) ([]favorites.FavoriteAsset, error)
	updateFavoriteFunc     func(ctx context.Context, userID, assetID string, params *favorites.UpdateFavoriteParams) (*favorites.FavoriteAsset, error)
	deleteFavoriteFunc     func(ctx context.Context, favoriteID, userID string) error
}

func (m *favoritesSvcMock) FavoriteAsset(ctx context.Context, params *favorites.FavoriteAssetParams) error {
	return m.favoriteAssetFunc(ctx, params)
}

func (m *favoritesSvcMock) FetchUserFavorites(ctx context.Context, userID string) ([]favorites.FavoriteAsset, error) {
	return m.fetchUserFavoritesFunc(ctx, userID)
}

func (m *favoritesSvcMock) UpdateFavorite(ctx context.Context, userID, assetID string, params *favorites.UpdateFavoriteParams) (*favorites.FavoriteAsset, error) {
	return m.updateFavoriteFunc(ctx, userID, assetID, params)
}

func (m *favoritesSvcMock) DeleteFavorite(ctx context.Context, favoriteID, userID string) error {
	return m.deleteFavoriteFunc(ctx, favoriteID, userID)
}

// error handler

var _ errorHandler = &errorHandlerMock{}

type errorHandlerMock struct {
	handleFunc func(ctx context.Context, w resterr.Writer, err error)
}

func (m *errorHandlerMock) Handle(ctx context.Context, w resterr.Writer, err error) {
	if m.handleFunc != nil {
		m.handleFunc(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
