// Package api provides HTTP handlers and middleware for the URL shortener service.
// It defines request handlers for shortening, expanding, retrieving, and deleting URLs,
// as well as health checks and batch operations.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
)

// URLHandler handles HTTP requests for the URL shortener service.
type URLHandler struct {
	shortener service.Shortener // Service for URL shortening logic
	config    *config.Config    // Application configuration
	logger    logger.Logger     // Logger for error and info messages
}

// NewURLHandler creates a new URLHandler with the given shortener service, configuration, and logger.
func NewURLHandler(shortener service.Shortener, config *config.Config, logger logger.Logger) *URLHandler {
	return &URLHandler{
		shortener: shortener,
		config:    config,
		logger:    logger,
	}
}

func (h *URLHandler) closeBody(r *http.Request) {
	if err := r.Body.Close(); err != nil {
		h.logger.Errorf("failed to close request body: %v", err)
	}
}

// ShortenURL handles plain text POST requests to shorten a URL.
// It expects the URL in the request body and returns the shortened URL as plain text.
func (h *URLHandler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		h.logger.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer h.closeBody(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w)
		return
	}

	if len(body) == 0 {
		writeError(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	shortURL, err := h.shortener.ShortenURL(r.Context(), string(body), userID)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExist) {
			w.WriteHeader(http.StatusConflict)
		} else {
			writeError(w)
		}
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		writeError(w)
	}
}

// ShortenURLJSON handles JSON POST requests to shorten a URL.
// It expects a JSON body with a URL field and returns the shortened URL in a JSON response.
func (h *URLHandler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		h.logger.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req models.ShortenRequest
	defer h.closeBody(r)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w)
		return
	}

	if len(req.URL) == 0 {
		writeError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	shortURL, err := h.shortener.ShortenURL(r.Context(), req.URL, userID)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExist) {
			w.WriteHeader(http.StatusConflict)
		} else {
			writeError(w)
		}
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	err = json.NewEncoder(w).Encode(models.ShortenResponse{
		Result: shortURL,
	})
	if err != nil {
		writeError(w)
	}
}

// ShortenBatch handles batch URL shortening requests.
// It expects a JSON array of URLs and returns a JSON array of shortened URLs.
func (h *URLHandler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		h.logger.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req models.BatchRequest
	defer h.closeBody(r)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w)
		return
	}

	if len(req) == 0 {
		writeError(w)
		return
	}

	resp, err := h.shortener.ShortenBatch(r.Context(), req, userID)
	if err != nil {
		writeError(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		writeError(w)
	}
}

// ExpandURL handles GET requests to expand a shortened URL.
// It redirects the client to the original URL if found.
func (h *URLHandler) ExpandURL(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		writeError(w)
		return
	}

	url, err := h.shortener.ExpandURL(r.Context(), shortURL)
	if err != nil {
		if errors.Is(err, storage.ErrDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		} else {
			writeError(w)
			return
		}
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// PingDB handles health check requests for the storage backend.
// It returns 200 OK if the storage is reachable, otherwise 500 Internal Server Error.
func (h *URLHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := h.shortener.PingStorage(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetURLs handles requests to retrieve all URLs for a user.
// It returns a JSON array of URLs or 204 No Content if none exist.
func (h *URLHandler) GetURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		h.logger.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	urls, err := h.shortener.GetAll(r.Context(), userID)
	if err != nil {
		h.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(urls)
	if err != nil {
		h.logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// DeleteURLs handles requests to delete multiple shortened URLs for a user.
// It accepts a JSON array of short URL IDs and processes deletion asynchronously.
func (h *URLHandler) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(string)
	if !ok {
		h.logger.Error("user ID not found in context")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var shortURLs []string
	defer h.closeBody(r)
	err := json.NewDecoder(r.Body).Decode(&shortURLs)
	if err != nil {
		writeError(w)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		err := h.shortener.DeleteMany(ctx, userID, shortURLs)
		if err != nil {
			h.logger.Error(err)
		}
	}()

	w.WriteHeader(http.StatusAccepted)
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
