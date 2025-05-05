package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
)

// NewRouter creates and configures a new chi.Router for the URL shortener API.
//
// It registers all API endpoints, applies middleware for logging, compression, and authentication,
// and sets up handlers for URL shortening, expansion, health checks, and user-specific operations.
//
// Parameters:
//
//	h      - pointer to URLHandler containing all endpoint handler methods
//	config - pointer to Config struct with application configuration (e.g., JWT secret)
//	logger - Logger interface for request logging
//
// Returns:
//
//	chi.Router - a fully configured router ready to be used by an HTTP server
func NewRouter(h *URLHandler, config *config.Config, logger logger.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(
		middleware.WithLogging(logger),
		middleware.WithCompressing(logger),
		middleware.Authenticate(config.JWTSecret, logger),
	)

	r.Post("/", h.ShortenURL)
	r.Get("/{id}", h.ExpandURL)
	r.Get("/ping", h.PingDB)
	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", h.ShortenURLJSON)
			r.Post("/batch", h.ShortenBatch)
		})
		r.Route("/user/urls", func(r chi.Router) {
			r.Get("/", h.GetURLs)
			r.Delete("/", h.DeleteURLs)
		})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	return r
}
