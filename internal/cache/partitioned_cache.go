package cache

import (
	"fmt"
	"hash/fnv"
	"unsafe"
)

const (
	DefaultPartitionsNum = 100
	DefaultPartitionSize = 100
)

///////////////////////////
/// stolen from runtime ///
///////////////////////////

// mh is an inlined combination of runtime._type and runtime.maptype.
type mh struct {
	_  uintptr
	_  uintptr
	_  uint32
	_  uint8
	_  uint8
	_  uint8
	_  uint8
	_  func(unsafe.Pointer, unsafe.Pointer) bool
	_  *byte
	_  int32
	_  int32
	_  unsafe.Pointer
	_  unsafe.Pointer
	_  unsafe.Pointer
	hf func(unsafe.Pointer, uintptr) uintptr
}

func StringHashFunc(key string, size int) int {
	h := fnv.New32()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32()) % size
}

func GenericHashFunc[K comparable](key K, size int) int {
	var m interface{} = make(map[K]struct{})
	hf := (*mh)(*(*unsafe.Pointer)(unsafe.Pointer(&m))).hf
	return int(uint(hf(unsafe.Pointer(&key), 0)) % uint(size))
}

type (
	Partitioner[K comparable, V any] interface {
		Get(key K) (V, error)
		Put(key K, value V) error
		Del(key K) error
		Len() int
	}

	PartitionedCache[K comparable, V any] struct {
		partitions       []Partitioner[K, V]
		hashFunc         HashFunc[K]
		partitionCreator PartitionCreator[K, V]
		partitionsNum    int
		partitionSize    int
	}

	Option[K comparable, V any] func(*PartitionedCache[K, V])

	HashFunc[K comparable]                func(key K, size int) int
	PartitionCreator[K comparable, V any] func() Partitioner[K, V]
)

func WithPartitionsNum[K comparable, V any](partitionsNum int) func(*PartitionedCache[K, V]) {
	return func(pm *PartitionedCache[K, V]) {
		pm.partitionsNum = partitionsNum
	}
}

func WithMapPartition[K comparable, V any](size int) func(*PartitionedCache[K, V]) {
	return func(pm *PartitionedCache[K, V]) {
		pm.partitionCreator = func() Partitioner[K, V] {
			return NewMapCache[K, V](size)
		}
	}
}

func WithHashFunc[K comparable, V any](hashFunc HashFunc[K]) func(*PartitionedCache[K, V]) {
	return func(pm *PartitionedCache[K, V]) {
		pm.hashFunc = hashFunc
	}
}

func MapCachePartitionCreator[K comparable, V any](size int) func() Partitioner[K, V] {
	return func() Partitioner[K, V] {
		return NewMapCache[K, V](size)
	}
}

func NewPartitionedCache[K comparable, V any](options ...Option[K, V]) *PartitionedCache[K, V] {
	pm := &PartitionedCache[K, V]{
		partitionsNum:    DefaultPartitionsNum,
		partitionSize:    DefaultPartitionSize,
		partitionCreator: MapCachePartitionCreator[K, V](DefaultPartitionSize),
		hashFunc:         GenericHashFunc[K],
	}

	for _, option := range options {
		option(pm)
	}

	for i := 0; i < pm.partitionsNum; i++ {
		pm.partitions = append(pm.partitions, pm.partitionCreator())
	}

	return pm
}

func (pm *PartitionedCache[K, V]) Put(k K, v V) error {
	const op = "PartitionedMap.Put"

	if err := pm.GetPartition(k).Put(k, v); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func (pm *PartitionedCache[K, V]) Get(k K) (V, error) {
	const op = "PartitionedMap.Get"

	v, err := pm.GetPartition(k).Get(k)
	if err != nil {
		return v, fmt.Errorf("%s: %v", op, err)
	}
	return v, nil
}

func (pm *PartitionedCache[K, V]) Del(k K) error {
	const op = "PartitionedMap.Del"

	if err := pm.GetPartition(k).Del(k); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (pm *PartitionedCache[K, V]) Len() int {
	return len(pm.partitions)
}

func (pm *PartitionedCache[K, V]) GetPartitionKey(k K) int {
	return pm.hashFunc(k, pm.Len())
}

func (pm *PartitionedCache[K, V]) GetPartition(k K) Partitioner[K, V] {
	pk := pm.GetPartitionKey(k)
	return pm.partitions[pk]
}

func (pm *PartitionedCache[K, V]) GetPartitions() []Partitioner[K, V] {
	return pm.partitions
}
