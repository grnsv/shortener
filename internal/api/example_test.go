package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
)

const (
	uri    = "http://localhost:8080"
	userID = "00000000-0000-0000-0000-000000000001"
)

// helper to create handler and config for examples
func exampleHandler() (*api.URLHandler, *config.Config) {
	mem, _ := storage.NewMemoryStorage(context.Background())
	cfg := config.New(
		config.WithAppEnv("testing"),
		config.WithServerAddress(config.NetAddress{Host: "localhost", Port: 8080}),
		config.WithBaseAddress(config.BaseURI{Scheme: "http://", Address: config.NetAddress{Host: "localhost", Port: 8080}}),
	)
	shortener := service.NewShortener(mem, mem, mem, mem, cfg.BaseAddress.String())
	log, _ := logger.New("testing")

	return api.NewURLHandler(shortener, cfg, log), cfg
}

// Example of shortening a URL via plain text POST
func Example_shortenURL_plain() {
	handler, _ := exampleHandler()

	req := httptest.NewRequest(http.MethodPost, uri, bytes.NewBufferString("https://example.com"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, userID))
	rec := httptest.NewRecorder()
	handler.ShortenURL(rec, req)
	fmt.Print(rec.Code)
	// Output: 201
}

// Example of shortening a URL via JSON POST
func Example_shortenURL_json() {
	handler, _ := exampleHandler()

	body := map[string]string{"url": "https://example.com"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, uri+"/api/shorten", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, userID))
	rec := httptest.NewRecorder()
	handler.ShortenURLJSON(rec, req)
	fmt.Print(rec.Code)
	// Output: 201
}

// Example of expanding a shortened URL
func Example_expandURL() {
	handler, _ := exampleHandler()

	req := httptest.NewRequest(http.MethodGet, uri, nil)
	rec := httptest.NewRecorder()
	handler.ExpandURL(rec, req)
	fmt.Print(rec.Code)
	// Output: 400
}

// Example of getting all user URLs
func Example_getUserURLs() {
	handler, _ := exampleHandler()

	req := httptest.NewRequest(http.MethodGet, uri+"/api/user/urls", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, userID))
	rec := httptest.NewRecorder()
	handler.GetURLs(rec, req)
	fmt.Print(rec.Code)
	// Output: 204
}

// Example of deleting URLs
func Example_deleteURLs() {
	handler, _ := exampleHandler()

	// Add a URL for the user
	reqShort := httptest.NewRequest(http.MethodPost, uri, bytes.NewBufferString("https://example.com"))
	reqShort = reqShort.WithContext(context.WithValue(reqShort.Context(), middleware.UserIDContextKey, userID))
	recShort := httptest.NewRecorder()
	handler.ShortenURL(recShort, reqShort)
	shortURL := strings.TrimPrefix(strings.TrimSpace(recShort.Body.String()), "http://localhost:8080/")

	ids := []string{shortURL}
	b, _ := json.Marshal(ids)
	req := httptest.NewRequest(http.MethodDelete, uri+"/api/user/urls", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, userID))
	rec := httptest.NewRecorder()
	handler.DeleteURLs(rec, req)
	fmt.Print(rec.Code)
	// Output: 202
}

// Example of pinging the DB
func Example_pingDB() {
	handler, _ := exampleHandler()

	req := httptest.NewRequest(http.MethodGet, uri+"/ping", nil)
	rec := httptest.NewRecorder()
	handler.PingDB(rec, req)
	fmt.Print(rec.Code)
	// Output: 200
}
