package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raymanovg/kvstorage/internal/cache"
	"github.com/raymanovg/kvstorage/internal/config"
	"github.com/raymanovg/kvstorage/internal/handler"
	"github.com/raymanovg/kvstorage/internal/logger"
	"github.com/raymanovg/kvstorage/internal/server"
)

const Version = "dev"

func main() {
	cfg := config.MustLoad()
	log := logger.NewLogger(cfg.Env, Version)

	c, err := cache.NewStrStrCache(cfg)
	if err != nil {
		log.Error("Creating cache error", slog.String("error", err.Error()))
		return
	}

	h := handler.NewHandler(c, log)
	srv := server.NewServer(cfg.Http, h)

	go func() {
		if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error", slog.String("error", err.Error()))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("HTTP shutdown error", slog.String("error", err.Error()))
	}

	log.Info("HTTP server shutdown gracefully")
}
