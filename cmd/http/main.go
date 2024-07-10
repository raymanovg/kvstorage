package main

import (
	"github.com/raymanovg/kvstorage/internal/cache"
	"github.com/raymanovg/kvstorage/internal/config"
	"github.com/raymanovg/kvstorage/internal/handler"
	"github.com/raymanovg/kvstorage/internal/server"
)

func main() {
	cfg := config.MustLoad()
	c, err := cache.NewStrStrCache(cfg)
	if err != nil {
		panic(err)
	}

	h := handler.NewHandler(c)
	srv := server.NewServer(cfg.Http, h)

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
