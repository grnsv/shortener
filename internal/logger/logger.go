package logger

import (
	"log"
	"net/http"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger = zap.NewNop().Sugar()

func Initialize(env string, opts ...zap.Option) {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build(opts...)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	Log = logger.Sugar()
}

func RequestLogger(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log.Debug("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		h(w, r)
	})
}
