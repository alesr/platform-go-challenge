package assets

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// AssetFactory is responsible for creating different types of assets
type AssetFactory struct{}

// NewAssetFactory creates a new instance of AssetFactory
func NewAssetFactory() *AssetFactory {
	return &AssetFactory{}
}

// CreateChart creates a new chart asset.
func (f *AssetFactory) CreateChart(title, xAxis, yAxis string, data []float64) ChartAsset {
	return ChartAsset{
		ID:        ulid.Make().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Data: chart{
			Title: title,
			XAxis: xAxis,
			YAxis: yAxis,
			Data:  data,
		},
		assetType: TypeAssetChart,
	}
}

// CreateInsight creates a new insight asset.
func (f *AssetFactory) CreateInsight(data string) InsightAsset {
	return InsightAsset{
		ID:        ulid.Make().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Data: insight{
			Insight: data,
		},
		assetType: TypeAssetInsight,
	}
}

// CreateAudience creates a new audience asset.
func (f *AssetFactory) CreateAudience(
	gender, birthCountry string, ageMin, ageMax, socialMediaHours, lastMonthPurchases int,
) AudienceAsset {
	return AudienceAsset{
		ID:        ulid.Make().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Data: audience{
			Gender:             gender,
			BirthCountry:       birthCountry,
			AgeMin:             ageMin,
			AgeMax:             ageMax,
			SocialMediaHours:   socialMediaHours,
			LastMonthPurchases: lastMonthPurchases,
		},
		assetType: TypeAssetAudience,
	}
}
