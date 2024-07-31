package cache

import (
	"container/list"
	"sync"
)

type Node[V any] struct {
	Value   V
	Element *list.Element
}

type LRUCache[K comparable, V any] struct {
	sync.RWMutex
	mp   map[K]Node[V]
	list *list.List
	cap  int
}

func (p *LRUCache[K, V]) Get(key K) (V, error) {
	p.RLock()
	defer p.RUnlock()

	if node, ok := p.mp[key]; ok {
		p.list.MoveToFront(node.Element)
		return node.Value, nil
	}

	var v V

	return v, NotFoundError
}

func (p *LRUCache[K, V]) Put(k K, v V) error {
	p.Lock()
	defer p.Unlock()

	if node, ok := p.mp[k]; !ok {
		if len(p.mp) >= p.cap {
			p.list.Remove(p.list.Back())
		}
		p.mp[k] = Node[V]{Value: v, Element: p.list.PushFront(k)}
	} else {
		p.list.MoveToFront(node.Element)
		node.Value = v
		p.mp[k] = node
	}

	return nil
}

func (p *LRUCache[K, V]) Del(k K) error {
	p.Lock()
	defer p.Unlock()
	if node, ok := p.mp[k]; ok {
		p.list.Remove(node.Element)
		delete(p.mp, k)
	}

	return nil
}

func (p *LRUCache[K, V]) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.mp)
}

func NewLRUCache[K comparable, V any](size int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		mp:   make(map[K]Node[V], size),
		list: list.New(),
		cap:  size,
	}
}
