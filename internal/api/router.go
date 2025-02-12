package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/logger"
)

func NewRouter(h *URLHandler, logger logger.Logger) chi.Router {
	r := chi.NewRouter()

	r.Use(
		middleware.WithLogging(logger),
		middleware.WithCompressing,
	)

	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLJSON)
	r.Post("/api/shorten/batch", h.ShortenBatch)
	r.Get("/{id}", h.ExpandURL)
	r.Get("/ping", h.PingDB)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	return r
}
