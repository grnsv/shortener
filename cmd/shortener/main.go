package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/grnsv/shortener/internal/app"
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

func main() {
	printBuildInfo()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	app, err := app.NewApplication(ctx)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	app.Run()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Graceful shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
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
