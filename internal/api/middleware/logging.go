package middleware

import (
	"net/http"
	"time"

	"github.com/grnsv/shortener/internal/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

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
