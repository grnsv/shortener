package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type URLHandler struct {
	shortener service.Shortener
}

func NewURLHandler(shortener service.Shortener) *URLHandler {
	return &URLHandler{shortener: shortener}
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

	shortURL, err := h.shortener.ShortenURL(string(body))
	if err != nil {
		writeError(w)
	}

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

	shortURL, err := h.shortener.ShortenURL(req.URL)
	if err != nil {
		writeError(w)
	}

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

func (h *URLHandler) PingDB(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("pgx", config.Get().DatabaseDSN)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}
