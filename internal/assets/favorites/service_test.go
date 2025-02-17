package favorites

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/logutil"
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_FavoriteAsset(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()

	testCases := []struct {
		name                          string
		givenParams                   FavoriteAssetParams
		givenFetchUserResult          func() (*users.User, error)
		givenStoreFavoriteAssetResult func() error
		expectedError                 error
	}{
		{
			name: "success",
			givenParams: FavoriteAssetParams{
				UserID:      userID.String(),
				AssetID:     "asset-123",
				Description: "my favorite asset",
			},
			givenFetchUserResult: func() (*users.User, error) {
				return &users.User{}, nil
			},
			givenStoreFavoriteAssetResult: func() error {
				return nil
			},
		},
		{
			name: "user not found",
			givenParams: FavoriteAssetParams{
				UserID:  userID.String(),
				AssetID: "asset-123",
			},
			givenFetchUserResult: func() (*users.User, error) {
				return nil, users.ErrUserNotFound
			},
			givenStoreFavoriteAssetResult: func() error {
				return nil
			},
			expectedError: users.ErrUserNotFound,
		},
		{
			name: "user service rand error",
			givenParams: FavoriteAssetParams{
				UserID:  userID.String(),
				AssetID: "asset-123",
			},
			givenFetchUserResult: func() (*users.User, error) {
				return nil, assert.AnError
			},
			givenStoreFavoriteAssetResult: func() error {
				return nil
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				userSvcCalled bool
				repoCalled    bool
			)

			userSvc := userSvcMock{
				fetchUserFunc: func(ctx context.Context, id string) (*users.User, error) {
					userSvcCalled = true
					assert.Equal(t, tc.givenParams.UserID, id)
					return tc.givenFetchUserResult()
				},
			}

			repo := repoMock{
				storeFavoriteAssetFunc: func(ctx context.Context, params *FavoriteAssetParams) error {
					repoCalled = true
					assert.Equal(t, &tc.givenParams, params)
					return tc.givenStoreFavoriteAssetResult()
				},
			}

			svc := NewService(logutil.NewNoop(), &repo, &userSvc)

			// create a context with timeout for the entire test
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := svc.FavoriteAsset(ctx, &tc.givenParams)

			assert.True(t, userSvcCalled)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)

			// wait for worker pool to process the task
			err = svc.Shutdown(ctx)
			require.NoError(t, err)

			// now we can check if the repository was called
			if tc.expectedError == nil {
				assert.True(t, repoCalled)
			}
		})
	}
}

func TestService_FetchUserFavorites(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()

	givenFavorites := []FavoriteAsset{
		{
			ID:          "fav-1",
			UserID:      userID.String(),
			AssetID:     "asset-1",
			Description: "my favorite asset 1",
		},
		{
			ID:          "fav-2",
			UserID:      userID.String(),
			AssetID:     "asset-2",
			Description: "my favorite asset 2",
		},
	}

	testCases := []struct {
		name                        string
		givenUserID                 string
		givenFetchUserResult        func() (*users.User, error)
		givenGetUserFavoritesResult func() ([]FavoriteAsset, error)
		expectedFavorites           []FavoriteAsset
		expectedError               error
	}{
		{
			name:        "success",
			givenUserID: userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return &users.User{}, nil
			},
			givenGetUserFavoritesResult: func() ([]FavoriteAsset, error) {
				return givenFavorites, nil
			},
			expectedFavorites: givenFavorites,
		},
		{
			name:        "user not found",
			givenUserID: userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return nil, users.ErrUserNotFound
			},
			givenGetUserFavoritesResult: func() ([]FavoriteAsset, error) {
				return nil, nil
			},
			expectedError: users.ErrUserNotFound,
		},
		{
			name:        "user service random error",
			givenUserID: userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return nil, assert.AnError
			},
			givenGetUserFavoritesResult: func() ([]FavoriteAsset, error) {
				return nil, nil
			},
			expectedError: assert.AnError,
		},
		{
			name:        "repository error",
			givenUserID: userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return &users.User{}, nil
			},
			givenGetUserFavoritesResult: func() ([]FavoriteAsset, error) {
				return nil, assert.AnError
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				userSvcCalled bool
				repoCalled    bool
			)

			userSvc := userSvcMock{
				fetchUserFunc: func(ctx context.Context, id string) (*users.User, error) {
					userSvcCalled = true
					assert.Equal(t, tc.givenUserID, id)
					return tc.givenFetchUserResult()
				},
			}

			repo := repoMock{
				getuserfavoritesFunc: func(ctx context.Context, userID string) ([]FavoriteAsset, error) {
					repoCalled = true
					assert.Equal(t, tc.givenUserID, userID)
					return tc.givenGetUserFavoritesResult()
				},
			}

			svc := NewService(logutil.NewNoop(), &repo, &userSvc)

			favorites, err := svc.FetchUserFavorites(context.TODO(), tc.givenUserID)

			assert.True(t, userSvcCalled)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				return
			}

			require.NoError(t, err)
			assert.True(t, repoCalled)
			assert.Equal(t, tc.expectedFavorites, favorites)
		})
	}
}

func TestService_UpdateFavorite(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	favoriteID := ulid.Make()

	givenParams := &UpdateFavoriteParams{
		Description: "updated description",
	}

	givenFavoriteAsset := &FavoriteAsset{
		ID:          favoriteID.String(),
		UserID:      userID.String(),
		AssetID:     "asset-123",
		Description: givenParams.Description,
	}

	testCases := []struct {
		name                      string
		givenFavoriteID           string
		givenUserID               string
		givenParams               *UpdateFavoriteParams
		givenUpdateFavoriteResult func() (*FavoriteAsset, error)
		expectedFavorite          *FavoriteAsset
		expectedError             error
	}{
		{
			name:            "success",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenParams:     givenParams,
			givenUpdateFavoriteResult: func() (*FavoriteAsset, error) {
				return givenFavoriteAsset, nil
			},
			expectedFavorite: givenFavoriteAsset,
		},
		{
			name:            "favorite not found",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenParams:     givenParams,
			givenUpdateFavoriteResult: func() (*FavoriteAsset, error) {
				return nil, ErrFavoriteAssetNotFound
			},
			expectedError: ErrFavoriteAssetNotFound,
		},
		{
			name:            "repository error",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenParams:     givenParams,
			givenUpdateFavoriteResult: func() (*FavoriteAsset, error) {
				return nil, assert.AnError
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repoCalled bool

			repo := repoMock{
				updatefavoriteFunc: func(ctx context.Context, favID, userID string, params *UpdateFavoriteParams) (*FavoriteAsset, error) {
					repoCalled = true
					assert.Equal(t, tc.givenFavoriteID, favID)
					assert.Equal(t, tc.givenUserID, userID)
					assert.Equal(t, tc.givenParams, params)
					return tc.givenUpdateFavoriteResult()
				},
			}

			svc := NewService(logutil.NewNoop(), &repo, nil)

			favorite, err := svc.UpdateFavorite(context.TODO(), tc.givenFavoriteID, tc.givenUserID, tc.givenParams)

			assert.True(t, repoCalled)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Nil(t, favorite)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, favorite)
			assert.Equal(t, tc.expectedFavorite, favorite)
		})
	}
}

func TestService_DeleteFavorite(t *testing.T) {
	t.Parallel()

	userID := ulid.Make()
	favoriteID := ulid.Make()

	testCases := []struct {
		name                      string
		givenFavoriteID           string
		givenUserID               string
		givenFetchUserResult      func() (*users.User, error)
		givenDeleteFavoriteResult func() error
		expectedError             error
	}{
		{
			name:            "success",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return &users.User{}, nil
			},
			givenDeleteFavoriteResult: func() error {
				return nil
			},
		},
		{
			name:            "user not found",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return nil, users.ErrUserNotFound
			},
			givenDeleteFavoriteResult: func() error {
				return nil
			},
			expectedError: users.ErrUserNotFound,
		},
		{
			name:            "user service error",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return nil, assert.AnError
			},
			givenDeleteFavoriteResult: func() error {
				return nil
			},
			expectedError: assert.AnError,
		},
		{
			name:            "favorite not found",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return &users.User{}, nil
			},
			givenDeleteFavoriteResult: func() error {
				return ErrFavoriteAssetNotFound
			},
			expectedError: ErrFavoriteAssetNotFound,
		},
		{
			name:            "repository error",
			givenFavoriteID: favoriteID.String(),
			givenUserID:     userID.String(),
			givenFetchUserResult: func() (*users.User, error) {
				return &users.User{}, nil
			},
			givenDeleteFavoriteResult: func() error {
				return assert.AnError
			},
			expectedError: assert.AnError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				userSvcCalled bool
				repoCalled    bool
			)

			userSvc := userSvcMock{
				fetchUserFunc: func(ctx context.Context, id string) (*users.User, error) {
					userSvcCalled = true
					assert.Equal(t, tc.givenUserID, id)
					return tc.givenFetchUserResult()
				},
			}

			repo := repoMock{
				deleteFavoriteFunc: func(ctx context.Context, favoriteID, userID string) error {
					repoCalled = true
					assert.Equal(t, tc.givenFavoriteID, favoriteID)
					assert.Equal(t, tc.givenUserID, userID)
					return tc.givenDeleteFavoriteResult()
				},
			}

			svc := NewService(logutil.NewNoop(), &repo, &userSvc)

			err := svc.DeleteFavorite(context.TODO(), tc.givenFavoriteID, tc.givenUserID)

			assert.True(t, userSvcCalled)

			if tc.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedError)
				// if error is from user service, repository should not be called
				if errors.Is(tc.expectedError, users.ErrUserNotFound) ||
					tc.expectedError == assert.AnError && !repoCalled {
					assert.False(t, repoCalled)
				}
				return
			}

			require.NoError(t, err)
			assert.True(t, repoCalled)
		})
	}
}
