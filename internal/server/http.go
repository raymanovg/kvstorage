package server

import (
	"github.com/raymanovg/kvstorage/internal/config"
	"net/http"
)

func NewServer(config config.Http, handler http.Handler) *http.Server {
	server := &http.Server{Addr: config.Addr, Handler: handler}

	return server
}
