package assets

import (
	"context"
)

// Repository mock

var _ Repository = &repoMock{}

type repoMock struct {
	storeAssetFunc func(ctx context.Context, asset Asseter) error
	listAssetsFunc func(ctx context.Context, params *ListAssetsParams) ([]Asseter, string, error)
}

func (m *repoMock) StoreAsset(ctx context.Context, asset Asseter) error {
	return m.storeAssetFunc(ctx, asset)
}

func (m *repoMock) ListAssets(ctx context.Context, params *ListAssetsParams) ([]Asseter, string, error) {
	return m.listAssetsFunc(ctx, params)
}
