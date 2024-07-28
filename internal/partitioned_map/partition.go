package partitioned_map

import "sync"

// Partition is a struct representing map with its own rw mutex
type Partition[K comparable, V any] struct {
	sync.RWMutex
	mp map[K]V
}

func (p *Partition[K, V]) Get(key K) (V, bool) {
	p.RLock()
	defer p.RUnlock()

	if v, ok := p.mp[key]; ok {
		return v, true
	}

	var v V

	return v, false
}

func (p *Partition[K, V]) Put(k K, v V) {
	p.Lock()
	defer p.Unlock()
	p.mp[k] = v
}

func (p *Partition[K, V]) Del(k K) {
	p.Lock()
	defer p.Unlock()
	delete(p.mp, k)
}

func (p *Partition[K, V]) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.mp)
}

func NewPartition[K comparable, V any](size int) *Partition[K, V] {
	return &Partition[K, V]{
		mp: make(map[K]V, size),
	}
}
