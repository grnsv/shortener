package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
)

//go:generate go mod tidy

func main() {
	cfg := config.Parse()
	log, err := logger.New(cfg.AppEnv)
	handleError(err)
	defer must(log.Sync)

	storage, err := storage.New(context.Background(), cfg)
	handleError(err)
	defer must(storage.Close)

	shortener := service.NewShortener(storage, storage, storage, storage, cfg.BaseAddress.String())
	handler := api.NewURLHandler(shortener, cfg, log)
	r := api.NewRouter(handler, cfg, log)
	handleError(http.ListenAndServe(cfg.ServerAddress.String(), r))
}

func must(fn func() error) {
	handleError(fn())
}

func handleError(err error) {
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
