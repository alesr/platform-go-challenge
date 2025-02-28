package e2e

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/internal/app/rest/handlers"
	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	time.Sleep(time.Second)

	testUser := getTestUser(t)
	fmt.Printf("User selected: %s\n\n", testUser.Name)

	testAsset := getTestAsset(t)
	assetID, assetType := extractAssetInfo(t, testAsset)
	fmt.Printf("Asset selected: %s (Type: %s)\n\n", assetID, assetType)

	testFavorite := createFavorite(t, testUser.ID, assetID)
	fmt.Printf("Favorite created with description: %q\n\n", testFavorite.Description)

	favorite := verifyFavoriteExists(t, testUser.ID)
	fmt.Printf("Favorite found: %s with description: %q\n\n", favorite.ID, favorite.Description)

	updatedFavorite := updateFavorite(t, testUser.ID, favorite.ID)
	fmt.Printf("Favorite updated successfully from %q to %q\n\n",
		favorite.Description,
		updatedFavorite.Description)

	deleteFavorite(t, testUser.ID, favorite.ID)
	verifyFavoriteDeletion(t, testUser.ID)
}

func getTestUser(t *testing.T) handlers.UserResponse {
	t.Helper()

	fmt.Println("Listing users...")
	var usersResp httputil.Response[handlers.ListUsersResponse]

	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodGet, testEnv.baseURL+"/users", nil)
		if resp.StatusCode != http.StatusOK {
			return false
		}
		decodeResponse(t, resp, &usersResp)
		return len(usersResp.Data.Data) > 0
	})
	require.True(t, success)
	return usersResp.Data.Data[0]
}

func getTestAsset(t *testing.T) any {
	t.Helper()

	fmt.Println("Listing assets...")
	var assetsResp httputil.Response[handlers.ListAssetsResponse]

	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/assets?pageSize=10&maxResults=100", testEnv.baseURL), nil)
		if resp.StatusCode != http.StatusOK {
			return false
		}
		decodeResponse(t, resp, &assetsResp)
		return len(assetsResp.Data.Items) > 0
	})
	require.True(t, success)
	return assetsResp.Data.Items[0]
}

func extractAssetInfo(t *testing.T, testAsset any) (string, string) {
	t.Helper()

	switch v := testAsset.(type) {
	case map[string]any:
		return v["id"].(string), v["type"].(string)
	case struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}:
		return v.ID, v.Type
	default:
		t.Fatal("Unknown asset type")
		return "", ""
	}
}

func createFavorite(t *testing.T, userID, assetID string) handlers.FavoriteAssetRequest {
	t.Helper()

	fmt.Println("Marking asset as favorite...")
	favoriteReq := handlers.FavoriteAssetRequest{
		UserID:      userID,
		AssetID:     assetID,
		Description: "Test favorite",
	}

	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodPost, testEnv.baseURL+"/assets/favorite", favoriteReq)
		return resp.StatusCode == http.StatusAccepted
	})
	require.True(t, success)
	return favoriteReq
}

func verifyFavoriteExists(t *testing.T, userID string) handlers.FavoriteAssetResponse {
	t.Helper()

	fmt.Println("Retrieving favorites...")
	var favoritesResp httputil.Response[handlers.ListUserFavoritesResponse]

	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodGet, fmt.Sprintf("%s/users/%s/favorites", testEnv.baseURL, userID), nil)
		if resp.StatusCode != http.StatusOK {
			return false
		}
		decodeResponse(t, resp, &favoritesResp)
		return len(favoritesResp.Data.Items) > 0
	})
	require.True(t, success)
	return favoritesResp.Data.Items[0]
}

func updateFavorite(t *testing.T, userID, favoriteID string) handlers.FavoriteAssetResponse {
	t.Helper()

	fmt.Println("Updating favorite...")
	updateReq := handlers.UpdateFavoriteRequest{
		Description: "Updated description",
	}

	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodPatch,
			fmt.Sprintf("%s/users/%s/favorites/%s", testEnv.baseURL, userID, favoriteID),
			updateReq)
		return resp.StatusCode == http.StatusOK
	})
	require.True(t, success)

	var updatedFavoriteResp httputil.Response[handlers.ListUserFavoritesResponse]
	success = retry(t, func() bool {
		resp := makeRequest(t, http.MethodGet,
			fmt.Sprintf("%s/users/%s/favorites", testEnv.baseURL, userID),
			nil)
		if resp.StatusCode != http.StatusOK {
			return false
		}
		decodeResponse(t, resp, &updatedFavoriteResp)
		return len(updatedFavoriteResp.Data.Items) > 0
	})
	require.True(t, success)
	return updatedFavoriteResp.Data.Items[0]
}

func deleteFavorite(t *testing.T, userID, favoriteID string) {
	t.Helper()

	fmt.Println("Deleting favorite...")
	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodDelete,
			fmt.Sprintf("%s/users/%s/favorites/%s", testEnv.baseURL, userID, favoriteID),
			nil)
		return resp.StatusCode == http.StatusNoContent
	})
	require.True(t, success)
}

func verifyFavoriteDeletion(t *testing.T, userID string) {
	t.Helper()

	fmt.Println("Verifying deletion...")
	success := retry(t, func() bool {
		resp := makeRequest(t, http.MethodGet,
			fmt.Sprintf("%s/users/%s/favorites", testEnv.baseURL, userID),
			nil)
		if resp.StatusCode != http.StatusOK {
			return false
		}
		var emptyFavoritesResp handlers.ListUserFavoritesResponse
		decodeResponse(t, resp, &emptyFavoritesResp)
		return len(emptyFavoritesResp.Items) == 0
	})
	require.True(t, success)
	fmt.Println("Favorite successfully deleted")
}
