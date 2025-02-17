package rest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/httputil/middleware"
)

const defaultShutdownTimeout = 30 * time.Second

type handlers interface {
	Shutdown(ctx context.Context) error
	ListAssets() http.HandlerFunc
	ListUsers() http.HandlerFunc
	FavoriteAsset() http.HandlerFunc
	GetUserFavorites() http.HandlerFunc
	UpdateFavorite() http.HandlerFunc
	DeleteFavorite() http.HandlerFunc
}

// App represents our RESTful application instance.
type App struct {
	logger          *slog.Logger
	shutdownTimeout time.Duration
	*http.Server
	loggingMiddleware  middleware.Middleware
	recoveryMiddleware middleware.Middleware
	handlers           handlers
}

// NewApp creates a new RESTful application instance.
func NewApp(logger *slog.Logger, srv *http.Server, hdlers handlers) *App {
	app := &App{
		logger:             logger.WithGroup("rest-app"),
		shutdownTimeout:    defaultShutdownTimeout,
		Server:             srv,
		loggingMiddleware:  middleware.Logging(logger),
		recoveryMiddleware: middleware.Recovery(logger),
		handlers:           hdlers,
	}
	app.Handler = http.NewServeMux()
	return app
}

// Start initializes the application's server and starts listening.
func (app *App) Start() error {
	// Register endpoints

	app.handleFuncWithMiddleware("GET /assets", app.handlers.ListAssets())
	app.handleFuncWithMiddleware("GET /users", app.handlers.ListUsers())
	app.handleFuncWithMiddleware("POST /assets/favorite", app.handlers.FavoriteAsset())
	app.handleFuncWithMiddleware("GET /users/{user_id}/favorites", app.handlers.GetUserFavorites())
	app.handleFuncWithMiddleware("PATCH /users/{user_id}/favorites/{favorite_id}", app.handlers.UpdateFavorite())
	app.handleFuncWithMiddleware("DELETE /users/{user_id}/favorites/{favorite_id}", app.handlers.DeleteFavorite())

	go func() {
		app.logger.Info("Starting server", slog.String("addr", app.Addr))
		if err := app.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.logger.Error("Could not start server", slog.String("error", err.Error()))
		}
	}()
	return nil
}

// Shutdown gracefully shuts down the server.
func (app *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), app.shutdownTimeout)
	defer cancel()

	app.logger.Info("Shutting down server")

	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not gracefully shutdown the server: %w", err)
	}

	if err := app.handlers.Shutdown(ctx); err != nil {
		return fmt.Errorf("could not gracefully shutdown handlers: %w", err)
	}
	return nil
}

// handleFuncWithMiddleware registers a handler function with the given path and applies the given middleware.
// By default, it applies the logging and recovery middlewares.
func (app *App) handleFuncWithMiddleware(
	path string,
	handler http.HandlerFunc,
	middlewares ...middleware.Middleware,
) {
	middlewares = append(
		[]middleware.Middleware{app.loggingMiddleware, app.recoveryMiddleware},
		middlewares...,
	)
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	app.Handler.(*http.ServeMux).HandleFunc(path, handler)
}
