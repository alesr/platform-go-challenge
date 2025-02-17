package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
)

const MaxDescriptionLength = 128

// FavoriteAssetRequest defines the data structure for a request to favorite an asset.
type FavoriteAssetRequest struct {
	UserID      string `json:"user_id"`
	AssetID     string `json:"asset_id"`
	Description string `json:"description"`
}

// ListUserFavoritesResponse defines the data structure for listing user favorites
type ListUserFavoritesResponse struct {
	Items []FavoriteAssetResponse `json:"items"`
}

// FavoriteAssetResponse defines the data structure item for a list of favorite assets.
type FavoriteAssetResponse struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type UpdateFavoriteRequest struct {
	Description string `json:"description"`
}

func (u *UpdateFavoriteRequest) validate() error {
	if len(u.Description) > MaxDescriptionLength {
		return ErrDescriptionMaxLen
	}
	return nil
}

// FavoriteAsset marks an asset as user favorite.
func (h *Handler) FavoriteAsset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var data FavoriteAssetRequest
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			h.errHandler.Handle(
				r.Context(), w,
				fmt.Errorf("could not decode request data: %w, %w",
					err, ErrInvalidFavoriteAssetPayload,
				),
			)
			return
		}

		params := favorites.FavoriteAssetParams{
			UserID:      data.UserID,
			AssetID:     data.AssetID,
			Description: data.Description,
		}

		if err := h.favoritesSvc.FavoriteAsset(r.Context(), &params); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not favorite asset: %w", err))
			return
		}

		// using empty struct since we don't need to return data
		httputil.RespondWithJSON(w, http.StatusAccepted, struct{}{})
	}
}

// GetUserFavorites returns a list of user's favorites (assets)
func (h *Handler) GetUserFavorites() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user_id")
		if err := validateID(userID); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not validate user ID: %w, %v", ErrInvalidUserID, err))
			return
		}

		favorites, err := h.favoritesSvc.FetchUserFavorites(r.Context(), userID)
		if err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not fetch user favorites: %w", err))
			return
		}

		httputil.RespondWithJSON(w, http.StatusOK, ListUserFavoritesResponse{
			Items: toFavoritesResponse(favorites...),
		})
	}
}

// UpdateFavorite updates a user's favorite asset.
func (h *Handler) UpdateFavorite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user_id")
		favoriteID := r.PathValue("favorite_id")

		if err := validateID(userID); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not validate user ID: %w, %v", ErrInvalidUserID, err))
			return
		}

		if err := validateID(favoriteID); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not validate favorite ID: %w, %v", ErrInvalidFavoriteID, err))
			return
		}

		var reqData UpdateFavoriteRequest
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not decode request data: %w", err))
			return
		}

		params := favorites.UpdateFavoriteParams{
			Description: reqData.Description,
		}

		favAsset, err := h.favoritesSvc.UpdateFavorite(r.Context(), favoriteID, userID, &params)
		if err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not update user favorite: %w", err))
			return
		}
		httputil.RespondWithJSON(w, http.StatusOK, toFavoritesResponse(*favAsset)[0])
	}
}

// DeleteFavorite deletes a user's favorite asset
func (h *Handler) DeleteFavorite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("user_id")
		favoriteID := r.PathValue("favorite_id")

		if err := validateID(userID); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not validate user ID: %w, %v", ErrInvalidUserID, err))
			return
		}

		if err := validateID(favoriteID); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not validate favorite ID: %w, %v", ErrInvalidFavoriteID, err))
			return
		}

		if err := h.favoritesSvc.DeleteFavorite(r.Context(), favoriteID, userID); err != nil {
			h.errHandler.Handle(r.Context(), w, fmt.Errorf("could not delete favorite: %w", err))
			return
		}
		httputil.RespondWithJSON[any](w, http.StatusNoContent, nil)
	}
}

func toFavoritesResponse(favorites ...favorites.FavoriteAsset) []FavoriteAssetResponse {
	var items []FavoriteAssetResponse
	for _, favorite := range favorites {
		items = append(items, FavoriteAssetResponse{
			ID:          favorite.ID,
			Description: favorite.Description,
		})
	}
	return items
}
