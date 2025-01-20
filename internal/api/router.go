package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()
	handler := NewURLHandler()
	r.Post("/", WithLogging(handler.ShortenURL))
	r.Post("/api/shorten", WithLogging(handler.ShortenURLJSON))
	r.Get("/{id}", WithLogging(handler.ExpandURL))
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return r
}
