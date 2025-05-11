package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
)

// buildVersion is set at compile time using -ldflags.
// Example:
//
//	go build -ldflags "-X 'main.buildVersion=1.0.0'"
var buildVersion string

// buildDate is the date of the build, injected via -ldflags.
// Example:
//
//	go build -ldflags "-X 'main.buildDate=2025-05-02'"
var buildDate string

// buildCommit is the Git commit hash at build time.
// Injected via:
//
//	go build -ldflags "-X 'main.buildCommit=abc1234'"
var buildCommit string

type application struct {
	Config    *config.Config
	Logger    logger.Logger
	Storage   storage.Storage
	Shortener service.Shortener
	Router    http.Handler
}

func newApplication(ctx context.Context) (*application, error) {
	var app application
	var err error

	if app.Config, err = config.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if app.Logger, err = logger.New(app.Config.AppEnv); err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	if app.Storage, err = storage.New(ctx, app.Config); err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	app.Shortener = service.NewShortener(app.Storage, app.Storage, app.Storage, app.Storage, app.Config.BaseAddress.String())
	app.initHandlers()

	return &app, nil
}

func (app *application) initHandlers() {
	handler := api.NewURLHandler(app.Shortener, app.Config, app.Logger)
	app.Router = api.NewRouter(handler, app.Config, app.Logger)
}

// Run starts the HTTP server using the application's configuration.
// It blocks until the server exits or fails.
func (app *application) Run() {
	var err error
	if app.Config.EnableHTTPS {
		err = http.ListenAndServeTLS(app.Config.ServerAddress.String(), app.Config.CertFile, app.Config.KeyFile, app.Router)
	} else {
		err = http.ListenAndServe(app.Config.ServerAddress.String(), app.Router)
	}

	if err != nil {
		app.Logger.Fatalf("Server failed: %v", err)
	}
}

// Close gracefully shuts down the application's resources,
// including storage and logger.
func (app *application) Close() {
	if err := app.Storage.Close(); err != nil {
		app.Logger.Fatalf("Failed to close storage: %v", err)
	}
	if err := app.Logger.Sync(); err != nil {
		log.Fatalf("Failed to sync logger: %v", err)
	}
}

func main() {
	printBuildInfo()
	app, err := newApplication(context.Background())
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer app.Close()

	app.Run()
}

func printBuildInfo() {
	println("Build version: " + nonEmpty(buildVersion))
	println("Build date: " + nonEmpty(buildDate))
	println("Build commit: " + nonEmpty(buildCommit))
}

func nonEmpty(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
