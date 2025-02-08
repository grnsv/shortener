package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grnsv/shortener/internal/mocks"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/service"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

var _ = Describe("ShortenBatch", func() {
	var (
		ctrl      *gomock.Controller
		storage   *mocks.MockStorage
		shortener *service.URLShortener
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		storage = mocks.NewMockStorage(ctrl)
		shortener = service.NewURLShortener(storage)
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

		storage.EXPECT().SaveMany(gomock.Any(), gomock.Len(2)).Return(nil)

		batchResponse, err := shortener.ShortenBatch(context.Background(), batchRequest, "")
		Expect(err).To(BeNil())
		Expect(batchResponse).To(HaveLen(2))
		Expect(batchResponse[0].CorrelationID).To(Equal("00000000-0000-0000-0000-000000000001"))
		Expect(batchResponse[1].CorrelationID).To(Equal("00000000-0000-0000-0000-000000000002"))
		Expect(batchResponse[0].ShortURL).To(HaveLen(8))
		Expect(batchResponse[1].ShortURL).To(HaveLen(8))
	})

	Context("when storage fails", func() {
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

			storage.EXPECT().SaveMany(gomock.Any(), gomock.Len(2)).Return(errors.New("storage error"))

			batchResponse, err := shortener.ShortenBatch(context.Background(), batchRequest, "")
			Expect(err).To(HaveOccurred())
			Expect(batchResponse).To(BeNil())
		})
	})
})
