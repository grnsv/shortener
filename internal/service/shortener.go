package service

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/google/uuid"
	"github.com/grnsv/shortener/internal/models"
	"github.com/grnsv/shortener/internal/storage"
)

const shortURLLength = 8

type Shortener interface {
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
	ShortenBatch(ctx context.Context, longs models.BatchRequest, userID string) (models.BatchResponse, error)
	ExpandURL(ctx context.Context, shortURL string) (string, error)
	PingStorage(ctx context.Context) error
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
	DeleteMany(ctx context.Context, userID string, shortURLs []string) error
}

type URLShortener struct {
	storage     storage.Storage
	baseAddress string
}

func NewURLShortener(storage storage.Storage, baseAddress string) *URLShortener {
	return &URLShortener{storage: storage, baseAddress: baseAddress}
}

func (s *URLShortener) generateShortURL(url string, userID string) models.URL {
	uuid := uuid.NewSHA1(uuid.NameSpaceURL, []byte(url))
	return models.URL{
		UUID:        uuid.String(),
		UserID:      userID,
		ShortURL:    base64.URLEncoding.EncodeToString(uuid[:])[:shortURLLength],
		OriginalURL: url,
	}
}

func (s *URLShortener) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	model := s.generateShortURL(url, userID)
	err := s.storage.Save(ctx, model)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExist) {
			return s.baseAddress + "/" + model.ShortURL, err
		}

		return "", err
	}

	return s.baseAddress + "/" + model.ShortURL, nil
}

func (s *URLShortener) ShortenBatch(ctx context.Context, longs models.BatchRequest, userID string) (models.BatchResponse, error) {
	length := len(longs)
	shorts := make([]models.BatchResponseItem, length)
	urls := make([]models.URL, length)

	for i, long := range longs {
		url := s.generateShortURL(long.OriginalURL, userID)
		urls[i] = url
		shorts[i] = models.BatchResponseItem{
			CorrelationID: long.CorrelationID,
			ShortURL:      s.baseAddress + "/" + url.ShortURL,
		}
	}

	err := s.storage.SaveMany(ctx, urls)
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

func (s *URLShortener) ExpandURL(ctx context.Context, shortURL string) (string, error) {
	return s.storage.Get(ctx, shortURL)
}

func (s *URLShortener) PingStorage(ctx context.Context) error {
	return s.storage.Ping(ctx)
}

func (s *URLShortener) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	urls, err := s.storage.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range urls {
		urls[i].ShortURL = s.baseAddress + "/" + urls[i].ShortURL
	}

	return urls, nil
}

func (s *URLShortener) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	return s.storage.DeleteMany(ctx, userID, shortURLs)
}
