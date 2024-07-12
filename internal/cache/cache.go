package cache

import (
	"errors"

	"github.com/raymanovg/kvstorage/internal/config"
	"github.com/raymanovg/kvstorage/internal/partitioned_map"
)

var (
	InvalidCacheDriver = errors.New("invalid cache driver")
)

type Driver[K comparable, V any] interface {
	Set(key K, value V) error
	Get(key K) (V, error)
	Del(key K) error
}

type Cache[K comparable, V any] struct {
	driver Driver[K, V]
}

func NewStrStrCache(conf config.Config) (*Cache[string, string], error) {
	switch conf.Cache.Driver {
	case "partitioned_map":
		return NewWithPartitionedMap[string, string](conf.PartitionedMap, partitioned_map.DefaultStringHashFunc), nil
	default:
		return nil, InvalidCacheDriver
	}
}

func NewWithDriver[K comparable, V any](driver Driver[K, V]) *Cache[K, V] {
	return &Cache[K, V]{driver: driver}
}

func NewWithPartitionedMap[K comparable, V any](conf config.PartitionedMap, hashFunc partitioned_map.HashFunc[K]) *Cache[K, V] {
	driver := partitioned_map.NewPartitionedMap[K, V](conf.Partitions, conf.PartitionSize, hashFunc)

	return NewWithDriver[K, V](driver)
}

func (c *Cache[K, V]) Set(key K, value V) error {
	return c.driver.Set(key, value)
}

func (c *Cache[K, V]) Get(key K) (V, error) {
	return c.driver.Get(key)
}

func (c *Cache[K, V]) Del(key K) error {
	return c.driver.Del(key)
}
