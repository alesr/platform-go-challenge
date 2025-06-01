package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
)

// Middleware chains multiple http.HandlerFunc.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Logging middleware logs request/response details.
func Logging(logger *slog.Logger) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := httputil.NewResponseWriter(w)

			next.ServeHTTP(rw, r)

			logger.InfoContext(
				r.Context(),
				"Request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.StatusCode),
				slog.Duration("duration", time.Since(start)),
			)
		}
	}
}

// Recovery middleware handles panics
func Recovery(logger *slog.Logger) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("Recovered from panic", slog.Any("error", err))
					http.Error(
						w,
						http.StatusText(http.StatusInternalServerError),
						http.StatusInternalServerError,
					)
				}
			}()
			next.ServeHTTP(w, r)
		}
	}
}
