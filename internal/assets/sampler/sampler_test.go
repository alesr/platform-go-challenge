package sampler

import (
	"testing"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/stretchr/testify/assert"
)

func TestSampleAssets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		amount int
	}{
		{
			name:   "Sample 0 assets",
			amount: 0,
		},
		{
			name:   "Sample 10 assets",
			amount: 10,
		},
		{
			name:   "Sample 100 assets",
			amount: 100,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SampleAssets(tc.amount)

			assert.Equal(t, tc.amount, len(got))

			for _, asset := range got {
				assert.NotEmpty(t, asset.Type())

				switch asset.Type() {
				case assets.TypeAssetChart:
					chartAsset, ok := asset.(assets.ChartAsset)
					assert.True(t, ok)
					assert.NotEmpty(t, chartAsset.ID)
					assert.NotEmpty(t, chartAsset.Data.Title)
					assert.NotEmpty(t, chartAsset.Data.XAxis)
					assert.NotEmpty(t, chartAsset.Data.YAxis)
					assert.NotEmpty(t, chartAsset.Data.Data)

				case assets.TypeAssetInsight:
					insightAsset, ok := asset.(assets.InsightAsset)
					assert.True(t, ok)
					assert.NotEmpty(t, insightAsset.ID)
					assert.NotEmpty(t, insightAsset.Data.Insight)

				case assets.TypeAssetAudience:
					audienceAsset, ok := asset.(assets.AudienceAsset)
					assert.True(t, ok)
					assert.NotEmpty(t, audienceAsset.ID)
					assert.NotEmpty(t, audienceAsset.Data.Gender)
					assert.NotEmpty(t, audienceAsset.Data.BirthCountry)
					assert.GreaterOrEqual(t, audienceAsset.Data.AgeMax, audienceAsset.Data.AgeMin)
					assert.GreaterOrEqual(t, audienceAsset.Data.SocialMediaHours, 0)
					assert.GreaterOrEqual(t, audienceAsset.Data.LastMonthPurchases, 0)

				default:
					t.Errorf("unexpected asset type: %v", asset.Type())
				}

				// Common assertions for all asset types
				assert.NotZero(t, asset.(assets.Asseter).Type())
			}
		})
	}
}
