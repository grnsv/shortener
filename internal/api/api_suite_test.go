package api_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/mocks"
	"github.com/grnsv/shortener/internal/models"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
}

func must(fn func() error) {
	handleError(fn())
}

func handleError(err error) {
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("PingDB Handler", func() {
	var (
		ctrl          *gomock.Controller
		mockShortener *mocks.MockShortener
		cfg           *config.Config
		log           logger.Logger
		handler       *api.URLHandler
		router        chi.Router
		ts            *httptest.Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockShortener = mocks.NewMockShortener(ctrl)
		cfg = config.New()
		log, _ = logger.New("testing")
		handler = api.NewURLHandler(mockShortener, cfg, log)
		router = api.NewRouter(handler, cfg, log)
		ts = httptest.NewServer(router)
	})

	AfterEach(func() {
		ts.Close()
		ctrl.Finish()
	})

	Context("when the database is reachable", func() {
		It("returns status 200 OK", func() {
			mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)

			resp, err := http.Get(ts.URL + "/ping")
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("when the database is unreachable", func() {
		It("returns status 500 Internal Server Error", func() {
			mockShortener.EXPECT().PingStorage(gomock.Any()).Return(errors.New("connection failed"))

			resp, err := http.Get(ts.URL + "/ping")
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
		})
	})
})

var _ = Describe("Authenticate", func() {
	var (
		ctrl          *gomock.Controller
		mockShortener *mocks.MockShortener
		cfg           *config.Config
		log           logger.Logger
		handler       *api.URLHandler
		router        chi.Router
		ts            *httptest.Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockShortener = mocks.NewMockShortener(ctrl)
		cfg = config.New(config.WithJWTSecret("secret"))
		log, _ = logger.New("testing")
		handler = api.NewURLHandler(mockShortener, cfg, log)
		router = api.NewRouter(handler, cfg, log)
		ts = httptest.NewServer(router)
	})

	AfterEach(func() {
		ts.Close()
		ctrl.Finish()
	})

	Context("when request does not have token", func() {
		It("returns new token in cookie", func() {
			mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)

			resp, err := http.Get(ts.URL + "/ping")
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var tokenCookie *http.Cookie
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "token" {
					tokenCookie = cookie
					break
				}
			}
			Expect(tokenCookie).NotTo(BeNil())
			Expect(tokenCookie.Value).NotTo(BeEmpty())
		})
	})

	Context("when request has invalid token", func() {
		It("returns new token in cookie", func() {
			mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)

			req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
			handleError(err)
			req.AddCookie(&http.Cookie{Name: "token", Value: "invalid"})
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var tokenCookie *http.Cookie
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "token" {
					tokenCookie = cookie
					break
				}
			}
			Expect(tokenCookie).NotTo(BeNil())
			Expect(tokenCookie.Value).NotTo(BeEmpty())
		})
	})

	Context("when request has empty token", func() {
		It("returns new token in cookie", func() {
			mockShortener.EXPECT().PingStorage(gomock.Any()).Return(nil)

			req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
			handleError(err)
			req.AddCookie(&http.Cookie{Name: "token", Value: ""})
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			var tokenCookie *http.Cookie
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "token" {
					tokenCookie = cookie
					break
				}
			}
			Expect(tokenCookie).NotTo(BeNil())
			Expect(tokenCookie.Value).NotTo(BeEmpty())
		})
	})

	Context("when request has empty user ID", func() {
		It("returns status 401 Unauthorized", func() {
			req, err := http.NewRequest("GET", ts.URL+"/ping", nil)
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", "")
			handleError(err)
			req.AddCookie(cookie)
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})
})

var _ = Describe("GetURLs", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl          *gomock.Controller
		mockShortener *mocks.MockShortener
		cfg           *config.Config
		log           logger.Logger
		handler       *api.URLHandler
		router        chi.Router
		ts            *httptest.Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockShortener = mocks.NewMockShortener(ctrl)
		cfg = config.New(config.WithJWTSecret("secret"))
		log, _ = logger.New("testing")
		handler = api.NewURLHandler(mockShortener, cfg, log)
		router = api.NewRouter(handler, cfg, log)
		ts = httptest.NewServer(router)
	})

	AfterEach(func() {
		ts.Close()
		ctrl.Finish()
	})

	Context("when storage does not have stored urls", func() {
		It("returns status 204 StatusNoContent", func() {
			mockShortener.EXPECT().GetAll(gomock.Any(), userID).Return(nil, nil)

			req, err := http.NewRequest("GET", ts.URL+"/api/user/urls", nil)
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", userID)
			handleError(err)
			req.AddCookie(cookie)
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
	})

	Context("when storage has stored urls", func() {
		It("returns status 200 OK", func() {
			urls := []models.URL{
				{
					UUID:        "00000000-0000-0000-0000-000000000001",
					ShortURL:    "00000001",
					OriginalURL: "http://example.com/1",
				},
				{
					UUID:        "00000000-0000-0000-0000-000000000002",
					ShortURL:    "00000002",
					OriginalURL: "http://example.com/2",
				},
			}
			mockShortener.EXPECT().GetAll(gomock.Any(), userID).Return(urls, nil)

			req, err := http.NewRequest("GET", ts.URL+"/api/user/urls", nil)
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", userID)
			handleError(err)
			req.AddCookie(cookie)
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

			var responseURLs []models.URL
			err = json.NewDecoder(resp.Body).Decode(&responseURLs)
			handleError(err)
			Expect(len(urls)).To(Equal(len(responseURLs)))
		})
	})
})

var _ = Describe("ShortenBatch", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl          *gomock.Controller
		mockShortener *mocks.MockShortener
		cfg           *config.Config
		log           logger.Logger
		handler       *api.URLHandler
		router        chi.Router
		ts            *httptest.Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockShortener = mocks.NewMockShortener(ctrl)
		cfg = config.New(config.WithJWTSecret("secret"))
		log, _ = logger.New("testing")
		handler = api.NewURLHandler(mockShortener, cfg, log)
		router = api.NewRouter(handler, cfg, log)
		ts = httptest.NewServer(router)
	})

	AfterEach(func() {
		ts.Close()
		ctrl.Finish()
	})

	Context("when batch request is valid", func() {
		It("returns status 201 Created and batch response", func() {
			batchReq := models.BatchRequest{
				{CorrelationID: "1", OriginalURL: "http://example.com/1"},
				{CorrelationID: "2", OriginalURL: "http://example.com/2"},
			}
			batchResp := models.BatchResponse{
				{CorrelationID: "1", ShortURL: "http://localhost:8080/short1"},
				{CorrelationID: "2", ShortURL: "http://localhost:8080/short2"},
			}
			mockShortener.EXPECT().ShortenBatch(gomock.Any(), batchReq, userID).Return(batchResp, nil)

			body, err := json.Marshal(batchReq)
			handleError(err)
			req, err := http.NewRequest("POST", ts.URL+"/api/shorten/batch", bytes.NewReader(body))
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", userID)
			handleError(err)
			req.AddCookie(cookie)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			var got models.BatchResponse
			err = json.NewDecoder(resp.Body).Decode(&got)
			handleError(err)
			Expect(got).To(Equal(batchResp))
		})
	})

	Context("when batch request is invalid", func() {
		It("returns status 400 BadRequest", func() {
			body := []byte(`invalid json`)
			req, err := http.NewRequest("POST", ts.URL+"/api/shorten/batch", bytes.NewReader(body))
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", userID)
			handleError(err)
			req.AddCookie(cookie)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})

var _ = Describe("DeleteURLs", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl          *gomock.Controller
		mockShortener *mocks.MockShortener
		cfg           *config.Config
		log           logger.Logger
		handler       *api.URLHandler
		router        chi.Router
		ts            *httptest.Server
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockShortener = mocks.NewMockShortener(ctrl)
		cfg = config.New(config.WithJWTSecret("secret"))
		log, _ = logger.New("testing")
		handler = api.NewURLHandler(mockShortener, cfg, log)
		router = api.NewRouter(handler, cfg, log)
		ts = httptest.NewServer(router)
	})

	AfterEach(func() {
		ts.Close()
		ctrl.Finish()
	})

	Context("when delete request is valid", func() {
		It("returns status 202 Accepted", func() {
			shortURLs := []string{"short1", "short2"}
			mockShortener.EXPECT().DeleteMany(gomock.Any(), userID, shortURLs).Return(nil)

			body, _ := json.Marshal(shortURLs)
			req, err := http.NewRequest("DELETE", ts.URL+"/api/user/urls", bytes.NewReader(body))
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", userID)
			handleError(err)
			req.AddCookie(cookie)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusAccepted))
		})
	})

	Context("when delete request is invalid", func() {
		It("returns status 400 BadRequest", func() {
			body := []byte(`invalid json`)
			req, err := http.NewRequest("DELETE", ts.URL+"/api/user/urls", bytes.NewReader(body))
			handleError(err)
			cookie, err := middleware.BuildAuthCookie("secret", userID)
			handleError(err)
			req.AddCookie(cookie)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			handleError(err)
			defer must(resp.Body.Close)

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})
