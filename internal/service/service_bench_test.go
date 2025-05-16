package service

import (
	"context"
	"testing"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/storage"
)

func BenchmarkService(b *testing.B) {
	cfg := config.New(
		config.WithDatabaseDSN("postgres://postgres:postgres@postgres:5432/praktikum?sslmode=disable"),
	)
	log, err := logger.New("testing")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		err = log.Sync()
		if err != nil {
			b.Fatal(err)
		}
	}()

	storage, err := storage.New(context.Background(), cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		err = storage.Close()
		if err != nil {
			b.Fatal(err)
		}
	}()

	shortener := NewShortener(storage, storage, storage, storage, cfg.BaseURL.String())

	b.ResetTimer()

	b.Run("ExpandURL", func(b *testing.B) {
		for b.Loop() {
			_, err := shortener.ExpandURL(context.Background(), "kv430TPx")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
