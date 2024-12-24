package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleShortenURL(t *testing.T) {
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
		request := httptest.NewRequest(tt.req.method, "/", strings.NewReader(tt.req.body))
		request.Header.Add("Content-Type", tt.req.contentType)
		w := httptest.NewRecorder()
		HandleShortenURL(w, request)

		res := w.Result()
		assert.Equal(t, tt.want.statusCode, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)

		require.NoError(t, err)
		assert.Contains(t, string(resBody), tt.want.body)
		assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
	}
}

func TestHandleExpandURL(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://practicum.yandex.ru/"))
	request.Header.Add("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	HandleShortenURL(w, request)
	res := w.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	shorten := string(resBody)

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
		request = httptest.NewRequest(tt.req.method, tt.req.target, nil)
		w = httptest.NewRecorder()
		HandleExpandURL(w, request)

		res = w.Result()
		defer res.Body.Close()
		assert.Equal(t, tt.want.statusCode, res.StatusCode)
		assert.Equal(t, tt.want.location, res.Header.Get("Location"))
	}
}
