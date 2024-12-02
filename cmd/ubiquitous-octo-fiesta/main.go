package main

import (
	"context"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/config"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/refresh"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/handlers/tokens"
	mwLogger "github.com/Gonnekone/ubiquitous-octo-fiesta/internal/http-server/middleware/logger"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/jwt"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/logger/handlers/slogpretty"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/lib/logger/sl"
	"github.com/Gonnekone/ubiquitous-octo-fiesta/internal/storage/postgres"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting up the application", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(log, cfg.Storage.DSN())
	if err != nil {
		log.Error("failed to create storage", sl.Err(err))
		os.Exit(1)
	}

	jwtService := jwt.New(cfg.SecretKey)

	router := chi.NewRouter()
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)

	router.Get("/get-tokens", tokens.New(log, storage, jwtService))
	router.Post("/refresh", refresh.New(log, storage, jwtService))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Debug("closing storage")

	storage.Close()
	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
