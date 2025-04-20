package middleware

import (
	"net/http"
	"time"

	"github.com/grnsv/shortener/internal/logger"
)

// loggingResponseWriter is a custom http.ResponseWriter that captures status and size for logging.
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

// Write writes the data to the connection as part of an HTTP reply and records the size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

// WriteHeader sends an HTTP response header with the provided status code and records the status.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

// WithLogging is a middleware that logs HTTP requests and responses using the provided logger.
func WithLogging(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := &loggingResponseWriter{ResponseWriter: w}
			next.ServeHTTP(lw, r)

			logger.Infoln(
				"uri", r.RequestURI,
				"method", r.Method,
				"duration", time.Since(start),
				"status", lw.status,
				"size", lw.size,
			)
		})
	}
}
