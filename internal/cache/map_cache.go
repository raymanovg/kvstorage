package cache

import (
	"sync"
)

// MapCache is a struct representing map with its own rw mutex
type MapCache[K comparable, V any] struct {
	sync.RWMutex
	mp map[K]V
}

func (p *MapCache[K, V]) Get(key K) (V, error) {
	p.RLock()
	defer p.RUnlock()

	if v, ok := p.mp[key]; ok {
		return v, nil
	}

	var v V

	return v, NotFoundError
}

func (p *MapCache[K, V]) Put(k K, v V) error {
	p.Lock()
	defer p.Unlock()
	p.mp[k] = v

	return nil
}

func (p *MapCache[K, V]) Del(k K) error {
	p.Lock()
	defer p.Unlock()
	delete(p.mp, k)

	return nil
}

func (p *MapCache[K, V]) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.mp)
}

func NewMapCache[K comparable, V any](size int) *MapCache[K, V] {
	return &MapCache[K, V]{
		mp: make(map[K]V, size),
	}
}
