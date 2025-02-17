package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alesr/platform-go-challenge/internal/pkg/logutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	t.Parallel()

	logger := logutil.NewNoop()
	testServer := httptest.NewServer(nil)
	handlers := handlersMock{}

	defer testServer.Close()

	got := NewApp(
		logger,
		testServer.Config,
		&handlers,
	)

	require.NotNil(t, got)

	assert.Equal(t, logger.WithGroup("rest-app"), got.logger)
	assert.Equal(t, defaultShutdownTimeout, got.shutdownTimeout)
	assert.Equal(t, testServer.Config, got.Server)
	assert.NotNil(t, got.loggingMiddleware)
	assert.NotNil(t, got.recoveryMiddleware)
	assert.Equal(t, http.NewServeMux(), got.Handler)
	assert.Equal(t, &handlers, got.handlers)
}

func TestApp_Start(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(nil)
	defer testServer.Close()

	handlers := handlersMock{
		shutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}

	app := NewApp(logutil.NewNoop(), testServer.Config, &handlers)

	app.handleFuncWithMiddleware("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	assert.NoError(t, app.Start())

	resp, err := http.Get(testServer.URL)
	assert.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)
	assert.NoError(t, app.Shutdown())
}

func TestApp_Shutdown(t *testing.T) {
	t.Parallel()

	testServer := httptest.NewServer(nil)
	defer testServer.Close()

	var handlersShutdownCalled bool
	handlers := handlersMock{
		shutdownFunc: func(ctx context.Context) error {
			handlersShutdownCalled = true
			return nil
		},
	}

	app := NewApp(logutil.NewNoop(), testServer.Config, &handlers)

	app.handleFuncWithMiddleware("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	assert.NoError(t, app.Start())

	// make a request to the server to ensure it's running
	resp, err := http.Get(testServer.URL)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusTeapot, resp.StatusCode)

	assert.NoError(t, app.Shutdown())

	// attempting another request should fail
	_, err = http.Get(testServer.URL)
	assert.Error(t, err)

	// attempting another shutdown should not cause issues
	assert.NoError(t, app.Shutdown())

	assert.True(t, handlersShutdownCalled)
}
