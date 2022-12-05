package lru_cache

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleNewCache() {
	c := NewCache(5*time.Minute, 10000)
	c.Set("key", "value")
	v, exist := c.Get("key")
	if !exist {
		// not exist
	}
	fmt.Println(v)
	fmt.Printf("Cache hit: %d, cache miss: %d", c.HitNumber(), c.MissNumber())
}

func TestCache_Set_and_Get(t *testing.T) {
	c := NewCache(0, 100)

	for i := 0; i < 100; i++ {
		c.Set(strconv.FormatInt(int64(i), 10), i)
	}

	for i := 0; i < 100; i++ {
		v, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.True(t, exist)
		assert.EqualValues(t, i, v)
	}
}

func TestCache_Set_Replace(t *testing.T) {
	c := NewCache(0, 100)

	for i := 0; i < 10; i++ {
		c.Set(strconv.FormatInt(int64(i), 10), i)
	}

	for i := 0; i < 10; i++ {
		c.SetEx(strconv.FormatInt(int64(i), 10), i*10, 3, 0)
	}

	for i := 1; i < 10; i++ {
		v, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.True(t, exist)
		assert.EqualValues(t, i*10, v)
	}
	assert.EqualValues(t, 30, c.used)
}

func TestCache_Set_expire(t *testing.T) {
	c := NewCache(0, 100)

	for i := 0; i < 10; i++ {
		c.SetEx(strconv.FormatInt(int64(i), 10), i*10, 3, 1*time.Second)
	}

	time.Sleep(200 * time.Millisecond)

	for i := 1; i < 10; i++ {
		v, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.True(t, exist)
		assert.EqualValues(t, i*10, v)
	}

	time.Sleep(1 * time.Second)

	for i := 1; i < 10; i++ {
		_, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.False(t, exist)
	}
}

func TestCache_Set_lru(t *testing.T) {
	c := NewCache(0, 100)

	for i := 0; i < 101; i++ {
		c.Set(strconv.FormatInt(int64(i), 10), i)
	}

	v, exist := c.Get("0")
	assert.False(t, exist)
	assert.Nil(t, v)

	for i := 1; i < 101; i++ {
		v, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.True(t, exist)
		assert.EqualValues(t, i, v)
	}
}

func TestCache_Remove(t *testing.T) {
	c := NewCache(0, 100)
	for i := 0; i < 10; i++ {
		c.Set(strconv.FormatInt(int64(i), 10), i)
	}
	for i := 0; i < 10; i++ {
		c.Remove(strconv.FormatInt(int64(i), 10))
	}
	for i := 0; i < 10; i++ {
		_, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.False(t, exist)
	}
}

func TestCache_Set_lru_2(t *testing.T) {
	c := NewCache(0, 100)

	for i := 0; i < 100; i++ {
		c.Set(strconv.FormatInt(int64(i), 10), i)
	}

	_, exist := c.Get("0")
	assert.True(t, exist)

	c.Set("100", 100)

	_, exist = c.Get("0")
	assert.True(t, exist)

	_, exist = c.Get("1")
	assert.False(t, exist)

	for i := 2; i < 101; i++ {
		v, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.True(t, exist)
		assert.EqualValues(t, i, v)
	}
}

func TestCache_Set_lru_3(t *testing.T) {
	c := NewCache(0, 100)

	for i := 0; i < 100; i++ {
		c.Set(strconv.FormatInt(int64(i), 10), i)
	}

	c.SetEx("100", 100, 90, 10*time.Second)

	for i := 0; i < 90; i++ {
		_, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.False(t, exist)
	}

	for i := 90; i < 101; i++ {
		v, exist := c.Get(strconv.FormatInt(int64(i), 10))
		assert.True(t, exist)
		assert.EqualValues(t, i, v)
	}

}

func BenchmarkCache_Set(b *testing.B) {
	c := NewCache(0, 60000000)
	concurrent := 1000
	times := 10000

	start := time.Now().UnixNano()

	wg := sync.WaitGroup{}
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for i := 0; i < times; i++ {
				key := strconv.FormatInt(int64(index*times+i), 10)
				value := []byte("fsdfsfasafdsadfasdfasdfsfasfdsaf" + key)
				c.SetEx(key, value, 1, 1000*time.Second)
			}
		}(i)
	}
	wg.Wait()

	//for index := 0; index < concurrent; index++ {
	//	for i := 0; i < times; i++ {
	//		key := strconv.FormatInt(int64(index*100000+i), 10)
	//		_, exist := c.Get(key)
	//		assert.True(b, exist)
	//		//assert.EqualValues(b, []byte("fsdfsfasafdsadfasdfasdfsfasfdsaf"+key), value)
	//	}
	//}
	end := time.Now().UnixNano()

	duration := end - start

	fmt.Printf("duration: %d\n", duration)
	fmt.Printf("%d ns/op \n", duration/int64(concurrent*times))
}
