package main

import (
	"net/http"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
)

func main() {
	logger.Initialize(config.Get().AppEnv)
	defer logger.Log.Sync()

	r := api.Router()
	if err := http.ListenAndServe(config.Get().ServerAddress.String(), r); err != nil {
		logger.Log.Fatalf("Server failed: %v", err)
	}
}
