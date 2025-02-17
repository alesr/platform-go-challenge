package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	var nextHandlerCalled bool
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	middleware := Logging(logger)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.True(t, nextHandlerCalled)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Contains(t, buf.String(), "Request")
	assert.Contains(t, buf.String(), "method=GET")
	assert.Contains(t, buf.String(), "path=/test")
	assert.Contains(t, buf.String(), "status=200")
	assert.Contains(t, buf.String(), "duration=")
}

func TestRecoveryMiddleware(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	var nextHandlerCalled bool
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		panic("test panic")
	})

	middleware := Recovery(logger)
	handler := middleware(nextHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	require.True(t, nextHandlerCalled)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Contains(t, buf.String(), "Recovered from panic")
	assert.Contains(t, buf.String(), "error=\"test panic\"")
}
