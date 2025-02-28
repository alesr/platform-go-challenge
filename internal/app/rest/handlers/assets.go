package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
	"github.com/oklog/ulid/v2"
)

type (
	assetResponse interface {
		chartAssetResponse | insightAssetResponse | audienceAssetResponse
	}

	// fields must be exported for json marshalling
	asset[T assetResponse] struct {
		ID        string    `json:"id"`
		Type      string    `json:"type"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Data      T         `json:"data"`
	}

	chartAssetResponse struct {
		Title string    `json:"title"`
		XAxis string    `json:"x_axis"`
		YAxis string    `json:"y_axis"`
		Data  []float64 `json:"data"`
	}

	insightAssetResponse struct {
		Insight string `json:"insight"`
	}

	audienceAssetResponse struct {
		Gender             string `json:"gender"`
		BirthCountry       string `json:"birth_country"`
		AgeMin             int    `json:"age_min"`
		AgeMax             int    `json:"age_max"`
		SocialMediaHours   int    `json:"social_media_hours"`
		LastMonthPurchases int    `json:"last_month_purchases"`
	}

	// type aliases for improved readability
	chartResponse    = asset[chartAssetResponse]
	insightResponse  = asset[insightAssetResponse]
	audienceResponse = asset[audienceAssetResponse]

	// ListAssetsResponse defines the data structure for listing assets
	ListAssetsResponse struct {
		Items         []any  `json:"items"`
		NextPageToken string `json:"next_page_token,omitempty"`
	}
)

// ListAssets returns a list of assets
func (h *Handler) ListAssets() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params, err := h.parseListAssetsParams(r)
		if err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not parse list assets params: %w", err))
			return
		}

		assets, nextPageToken, err := h.assetsSvc.ListAssets(r.Context(), params)
		if err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not list assets: %w", err))
			return
		}

		items := make([]any, 0, len(assets))
		for _, asset := range assets {
			if transportItem := toTransportAsset(asset); transportItem != nil {
				items = append(items, transportItem)
			}
		}

		httputil.RespondWithJSON(w, http.StatusOK, ListAssetsResponse{
			Items:         items,
			NextPageToken: nextPageToken,
		})
	}
}

const (
	defaultPageSize   = 10
	defaultMaxResults = 100
)

func (h *Handler) parseListAssetsParams(r *http.Request) (*assets.ListAssetsParams, error) {
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPageSize, err)
	}

	maxResults, err := strconv.Atoi(r.URL.Query().Get("maxResults"))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPageMaxResults, err)
	}

	pageToken := r.URL.Query().Get("pageToken")
	if pageToken != "" {
		if _, err := ulid.Parse(pageToken); err != nil {
			return nil, ErrInvalidPageToken
		}
	}

	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	if maxResults <= 0 {
		maxResults = defaultMaxResults
	}

	return &assets.ListAssetsParams{
		PageSize:   pageSize,
		PageToken:  pageToken,
		MaxResults: maxResults,
	}, nil
}

func toTransportAsset(a assets.Asseter) any {
	switch v := a.(type) {
	case assets.ChartAsset:
		return chartResponse{
			ID:        v.ID,
			Type:      string(a.Type()),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Data: chartAssetResponse{
				Title: v.Data.Title,
				XAxis: v.Data.XAxis,
				YAxis: v.Data.YAxis,
				Data:  v.Data.Data,
			},
		}
	case assets.InsightAsset:
		return insightResponse{
			ID:        v.ID,
			Type:      string(a.Type()),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Data: insightAssetResponse{
				Insight: v.Data.Insight,
			},
		}
	case assets.AudienceAsset:
		return audienceResponse{
			ID:        v.ID,
			Type:      string(a.Type()),
			CreatedAt: v.CreatedAt,
			UpdatedAt: v.UpdatedAt,
			Data: audienceAssetResponse{
				Gender:             v.Data.Gender,
				BirthCountry:       v.Data.BirthCountry,
				AgeMin:             v.Data.AgeMin,
				AgeMax:             v.Data.AgeMax,
				SocialMediaHours:   v.Data.SocialMediaHours,
				LastMonthPurchases: v.Data.LastMonthPurchases,
			},
		}
	}
	return nil
}
