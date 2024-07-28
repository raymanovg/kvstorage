package partitioned_map

import (
	"container/list"
	"sync"
)

type Node[V any] struct {
	Value   V
	Element *list.Element
}

type LRUPartition[K comparable, V any] struct {
	sync.RWMutex
	mp   map[K]Node[V]
	list *list.List
	cap  int
}

func (p *LRUPartition[K, V]) Get(key K) (V, bool) {
	p.RLock()
	defer p.RUnlock()

	if node, ok := p.mp[key]; ok {
		p.list.MoveToFront(node.Element)
		return node.Value, true
	}

	var v V

	return v, false
}

func (p *LRUPartition[K, V]) Put(k K, v V) {
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
}

func (p *LRUPartition[K, V]) Del(k K) {
	p.Lock()
	defer p.Unlock()
	if node, ok := p.mp[k]; ok {
		p.list.Remove(node.Element)
		delete(p.mp, k)
	}
}

func (p *LRUPartition[K, V]) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.mp)
}

func NewLRUPartition[K comparable, V any](size int) *LRUPartition[K, V] {
	return &LRUPartition[K, V]{
		mp:   make(map[K]Node[V], size),
		list: list.New(),
		cap:  size,
	}
}
