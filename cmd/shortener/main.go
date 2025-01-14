package main

import (
	"log"
	"net/http"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/handlers"
)

func main() {
	r := handlers.Router()
	if err := http.ListenAndServe(config.Get().ServerAddress.String(), r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
