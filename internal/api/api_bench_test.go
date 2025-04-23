package api

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
)

func BenchmarkApi(b *testing.B) {
	const (
		userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
		secret = "secret"
		n      = 1000
	)

	cfg := config.New(
		config.WithJWTSecret(secret),
		config.WithDatabaseDSN("postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"),
	)
	log, err := logger.New("testing")
	if err != nil {
		b.Fatal(err)
	}
	defer log.Sync()

	storage, err := storage.New(context.Background(), cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	shortener := service.NewShortener(storage, storage, storage, storage, cfg.BaseAddress.String())
	handler := NewURLHandler(shortener, cfg, log)
	router := NewRouter(handler, cfg, log)
	cookie, err := middleware.BuildAuthCookie(secret, userID)
	if err != nil {
		b.Fatal(err)
	}

	ts := httptest.NewServer(router)
	defer ts.Close()
	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	makeRequest := func(method, path, body, contentType string) *http.Request {
		b.Helper()
		req, err := http.NewRequest(method, path, strings.NewReader(body))
		if err != nil {
			b.Fatal(err)
		}
		req.AddCookie(cookie)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		return req
	}

	doRequest := func(req *http.Request) *http.Response {
		b.Helper()
		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		return resp
	}

	req := makeRequest(http.MethodPost, ts.URL, "https://practicum.yandex.ru/", "text/plain")
	resp := doRequest(req)
	resp.Body.Close()

	b.ResetTimer()

	b.Run("ShortenURL", func(b *testing.B) {
		for i := 0; i < n; i++ {
			url := fmt.Sprintf("https://app.pachca.com/chats/%d", rand.Intn(20_000_000))
			req := makeRequest(http.MethodPost, ts.URL, url, "text/plain")
			resp := doRequest(req)
			resp.Body.Close()
		}
	})

	b.Run("ExpandURL", func(b *testing.B) {
		for i := 0; i < n; i++ {
			resp := doRequest(makeRequest(http.MethodGet, ts.URL+"/kv430TPx", "", ""))
			resp.Body.Close()
		}
	})

	b.Run("PingDB", func(b *testing.B) {
		for i := 0; i < n; i++ {
			resp := doRequest(makeRequest(http.MethodGet, ts.URL+"/ping", "", ""))
			resp.Body.Close()
		}
	})

	b.Run("ShortenURLJSON", func(b *testing.B) {
		for i := 0; i < n; i++ {
			body := fmt.Sprintf(`{"url":"https://app.pachca.com/chats/%d"}`, rand.Intn(20_000_000))
			req := makeRequest(http.MethodPost, ts.URL+"/api/shorten", body, "application/json")
			resp := doRequest(req)
			resp.Body.Close()
		}
	})

	b.Run("ShortenBatch", func(b *testing.B) {
		body := `[
			{"correlation_id":"38_go_info","original_url":"https://app.pachca.com/chats/16676594"},
			{"correlation_id":"38_go_group1_study","original_url":"https://app.pachca.com/chats/16676765"},
			{"correlation_id":"38_go_community","original_url":"https://app.pachca.com/chats/16676617"},
			{"correlation_id":"38_go_library","original_url":"https://app.pachca.com/chats/16676656"}
		]`
		for i := 0; i < n; i++ {
			req := makeRequest(http.MethodPost, ts.URL+"/api/shorten/batch", body, "application/json")
			resp := doRequest(req)
			resp.Body.Close()
		}
	})

	b.Run("GetURLs", func(b *testing.B) {
		for i := 0; i < n; i++ {
			req := makeRequest(http.MethodGet, ts.URL+"/api/user/urls", "", "")
			resp := doRequest(req)
			resp.Body.Close()
		}
	})

	b.Run("DeleteURLs", func(b *testing.B) {
		body := `[
			"jd4Hd3pG",
			"XlU8ZuE8",
			"HecihuYE",
			"wG8NPSAf"
			]`
		for i := 0; i < n; i++ {
			req := makeRequest(http.MethodDelete, ts.URL+"/api/user/urls", body, "application/json")
			resp := doRequest(req)
			resp.Body.Close()
		}
	})
}
