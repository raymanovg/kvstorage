package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raymanovg/kvstorage/internal/cache"
	"github.com/raymanovg/kvstorage/internal/config"
	"github.com/raymanovg/kvstorage/internal/handler"
	"github.com/raymanovg/kvstorage/internal/server"
)

func main() {
	cfg := config.MustLoad()
	c, err := cache.NewStrStrCache(cfg)
	if err != nil {
		log.Fatalf("Creating cache error: %v", err)
	}

	h := handler.NewHandler(c)
	srv := server.NewServer(cfg.Http, h)

	go func() {
		if err = srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}

	log.Println("HTTP server shutdown gracefully")
}
