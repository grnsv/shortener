package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grnsv/shortener/internal/mocks"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

var _ = Describe("ShortenBatch", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl      *gomock.Controller
		store     *mocks.MockStorage
		shortener service.Shortener
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockStorage(ctrl)
		shortener = service.NewShortener(store, store, store, store, "")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should shorten a batch of URLs", func() {
		batchRequest := models.BatchRequest{
			models.BatchRequestItem{
				CorrelationID: "00000000-0000-0000-0000-000000000001",
				OriginalURL:   "http://example.com/1",
			},
			models.BatchRequestItem{
				CorrelationID: "00000000-0000-0000-0000-000000000002",
				OriginalURL:   "http://example.com/2",
			},
		}

		store.EXPECT().SaveMany(gomock.Any(), gomock.Len(2)).Return(nil)

		batchResponse, err := shortener.ShortenBatch(context.Background(), batchRequest, userID)
		Expect(err).To(BeNil())
		Expect(batchResponse).To(HaveLen(2))
		Expect(batchResponse[0].CorrelationID).To(Equal("00000000-0000-0000-0000-000000000001"))
		Expect(batchResponse[1].CorrelationID).To(Equal("00000000-0000-0000-0000-000000000002"))
	})

	When("storage fails", func() {
		It("returns an error", func() {
			batchRequest := models.BatchRequest{
				models.BatchRequestItem{
					CorrelationID: "00000000-0000-0000-0000-000000000001",
					OriginalURL:   "http://example.com/1",
				},
				models.BatchRequestItem{
					CorrelationID: "00000000-0000-0000-0000-000000000002",
					OriginalURL:   "http://example.com/2",
				},
			}

			store.EXPECT().SaveMany(gomock.Any(), gomock.Len(2)).Return(errors.New("storage error"))

			batchResponse, err := shortener.ShortenBatch(context.Background(), batchRequest, userID)
			Expect(err).To(HaveOccurred())
			Expect(batchResponse).To(BeNil())
		})
	})
})

var _ = Describe("ShortenURL", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl      *gomock.Controller
		store     *mocks.MockStorage
		shortener service.Shortener
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockStorage(ctrl)
		shortener = service.NewShortener(store, store, store, store, "http://short")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should shorten a single URL", func() {
		store.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
		shortURL, err := shortener.ShortenURL(context.Background(), "http://example.com/1", userID)
		Expect(err).To(BeNil())
		Expect(shortURL).To(HavePrefix("http://short/"))
	})

	It("should return existing short URL if already exists", func() {
		store.EXPECT().Save(gomock.Any(), gomock.Any()).Return(storage.ErrAlreadyExist)
		shortURL, err := shortener.ShortenURL(context.Background(), "http://example.com/1", userID)
		Expect(err).To(HaveOccurred())
		Expect(shortURL).To(HavePrefix("http://short/"))
	})
})

var _ = Describe("ExpandURL", func() {
	var (
		ctrl      *gomock.Controller
		store     *mocks.MockStorage
		shortener service.Shortener
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockStorage(ctrl)
		shortener = service.NewShortener(store, store, store, store, "http://short")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should expand a short URL", func() {
		store.EXPECT().Get(gomock.Any(), "short123").Return("http://example.com/1", nil)
		orig, err := shortener.ExpandURL(context.Background(), "short123")
		Expect(err).To(BeNil())
		Expect(orig).To(Equal("http://example.com/1"))
	})

	It("should return error if not found", func() {
		store.EXPECT().Get(gomock.Any(), "short404").Return("", errors.New("not found"))
		orig, err := shortener.ExpandURL(context.Background(), "short404")
		Expect(err).To(HaveOccurred())
		Expect(orig).To(BeEmpty())
	})
})

var _ = Describe("GetAll", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl      *gomock.Controller
		store     *mocks.MockStorage
		shortener service.Shortener
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockStorage(ctrl)
		shortener = service.NewShortener(store, store, store, store, "http://short")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return all URLs for a user", func() {
		urls := []models.URL{
			{ShortURL: "abc123", OriginalURL: "http://example.com/1"},
			{ShortURL: "def456", OriginalURL: "http://example.com/2"},
		}
		store.EXPECT().GetAll(gomock.Any(), userID).Return(urls, nil)
		result, err := shortener.GetAll(context.Background(), userID)
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(2))
		Expect(result[0].ShortURL).To(Equal("http://short/abc123"))
		Expect(result[1].ShortURL).To(Equal("http://short/def456"))
	})

	It("should return error if storage fails", func() {
		store.EXPECT().GetAll(gomock.Any(), userID).Return(nil, errors.New("fail"))
		result, err := shortener.GetAll(context.Background(), userID)
		Expect(err).To(HaveOccurred())
		Expect(result).To(BeNil())
	})
})

var _ = Describe("DeleteMany", func() {
	const userID = "ffffffff-ffff-ffff-ffff-ffffffffffff"
	var (
		ctrl      *gomock.Controller
		store     *mocks.MockStorage
		shortener service.Shortener
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockStorage(ctrl)
		shortener = service.NewShortener(store, store, store, store, "http://short")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should delete many URLs", func() {
		shorts := []string{"abc123", "def456"}
		store.EXPECT().DeleteMany(gomock.Any(), userID, shorts).Return(nil)
		err := shortener.DeleteMany(context.Background(), userID, shorts)
		Expect(err).To(BeNil())
	})

	It("should return error if delete fails", func() {
		shorts := []string{"abc123", "def456"}
		store.EXPECT().DeleteMany(gomock.Any(), userID, shorts).Return(errors.New("fail"))
		err := shortener.DeleteMany(context.Background(), userID, shorts)
		Expect(err).To(HaveOccurred())
	})
})
