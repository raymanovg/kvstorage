package partitioned_map

import "errors"

const (
	DefaultPartitionsNum = 100
	DefaultPartitionSize = 100
)

var (
	NotFoundError = errors.New("key not found")
)

type (
	Partitioner[K comparable, V any] interface {
		Get(key K) (V, bool)
		Put(key K, value V)
		Del(key K)
		Len() int
	}

	PartitionedMap[K comparable, V any] struct {
		partitions       []Partitioner[K, V]
		hashFunc         HashFunc[K]
		partitionCreator PartitionCreator[K, V]
		partitionsNum    int
		partitionSize    int
	}

	Option[K comparable, V any] func(*PartitionedMap[K, V])

	HashFunc[K comparable]                func(key K) int
	PartitionCreator[K comparable, V any] func() Partitioner[K, V]
)

func DefaultStringHashFunc(k string) int {
	return len(k)
}

func WithPartitionsNum[K comparable, V any](partitionsNum int) func(*PartitionedMap[K, V]) {
	return func(pm *PartitionedMap[K, V]) {
		pm.partitionsNum = partitionsNum
	}
}

func WithLRUPartition[K comparable, V any](size int) func(*PartitionedMap[K, V]) {
	return func(pm *PartitionedMap[K, V]) {
		pm.partitionCreator = func() Partitioner[K, V] {
			return NewLRUPartition[K, V](size)
		}
	}
}

func WithDefaultPartition[K comparable, V any](size int) func(*PartitionedMap[K, V]) {
	return func(pm *PartitionedMap[K, V]) {
		pm.partitionCreator = func() Partitioner[K, V] {
			return NewPartition[K, V](size)
		}
	}
}

func NewPartitionedMap[K comparable, V any](hashFunc HashFunc[K], options ...Option[K, V]) *PartitionedMap[K, V] {
	pm := &PartitionedMap[K, V]{
		partitionsNum: DefaultPartitionsNum,
		partitionSize: DefaultPartitionSize,
		partitionCreator: func() Partitioner[K, V] {
			return NewPartition[K, V](DefaultPartitionSize)
		},
		hashFunc: hashFunc,
	}

	for _, option := range options {
		option(pm)
	}

	for i := 0; i < pm.partitionsNum; i++ {
		pm.partitions = append(pm.partitions, pm.partitionCreator())
	}

	return pm
}

func (pm *PartitionedMap[K, V]) Put(k K, v V) error {
	pm.GetPartition(k).Put(k, v)
	return nil
}

func (pm *PartitionedMap[K, V]) Get(k K) (V, error) {
	if v, ok := pm.GetPartition(k).Get(k); ok {
		return v, nil
	}

	var v V

	return v, NotFoundError
}

func (pm *PartitionedMap[K, V]) Del(k K) error {
	pm.GetPartition(k).Del(k)
	return nil
}

func (pm *PartitionedMap[K, V]) GetPartitionKey(k K) int {
	return pm.hashFunc(k) % pm.Len()
}

func (pm *PartitionedMap[K, V]) Len() int {
	return len(pm.partitions)
}

func (pm *PartitionedMap[K, V]) GetPartition(k K) Partitioner[K, V] {
	pk := pm.GetPartitionKey(k)
	return pm.partitions[pk]
}
