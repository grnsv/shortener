package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func requireNoError(t *testing.T, fn func() error, msgAndArgs ...any) {
	require.NoError(t, fn(), msgAndArgs...)
}

func closeBody(t *testing.T, r *http.Response, msgAndArgs ...any) {
	require.NoError(t, r.Body.Close(), msgAndArgs...)
}

func TestHandleShortenURL(t *testing.T) {
	storage, err := storage.NewMemoryStorage(context.Background())
	defer requireNoError(t, storage.Close)
	require.NoError(t, err)
	cfg := config.New(
		config.WithAppEnv("testing"),
		config.WithServerAddress(config.NetAddress{Host: "localhost", Port: 8080}),
		config.WithBaseURL(config.BaseURL{Scheme: "http://", Address: config.NetAddress{Host: "localhost", Port: 8080}}),
	)
	shortener := service.NewShortener(storage, storage, storage, storage, cfg.BaseURL.String())
	log, err := logger.New("testing")
	require.NoError(t, err)
	handler := NewURLHandler(shortener, cfg, log)
	ts := httptest.NewServer(NewRouter(handler, cfg, log))
	defer ts.Close()

	type req struct {
		method      string
		body        string
		contentType string
	}
	type want struct {
		statusCode  int
		body        string
		contentType string
	}
	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			name: "positive test",
			req: req{
				method:      http.MethodPost,
				body:        "https://practicum.yandex.ru/",
				contentType: "text/plain",
			},
			want: want{
				statusCode:  http.StatusCreated,
				body:        "http://",
				contentType: "text/plain",
			},
		},
		{
			name: "invalid method",
			req: req{
				method:      http.MethodGet,
				body:        "https://practicum.yandex.ru/",
				contentType: "text/plain",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty body",
			req: req{
				method:      http.MethodPost,
				body:        "",
				contentType: "text/plain",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		request, err := http.NewRequest(tt.req.method, ts.URL, strings.NewReader(tt.req.body))
		require.NoError(t, err, tt.name)

		request.Header.Add("Content-Type", tt.req.contentType)

		res, err := ts.Client().Do(request)
		require.NoError(t, err, tt.name)
		defer closeBody(t, res, tt.name)

		assert.Equal(t, tt.want.statusCode, res.StatusCode, tt.name)
		if tt.want.body != "" {
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err, tt.name)
			assert.Contains(t, string(resBody), tt.want.body, tt.name)
		}

		assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"), tt.name)
	}
}

func TestHandleExpandURL(t *testing.T) {
	storage, err := storage.NewMemoryStorage(context.Background())
	defer requireNoError(t, storage.Close)
	require.NoError(t, err)
	cfg := config.New(
		config.WithAppEnv("testing"),
		config.WithServerAddress(config.NetAddress{Host: "localhost", Port: 8080}),
		config.WithBaseURL(config.BaseURL{Scheme: "http://", Address: config.NetAddress{Host: "localhost", Port: 8080}}),
	)
	shortener := service.NewShortener(storage, storage, storage, storage, cfg.BaseURL.String())
	log, err := logger.New("testing")
	require.NoError(t, err)
	handler := NewURLHandler(shortener, cfg, log)
	ts := httptest.NewServer(NewRouter(handler, cfg, log))
	defer ts.Close()

	client := ts.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	request, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader("https://practicum.yandex.ru/"))
	require.NoError(t, err)

	request.Header.Add("Content-Type", "text/plain")

	res, err := client.Do(request)
	require.NoError(t, err)
	defer closeBody(t, res)

	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	shorten := strings.Split(string(resBody), cfg.BaseURL.String())[1]

	type req struct {
		method string
		target string
	}
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			name: "positive test",
			req: req{
				method: http.MethodGet,
				target: shorten,
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "https://practicum.yandex.ru/",
			},
		},
		{
			name: "invalid method",
			req: req{
				method: http.MethodHead,
				target: shorten,
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "empty path",
			req: req{
				method: http.MethodGet,
				target: "/",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "invalid path",
			req: req{
				method: http.MethodGet,
				target: "/practicum.yandex.ru",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		request, err := http.NewRequest(tt.req.method, ts.URL+tt.req.target, nil)
		require.NoError(t, err, tt.name)

		res, err := client.Do(request)
		require.NoError(t, err, tt.name)
		defer closeBody(t, res, tt.name)

		assert.Equal(t, tt.want.statusCode, res.StatusCode, tt.name)
		assert.Equal(t, tt.want.location, res.Header.Get("Location"), tt.name)
	}
}

func TestHandleShortenURLJSON(t *testing.T) {
	storage, err := storage.NewMemoryStorage(context.Background())
	defer requireNoError(t, storage.Close)
	require.NoError(t, err)
	cfg := config.New(
		config.WithAppEnv("testing"),
		config.WithServerAddress(config.NetAddress{Host: "localhost", Port: 8080}),
		config.WithBaseURL(config.BaseURL{Scheme: "http://", Address: config.NetAddress{Host: "localhost", Port: 8080}}),
	)
	shortener := service.NewShortener(storage, storage, storage, storage, cfg.BaseURL.String())
	log, err := logger.New("testing")
	require.NoError(t, err)
	handler := NewURLHandler(shortener, cfg, log)
	ts := httptest.NewServer(NewRouter(handler, cfg, log))
	defer ts.Close()

	type req struct {
		method      string
		body        models.ShortenRequest
		contentType string
	}
	type want struct {
		statusCode  int
		body        models.ShortenResponse
		contentType string
	}
	tests := []struct {
		name string
		req  req
		want want
	}{
		{
			name: "positive test",
			req: req{
				method:      http.MethodPost,
				body:        models.ShortenRequest{URL: "https://practicum.yandex.ru/"},
				contentType: "application/json",
			},
			want: want{
				statusCode: http.StatusCreated,
				body: models.ShortenResponse{
					Result: "http://",
				},
				contentType: "application/json",
			},
		},
		{
			name: "invalid method",
			req: req{
				method:      http.MethodGet,
				body:        models.ShortenRequest{URL: "https://practicum.yandex.ru/"},
				contentType: "application/json",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/x-gzip",
			},
		},
		{
			name: "empty body",
			req: req{
				method:      http.MethodPost,
				body:        models.ShortenRequest{URL: ""},
				contentType: "application/json",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/x-gzip",
			},
		},
	}
	for _, tt := range tests {
		body, err := json.Marshal(tt.req.body)
		require.NoError(t, err, tt.name)

		request, err := http.NewRequest(tt.req.method, ts.URL+"/api/shorten", bytes.NewReader(body))
		require.NoError(t, err, tt.name)

		request.Header.Add("Content-Type", tt.req.contentType)

		res, err := ts.Client().Do(request)
		require.NoError(t, err, tt.name)
		defer closeBody(t, res, tt.name)

		assert.Equal(t, tt.want.statusCode, res.StatusCode, tt.name)
		if tt.want.body.Result != "" {
			var resp models.ShortenResponse
			err = json.NewDecoder(res.Body).Decode(&resp)

			require.NoError(t, err, tt.name)
			assert.Contains(t, resp.Result, tt.want.body.Result, tt.name)
		}

		assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"), tt.name)
	}
}
