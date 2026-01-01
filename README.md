# LRU Cache

A high-performance, thread-safe, generic LRU (Least Recently Used) cache implementation in Go.

## Features

- **Generic Types**: Supports any comparable key type and any value type using Go generics
- **Thread-Safe**: All operations are protected by `sync.RWMutex` for safe concurrent access
- **O(1) Operations**: Both `Get` and `Put` operations run in constant time
- **Statistics Tracking**: Built-in tracking for cache hits, misses, and evictions
- **Zero Dependencies**: Uses only Go standard library (`container/list`, `sync`, `errors`)

## Installation

```bash
go get github.com/aditya1944/lru-cache
```

## Quick Start

```go
package main

import (
    "fmt"
    lrucache "github.com/aditya1944/lru-cache"
)

func main() {
    // Create a cache with capacity of 100 items
    cache, err := lrucache.New[string, int](100)
    if err != nil {
        panic(err)
    }

    // Store values
    cache.Put("user:1", 42)
    cache.Put("user:2", 100)

    // Retrieve values
    if value, ok := cache.Get("user:1"); ok {
        fmt.Printf("Found: %d\n", value) // Output: Found: 42
    }

    // Check statistics
    stats := cache.Stats()
    fmt.Printf("Hits: %d, Misses: %d, Evictions: %d\n",
        stats.Hits, stats.Misses, stats.Evictions)
}
```

## API Reference

### Creating a Cache

```go
func New[K comparable, V any](capacity uint) (*cache[K, V], error)
```

Creates a new LRU cache with the specified capacity. Returns an error if capacity is 0.

**Parameters:**
- `capacity`: Maximum number of items the cache can hold (must be > 0)

**Returns:**
- A pointer to the cache instance
- An error if capacity is 0

**Example:**
```go
// String keys, integer values
cache, err := lrucache.New[string, int](1000)

// Integer keys, struct values
type User struct {
    Name string
    Age  int
}
userCache, err := lrucache.New[int, User](500)
```

---

### Get

```go
func (c *cache[K, V]) Get(key K) (value V, ok bool)
```

Retrieves a value from the cache. If found, the item is moved to the front (most recently used).

**Parameters:**
- `key`: The key to look up

**Returns:**
- `value`: The stored value (zero value if not found)
- `ok`: `true` if the key exists, `false` otherwise

**Behavior:**
- Increments `Hits` stat on cache hit
- Increments `Misses` stat on cache miss
- Moves accessed item to front of LRU list

**Example:**
```go
value, ok := cache.Get("mykey")
if ok {
    fmt.Println("Found:", value)
} else {
    fmt.Println("Not found")
}
```

---

### Put

```go
func (c *cache[K, V]) Put(key K, value V)
```

Stores a key-value pair in the cache. If the key already exists, its value is updated. If the cache is at capacity, the least recently used item is evicted.

**Parameters:**
- `key`: The key to store
- `value`: The value to associate with the key

**Behavior:**
- If key exists: updates value and moves to front
- If key is new and cache is full: evicts LRU item, increments `Evictions` stat
- New items are always placed at the front (most recently used)

**Example:**
```go
cache.Put("session:abc123", sessionData)
```

---

### Delete

```go
func (c *cache[K, V]) Delete(key K)
```

Removes an item from the cache. No-op if the key doesn't exist.

**Parameters:**
- `key`: The key to remove

**Example:**
```go
cache.Delete("expired:token")
```

---

### Len

```go
func (c *cache[K, V]) Len() int
```

Returns the current number of items in the cache.

**Example:**
```go
fmt.Printf("Cache contains %d items\n", cache.Len())
```

---

### Clear

```go
func (c *cache[K, V]) Clear()
```

Removes all items from the cache and resets all statistics to zero.

**Example:**
```go
cache.Clear()
// cache.Len() == 0
// cache.Stats() == Stats{Hits: 0, Misses: 0, Evictions: 0}
```

---

### Stats

```go
func (c *cache[K, V]) Stats() Stats
```

Returns the current cache statistics.

**Returns:**
```go
type Stats struct {
    Hits      uint  // Number of successful Get operations
    Misses    uint  // Number of Get operations for non-existent keys
    Evictions uint  // Number of items evicted due to capacity
}
```

**Example:**
```go
stats := cache.Stats()
hitRate := float64(stats.Hits) / float64(stats.Hits + stats.Misses) * 100
fmt.Printf("Hit rate: %.2f%%\n", hitRate)
```

## How It Works

### Data Structures

The LRU cache uses two complementary data structures:

1. **Hash Map** (`map[K]*list.Element`): Provides O(1) key lookup
2. **Doubly Linked List** (`container/list.List`): Maintains access order for O(1) eviction

```
┌─────────────────────────────────────────────────────────┐
│                       Hash Map                          │
│  ┌───────┬───────┬───────┬───────┬───────┐             │
│  │ key_a │ key_b │ key_c │ key_d │ key_e │             │
│  └───┬───┴───┬───┴───┬───┴───┬───┴───┬───┘             │
│      │       │       │       │       │                  │
│      ▼       ▼       ▼       ▼       ▼                  │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Doubly Linked List                   │  │
│  │  ┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐  │  │
│  │  │  a  │◄─►│  b  │◄─►│  c  │◄─►│  d  │◄─►│  e  │  │  │
│  │  └─────┘   └─────┘   └─────┘   └─────┘   └─────┘  │  │
│  │   FRONT                                   BACK    │  │
│  │   (MRU)                                   (LRU)   │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Algorithm

**Get Operation:**
1. Look up key in hash map → O(1)
2. If found, move element to front of list → O(1)
3. Return value

**Put Operation:**
1. Check if key exists in hash map → O(1)
2. If exists: update value, move to front → O(1)
3. If not exists:
   - If at capacity: remove back element (LRU), delete from map → O(1)
   - Create new element, add to front, add to map → O(1)

### Thread Safety

All public methods acquire appropriate locks:
- **Write lock** (`Lock`): `Get`, `Put`, `Delete`, `Clear`
- **Read lock** (`RLock`): `Len`, `Stats`

Note: `Get` uses a write lock because it modifies the LRU order.

## Benchmarks

Benchmarks run on Apple M1 Pro, Go 1.25.5:

```
goos: darwin
goarch: arm64
cpu: Apple M1 Pro
BenchmarkPut-10    98,779,692    112.9 ns/op    64 B/op    2 allocs/op
BenchmarkGet-10   618,788,148     19.42 ns/op    0 B/op    0 allocs/op
```

| Operation | Throughput | Latency | Allocations |
|-----------|------------|---------|-------------|
| `Put` | ~8.8M ops/sec | 112.9 ns | 2 allocs (64 B) |
| `Get` | ~51.5M ops/sec | 19.42 ns | 0 allocs |

### Running Benchmarks

```bash
go test -bench=. -benchmem -benchtime=10s
```

## Thread Safety Examples

### Concurrent Access

```go
cache, _ := lrucache.New[string, int](1000)
var wg sync.WaitGroup

// Concurrent writers
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        cache.Put(fmt.Sprintf("key-%d", n), n)
    }(i)
}

// Concurrent readers
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        cache.Get(fmt.Sprintf("key-%d", n))
    }(i)
}

wg.Wait()
```

### Race Detection

Verify thread safety with Go's race detector:

```bash
go test -race ./...
```

## Use Cases

- **Session Storage**: Cache user sessions with automatic expiration of inactive sessions
- **Database Query Cache**: Store frequently accessed query results
- **API Response Cache**: Cache external API responses to reduce latency
- **Configuration Cache**: Store parsed configuration with bounded memory
- **DNS Cache**: Cache DNS lookups with automatic eviction

## Limitations

- **No TTL Support**: Items are only evicted based on access patterns, not time
- **No Size-Based Eviction**: Capacity is based on item count, not memory size
- **Single Lock**: All operations share one mutex (consider sharding for very high concurrency)

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with race detection:

```bash
go test -race ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

## License

Apache License 2.0
