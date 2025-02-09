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
	ShortenURL(ctx context.Context, url string) (string, error)
	ShortenBatch(ctx context.Context, longs models.BatchRequest, baseAddress string) (models.BatchResponse, error)
	ExpandURL(ctx context.Context, shortURL string) (string, error)
	PingStorage(ctx context.Context) error
}

type URLShortener struct {
	storage storage.Storage
}

func NewURLShortener(storage storage.Storage) *URLShortener {
	return &URLShortener{storage: storage}
}

func (s *URLShortener) generateShortURL(url string) models.URL {
	uuid := uuid.NewSHA1(uuid.NameSpaceURL, []byte(url))
	return models.URL{
		UUID:        uuid.String(),
		ShortURL:    base64.URLEncoding.EncodeToString(uuid[:])[:shortURLLength],
		OriginalURL: url,
	}
}

func (s *URLShortener) ShortenURL(ctx context.Context, url string) (string, error) {
	model := s.generateShortURL(url)
	err := s.storage.Save(ctx, model)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExist) {
			return model.ShortURL, err
		}

		return "", err
	}

	return model.ShortURL, nil
}

func (s *URLShortener) ShortenBatch(ctx context.Context, longs models.BatchRequest, baseAddress string) (models.BatchResponse, error) {
	length := len(longs)
	shorts := make([]models.BatchResponseItem, length)
	urls := make([]models.URL, length)

	for i, long := range longs {
		url := s.generateShortURL(long.OriginalURL)
		urls[i] = url
		shorts[i] = models.BatchResponseItem{
			CorrelationID: long.CorrelationID,
			ShortURL:      baseAddress + url.ShortURL,
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
