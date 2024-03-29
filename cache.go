package lru_cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// 简单的内存缓存实现：
//  * 过期剔除
//  * 容量限制的LRU剔除。

type (
	Cache struct {
		defaultExpireTime time.Duration
		capacity          int
		used              int

		//lru linked list
		ll *list.List

		//expired linked list
		el *list.List

		//elements hash map
		items map[string]*entry

		mu sync.RWMutex

		hit  int64
		miss int64
	}

	entry struct {
		key    string
		value  interface{}
		expire int64
		size   int

		//lru list element
		lle *list.Element

		//expired list element
		ele *list.Element
	}
)

func (e *entry) isExpire() bool {
	return e.expire != 0 && e.expire < time.Now().UnixNano()
}

// NewCache create a new Cache instance, you can set the capacity to 0 if you don't want to limit the cache size
func NewCache(defaultExpireTime time.Duration, capacity int) *Cache {
	c := &Cache{
		defaultExpireTime: defaultExpireTime,
		capacity:          capacity,
		used:              0,

		ll:    list.New(),
		el:    list.New(),
		items: make(map[string]*entry),

		mu: sync.RWMutex{},
	}
	c.startExpireJanitor()
	return c
}

// Set is an easy set method. Set a key-value to memory with default expire time and size 1
func (c *Cache) Set(key string, value interface{}) {
	c.SetEx(key, value, 1, c.defaultExpireTime)
}

// SetEx sets a key-value to memory with size and expire time.
func (c *Cache) SetEx(key string, value interface{}, size int, expire time.Duration) {
	// You can't set an element whose size larger than the capacity, we don't return an error
	// since this operation is just like this huge element was eliminated as soon as it was set in.
	if c.capacity > 0 && size > c.capacity {
		return
	}

	item := &entry{
		key:    key,
		value:  value,
		size:   size,
		expire: 0,
	}

	if expire != 0 {
		item.expire = time.Now().UnixNano() + expire.Nanoseconds()
	}

	c.mu.Lock()

	//do lru
	for c.capacity > 0 && c.used+item.size > c.capacity {
		c.removeLeastRecentUsed()
	}

	c.used += item.size
	// replace if the key is existed.
	if origin, exist := c.items[key]; exist {
		c.used -= origin.size

		origin.size = item.size
		origin.value = item.value
		origin.expire = item.expire

		c.ll.MoveToFront(origin.lle)
		c.el.MoveToBack(origin.ele)
		c.mu.Unlock()
		return
	}

	item.lle = c.ll.PushFront(item)
	item.ele = c.el.PushBack(item)
	c.items[key] = item

	c.mu.Unlock()
}

// Get gets the value of key. if not exist will return ( nil, false)
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()

	item, exist := c.items[key]
	if !exist || item.isExpire() {
		atomic.AddInt64(&c.miss, 1)
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()
	atomic.AddInt64(&c.hit, 1)

	c.mu.Lock()
	c.ll.MoveToFront(item.lle)
	c.mu.Unlock()

	return item.value, true
}

func (c *Cache) Remove(key string) {
	c.mu.Lock()
	item, exist := c.items[key]
	if exist {
		c.remove(item)
	}
	c.mu.Unlock()
}

func (c *Cache) remove(item *entry) {
	delete(c.items, item.key)
	c.ll.Remove(item.lle)
	c.el.Remove(item.ele)
	c.used -= item.size
}

func (c *Cache) removeExpired() {
	for c.el.Front() != nil && c.el.Front().Value.(*entry).isExpire() {
		c.remove(c.el.Front().Value.(*entry))
	}
}

func (c *Cache) removeLeastRecentUsed() {
	e := c.ll.Back()
	if e != nil {
		c.remove(e.Value.(*entry))
	}
}

// HitNumber returns the hit number.
func (c *Cache) HitNumber() int64 {
	return atomic.LoadInt64(&c.hit)
}

// MissNumber returns the miss number.
func (c *Cache) MissNumber() int64 {
	return atomic.LoadInt64(&c.miss)
}

func (c *Cache) startExpireJanitor() {
	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)
			c.mu.Lock()
			c.removeExpired()
			c.mu.Unlock()
		}
	}()
}
