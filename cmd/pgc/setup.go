package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/alesr/platform-go-challenge/internal/users/inmemorydb"
	"github.com/alesr/resterr"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Note: these could be moved to a configuration struct,
// but I'll keep them here for brevity.
const (
	// DB settings - default values
	defaultDBName     = "pgc"
	defaultDBUser     = "postgres"
	defaultDBPassword = "postgres"
	defaultDBHost     = "localhost:5432"

	// HTTP server settings
	httpAddr         = ":8090"
	httpReadTimeout  = 5 * time.Second
	httpWriteTimeout = 10 * time.Second
	httpIdleTimeout  = 15 * time.Second

	// Number of assets and users we populate the DB with
	preloadedAssets = 100
	preloadedusers  = 50
)

func setupLogger() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	return logger
}

func setupUTC() error {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		return fmt.Errorf("could not load UTC location: %w", err)
	}
	time.Local = loc
	return nil
}

func setupDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	dbName := envutil.GetEnv("DB_NAME", defaultDBName)
	dbUser := envutil.GetEnv("DB_USER", defaultDBUser)
	dbPassword := envutil.GetEnv("DB_PASSWORD", defaultDBPassword)
	dbHost := envutil.GetEnv("DB_HOST", defaultDBHost)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbName,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not parse pool config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create connection pool: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("could notopen db connection: %w", err)
	}
	defer db.Close()

	if err := dbmigrations.Run(db, "pgc", filepath.Join(".", "migrations")); err != nil {
		return nil, fmt.Errorf("could not run migrations: %w", err)
	}
	return pool, nil
}

func setupUsersService(logger *slog.Logger, usersSample []users.User) *users.Service {
	usersRepo := inmemorydb.NewRepository(usersSample)
	return users.NewService(logger, usersRepo)
}

func setupAssetsRepository(pool *pgxpool.Pool) *postgres.Repository {
	return postgres.NewRepository(pool)
}

func setupAssetsService(logger *slog.Logger, repo *postgres.Repository) *assets.Service {
	return assets.NewService(logger, repo)
}

func setupFavoritesService(logger *slog.Logger, repo *postgres.Repository, usersSvc *users.Service) *favorites.Service {
	return favorites.NewService(logger, repo, usersSvc)
}

func populateDatabase(ctx context.Context, assetsSvc *assets.Service) error {
	if err := assetsSvc.StoreAssets(
		ctx, sampler.SampleAssets(preloadedAssets),
	); err != nil {
		return fmt.Errorf("could not populate assets tables: %w", err)
	}
	return nil
}

func setupHTTPServer(
	logger *slog.Logger,
	usersSvc *users.Service,
	assetsSvc *assets.Service,
	favSvc *favorites.Service,
) (*rest.App, error) {
	httpSrv := http.Server{
		Addr:         httpAddr,
		ReadTimeout:  httpReadTimeout,
		WriteTimeout: httpWriteTimeout,
		IdleTimeout:  httpIdleTimeout,
	}

	errHandler, err := resterr.NewHandler(
		logger,
		resterrors.ErrorMap,
		resterr.WithValidationFn(resterrors.ValidateRestErr),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create error handler: %w", err)
	}

	restHandlers := handlers.New(logger, errHandler, usersSvc, assetsSvc, favSvc)
	restApp := rest.NewApp(logger, &httpSrv, restHandlers)

	if err := restApp.Start(); err != nil {
		return nil, fmt.Errorf("could notstart server: %w", err)
	}
	return restApp, nil
}
