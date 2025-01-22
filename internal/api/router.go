package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/api/middleware"
)

func Router() chi.Router {
	r := chi.NewRouter()
	handler := NewURLHandler()
	r.Post("/", middleware.WithDefaults(handler.ShortenURL))
	r.Post("/api/shorten", middleware.WithDefaults(handler.ShortenURLJSON))
	r.Get("/{id}", middleware.WithDefaults(handler.ExpandURL))
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return r
}
