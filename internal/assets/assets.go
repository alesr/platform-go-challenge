package assets

import (
	"time"
)

const (
	// Enumerate asset types

	TypeAssetChart    AssetType = "CHART"
	TypeAssetInsight  AssetType = "INSIGHT"
	TypeAssetAudience AssetType = "AUDIENCE"

	BackgroundCtxTimeout = time.Second * 15
)

type (
	// Type aliases for brevity when working with generic assets.
	// Apart the interface and the asset type (string), these are
	// the only asset types that are part of the service API.
	ChartAsset    = asset[chart]
	InsightAsset  = asset[insight]
	AudienceAsset = asset[audience]

	// AssetType represent the type of an asset (Chart, Insight, or Audience)
	AssetType string

	// Asseter defines the interface an asset must implement
	// to be passed between services and layers.
	Asseter interface{ Type() AssetType }

	// assetData is our type constraint for asset data.
	// Any new asset type must be registered here.
	assetData interface {
		chart | insight | audience
	}
)

// asset defines the generic asset which implements the Asseter interface.
type asset[T assetData] struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      T
	assetType AssetType
}

// Type implements the Asseter interface.
func (a asset[T]) Type() AssetType {
	return a.assetType
}

// Chart defines the data structure of a chart asset.
type chart struct {
	Title string
	XAxis string
	YAxis string
	Data  []float64
}

// insight defines the data structure of an insight asset.
type insight struct{ Insight string }

// audience defines the data structure of an audience.
type audience struct {
	Gender             string
	BirthCountry       string
	AgeMin             int
	AgeMax             int
	SocialMediaHours   int
	LastMonthPurchases int
}

// ListAssetsParams defines pagination parameters for listing assets.
type ListAssetsParams struct {
	PageSize   int
	PageToken  string
	MaxResults int
}
