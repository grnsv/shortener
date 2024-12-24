package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/util"
)

var (
	urlMap = make(map[string]string)
	mu     sync.RWMutex
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w)
		return
	}

	if len(body) == 0 {
		writeError(w)
		return
	}

	shortURL := util.GenerateShortURL(body)

	mu.Lock()
	urlMap[shortURL] = string(body)
	mu.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("http://%s/%s", config.ServerAddress, shortURL)))
}

func HandleExpandURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	if shortURL == "" {
		writeError(w)
		return
	}

	mu.RLock()
	url, exists := urlMap[shortURL]
	mu.RUnlock()

	if !exists {
		writeError(w)
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func writeError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}