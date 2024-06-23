# xlander-io Cache

```high-speed```
```thread-safe```
```key-value```
```all data in memory```
```auto recycling ```

## Description
```
Cache is a reference system, not a typical cache, 
as values must be of a pointer type that implements the CacheItem interface.
No deep copy occurs during the set process.

deep copy won't happen in set process
```

## support Type
value type can only be pointer type with implementation of interface `CacheItem`.

## usage

```go
package main

import (
	"fmt"
	"unsafe"

	"github.com/xlander-io/cache"
)

type Person struct {
	Name     string
	Age      int32
	Location string
}

// only support pointer type with CacheBytes interface support
func (p *Person) CacheBytes() int {
	return int(unsafe.Sizeof(*p))
}

func main() {
	local_cache, _ := cache.New(nil)                                                 //nil for default config
	local_cache.Set("key", &Person{Name: "testname", Age: 1, Location: "world"}, 10) 
	item, _ := local_cache.Get("key")
	fmt.Println(item.(*Person))
	fmt.Println(item.CacheBytes())
}
```

### default config

```go
var cache_config = &CacheConfig{
	CacheBytesLimit:          1024 * 1024 * 50, // 50M bytes
	MaxTtlSecs:               7200,             // 2 hours
	RecycleCheckIntervalSecs: 5,                // 5 secs for high efficiency
	RecycleRatioThreshold:    80,               // 80% usage will trigger recycling
	RecycleBatchSize:         10000,            // 10000 items recycled in a batch
	SkipListBufferSize:       20000,            // 20000 commands for chan buffer between internal map and skiplist
	DefaultTtlSecs:           30,               // default cache item duration is 30 secs
}
```

### auto recycling

`RecycleRatioThreshold` or `RecycleBatchSize` of records will be recycled automatically
if `CacheBytesLimit` is reached.

### custom config

```go
// modify duplication of the default config is convenience
config := cache.DupDefaultConfig()
	
config.CacheBytesLimit = 1024 * 1024 * 50 * 4
config.MaxTtlSecs = 7200*2
// ... ...

local_cache := cache.New(&config)
```

## Benchmark

### set

```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalReference_SetPointer
BenchmarkLocalReference_SetPointer-8   	 1000000	      1495 ns/op	     347 B/op	      9 allocs/op
PASS
```

### get

```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalReference_GetPointer
BenchmarkLocalReference_GetPointer-8   	 9931429	       28.48 ns/op	       0 B/op	       0 allocs/op
PASS
```
