package main

import (
	"net/http"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
)

func main() {
	logger.Initialize(config.Get().AppEnv)
	defer logger.Log.Sync()

	storage, err := service.NewStorage("file")
	if err != nil {
		fatal(err)
	}
	defer storage.Close()

	r := api.NewRouter(service.NewURLShortener(storage))

	if err := http.ListenAndServe(config.Get().ServerAddress.String(), r); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	logger.Log.Fatalf("Server failed: %v", err)
}
