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
		stmt *mocks.MockStmt
		s    storage.Storage
		err  error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		db = mocks.NewMockDB(ctrl)
		stmt = mocks.NewMockStmt(ctrl)
		db.EXPECT().ExecContext(gomock.Any(), gomock.Any()).Return(nil, nil)
		db.EXPECT().PreparexContext(gomock.Any(), gomock.Any()).Return(stmt, nil).Times(5)
		s, err = storage.NewDBStorage(context.Background(), db)
		Expect(err).To(BeNil())
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

		db.EXPECT().NamedExecContext(gomock.Any(), gomock.Any(), gomock.Len(2)).Return(nil, nil)
		err = s.SaveMany(context.Background(), models)
		Expect(err).To(BeNil())
	})

	When("db fails", func() {
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

			db.EXPECT().
				NamedExecContext(gomock.Any(), gomock.Any(), gomock.Len(2)).
				Return(nil, errors.New("db error"))
			err = s.SaveMany(context.Background(), models)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("DBStorage_Get", func() {
	var (
		ctrl *gomock.Controller
		db   *mocks.MockDB
		stmt *mocks.MockStmt
		s    storage.Storage
		err  error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		db = mocks.NewMockDB(ctrl)
		stmt = mocks.NewMockStmt(ctrl)
		db.EXPECT().ExecContext(gomock.Any(), gomock.Any()).Return(nil, nil)
		db.EXPECT().PreparexContext(gomock.Any(), gomock.Any()).Return(stmt, nil).Times(5)
		s, err = storage.NewDBStorage(context.Background(), db)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should get the original URL", func() {
		short := "short1"
		url := models.URL{OriginalURL: "http://original.com", IsDeleted: false}
		stmt.EXPECT().GetContext(gomock.Any(), gomock.Any(), short).DoAndReturn(
			func(ctx context.Context, dest any, args ...any) error {
				ptr := dest.(*models.URL)
				*ptr = url
				return nil
			},
		)
		orig, err := s.Get(context.Background(), short)
		Expect(err).To(BeNil())
		Expect(orig).To(Equal("http://original.com"))
	})

	It("should return ErrDeleted if url is deleted", func() {
		short := "short2"
		url := models.URL{OriginalURL: "http://deleted.com", IsDeleted: true}
		stmt.EXPECT().GetContext(gomock.Any(), gomock.Any(), short).DoAndReturn(
			func(ctx context.Context, dest any, args ...any) error {
				ptr := dest.(*models.URL)
				*ptr = url
				return nil
			},
		)
		_, err := s.Get(context.Background(), short)
		Expect(err).To(MatchError(storage.ErrDeleted))
	})

	It("should return error if GetContext fails", func() {
		short := "short3"
		stmt.EXPECT().GetContext(gomock.Any(), gomock.Any(), short).Return(errors.New("get error"))
		_, err := s.Get(context.Background(), short)
		Expect(err).To(MatchError("get error"))
	})
})

var _ = Describe("DBStorage_GetAll", func() {
	var (
		ctrl *gomock.Controller
		db   *mocks.MockDB
		stmt *mocks.MockStmt
		s    storage.Storage
		err  error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		db = mocks.NewMockDB(ctrl)
		stmt = mocks.NewMockStmt(ctrl)
		db.EXPECT().ExecContext(gomock.Any(), gomock.Any()).Return(nil, nil)
		db.EXPECT().PreparexContext(gomock.Any(), gomock.Any()).Return(stmt, nil).Times(5)
		s, err = storage.NewDBStorage(context.Background(), db)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should get all URLs for a user", func() {
		userID := "user-123"
		urls := []models.URL{
			{ShortURL: "s1", OriginalURL: "o1"},
			{ShortURL: "s2", OriginalURL: "o2"},
		}
		stmt.EXPECT().SelectContext(gomock.Any(), gomock.Any(), userID).DoAndReturn(
			func(ctx context.Context, dest any, args ...any) error {
				ptr := dest.(*[]models.URL)
				*ptr = urls
				return nil
			},
		)
		result, err := s.GetAll(context.Background(), userID)
		Expect(err).To(BeNil())
		Expect(result).To(Equal(urls))
	})

	It("should return error if SelectContext fails", func() {
		userID := "user-err"
		stmt.EXPECT().SelectContext(gomock.Any(), gomock.Any(), userID).Return(errors.New("select error"))
		_, err := s.GetAll(context.Background(), userID)
		Expect(err).To(MatchError("select error"))
	})
})

var _ = Describe("DBStorage_DeleteMany", func() {
	var (
		ctrl *gomock.Controller
		db   *mocks.MockDB
		stmt *mocks.MockStmt
		s    storage.Storage
		err  error
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		db = mocks.NewMockDB(ctrl)
		stmt = mocks.NewMockStmt(ctrl)
		db.EXPECT().ExecContext(gomock.Any(), gomock.Any()).Return(nil, nil)
		db.EXPECT().PreparexContext(gomock.Any(), gomock.Any()).Return(stmt, nil).Times(5)
		s, err = storage.NewDBStorage(context.Background(), db)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should delete many URLs", func() {
		userID := "user-del"
		shorts := []string{"s1", "s2"}
		stmt.EXPECT().ExecContext(gomock.Any(), userID, gomock.Any()).Return(nil, nil)
		err := s.DeleteMany(context.Background(), userID, shorts)
		Expect(err).To(BeNil())
	})

	It("should return error if ExecContext fails", func() {
		userID := "user-del-err"
		shorts := []string{"s3"}
		stmt.EXPECT().ExecContext(gomock.Any(), userID, gomock.Any()).Return(nil, errors.New("delete error"))
		err := s.DeleteMany(context.Background(), userID, shorts)
		Expect(err).To(MatchError("delete error"))
	})
})
