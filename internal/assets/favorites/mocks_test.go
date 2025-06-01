package favorites

import (
	"context"

	"github.com/alesr/platform-go-challenge/internal/users"
)

// Repository mock

var _ Repository = &repoMock{}

type repoMock struct {
	storeFavoriteAssetFunc func(ctx context.Context, params *FavoriteAssetParams) error
	getuserfavoritesFunc   func(ctx context.Context, userID string) ([]FavoriteAsset, error)
	updatefavoriteFunc     func(ctx context.Context, favID, userID string, params *UpdateFavoriteParams) (*FavoriteAsset, error)
	deleteFavoriteFunc     func(ctx context.Context, favoriteID, userID string) error
}

func (m *repoMock) StoreFavoriteAsset(ctx context.Context, params *FavoriteAssetParams) error {
	return m.storeFavoriteAssetFunc(ctx, params)
}

func (m *repoMock) GetUserFavorites(ctx context.Context, userID string) ([]FavoriteAsset, error) {
	return m.getuserfavoritesFunc(ctx, userID)
}

func (m *repoMock) UpdateFavorite(ctx context.Context, favID, userID string, params *UpdateFavoriteParams) (*FavoriteAsset, error) {
	return m.updatefavoriteFunc(ctx, favID, userID, params)
}

func (m *repoMock) DeleteFavorite(ctx context.Context, favoriteID, userID string) error {
	return m.deleteFavoriteFunc(ctx, favoriteID, userID)
}

// User service mock

var _ usersService = &userSvcMock{}

type userSvcMock struct {
	fetchUserFunc func(ctx context.Context, id string) (*users.User, error)
}

func (m *userSvcMock) FetchUser(ctx context.Context, id string) (*users.User, error) {
	return m.fetchUserFunc(ctx, id)
}
