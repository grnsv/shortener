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

func newApplication(ctx context.Context) *application {
	var app application

	app.initConfig()
	app.initLogger()
	app.initStorage(ctx)
	app.initService()
	app.initHandlers()

	return &app
}

func (app *application) initConfig() {
	var err error
	app.Config, err = config.Parse()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
}

func (app *application) initLogger() {
	var err error
	app.Logger, err = logger.New(app.Config.AppEnv)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
}

func (app *application) initStorage(ctx context.Context) {
	var err error
	app.Storage, err = storage.New(ctx, app.Config)
	if err != nil {
		app.Logger.Fatalf("Failed to create storage: %v", err)
	}
}

func (app *application) initService() {
	app.Shortener = service.NewShortener(app.Storage, app.Storage, app.Storage, app.Storage, app.Config.BaseAddress.String())
}

func (app *application) initHandlers() {
	handler := api.NewURLHandler(app.Shortener, app.Config, app.Logger)
	app.Router = api.NewRouter(handler, app.Config, app.Logger)
}

// Run starts the HTTP server using the application's configuration.
// It blocks until the server exits or fails.
func (app *application) Run() {
	if err := http.ListenAndServe(app.Config.ServerAddress.String(), app.Router); err != nil {
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
	app := newApplication(context.Background())
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
