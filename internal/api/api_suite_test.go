package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/mocks"
)

func TestApi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Suite")
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
		router = api.NewRouter(handler, log)
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
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("when the database is unreachable", func() {
		It("returns status 500 Internal Server Error", func() {
			mockShortener.EXPECT().PingStorage(gomock.Any()).Return(errors.New("connection failed"))

			resp, err := http.Get(ts.URL + "/ping")
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
		})
	})
})
