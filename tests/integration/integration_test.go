/*
The idea of integration tests on the context of this challenge is to test
the assets package repository implementation with a real test database.

As the project grows and other repositories are added, these tests could/should be
split into separate files, each testing a specific repository. Or for example,
testing HTTP client implementations towards external services like stub APIs.

I'm not being exhaustive on the test coverage here, I just want to make sure that the
repository remains functional while I continue to develop and to give you an idea of
how I write integration tests.

Refer to the Makefile for how to run these tests.
*/
package integration

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/alesr/platform-go-challenge/internal/assets/postgres"
	"github.com/alesr/platform-go-challenge/internal/pkg/dbmigrations"
	"github.com/alesr/platform-go-challenge/internal/pkg/envutil"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDBHost     = "localhost:5433" // will be overridden by TEST_DB_HOST env var
	testDBName     = "pgc_test"
	testDBUser     = "postgres"
	testDBPassword = "postgres"
)

// global so each test can use it
var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		fmt.Println("skipping integration tests")
		return
	}

	pool = setupTestDB()
	defer cleanUp()
	code := m.Run()
	os.Exit(code)
}

func setupTestDB() *pgxpool.Pool {
	dbHost := envutil.GetEnv("TEST_DB_HOST", testDBHost)
	dbName := envutil.GetEnv("TEST_DB_NAME", testDBName)
	dbUser := envutil.GetEnv("TEST_DB_USER", testDBUser)
	dbPassword := envutil.GetEnv("TEST_DB_PASSWORD", testDBPassword)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	path := filepath.Join("..", "..", "migrations")
	if err := dbmigrations.Run(db, testDBName, path); err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalln(err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalln(err)
	}

	// to start with a clean slate
	if _, err := pool.Exec(ctx, `
		TRUNCATE chart_assets, insight_assets, audience_assets, user_favorites CASCADE
	`); err != nil {
		log.Fatalln(err)
	}
	return pool
}

func cleanUp() {
	defer pool.Close()
	if _, err := pool.Exec(context.Background(), `
		TRUNCATE chart_assets, insight_assets, audience_assets, user_favorites CASCADE
	`); err != nil {
		log.Fatalln(err)
	}
}

func TestRepository_StoreAndListAssets(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	factory := assets.NewAssetFactory()

	chartAsset := factory.CreateChart("Test Chart", "X", "Y", []float64{1.0, 2.0})
	chartAsset.ID = "test-chart-123"

	insightAsset := factory.CreateInsight("Test Insight")
	insightAsset.ID = "test-insight-123"

	audienceAsset := factory.CreateAudience("M", "IT", 20, 30, 2, 5)
	audienceAsset.ID = "test-audience-123"

	require.NoError(t, repo.StoreAsset(ctx, chartAsset))
	require.NoError(t, repo.StoreAsset(ctx, insightAsset))
	require.NoError(t, repo.StoreAsset(ctx, audienceAsset))

	returnedAssets, nextToken, err := repo.ListAssets(ctx, &assets.ListAssetsParams{
		PageSize: 10,
	})

	require.NoError(t, err)
	require.NotEmpty(t, nextToken)

	foundAssets := make(map[string]struct{})

	for _, a := range returnedAssets {
		switch v := a.(type) {
		case assets.ChartAsset:
			if v.ID == chartAsset.ID {
				foundAssets[v.ID] = struct{}{}
				assert.Equal(t, "Test Chart", v.Data.Title)
				assert.Equal(t, assets.TypeAssetChart, v.Type())
			}
		case assets.InsightAsset:
			if v.ID == insightAsset.ID {
				foundAssets[v.ID] = struct{}{}
				assert.Equal(t, "Test Insight", v.Data.Insight)
				assert.Equal(t, assets.TypeAssetInsight, v.Type())
			}
		case assets.AudienceAsset:
			if v.ID == audienceAsset.ID {
				foundAssets[v.ID] = struct{}{}
				assert.Equal(t, "M", v.Data.Gender)
				assert.Equal(t, assets.TypeAssetAudience, v.Type())
			}
		}
	}

	assert.Contains(t, foundAssets, chartAsset.ID)
	assert.Contains(t, foundAssets, insightAsset.ID)
	assert.Contains(t, foundAssets, audienceAsset.ID)
}

func TestRepository_FavoriteAssets(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	repo := postgres.NewRepository(pool)
	ctx := context.Background()

	// create and store a test asset
	factory := assets.NewAssetFactory()
	chartAsset := factory.CreateChart("Test Chart", "X", "Y", []float64{1.0, 2.0})
	chartAsset.ID = "chart-1"

	require.NoError(t, repo.StoreAsset(ctx, chartAsset))

	t.Run("favorite non-existent asset", func(t *testing.T) {
		t.Parallel()

		favParams := favorites.FavoriteAssetParams{
			UserID:      "test-user",
			AssetID:     "non-existent-asset-id",
			Description: "This should fail",
		}

		err := repo.StoreFavoriteAsset(ctx, &favParams)
		require.ErrorIs(t, err, assets.ErrAssetNotFound)
	})

	t.Run("favorite existing asset", func(t *testing.T) {
		t.Parallel()

		// test storing a favorite
		favParams := favorites.FavoriteAssetParams{
			UserID:      "test-user",
			AssetID:     chartAsset.ID,
			Description: "Foo chart",
		}

		require.NoError(t, repo.StoreFavoriteAsset(ctx, &favParams))

		// test getting favorites
		favs, err := repo.GetUserFavorites(ctx, "test-user")
		require.NoError(t, err)
		require.Len(t, favs, 1)
		require.Equal(t, chartAsset.ID, favs[0].AssetID)
		require.Equal(t, "Foo chart", favs[0].Description)

		// test updating a favorite
		updateParams := favorites.UpdateFavoriteParams{
			Description: "Updated description",
		}

		updatedFav, err := repo.UpdateFavorite(ctx, favs[0].ID, "test-user", &updateParams)
		require.NoError(t, err)
		require.Equal(t, "Updated description", updatedFav.Description)

		// verify the update
		favs, err = repo.GetUserFavorites(ctx, "test-user")
		require.NoError(t, err)
		require.Len(t, favs, 1)
		require.Equal(t, "Updated description", favs[0].Description)
	})
}
