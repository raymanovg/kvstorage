package partitioned_map

import (
	"fmt"
	"sync"
)

// HashFunc is func to convert comparable key to int to find a partition key
type HashFunc[K comparable] func(key K) int

func DefaultStringHashFunc(k string) int {
	return len(k)
}

// Partition is a struct representing map with its own rw mutex
type Partition[K comparable, V any] struct {
	mu sync.RWMutex
	mp map[K]V
}

type PartitionedMap[K comparable, V any] struct {
	partitions []Partition[K, V]
	hf         HashFunc[K]
}

func NewPartitionedMap[K comparable, V any](partitions, size uint64, hashFunc HashFunc[K]) *PartitionedMap[K, V] {
	pm := PartitionedMap[K, V]{
		partitions: make([]Partition[K, V], partitions),
		hf:         hashFunc,
	}

	for i := 0; i < int(partitions); i++ {
		pm.partitions[i] = Partition[K, V]{
			mu: sync.RWMutex{},
			mp: make(map[K]V, size),
		}
	}

	return &pm
}

func (pm *PartitionedMap[K, V]) Set(k K, v V) error {
	pk := pm.GetPartitionKey(k)

	pm.partitions[pk].mu.Lock()
	defer pm.partitions[pk].mu.Unlock()

	pm.partitions[pk].mp[k] = v

	return nil
}

func (pm *PartitionedMap[K, V]) Get(k K) (V, error) {
	pk := pm.GetPartitionKey(k)

	pm.partitions[pk].mu.RLock()
	defer pm.partitions[pk].mu.RUnlock()

	v, ok := pm.partitions[pk].mp[k]
	if !ok {
		return v, fmt.Errorf("key not found")
	}

	return v, nil
}

func (pm *PartitionedMap[K, V]) Del(k K) error {
	pk := pm.GetPartitionKey(k)

	pm.partitions[pk].mu.Lock()
	defer pm.partitions[pk].mu.Unlock()

	delete(pm.partitions[pk].mp, k)

	return nil
}

func (pm *PartitionedMap[K, V]) GetPartitionKey(k K) int {
	return pm.hf(k) % len(pm.partitions)
}
