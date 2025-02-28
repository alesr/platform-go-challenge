package assets

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssetFactory_CreateChart(t *testing.T) {
	t.Parallel()

	factory := NewAssetFactory()

	testCases := []struct {
		name       string
		givenTitle string
		givenXAxis string
		givenYAxis string
		givenData  []float64
		expectErr  bool
	}{
		{
			name:       "valid chart creation",
			givenTitle: "Monthly Revenue",
			givenXAxis: "Month",
			givenYAxis: "Revenue",
			givenData:  []float64{100, 200, 300},
		},
		{
			name:       "empty title",
			givenTitle: "",
			givenXAxis: "Month",
			givenYAxis: "Revenue",
			givenData:  []float64{100, 200, 300},
		},
		{
			name:       "empty data",
			givenTitle: "Monthly Revenue",
			givenXAxis: "Month",
			givenYAxis: "Revenue",
			givenData:  []float64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := factory.CreateChart(tc.givenTitle, tc.givenXAxis, tc.givenYAxis, tc.givenData)

			require.NotEmpty(t, got.ID)
			assert.Equal(t, TypeAssetChart, got.Type())
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tc.givenTitle, got.Data.Title)
			assert.Equal(t, tc.givenXAxis, got.Data.XAxis)
			assert.Equal(t, tc.givenYAxis, got.Data.YAxis)
			assert.Equal(t, tc.givenData, got.Data.Data)

			// timestamps are within the last second
			assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
			assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
		})
	}
}

func TestAssetFactory_CreateInsight(t *testing.T) {
	t.Parallel()

	factory := NewAssetFactory()

	testCases := []struct {
		givenName string
		givenData string
	}{
		{
			givenName: "valid insight creation",
			givenData: "Revenue increased by 25% in Q4",
		},
		{
			givenName: "empty data",
			givenData: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.givenName, func(t *testing.T) {
			t.Parallel()

			got := factory.CreateInsight(tc.givenData)

			require.NotEmpty(t, got.ID)
			assert.Equal(t, TypeAssetInsight, got.Type())
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tc.givenData, got.Data.Insight)
			assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
			assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
		})
	}
}

func TestAssetFactory_CreateAudience(t *testing.T) {
	t.Parallel()

	factory := NewAssetFactory()

	testCases := []struct {
		givenName               string
		givenGender             string
		givenBirthCountry       string
		givenAgeMin             int
		givenAgeMax             int
		givenSocialMediaHours   int
		givenLastMonthPurchases int
	}{
		{
			givenName:               "valid audience creation",
			givenGender:             "Female",
			givenBirthCountry:       "US",
			givenAgeMin:             25,
			givenAgeMax:             34,
			givenSocialMediaHours:   3,
			givenLastMonthPurchases: 5,
		},
		{
			givenName:               "zero values",
			givenGender:             "",
			givenBirthCountry:       "",
			givenAgeMin:             0,
			givenAgeMax:             0,
			givenSocialMediaHours:   0,
			givenLastMonthPurchases: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.givenName, func(t *testing.T) {
			t.Parallel()

			got := factory.CreateAudience(
				tc.givenGender,
				tc.givenBirthCountry,
				tc.givenAgeMin,
				tc.givenAgeMax,
				tc.givenSocialMediaHours,
				tc.givenLastMonthPurchases,
			)

			require.NotEmpty(t, got.ID)
			assert.Equal(t, TypeAssetAudience, got.Type())
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, tc.givenGender, got.Data.Gender)
			assert.Equal(t, tc.givenBirthCountry, got.Data.BirthCountry)
			assert.Equal(t, tc.givenAgeMin, got.Data.AgeMin)
			assert.Equal(t, tc.givenAgeMax, got.Data.AgeMax)
			assert.Equal(t, tc.givenSocialMediaHours, got.Data.SocialMediaHours)
			assert.Equal(t, tc.givenLastMonthPurchases, got.Data.LastMonthPurchases)

			assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
			assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
		})
	}
}
