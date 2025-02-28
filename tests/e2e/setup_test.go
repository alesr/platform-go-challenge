// Demo area!
// This file contains the setup for the e2e tests which I also want to use as a demo of the platform-go-challenge project.
package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/api/resterrors"
	"github.com/alesr/platform-go-challenge/internal/app/rest"
	"github.com/alesr/platform-go-challenge/internal/app/rest/handlers"
	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/alesr/platform-go-challenge/internal/assets/postgres"
	"github.com/alesr/platform-go-challenge/internal/assets/sampler"
	"github.com/alesr/platform-go-challenge/internal/pkg/dbmigrations"
	"github.com/alesr/platform-go-challenge/internal/pkg/envutil"
	"github.com/alesr/platform-go-challenge/internal/pkg/logutil"
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/alesr/platform-go-challenge/internal/users/inmemorydb"
	usrsampler "github.com/alesr/platform-go-challenge/internal/users/sampler"
	"github.com/alesr/resterr"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

const (
	testDBHost     = "localhost:5433" // will be overridden by TEST_DB_HOST env var
	testDBUser     = "postgres"
	testDBPassword = "postgres"
	testDBName     = "pgc_test"
	testHTTPAddr   = ":8091"

	preloadedTestAssets = 10
	preloadedTestUsers  = 5

	retryAttempts = 5
	retryInterval = 100 * time.Millisecond
)

var testEnv *testEnvironment

type testEnvironment struct {
	app       *rest.App
	dbPool    *pgxpool.Pool
	usersSvc  *users.Service
	assetsSvc *assets.Service
	favSvc    *favorites.Service
	baseURL   string
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		fmt.Println("skipping integration tests")
		return
	}

	ctx := context.Background()

	var err error
	testEnv, err = setupTestEnvironment(ctx)
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v\n", err)
	}

	defer func() {
		if err := teardownTestEnvironment(ctx, testEnv); err != nil {
			log.Fatalf("Failed to teardown test environment: %v\n", err)
		}
	}()

	code := m.Run()
	os.Exit(code)
}

func setupTestEnvironment(ctx context.Context) (*testEnvironment, error) {
	logger := logutil.NewNoop()

	dbPool, err := setupTestDatabase(ctx)
	if err != nil {
		return nil, err
	}

	usersSamples, err := usrsampler.SampleUsers(preloadedTestUsers)
	if err != nil {
		return nil, fmt.Errorf("generate sample users: %w", err)
	}

	usersSvc := setupTestUsersService(logger, usersSamples)
	assetsRepo := postgres.NewRepository(dbPool)
	assetsSvc := assets.NewService(logger, assetsRepo)
	favSvc := favorites.NewService(logger, assetsRepo, usersSvc)

	if err := populateTestDatabase(ctx, assetsSvc); err != nil {
		return nil, fmt.Errorf("populate test database: %w", err)
	}

	app, err := setupTestHTTPServer(logger, usersSvc, assetsSvc, favSvc)
	if err != nil {
		return nil, fmt.Errorf("setup test HTTP server: %w", err)
	}

	return &testEnvironment{
		app:       app,
		dbPool:    dbPool,
		usersSvc:  usersSvc,
		assetsSvc: assetsSvc,
		favSvc:    favSvc,
		baseURL:   "http://localhost" + testHTTPAddr,
	}, nil
}

func teardownTestEnvironment(ctx context.Context, env *testEnvironment) error {
	if env.app != nil {
		if err := env.app.Shutdown(); err != nil {
			return fmt.Errorf("shutdown app: %w", err)
		}
	}

	if env.dbPool != nil {
		if _, err := env.dbPool.Exec(ctx, `
            TRUNCATE TABLE chart_assets CASCADE;
            TRUNCATE TABLE insight_assets CASCADE;
            TRUNCATE TABLE audience_assets CASCADE;
            TRUNCATE TABLE user_favorites CASCADE;
        `); err != nil {
			return fmt.Errorf("clean database tables: %w", err)
		}
		env.dbPool.Close()
	}
	return nil
}

func setupTestDatabase(ctx context.Context) (*pgxpool.Pool, error) {
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
		return nil, fmt.Errorf("open db connection: %w", err)
	}
	defer db.Close()

	if err := dbmigrations.Run(db, testDBName, filepath.Join("..", "..", "migrations")); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if _, err := pool.Exec(ctx, `
        TRUNCATE TABLE chart_assets CASCADE;
        TRUNCATE TABLE insight_assets CASCADE;
        TRUNCATE TABLE audience_assets CASCADE;
        TRUNCATE TABLE user_favorites CASCADE;
    `); err != nil {
		return nil, fmt.Errorf("clean database tables: %w", err)
	}
	return pool, nil
}

func setupTestUsersService(logger *slog.Logger, usersSample []users.User) *users.Service {
	usersRepo := inmemorydb.NewRepository(usersSample)
	return users.NewService(logger, usersRepo)
}

func populateTestDatabase(ctx context.Context, assetsSvc *assets.Service) error {
	return assetsSvc.StoreAssets(ctx, sampler.SampleAssets(preloadedTestAssets))
}

func setupTestHTTPServer(
	logger *slog.Logger,
	usersSvc *users.Service,
	assetsSvc *assets.Service,
	favSvc *favorites.Service,
) (*rest.App, error) {
	httpSrv := http.Server{
		Addr:         testHTTPAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	errHandler, err := resterr.NewHandler(
		logger,
		resterrors.ErrorMap,
		resterr.WithValidationFn(resterrors.ValidateRestErr),
	)
	if err != nil {
		return nil, fmt.Errorf("create error handler: %w", err)
	}

	restHandlers := handlers.New(logger, errHandler, usersSvc, assetsSvc, favSvc)
	restApp := rest.NewApp(logger, &httpSrv, restHandlers)

	if err := restApp.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}
	return restApp, nil
}

func makeRequest(t *testing.T, method, url string, body any) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeResponse(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	require.NoError(t, json.NewDecoder(resp.Body).Decode(v))
	defer resp.Body.Close()
}

func retry(t *testing.T, f func() bool) bool {
	t.Helper()
	for i := 0; i < retryAttempts; i++ {
		if f() {
			return true
		}
		if i < retryAttempts-1 {
			time.Sleep(retryInterval)
		}
	}
	return false
}
