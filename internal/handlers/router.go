package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()
	handler := NewURLHandler()
	r.Post("/", handler.ShortenURL)
	r.Get("/{id}", handler.ExpandURL)
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return r
}
