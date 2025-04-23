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
	defer log.Sync()

	storage, err := storage.New(context.Background(), cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer storage.Close()

	shortener := NewShortener(storage, storage, storage, storage, cfg.BaseAddress.String())

	b.ResetTimer()

	b.Run("ExpandURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			shortener.ExpandURL(context.Background(), "kv430TPx")
		}
	})
}
