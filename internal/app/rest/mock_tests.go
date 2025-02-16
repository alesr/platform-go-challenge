package rest

import (
	"context"
	"net/http"
)

var fallbackHandlerFunc func(w http.ResponseWriter, r *http.Request)

type handlersMock struct {
	shutdownFunc         func(ctx context.Context) error
	listAssetsFunc       func() http.HandlerFunc
	listUsersFunc        func() http.HandlerFunc
	favoriteAssetFunc    func() http.HandlerFunc
	getuserFavoritesFunc func() http.HandlerFunc
	updateFavoriteFunc   func() http.HandlerFunc
	deleteFavoriteFunc   func() http.HandlerFunc
}

func (m *handlersMock) Shutdown(ctx context.Context) error {
	return m.shutdownFunc(ctx)
}

func (m *handlersMock) ListAssets() http.HandlerFunc {
	if m.listAssetsFunc == nil {
		return fallbackHandlerFunc
	}
	return m.listAssetsFunc()
}

func (m *handlersMock) ListUsers() http.HandlerFunc {
	if m.listUsersFunc == nil {
		return fallbackHandlerFunc
	}
	return m.listUsersFunc()
}

func (m *handlersMock) FavoriteAsset() http.HandlerFunc {
	if m.favoriteAssetFunc == nil {
		return fallbackHandlerFunc
	}
	return m.favoriteAssetFunc()
}

func (m *handlersMock) GetUserFavorites() http.HandlerFunc {
	if m.getuserFavoritesFunc == nil {
		return fallbackHandlerFunc
	}
	return m.getuserFavoritesFunc()
}

func (m *handlersMock) UpdateFavorite() http.HandlerFunc {
	if m.updateFavoriteFunc == nil {
		return fallbackHandlerFunc
	}
	return m.updateFavoriteFunc()
}

func (m *handlersMock) DeleteFavorite() http.HandlerFunc {
	if m.deleteFavoriteFunc == nil {
		return fallbackHandlerFunc
	}
	return m.deleteFavoriteFunc()
}
