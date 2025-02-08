package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
)

type URLHandler struct {
	shortener service.Shortener
	config    *config.Config
	logger    logger.Logger
}

func NewURLHandler(shortener service.Shortener, config *config.Config, logger logger.Logger) *URLHandler {
	return &URLHandler{
		shortener: shortener,
		config:    config,
		logger:    logger,
	}
}

func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
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

	shortURL, err := h.shortener.ShortenURL(r.Context(), string(body))
	if err != nil {
		writeError(w)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(h.config.BaseAddress.String() + "/" + shortURL))
	if err != nil {
		writeError(w)
	}
}

func (h *URLHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
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

	shortURL, err := h.shortener.ShortenURL(r.Context(), req.URL)
	if err != nil {
		writeError(w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(models.ShortenResponse{
		Result: h.config.BaseAddress.String() + "/" + shortURL,
	})
	if err != nil {
		writeError(w)
	}
}

func (h *URLHandler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	var req models.BatchRequest
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w)
		return
	}

	if len(req) == 0 {
		writeError(w)
		return
	}

	resp, err := h.shortener.ShortenBatch(r.Context(), req, h.config.BaseAddress.String()+"/")
	if err != nil {
		writeError(w)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		writeError(w)
	}
}

func (h *URLHandler) ExpandURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		writeError(w)
		return
	}

	url, err := h.shortener.ExpandURL(r.Context(), shortURL)
	if err != nil {
		writeError(w)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *URLHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.shortener.PingStorage(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
