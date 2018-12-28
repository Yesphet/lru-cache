## lru-cahce

Golang 内存缓存实现：

* expire 过期剔除
* capacity lru 基于容量的LRU剔除

## Example


```golang
// new cache instance
c := NewCache(5*time.Minute, 10000)

// easy set
c.Set("key", "value")

// set the value, size and expiration with a key
c.SetEx("key","value", 5, 5*time.Minute)
v, exist := c.Get("key")
if !exist {
	// not exist
}
fmt.Printf("Cache hit: %d, cache miss: %d", c.HitNumber(), c.MissNumber())
```

## Performance

### Set

| Cached number | Concurrent| Performance |
---|---|---
500,000 | 100 | 1375 ns/op
1,000,000 | 100 | 1456 ns/op
5,000,000| 100 | 1496 ns/op
10,000,000| 100 | 1751 ns/op
50,000,000| 100| 2264 ns/op