package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grnsv/shortener/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWithLogging(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockLogger.EXPECT().Infoln(
		"uri", gomock.Any(),
		"method", gomock.Any(),
		"duration", gomock.Any(),
		"status", gomock.Any(),
		"size", gomock.Any(),
	).Times(1)

	handler := WithLogging(mockLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestWithCompressing(t *testing.T) {
	t.Run("gzip supported and content-type supported", func(t *testing.T) {
		handler := WithCompressing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/html")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		gz, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
		if assert.NoError(t, err) {
			decompressed, err := io.ReadAll(gz)
			assert.NoError(t, err)
			assert.Equal(t, "OK", string(decompressed))
			gz.Close()
		}
	})

	t.Run("gzip supported but content-type not supported", func(t *testing.T) {
		handler := WithCompressing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/xml")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEqual(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "OK", rec.Body.String())
	})

	t.Run("gzip not supported", func(t *testing.T) {
		handler := WithCompressing(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}))
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Content-Type", "text/html")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEqual(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "OK", rec.Body.String())
	})
}

func TestAuthenticate(t *testing.T) {
	const secret = "test-secret"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLogger(ctrl)

	// Helper handler to check userID in context
	checkUserIDHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDContextKey).(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userID))
	})

	middleware := Authenticate(secret, mockLogger)

	t.Run("no token: should set new token and userID", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		wrapped := middleware(checkUserIDHandler)
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		resp := rec.Result()
		defer resp.Body.Close()
		cookie := resp.Cookies()
		found := false
		for _, c := range cookie {
			if c.Name == "token" && c.Value != "" {
				found = true
			}
		}
		assert.True(t, found, "should set token cookie")
	})

	t.Run("valid token: should refresh token and keep userID", func(t *testing.T) {
		userID := "test-user-id"
		cookie, err := BuildAuthCookie(secret, userID)
		assert.NoError(t, err)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		wrapped := middleware(checkUserIDHandler)
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), userID)
		resp := rec.Result()
		defer resp.Body.Close()
		// Should refresh token
		updated := false
		for _, c := range resp.Cookies() {
			if c.Name == "token" && c.Value != "" {
				updated = true
			}
		}
		assert.True(t, updated, "should refresh token cookie")
	})

	t.Run("invalid token: should set new token and userID", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "invalid-token"})
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		wrapped := middleware(checkUserIDHandler)
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		resp := rec.Result()
		defer resp.Body.Close()
		cookie := resp.Cookies()
		found := false
		for _, c := range cookie {
			if c.Name == "token" && c.Value != "" {
				found = true
			}
		}
		assert.True(t, found, "should set new token cookie")
	})

	t.Run("token with empty userID: should return 401", func(t *testing.T) {
		cookie, err := BuildAuthCookie(secret, "")
		assert.NoError(t, err)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(cookie)
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		wrapped := middleware(checkUserIDHandler)
		wrapped.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}
