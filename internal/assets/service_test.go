package assets

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/alesr/platform-go-challenge/internal/pkg/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	logger := logutil.NewNoop()
	repo := repoMock{}

	got := NewService(logger, &repo)
	require.NotNil(t, got)

	assert.Equal(t, logger.WithGroup("assets-service"), got.logger)
	assert.Equal(t, &repo, got.repository)
}

func TestService_ListAssets(t *testing.T) {
	t.Parallel()

	givenParams := ListAssetsParams{
		PageSize:   100,
		MaxResults: 1000,
		PageToken:  "foo-page-tkn",
	}

	// create test assets
	assets := createTestAssetsHelper(t)

	testCases := []struct {
		name            string
		givenMockResult func() ([]Asseter, string, error)
		expectedAssets  []Asseter
		expectedToken   string
		expectedError   error
	}{
		{
			name: "success",
			givenMockResult: func() ([]Asseter, string, error) {
				return []Asseter{assets.charts[0], assets.insights[0]}, "bar-page-tkn", nil
			},
			expectedAssets: []Asseter{assets.charts[0], assets.insights[0]},
			expectedToken:  "bar-page-tkn",
		},
		{
			name: "repository returns error",
			givenMockResult: func() ([]Asseter, string, error) {
				return nil, "", assert.AnError
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repoCalled bool
			repo := repoMock{
				listAssetsFunc: func(ctx context.Context, params *ListAssetsParams) ([]Asseter, string, error) {
					repoCalled = true

					assert.Equal(t, givenParams.PageSize, params.PageSize)
					assert.Equal(t, givenParams.MaxResults, params.MaxResults)
					assert.Equal(t, givenParams.PageToken, params.PageToken)
					return tc.givenMockResult()
				},
			}

			svc := Service{repository: &repo}

			assets, nextToken, err := svc.ListAssets(context.TODO(), &givenParams)

			if tc.expectedError != nil {
				assert.ErrorIs(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
			require.True(t, repoCalled)
			assert.Equal(t, tc.expectedAssets, assets)
			assert.Equal(t, tc.expectedToken, nextToken)
		})
	}
}

func TestStoreAssets(t *testing.T) {
	t.Parallel()

	assets := createTestAssetsHelper(t)

	testCases := []struct {
		name        string
		givenAssets []Asseter
		expectError bool
		repoError   error
	}{
		{
			name: "successfully store assets",
			givenAssets: []Asseter{
				assets.charts[0],
				assets.insights[0],
				assets.audiences[0],
			},
			expectError: false,
			repoError:   nil,
		},
		{
			name: "repository returns an error",
			givenAssets: []Asseter{
				assets.charts[0],
				assets.insights[0],
			},
			expectError: true,
			repoError:   assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var storedAssets atomic.Int32
			repo := repoMock{
				storeAssetFunc: func(ctx context.Context, asset Asseter) error {
					// if we expect an error, return it on the first asset
					if tc.repoError != nil && storedAssets.Load() == 0 {
						return tc.repoError
					}
					storedAssets.Add(1)
					return nil
				},
			}

			svc := Service{repository: &repo}

			err := svc.StoreAssets(context.TODO(), tc.givenAssets)

			if tc.expectError {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.repoError)
				// verify that we stopped processing after the error
				assert.Equal(t, int32(0), storedAssets.Load())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, int32(len(tc.givenAssets)), storedAssets.Load())
		})
	}
}

type testAssets struct {
	charts    []ChartAsset
	insights  []InsightAsset
	audiences []AudienceAsset
}

func createTestAssetsHelper(t *testing.T) testAssets {
	t.Helper()

	factory := NewAssetFactory()
	return testAssets{
		charts: []ChartAsset{
			factory.CreateChart("Should I get hired?", "contributions", "bugs", []float64{100, 3000, 2}),
			// Btw, now I realized that I should have chosen two dimensions for the data points.
			factory.CreateChart("Colleagues coming for a gyros in Crete", "weeks", "number of visits", []float64{1500, 2500, 3500}),
		},
		insights: []InsightAsset{
			factory.CreateInsight("I'm a very chill and friendly dev"),
			factory.CreateInsight("I think I'm gonna know if you read the tests throughly =]"),
		},
		audiences: []AudienceAsset{
			factory.CreateAudience("Female", "BR", 25, 34, 3, 5),
			factory.CreateAudience("Male", "IT", 18, 24, 5, 8),
		},
	}
}
