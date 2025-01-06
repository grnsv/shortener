package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/service"
)

func HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		writeError(w)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w)
		return
	}

	if len(body) == 0 {
		writeError(w)
		return
	}

	shortURL := service.ShortenURL(string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.Get().BaseAddress.String() + "/" + shortURL))
}

func HandleExpandURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w)
		return
	}

	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		writeError(w)
		return
	}

	url, exists := service.ExpandURL(shortURL)

	if !exists {
		writeError(w)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
