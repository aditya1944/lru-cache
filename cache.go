package lrucache

import (
	"container/list"
	"errors"
	"sync"
)

type Stats struct {
	Hits      uint
	Misses    uint
	Evictions uint
}

type container[K comparable, V any] struct {
	key   K
	value V
}

type cache[K comparable, V any] struct {
	capacity uint

	orderList *list.List
	m         map[K]*list.Element

	lock sync.RWMutex

	stats Stats
}

func (c *cache[K, V]) Get(key K) (value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	element, ok := c.m[key]

	if !ok {
		c.stats.Misses++
		var zero V
		return zero, false
	}
	c.stats.Hits++

	cvalue, ok := element.Value.(*container[K, V])

	if !ok {
		panic("list value is not of container type")
	}

	c.orderList.MoveToFront(element)

	return cvalue.value, true
}

func (c *cache[K, V]) Put(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// check if key is already existing in cache
	val, ok := c.m[key]
	if ok {
		cVal := val.Value.(*container[K, V])
		cVal.value = value
		c.orderList.MoveToFront(val)
		return
	}
	// key does not exist, first check capacity
	if uint(len(c.m)) == c.capacity {
		// evict last key
		lastC := c.orderList.Back().Value
		val, ok := lastC.(*container[K, V])
		if !ok {
			panic("element value not of container type")
		}
		// first delete from map
		// then delete from linked list
		c.stats.Evictions++
		delete(c.m, val.key)
		c.orderList.Remove(c.orderList.Back())
	}

	newC := &container[K, V]{
		key:   key,
		value: value,
	}

	c.m[key] = c.orderList.PushFront(newC)
}

func (c *cache[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.m)
}

func (c *cache[K, V]) Delete(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.m[key]
	if !ok {
		return
	}
	delete(c.m, key)
	c.orderList.Remove(val)
}

func (c *cache[K, V]) Stats() Stats {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.stats
}

func (c *cache[K, V]) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.stats = Stats{}
	clear(c.m)
	c.orderList.Init()
}

func New[K comparable, V any](capacity uint) (*cache[K, V], error) {
	if capacity == 0 {
		return nil, errors.New("capacity should be greater than 0")
	}
	return &cache[K, V]{
		capacity:  capacity,
		orderList: list.New(),
		m:         make(map[K]*list.Element, capacity),

		stats: Stats{},
	}, nil
}
