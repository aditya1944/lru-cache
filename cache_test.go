package lrucache

import (
	"fmt"
	"sync"
	"testing"
)

func TestZeroCapacity(t *testing.T) {
	t.Parallel()
	_, err := New[int, string](0)
	if err == nil {
		t.Error("New should return error when capacity is 0")
	}
}

func TestCache(t *testing.T) {
	t.Parallel()
	cache, _ := New[string, string](1)
	cache.Put("key", "value")

	val, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exists, but key do not exists")
	}
	if val != "value" {
		t.Errorf("expected value to be equal to: `value`, but got: `%s`", val)
	}
	if cache.Len() != 1 {
		t.Errorf("expected cache length to be 1, but got: %d", cache.Len())
	}
	cache.Clear()
	if cache.Len() != 0 {
		t.Errorf("expected cache length to be 0, but got : %d", cache.Len())
	}
	cache.Put("key", "value")
	cache.Delete("key")
	if cache.Len() != 0 {
		t.Errorf("expected cache length to be 0, but got : %d", cache.Len())
	}

	hits, misses, evictions := cache.Stats()
	if hits != 0 {
		t.Errorf("expected hits to be 0, but got: %d", hits)
	}
	if misses != 0 {
		t.Errorf("expected misses to be 0, but got: %d", misses)
	}
	if evictions != 0 {
		t.Errorf("expected evictions to be 0, but got: %d", evictions)
	}
}

func TestCacheConcurrency(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	cache, _ := New[string, string](1000)

	for i := range 1000 {
		key, value := fmt.Sprintf("key-%d", i), fmt.Sprintf("value-%d", i)
		wg.Go(func() {
			cache.Put(key, value)
		})
	}

	wg.Wait()

	for i := range 1000 {
		wg.Go(func() {
			key := fmt.Sprintf("key-%d", i)
			val, ok := cache.Get(key)
			if !ok {
				t.Errorf("expected value to exists in map for key: %s , but not exists", key)
			}
			expectedVal := fmt.Sprintf("value-%d", i)
			if val != expectedVal {
				t.Errorf("expected value for key : %s, to be: %s , but got val: %s", key, expectedVal, val)
			}
		})
	}

	wg.Wait()

	hits, misses, evictions := cache.Stats()
	if hits != 1000 {
		t.Errorf("expected hits to be 1000, but got: %d", hits)
	}
	if misses != 0 {
		t.Errorf("expected misses to be 0, but got: %d", misses)
	}
	if evictions != 0 {
		t.Errorf("expected evictions to be 0, but got: %d", evictions)
	}
}

func TestCacheEviction(t *testing.T) {
	t.Parallel()
	// create a capacity of 1
	// insert 2 keys
	// check if old key is evicted or not

	cache, _ := New[string, string](1)

	cache.Put("key1", "value1")
	cache.Put("key2", "value2")

	// key1 should have been evicted
	val, ok := cache.Get("key1")
	if ok || val != "" {
		t.Errorf("key: `key1` should have been evicted, but still exists")
	}

	hits, misses, evictions := cache.Stats()
	if hits != 0 {
		t.Errorf("expected hits to be 0, but got: %d", hits)
	}
	if misses != 1 {
		t.Errorf("expected misses to be 1, but got: %d", misses)
	}
	if evictions != 1 {
		t.Errorf("expected evictions to be 1, but got: %d", evictions)
	}
}

func TestLRUOrdering(t *testing.T) {
	t.Parallel()
	cache, _ := New[string, string](2)

	cache.Put("a", "1")
	cache.Put("b", "2")
	cache.Get("a")      // "a" is now most recently used
	cache.Put("c", "3") // should evict "b"

	_, ok := cache.Get("a")
	if !ok {
		t.Error("'a' should still exist (was recently accessed)")
	}

	_, ok = cache.Get("b")
	if ok {
		t.Error("'b' should have been evicted (least recently used)")
	}
}

func TestSameKeyInsertion(t *testing.T) {
	t.Parallel()
	cache, _ := New[string, string](1)

	cache.Put("key", "value")
	cache.Put("key", "value1")

	val, ok := cache.Get("key")
	if !ok {
		t.Errorf("entry with key: `key` should exists in map, but do not exists")
	}

	if val != "value1" {
		t.Errorf("expected value to be %s, but got: %s", "value1", val)
	}

	hits, misses, evictions := cache.Stats()
	if hits != 1 {
		t.Errorf("expected hits to be 1, but got: %d", hits)
	}
	if misses != 0 {
		t.Errorf("expected misses to be 0, but got: %d", misses)
	}
	if evictions != 0 {
		t.Errorf("expected evictions to be 0, but got: %d", evictions)
	}
}

// TestDeleteNotExistentKey verifies if deleting not existent key doesn't panic
func TestDeleteNotExistentKey(t *testing.T) {
	t.Parallel()
	cache, _ := New[string, string](1)
	cache.Delete("key")
}

func TestConcurrentReadWrite(t *testing.T) {
	t.Parallel()
	cache, _ := New[int, int](100)
	var wg sync.WaitGroup

	for i := range 1000 {
		wg.Go(func() {
			cache.Put(i, i)
		})
	}

	for i := range 1000 {
		wg.Go(func() {
			cache.Get(i)
		})
	}

	wg.Wait()
}

func TestStatsConcurrency(t *testing.T) {
	cache, _ := New[int, int](100)
	var wg sync.WaitGroup

	for i := range 1000 {
		wg.Go(func() {
			for j := range 100 {
				cache.Put(i, j)
				cache.Get(i)
			}
		})
	}

	for range 100 {
		wg.Go(func() {
			for range 1000 {
				_, _, _ = cache.Stats()
			}
		})
	}

	wg.Wait()
}

func BenchmarkPut(b *testing.B) {
	cache, _ := New[int, int](1000)

	i := 0
	for b.Loop() {
		cache.Put(i, i)
		i++
	}
}

func BenchmarkGet(b *testing.B) {
	cache, _ := New[int, string](1000)

	for i := range 1000 {
		cache.Put(i%1000, "value")
	}

	b.ResetTimer()
	i := 0
	for b.Loop() {
		cache.Get(i)
		i++
		i = i % 1000
	}
}
