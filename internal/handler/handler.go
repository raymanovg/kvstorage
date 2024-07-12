package handler

import (
	"encoding/json"
	"log/slog"
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
	log   *slog.Logger
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.chi.ServeHTTP(rw, r)
}

func NewHandler(c *cache.Cache[string, string], log *slog.Logger) *Handler {
	h := &Handler{
		cache: c,
		log:   log,
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Heartbeat("/ping"))
	router.Use(middleware.Timeout(2 * time.Second))
	router.Use(middleware.Recoverer)

	router.Route("/", func(rt chi.Router) {
		rt.Get("/_version", h.Version)
	})

	router.Route("/cache/{key}", func(rt chi.Router) {
		rt.Get("/", h.Get)
		rt.Put("/", h.Set)
		rt.Delete("/", h.Del)
	})

	h.chi = router

	return h
}

func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(config.VERSION)); err != nil {
		h.log.Error("failed to write response", slog.String("error", err.Error()))
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

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
	if _, err = w.Write(b); err != nil {
		h.log.Error("failed to write response", slog.String("error", err.Error()))
	}
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	var value Value

	defer func() {
		if err := r.Body.Close(); err != nil {
			h.log.Error("failed to close request body", slog.String("error", err.Error()))
		}
	}()

	err := json.NewDecoder(r.Body).Decode(&value)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.cache.Set(key, value.Value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Del(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.cache.Del(key); err != nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
