package favorites

import (
	"time"

	"github.com/alesr/platform-go-challenge/internal/assets"
)

// FavoriteAsset defines the information contained in asset marked as favorite.
type FavoriteAsset struct {
	ID          string
	UserID      string
	AssetID     string
	AssetType   assets.AssetType
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FavoriteAssetParams defines the information needed to mark an asset as favorite.
type FavoriteAssetParams struct {
	UserID      string
	AssetID     string
	Description string
}

// UpdateFavoriteParams defines the information needed
// to update an existing asset marked as favorite.
type UpdateFavoriteParams struct {
	Description string
}
