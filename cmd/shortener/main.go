package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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
	Server    *http.Server
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

	app.Shortener = service.NewShortener(app.Storage, app.Storage, app.Storage, app.Storage, app.Config.BaseURL.String())
	app.initHandlers()
	app.initServer()

	return &app, nil
}

func (app *application) initHandlers() {
	handler := api.NewURLHandler(app.Shortener, app.Config, app.Logger)
	app.Router = api.NewRouter(handler, app.Config, app.Logger)
}

func (app *application) initServer() {
	app.Server = &http.Server{
		Addr:         app.Config.ServerAddress.String(),
		Handler:      app.Router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}

// Run starts the HTTP server using the application's configuration.
// It blocks until the server exits or fails.
func (app *application) Run(ctx context.Context) {

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := app.Close(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
		} else {
			log.Println("Server shutdown gracefully")
		}
	}()

	var err error
	if app.Config.EnableHTTPS {
		err = app.Server.ListenAndServeTLS(app.Config.CertFile, app.Config.KeyFile)
	} else {
		err = app.Server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		app.Logger.Fatalf("Server failed: %v", err)
	}
}

// Close gracefully shuts down the application's server, storage, and logger.
func (app *application) Close(ctx context.Context) error {
	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	if err := app.Storage.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}
	if err := app.Logger.Sync(); err != nil && err.Error() != "sync /dev/stderr: invalid argument" {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	return nil
}

// MustClose calls Close and exits the application if an error occurs.
func (app *application) MustClose(ctx context.Context) {
	if err := app.Close(ctx); err != nil {
		log.Fatalf("Failed to close application: %v", err)
	}
}

func main() {
	printBuildInfo()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	app, err := newApplication(ctx)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer app.MustClose(ctx)

	app.Run(ctx)
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
