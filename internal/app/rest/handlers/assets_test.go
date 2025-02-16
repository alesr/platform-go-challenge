package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
	"github.com/alesr/resterr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAssets(t *testing.T) {
	t.Parallel()

	assetFactory := assets.NewAssetFactory()
	givenChart := assetFactory.CreateChart("Foo Chart", "Bar Axis", "Qux Axis", []float64{1, 2, 3})
	givenInsight := assetFactory.CreateInsight("Bar Insight")
	givenAudience := assetFactory.CreateAudience("male", "BR", 18, 35, 2, 5)

	testCases := []struct {
		name                      string
		givenListAssetsMockResult func() ([]assets.Asseter, string, error)
		expect                    httputil.Response[ListAssetsResponse]
		expectErr                 bool
	}{
		{
			name: "no assets",
			givenListAssetsMockResult: func() ([]assets.Asseter, string, error) {
				return []assets.Asseter{}, "foo-token", nil
			},
			expect: httputil.Response[ListAssetsResponse]{
				Status: "success",
				Data: ListAssetsResponse{
					Items:         []any{},
					NextPageToken: "foo-token",
				},
			},
		},
		{
			name: "chart asset",
			givenListAssetsMockResult: func() ([]assets.Asseter, string, error) {
				return []assets.Asseter{givenChart}, "chart-token", nil
			},
			expect: httputil.Response[ListAssetsResponse]{
				Status: "success",
				Data: ListAssetsResponse{
					Items: []any{
						chartResponse{
							ID:        givenChart.ID,
							Type:      string(givenChart.Type()),
							CreatedAt: givenChart.CreatedAt,
							UpdatedAt: givenChart.UpdatedAt,
							Data: chartAssetResponse{
								Title: givenChart.Data.Title,
								XAxis: givenChart.Data.XAxis,
								YAxis: givenChart.Data.YAxis,
								Data:  givenChart.Data.Data,
							},
						},
					},
					NextPageToken: "chart-token",
				},
			},
		},
		{
			name: "insight asset",
			givenListAssetsMockResult: func() ([]assets.Asseter, string, error) {
				return []assets.Asseter{givenInsight}, "insight-token", nil
			},
			expect: httputil.Response[ListAssetsResponse]{
				Status: "success",
				Data: ListAssetsResponse{
					Items: []any{
						insightResponse{
							ID:        givenInsight.ID,
							Type:      string(givenInsight.Type()),
							CreatedAt: givenInsight.CreatedAt,
							UpdatedAt: givenInsight.UpdatedAt,
							Data: insightAssetResponse{
								Insight: givenInsight.Data.Insight,
							},
						},
					},
					NextPageToken: "insight-token",
				},
			},
		},
		{
			name: "audience asset",
			givenListAssetsMockResult: func() ([]assets.Asseter, string, error) {
				return []assets.Asseter{givenAudience}, "audience-token", nil
			},
			expect: httputil.Response[ListAssetsResponse]{
				Status: "success",
				Data: ListAssetsResponse{
					Items: []any{
						audienceResponse{
							ID:        givenAudience.ID,
							Type:      string(givenAudience.Type()),
							CreatedAt: givenAudience.CreatedAt,
							UpdatedAt: givenAudience.UpdatedAt,
							Data: audienceAssetResponse{
								Gender:             givenAudience.Data.Gender,
								BirthCountry:       givenAudience.Data.BirthCountry,
								AgeMin:             givenAudience.Data.AgeMin,
								AgeMax:             givenAudience.Data.AgeMax,
								SocialMediaHours:   givenAudience.Data.SocialMediaHours,
								LastMonthPurchases: givenAudience.Data.LastMonthPurchases,
							},
						},
					},
					NextPageToken: "audience-token",
				},
			},
		},
		{
			name: "all asset types",
			givenListAssetsMockResult: func() ([]assets.Asseter, string, error) {
				return []assets.Asseter{givenChart, givenInsight, givenAudience}, "all-token", nil
			},
			expect: httputil.Response[ListAssetsResponse]{
				Status: "success",
				Data: ListAssetsResponse{
					Items: []any{
						chartResponse{
							ID:        givenChart.ID,
							Type:      string(givenChart.Type()),
							CreatedAt: givenChart.CreatedAt,
							UpdatedAt: givenChart.UpdatedAt,
							Data: chartAssetResponse{
								Title: givenChart.Data.Title,
								XAxis: givenChart.Data.XAxis,
								YAxis: givenChart.Data.YAxis,
								Data:  givenChart.Data.Data,
							},
						},
						insightResponse{
							ID:        givenInsight.ID,
							Type:      string(givenInsight.Type()),
							CreatedAt: givenInsight.CreatedAt,
							UpdatedAt: givenInsight.UpdatedAt,
							Data: insightAssetResponse{
								Insight: givenInsight.Data.Insight,
							},
						},
						audienceResponse{
							ID:        givenAudience.ID,
							Type:      string(givenAudience.Type()),
							CreatedAt: givenAudience.CreatedAt,
							UpdatedAt: givenAudience.UpdatedAt,
							Data: audienceAssetResponse{
								Gender:             givenAudience.Data.Gender,
								BirthCountry:       givenAudience.Data.BirthCountry,
								AgeMin:             givenAudience.Data.AgeMin,
								AgeMax:             givenAudience.Data.AgeMax,
								SocialMediaHours:   givenAudience.Data.SocialMediaHours,
								LastMonthPurchases: givenAudience.Data.LastMonthPurchases,
							},
						},
					},
					NextPageToken: "all-token",
				},
			},
		},
		{
			name: "error from service",
			givenListAssetsMockResult: func() ([]assets.Asseter, string, error) {
				return nil, "", assert.AnError
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assetsSvc := &assetsSvcMock{
				listAssetsFunc: func(ctx context.Context, params *assets.ListAssetsParams) ([]assets.Asseter, string, error) {
					return tc.givenListAssetsMockResult()
				},
			}

			errHandler := &errorHandlerMock{
				handleFunc: func(ctx context.Context, w resterr.Writer, err error) {
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				},
			}

			handler := Handler{
				assetsSvc:  assetsSvc,
				errHandler: errHandler,
			}

			req := httptest.NewRequest(http.MethodGet, "/?pageSize=10&maxResults=100", nil)
			rec := httptest.NewRecorder()

			handler.ListAssets().ServeHTTP(rec, req)

			if tc.expectErr {
				assert.Equal(t, http.StatusInternalServerError, rec.Code)
				return
			}

			assert.Equal(t, http.StatusOK, rec.Code)

			var resp httputil.Response[ListAssetsResponse]
			err := json.NewDecoder(rec.Body).Decode(&resp)
			require.NoError(t, err)

			assert.Equal(t, tc.expect.Status, resp.Status)
			assert.Equal(t, tc.expect.Data.NextPageToken, resp.Data.NextPageToken)
			assert.Len(t, resp.Data.Items, len(tc.expect.Data.Items))
		})
	}
}

func TestListAssets_Pagination(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		givenURL         string
		expectPageSize   int
		expectPageToken  string
		expectMaxResults int
		expectStatusCode int
		expectErr        error
	}{
		{
			name:             "custom pagination values",
			givenURL:         "/?pageSize=20&maxResults=200",
			expectPageSize:   20,
			expectPageToken:  "",
			expectMaxResults: 200,
			expectStatusCode: http.StatusOK,
		},
		{
			name:             "invalid page size",
			givenURL:         "/?pageSize=invalid&maxResults=100",
			expectStatusCode: http.StatusBadRequest,
			expectErr:        ErrInvalidPageSize,
		},
		{
			name:             "invalid max results",
			givenURL:         "/?pageSize=10&maxResults=invalid",
			expectStatusCode: http.StatusBadRequest,
			expectErr:        ErrInvalidPageMaxResults,
		},
		{
			name:             "invalid page token",
			givenURL:         "/?pageSize=10&maxResults=100&pageToken=invalid",
			expectStatusCode: http.StatusBadRequest,
			expectErr:        ErrInvalidPageToken,
		},
		{
			name:             "to default pagination values",
			givenURL:         "/?pageSize=10&maxResults=100&pageToken=invalid",
			expectStatusCode: http.StatusBadRequest,
			expectErr:        ErrInvalidPageToken,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				capturedParams *assets.ListAssetsParams
				capturedError  error
			)

			assetsSvc := &assetsSvcMock{
				listAssetsFunc: func(ctx context.Context, params *assets.ListAssetsParams) ([]assets.Asseter, string, error) {
					if tc.expectStatusCode == http.StatusOK {
						capturedParams = params
						return []assets.Asseter{}, "", nil
					}
					return nil, "", assert.AnError
				},
			}

			errHandler := &errorHandlerMock{
				handleFunc: func(ctx context.Context, w resterr.Writer, err error) {
					capturedError = err
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				},
			}

			handler := Handler{
				assetsSvc:  assetsSvc,
				errHandler: errHandler,
			}

			req := httptest.NewRequest(http.MethodGet, tc.givenURL, nil)
			rec := httptest.NewRecorder()

			handler.ListAssets().ServeHTTP(rec, req)

			assert.Equal(t, tc.expectStatusCode, rec.Code)

			if tc.expectErr != nil {
				assert.True(t, errors.Is(capturedError, tc.expectErr))
				return
			}

			if capturedParams != nil {
				assert.Equal(t, tc.expectPageSize, capturedParams.PageSize)
				assert.Equal(t, tc.expectPageToken, capturedParams.PageToken)
				assert.Equal(t, tc.expectMaxResults, capturedParams.MaxResults)
			}
		})
	}
}
