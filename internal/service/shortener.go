package service

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/grnsv/shortener/internal/models"
)

const shortURLLength = 8

type Shortener interface {
	ShortenURL(url string) (string, error)
	ExpandURL(shortURL string) (string, bool)
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

func (s *URLShortener) ShortenURL(url string) (string, error) {
	model := s.generateShortURL(url)
	err := s.storage.Save(model)
	if err != nil {
		return "", err
	}

	return model.ShortURL, nil
}

func (s *URLShortener) ExpandURL(shortURL string) (string, bool) {
	return s.storage.Get(shortURL)
}
