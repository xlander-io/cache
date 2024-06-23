# xlander-io Cache

```high-speed```
```thread-safe```
```key-value```
```all data in memory```
```not-persistent```
```auto recycling ```

## Description
```
cache is a reference system, it is NOT an USUAL cache system
because the value can only be reference type with implementation
of interface CacheItem support.

deep copy won't happen in set process
```

## support Type
value type can only be reference type with implementation of interface `CacheItem` support.

## usage

```go
//import
import (
    "github.com/xlander-io/cache"
)
```

### example

```go
package main

import (
	"log"
	"time"

	"github.com/xlander-io/cache"
)

/***************************
type CacheItem interface {
	CacheBytes() int
}
***************************/

type Person struct {
	Name     string
	Age      int32
	Location string
}

// Only support reference type which implement interface CacheItem
func (p *Person) CacheBytes() int {
	return 100
}

func main() {

	// modify duplication of the default config is convenient
	config := cache.DupDefaultConfig()
	config.CacheBytesLimit = 1024 * 1024 * 50 * 4

	local_cache := cache.New(&config)
	// local_cache := cache.New(nil)

	// set struct pointer
	err = local_cache.Set("struct*", &Person{"Jack", 18, "London"}, 300)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	// get
	log.Println("---get---")
	log.Println(local_cache.Get("struct*"))

	// overwrite
	log.Println("---set overwrite---")
	log.Println(local_cache.Get("struct*"))
	err = local_cache.Set("struct*", &Person{"Tom", 38, "London"}, 10)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
	log.Println(local_cache.Get("struct*"))

	// test ttl
	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Println(local_cache.Get("struct*"))
		}
	}()

	time.Sleep(20 * time.Second)

	// if not a pointer cause error
	// err = local_cache.Set("int", 10, 10)
	// if err != nil {
	// 	log.Fatalln("reference set error:", err)
	// }
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
