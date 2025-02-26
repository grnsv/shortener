package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
)

func NewRouter(h *URLHandler, config *config.Config, logger logger.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(
		middleware.WithLogging(logger),
		middleware.WithCompressing,
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
