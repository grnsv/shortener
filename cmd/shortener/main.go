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

func main() {
	cfg := config.Parse()
	log, err := logger.New(cfg.AppEnv)
	if err != nil {
		fatal(err)
	}
	defer log.Sync()

	storage, err := storage.New(context.Background(), cfg)
	if err != nil {
		fatal(err)
	}
	defer storage.Close()

	shortener := service.NewURLShortener(storage)
	handler := api.NewURLHandler(shortener, cfg, log)
	r := api.NewRouter(handler, log)
	if err := http.ListenAndServe(cfg.ServerAddress.String(), r); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	log.Fatalf("Server failed: %v", err)
}
