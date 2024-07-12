package server

import (
	"net/http"

	"github.com/raymanovg/kvstorage/internal/config"
)

func NewServer(config config.Http, handler http.Handler) *http.Server {
	server := &http.Server{Addr: config.Addr, Handler: handler}

	return server
}
