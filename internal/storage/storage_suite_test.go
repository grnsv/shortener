package storage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/grnsv/shortener/internal/mocks"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

var _ = Describe("DBStorage_SaveMany", func() {
	var (
		ctrl *gomock.Controller
		db   *mocks.MockDB
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		db = mocks.NewMockDB(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should save a batch of URLs", func() {
		models := []models.URL{
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

		db.EXPECT().ExecContext(gomock.Any(), gomock.Any()).Return(nil, nil)
		s, err := storage.NewDBStorage(context.Background(), db)
		Expect(err).To(BeNil())
		db.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Len(2)).Return(nil, nil)

		err = s.SaveMany(context.Background(), models)
		Expect(err).To(BeNil())
	})

	Context("when db fails", func() {
		It("returns an error", func() {
			models := []models.URL{
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

			db.EXPECT().ExecContext(gomock.Any(), gomock.Any()).Return(nil, nil)
			s, err := storage.NewDBStorage(context.Background(), db)
			Expect(err).To(BeNil())
			db.EXPECT().
				NamedExecContext(gomock.Any(), gomock.Any(), gomock.Len(2)).
				Return(nil, errors.New("db error"))

			err = s.SaveMany(context.Background(), models)
			Expect(err).To(HaveOccurred())
		})
	})
})
