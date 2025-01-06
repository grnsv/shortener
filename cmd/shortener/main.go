package main

import (
	"log"
	"net/http"

	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HandleShortenURL)
	mux.HandleFunc("/{id}", handlers.HandleExpandURL)

	if err := http.ListenAndServe(config.ServerAddress, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
