package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alesr/platform-go-challenge/internal/app/rest"
	usrsampler "github.com/alesr/platform-go-challenge/internal/users/sampler"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	ExitTimezoneSetupError = iota + 1
	ExitDatabaseSetupError
	ExitUserPopulationError
	ExitAssetPopulationError
	ExitServerSetupError
	ExitShutdownError
)

func main() {
	logger := setupLogger()

	if err := setupUTC(); err != nil {
		logger.Error("Failed to setup UTC timezone", slog.String("error", err.Error()))
		os.Exit(ExitTimezoneSetupError)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	dbPool, err := setupDatabase(ctx)
	if err != nil {
		logger.Error("Failed to setup database", slog.String("error", err.Error()))
		os.Exit(ExitDatabaseSetupError)
	}
	defer dbPool.Close()

	logger.Info("Populating users database...")
	usersSamples, err := usrsampler.SampleUsers(preloadedusers)
	if err != nil {
		logger.Error("Failed to populate database", slog.String("error", err.Error()))
		os.Exit(ExitUserPopulationError)
	}
	logger.Info("Users database populated", slog.Int("number_of_users", preloadedusers))

	usersSvc := setupUsersService(logger, usersSamples)

	assetsRepo := setupAssetsRepository(dbPool)

	assetsSvc := setupAssetsService(logger, assetsRepo)

	favoritesSvc := setupFavoritesService(logger, assetsRepo, usersSvc)

	logger.Info("Populating assets database...")
	if err := populateDatabase(ctx, assetsSvc); err != nil {
		logger.Error("Failed to populate database", slog.String("error", err.Error()))
		os.Exit(ExitAssetPopulationError)
	}
	logger.Info("Assets database populated", slog.Int("number_of_assets", preloadedAssets))

	restApp, err := setupHTTPServer(logger, usersSvc, assetsSvc, favoritesSvc)
	if err != nil {
		logger.Error("Failed to setup HTTP server", slog.String("error", err.Error()))
		os.Exit(ExitServerSetupError)
	}

	if err := waitForShutdown(restApp); err != nil {
		logger.Error("Failed during shutdown", slog.String("error", err.Error()))
		os.Exit(ExitShutdownError)
	}
}

func waitForShutdown(restApp *rest.App) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := restApp.Shutdown(); err != nil {
		return fmt.Errorf("could not shutdown server: %w", err)
	}
	return nil
}
