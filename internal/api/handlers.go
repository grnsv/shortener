package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
)

type URLHandler struct {
	shortener *service.URLShortener
}

func NewURLHandler() *URLHandler {
	return &URLHandler{shortener: service.NewURLShortener()}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	shortURL := h.shortener.ShortenURL(string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(config.Get().BaseAddress.String() + "/" + shortURL))
	if err != nil {
		writeError(w)
	}
}

func (h *URLHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w)
		return
	}

	var req models.ShortenRequest
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w)
		return
	}

	if len(req.URL) == 0 {
		writeError(w)
		return
	}

	shortURL := h.shortener.ShortenURL(req.URL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(models.ShortenResponse{
		Result: config.Get().BaseAddress.String() + "/" + shortURL,
	})
	if err != nil {
		writeError(w)
	}
}

func (h *URLHandler) ExpandURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w)
		return
	}

	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		writeError(w)
		return
	}

	url, exists := h.shortener.ExpandURL(shortURL)

	if !exists {
		writeError(w)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
