package main

import (
	"context"
	"net/http"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
)

func main() {
	cfg := config.Get()
	logger.Initialize(cfg.AppEnv)
	defer logger.Log.Sync()

	storage, err := service.NewStorage(context.Background())
	if err != nil {
		fatal(err)
	}
	defer storage.Close()

	r := api.NewRouter(service.NewURLShortener(storage))
	if err := http.ListenAndServe(cfg.ServerAddress.String(), r); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	logger.Log.Fatalf("Server failed: %v", err)
}
