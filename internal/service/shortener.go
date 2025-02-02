package service

import (
	"context"
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/grnsv/shortener/internal/models"
)

const shortURLLength = 8

type Shortener interface {
	ShortenURL(ctx context.Context, url string) (string, error)
	ExpandURL(ctx context.Context, shortURL string) (string, error)
}

type URLShortener struct {
	storage Storage
}

func NewURLShortener(storage Storage) *URLShortener {
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
		return "", err
	}

	return model.ShortURL, nil
}

func (s *URLShortener) ExpandURL(ctx context.Context, shortURL string) (string, error) {
	return s.storage.Get(ctx, shortURL)
}
