// Package app provides the core application structure and server initialization logic
// for the URL shortener service. It manages configuration, logging, storage, and
// the startup and graceful shutdown of HTTP and gRPC servers.
package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grnsv/shortener/internal/api"
	"github.com/grnsv/shortener/internal/api/middleware"
	"github.com/grnsv/shortener/internal/api/pb"
	"github.com/grnsv/shortener/internal/config"
	"github.com/grnsv/shortener/internal/logger"
	"github.com/grnsv/shortener/internal/service"
	"github.com/grnsv/shortener/internal/storage"
	"google.golang.org/grpc"
)

// Application encapsulates the main components and servers of the URL shortener application.
type Application struct {
	Config     *config.Config
	Logger     logger.Logger
	Storage    storage.Storage
	Shortener  service.Shortener
	HTTPServer *http.Server
	GRPCServer *grpc.Server
}

// NewApplication creates and initializes a new Application instance.
// It parses configuration, sets up logging, storage, and prepares HTTP and gRPC servers.
func NewApplication(ctx context.Context) (*Application, error) {
	var app Application
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
	app.initServers()

	return &app, nil
}

func (app *Application) initServers() {
	app.initHTTP()
	app.initGRPC()
}

func (app *Application) initHTTP() {
	handler := api.NewURLHandler(app.Shortener, app.Config, app.Logger)
	router := api.NewRouter(handler, app.Config, app.Logger)
	app.HTTPServer = &http.Server{
		Addr:         app.Config.ServerAddress.String(),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
}

func (app *Application) initGRPC() {
	authenticate := middleware.GRPCAuthenticateInterceptor(app.Config.JWTSecret, app.Logger)
	app.GRPCServer = grpc.NewServer(grpc.UnaryInterceptor(authenticate))
	pb.RegisterShortenerServer(app.GRPCServer, pb.NewGRPCShortenerServer(app.Shortener, app.Logger))
}

// Run starts the HTTP and gRPC servers of the application.
func (app *Application) Run() {
	go app.runHTTP()
	go app.runGRPC()
}

func (app *Application) runHTTP() {
	var err error
	if app.Config.EnableHTTPS {
		err = app.HTTPServer.ListenAndServeTLS(app.Config.CertFile, app.Config.KeyFile)
	} else {
		err = app.HTTPServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		app.Logger.Fatalf("HTTP server failed: %v", err)
	}
}

func (app *Application) runGRPC() {
	listener, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port 3200: %v", err)
	}
	if err := app.GRPCServer.Serve(listener); err != nil {
		app.Logger.Fatalf("gRPC server failed: %v", err)
	}
}

// Shutdown gracefully shuts down the application's servers, storage, and logger.
func (app *Application) Shutdown(ctx context.Context) error {
	app.GRPCServer.GracefulStop()
	if err := app.HTTPServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}
	if err := app.Storage.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}
	if err := app.Logger.Sync(); err != nil &&
		err.Error() != "sync /dev/stderr: invalid argument" &&
		err.Error() != "sync /dev/stdout: invalid argument" {
		return fmt.Errorf("failed to sync logger: %w", err)
	}

	return nil
}
