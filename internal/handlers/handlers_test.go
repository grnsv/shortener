package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grnsv/shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleShortenURL(t *testing.T) {
	ts := httptest.NewServer(Router())
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
			name: "invalid content type",
			req: req{
				method:      http.MethodPost,
				body:        "https://practicum.yandex.ru/",
				contentType: "application/json",
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

		assert.Equal(t, tt.want.statusCode, res.StatusCode, tt.name)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		assert.Contains(t, string(resBody), tt.want.body, tt.name)
		assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"), tt.name)
	}
}

func TestHandleExpandURL(t *testing.T) {
	ts := httptest.NewServer(Router())
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

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	shorten := strings.Split(string(resBody), config.ServerAddress)[1]

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
		defer res.Body.Close()

		assert.Equal(t, tt.want.statusCode, res.StatusCode, tt.name)
		assert.Equal(t, tt.want.location, res.Header.Get("Location"), tt.name)
	}
}
