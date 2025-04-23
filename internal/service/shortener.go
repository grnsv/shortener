// Package service provides URL shortening, expansion, and management services.
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

// Shortener aggregates all URL shortening and management interfaces.
type Shortener interface {
	URLShortener
	BatchShortener
	URLExpander
	StoragePinger
	URLLister
	URLDeleter
}

// URLShortener provides a method to shorten a single URL.
type URLShortener interface {
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
}

// BatchShortener provides a method to shorten a batch of URLs.
type BatchShortener interface {
	ShortenBatch(ctx context.Context, longs models.BatchRequest, userID string) (models.BatchResponse, error)
}

// URLExpander provides a method to expand a shortened URL to its original form.
type URLExpander interface {
	ExpandURL(ctx context.Context, shortURL string) (string, error)
}

// StoragePinger provides a method to check the availability of the underlying storage.
type StoragePinger interface {
	PingStorage(ctx context.Context) error
}

// URLLister provides a method to list all URLs for a user.
type URLLister interface {
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
}

// URLDeleter provides a method to delete multiple shortened URLs for a user.
type URLDeleter interface {
	DeleteMany(ctx context.Context, userID string, shortURLs []string) error
}

// Service implements the Shortener interface and provides URL shortening services.
type Service struct {
	saver       storage.Saver
	retriever   storage.Retriever
	deleter     storage.Deleter
	pinger      storage.Pinger
	baseAddress string
}

// NewShortener creates a new Service implementing the Shortener interface.
func NewShortener(
	saver storage.Saver,
	retriever storage.Retriever,
	deleter storage.Deleter,
	pinger storage.Pinger,
	baseAddress string,
) Shortener {
	return &Service{
		saver:       saver,
		retriever:   retriever,
		deleter:     deleter,
		pinger:      pinger,
		baseAddress: baseAddress,
	}
}

func (s *Service) generateShortURL(url string, userID string) models.URL {
	uuid := uuid.NewSHA1(uuid.NameSpaceURL, []byte(url))
	return models.URL{
		UUID:        uuid.String(),
		UserID:      userID,
		ShortURL:    base64.URLEncoding.EncodeToString(uuid[:])[:shortURLLength],
		OriginalURL: url,
	}
}

// ShortenURL shortens the given URL for the specified user and returns the shortened URL.
func (s *Service) ShortenURL(ctx context.Context, url string, userID string) (string, error) {
	model := s.generateShortURL(url, userID)
	err := s.saver.Save(ctx, model)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExist) {
			return s.baseAddress + "/" + model.ShortURL, err
		}

		return "", err
	}

	return s.baseAddress + "/" + model.ShortURL, nil
}

// ShortenBatch shortens a batch of URLs for the specified user and returns the batch response.
func (s *Service) ShortenBatch(ctx context.Context, longs models.BatchRequest, userID string) (models.BatchResponse, error) {
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

	err := s.saver.SaveMany(ctx, urls)
	if err != nil {
		return nil, err
	}

	return shorts, nil
}

// ExpandURL expands the given shortened URL to its original URL.
func (s *Service) ExpandURL(ctx context.Context, shortURL string) (string, error) {
	return s.retriever.Get(ctx, shortURL)
}

// PingStorage checks the availability of the underlying storage.
func (s *Service) PingStorage(ctx context.Context) error {
	return s.pinger.Ping(ctx)
}

// GetAll returns all URLs associated with the specified user.
func (s *Service) GetAll(ctx context.Context, userID string) ([]models.URL, error) {
	urls, err := s.retriever.GetAll(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range urls {
		urls[i].ShortURL = s.baseAddress + "/" + urls[i].ShortURL
	}

	return urls, nil
}

// DeleteMany deletes multiple shortened URLs for the specified user.
func (s *Service) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	return s.deleter.DeleteMany(ctx, userID, shortURLs)
}
