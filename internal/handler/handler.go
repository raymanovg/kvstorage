package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/raymanovg/kvstorage/internal/cache"
	"github.com/raymanovg/kvstorage/internal/config"
)

type Value struct {
	Value string `json:"value"`
}

type Handler struct {
	chi   chi.Router
	cache *cache.Cache[string, string]
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.chi.ServeHTTP(rw, r)
}

func NewHandler(c *cache.Cache[string, string]) *Handler {
	h := &Handler{
		cache: c,
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.Timeout(2 * time.Second))
	router.Use(middleware.Recoverer)

	router.Route("/", func(rt chi.Router) {
		rt.Get("/_version", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(config.VERSION))
		})
	})

	router.Route("/cache/{key}", func(rt chi.Router) {
		rt.Get("/", h.Get)
		rt.Put("/", h.Set)
		rt.Delete("/", h.Del)
	})

	h.chi = router

	return h
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	v, err := h.cache.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := map[string]interface{}{
		"key": key,
		"v":   v,
	}

	b, _ := json.Marshal(&resp)

	_, _ = w.Write(b)
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	var value Value
	err := json.NewDecoder(r.Body).Decode(&value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := chi.URLParam(r, "key")
	if err := h.cache.Set(key, value.Value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Del(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if err := h.cache.Del(key); err != nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
