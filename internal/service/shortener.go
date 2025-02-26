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
	URLShortener
	BatchShortener
	URLExpander
	StoragePinger
	URLLister
	URLDeleter
}

type URLShortener interface {
	ShortenURL(ctx context.Context, url string, userID string) (string, error)
}

type BatchShortener interface {
	ShortenBatch(ctx context.Context, longs models.BatchRequest, userID string) (models.BatchResponse, error)
}

type URLExpander interface {
	ExpandURL(ctx context.Context, shortURL string) (string, error)
}

type StoragePinger interface {
	PingStorage(ctx context.Context) error
}

type URLLister interface {
	GetAll(ctx context.Context, userID string) ([]models.URL, error)
}

type URLDeleter interface {
	DeleteMany(ctx context.Context, userID string, shortURLs []string) error
}

type Service struct {
	saver       storage.Saver
	retriever   storage.Retriever
	deleter     storage.Deleter
	pinger      storage.Pinger
	baseAddress string
}

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

func (s *Service) ExpandURL(ctx context.Context, shortURL string) (string, error) {
	return s.retriever.Get(ctx, shortURL)
}

func (s *Service) PingStorage(ctx context.Context) error {
	return s.pinger.Ping(ctx)
}

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

func (s *Service) DeleteMany(ctx context.Context, userID string, shortURLs []string) error {
	return s.deleter.DeleteMany(ctx, userID, shortURLs)
}
