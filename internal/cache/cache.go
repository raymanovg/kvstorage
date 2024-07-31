package cache

import "errors"

var (
	NotFoundError = errors.New("key not found")
)

type Cache[K comparable, V any] interface {
	Get(key K) (V, error)
	Put(key K, value V) error
	Del(key K) error
}
