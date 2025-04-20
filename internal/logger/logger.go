// Package logger provides a unified logging interface and a constructor
// for creating zap-based loggers with different configurations depending
// on the environment.
package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger defines a generic logging interface with methods for various
// log levels and line-ending variants, as well as a Sync method to flush logs.
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	DPanic(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	DPanicln(args ...interface{})
	Panicln(args ...interface{})
	Fatalln(args ...interface{})

	Sync() error
}

// New creates and returns a new Logger instance based on the provided
// environment string. Supported environments are "production", "development",
// and "testing". Additional zap options can be passed as variadic arguments.
// Returns an error if logger initialization fails.
func New(env string, opts ...zap.Option) (Logger, error) {
	if env == "testing" {
		return zap.NewNop().Sugar(), nil
	}

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	return logger.Sugar(), nil
}
