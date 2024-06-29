package cache

import (
	"errors"
	"sync"
	"time"
)

type CacheItem interface {
	CacheBytes() int
}

type cache_element struct {
	Score int64
	Value interface{}
}

type CacheConfig struct {
	CacheBytesLimit          int64 // max cache size in bytes
	MaxTtlSecs               int64 // max cache item duration in secs
	DefaultTtlSecs           int64 // default set ttl
	RecycleCheckIntervalSecs int   // recycle process cycle in secs
	RecycleRatioThreshold    int   // 1-100, old items are recycled once cache reach limit, e.g: 30 for 30% percentage
	RecycleBatchSize         int   // number of items to be recycled in a batch
	SkipListBufferSize       int   // chan buffer between internal map and skiplist
}

type Cache struct {
	//config
	cache_config            *CacheConfig
	recycle_bytes_threshold int64
	//
	sync_map   sync.Map
	skip_list  *skiplist
	lock       sync.Mutex
	sl_channel chan func()
	//
	element_count int32 //total element number
	element_bytes int64 //total element bytes
	now_unixtime  int64
}

// if passed user_config is nil, then use the default cache_config.
func New(user_config *CacheConfig) (*Cache, error) {

	cache_config := &CacheConfig{
		CacheBytesLimit:          1024 * 1024 * 50, // 50M bytes
		MaxTtlSecs:               7200,             // 2 hours
		DefaultTtlSecs:           30,               // 30 secs for default ttl secs
		RecycleCheckIntervalSecs: 5,                // 5 secs for high efficiency
		RecycleRatioThreshold:    80,               // 80% usage will trigger recycling
		RecycleBatchSize:         100,              // 100 items recycled in a batch
		SkipListBufferSize:       20000,            // 20000 commands for chan buffer between internal map and skiplist
	}

	//new a cache with default config
	if user_config != nil {
		//
		if user_config.CacheBytesLimit < 0 {
			return nil, errors.New("config CacheBytesLimit error, val < 0")
		} else if user_config.CacheBytesLimit == 0 {
			//bypass using default value
		} else {
			cache_config.CacheBytesLimit = user_config.CacheBytesLimit
		}

		//
		if user_config.MaxTtlSecs < 0 {
			return nil, errors.New("config MaxTtlSecs error, val < 0")
		} else if user_config.MaxTtlSecs == 0 {
			//bypass using default value
		} else {
			cache_config.MaxTtlSecs = user_config.MaxTtlSecs
		}

		//
		if user_config.DefaultTtlSecs < 0 {
			return nil, errors.New("config DefaultTtlSecs error, val < 0")
		} else if user_config.DefaultTtlSecs == 0 {
			//bypass using default value
		} else {
			cache_config.DefaultTtlSecs = user_config.DefaultTtlSecs
		}

		//
		if cache_config.DefaultTtlSecs > cache_config.MaxTtlSecs {
			return nil, errors.New("config DefaultTtlSecs || MaxTtlSecs error, DefaultTtlSecs > MaxTtlSecs")
		}

		//
		if user_config.RecycleCheckIntervalSecs < 0 {
			return nil, errors.New("config RecycleCheckIntervalSecs error: val < 0")
		} else if user_config.RecycleCheckIntervalSecs == 0 {
			//bypass using default value
		} else {
			cache_config.RecycleCheckIntervalSecs = user_config.RecycleCheckIntervalSecs
		}

		//
		if user_config.RecycleRatioThreshold < 0 || user_config.RecycleRatioThreshold > 100 {
			return nil, errors.New("config RecycleRatioThreshold error, val between [1,100]")
		} else if user_config.RecycleRatioThreshold == 0 {
			//bypass using default value
		} else {
			cache_config.RecycleRatioThreshold = user_config.RecycleRatioThreshold
		}

		//
		if user_config.RecycleBatchSize < 0 {
			return nil, errors.New("config RecycleBatchSize error: val < 0")
		} else if user_config.RecycleBatchSize == 0 {
			//bypass using default value
		} else {
			cache_config.RecycleBatchSize = user_config.RecycleBatchSize
		}

		//
		if user_config.SkipListBufferSize < 0 {
			return nil, errors.New("config SkipListBufferSize error: val < 0")
		} else if user_config.SkipListBufferSize == 0 {
			//bypass using default value
		} else {
			cache_config.SkipListBufferSize = user_config.SkipListBufferSize
		}

	}

	var config_recycle_bytes_threshold int64 = (cache_config.CacheBytesLimit * int64(cache_config.RecycleRatioThreshold) / 100)
	if config_recycle_bytes_threshold < 1 {
		return nil, errors.New("CacheBytesLimit * RecycleRatioThreshold / 100 must >= 1")
	}

	cache := &Cache{
		cache_config:            cache_config,
		recycle_bytes_threshold: config_recycle_bytes_threshold,
		now_unixtime:            time.Now().Unix(),
		skip_list:               makeSkiplist(),
		element_count:           0,
		element_bytes:           0,
		sl_channel:              make(chan func(), cache_config.SkipListBufferSize),
	}

	//for efficiency update the unixtime using a go-routine
	go func() {
		for {
			time.Sleep(1 * time.Second)
			cache.now_unixtime = time.Now().Unix()
		}
	}()

	// routine for processing commands which transfered from sl_channel for skiplist
	go func() {
		for {
			f := <-cache.sl_channel
			f()
		}
	}()

	//start the recycle go routine
	safeInfiLoop(func() {

		// remove expired keys
		keys := cache.skip_list.GetRangeByScore(0, cache.now_unixtime)
		for _, key := range keys {
			cache.Delete(key)
		}

		// check overlimit
		for int64(cache.Bytes()) >= cache.recycle_bytes_threshold {
			keys := cache.skip_list.GetRangeByRank(0, int64(cache.cache_config.RecycleBatchSize))
			for _, key := range keys {
				cache.Delete(key)
			}
		}

	}, nil, int64(cache.cache_config.RecycleCheckIntervalSecs), 30)

	//
	return cache, nil
}

// get the config of this cache
func (cache *Cache) GetConfig() *CacheConfig {
	return &CacheConfig{
		CacheBytesLimit:          cache.cache_config.CacheBytesLimit,
		MaxTtlSecs:               cache.cache_config.MaxTtlSecs,
		DefaultTtlSecs:           cache.cache_config.DefaultTtlSecs,
		RecycleCheckIntervalSecs: cache.cache_config.RecycleCheckIntervalSecs,
		RecycleRatioThreshold:    cache.cache_config.RecycleRatioThreshold,
		RecycleBatchSize:         cache.cache_config.RecycleBatchSize,
		SkipListBufferSize:       cache.cache_config.SkipListBufferSize,
	}
}

// get current unix time in the cache
func (cache *Cache) GetUnixTime() int64 {
	return cache.now_unixtime
}

// todo write docs
func (cache *Cache) Get(key string) (value CacheItem, ttl int64) {
	prev_ele_, pre_ele_exist_ := cache.sync_map.Load(key)
	if !pre_ele_exist_ {
		return nil, 0
	} else {
		pre_ele := prev_ele_.(*cache_element)
		if pre_ele.Score <= cache.now_unixtime {
			return nil, 0
		} else {
			return pre_ele.Value.(CacheItem), pre_ele.Score - cache.now_unixtime
		}
	}
}

func (cache *Cache) Set(key string, value CacheItem) error {
	return cache.set_(key, value, cache.cache_config.DefaultTtlSecs)
}

func (cache *Cache) SetTTL(key string, value CacheItem, ttlSecond int64) error {
	if ttlSecond <= 0 {
		return errors.New("ttl <=0 err")
	}
	return cache.set_(key, value, ttlSecond)
}

func (cache *Cache) Keep(key string, value CacheItem) error {
	return cache.set_(key, value, 0)
}

// for ttlSecond < 0 not allowed
// for ttlSecond == 0
// -----1. keep the ttl of previous val if previous value exist
// -----2. nothing will be set if previous value not exist
// for ttlSecond > MaxTtlSecs, ttl will be adjusted to MaxTtlSecs
func (cache *Cache) set_(key string, value CacheItem, ttlSecond int64) error {

	if value == nil {
		return errors.New("value can not be nil")
	}

	if ttlSecond < 0 {
		return errors.New("ttl < 0 error")
	}

	if ttlSecond > cache.cache_config.MaxTtlSecs {
		ttlSecond = cache.cache_config.MaxTtlSecs
	}

	cache.lock.Lock()
	defer cache.lock.Unlock()

	//default expire time
	expire_time := cache.now_unixtime + ttlSecond

	//
	var pre_ele *cache_element = nil
	prev_ele_, pre_ele_exist_ := cache.sync_map.Load(key)
	if pre_ele_exist_ {
		pre_ele = prev_ele_.(*cache_element)
	}

	//not meaningful
	if !pre_ele_exist_ && ttlSecond == 0 {
		return nil
	}

	//keep old ttl
	if pre_ele_exist_ && ttlSecond == 0 {
		expire_time = pre_ele.Score
	}

	//set to map
	cache.sync_map.Store(key, &cache_element{
		Score: expire_time,
		Value: value,
	})

	//statistics
	if pre_ele_exist_ {
		//cache.element_count--
		//cache.element_count++
		cache.element_bytes -= int64(pre_ele.Value.(CacheItem).CacheBytes())
		cache.element_bytes += int64(value.CacheBytes())
	} else {
		cache.element_count++
		cache.element_bytes += int64(value.CacheBytes())
	}

	//dispatch update msg to chan
	cache.sl_channel <- func() {
		if pre_ele_exist_ {
			cache.skip_list.remove(key, pre_ele.Score)
		}
		cache.skip_list.insert(key, expire_time)
	}

	return nil
}

func (cache *Cache) Delete(key string) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	prev_ele_, pre_ele_exist_ := cache.sync_map.Load(key)
	if pre_ele_exist_ {
		//
		pre_ele := prev_ele_.(*cache_element)
		cache.sync_map.Delete(key)

		//statistics
		cache.element_count--
		cache.element_bytes -= int64(pre_ele.Value.(CacheItem).CacheBytes())

		//dispatch update msg to chan
		cache.sl_channel <- func() {
			cache.skip_list.remove(key, pre_ele.Score)
		}
	}
}

func (cache *Cache) Items() int32 {
	return cache.element_count
}

func (cache *Cache) Bytes() int32 {
	return int32(cache.element_bytes)
}

func safeInfiLoop(todo func(), onPanic func(err interface{}), interval int64, redoDelaySec int64) {
	runChannel := make(chan struct{})
	go func() {
		for {
			<-runChannel
			go func() {
				defer func() {
					if err := recover(); err != nil {
						if onPanic != nil {
							onPanic(err)
						}
						time.Sleep(time.Duration(redoDelaySec) * time.Second)
						runChannel <- struct{}{}
					}
				}()
				for {
					todo()
					time.Sleep(time.Duration(interval) * time.Second)
				}
			}()
		}
	}()
	runChannel <- struct{}{}
}
