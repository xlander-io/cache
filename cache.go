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
	RecycleCheckIntervalSecs int   // recycle process cycle in secs
	RecycleRatioThreshold    int   // 1-100, old items are recycled once cache reach limit, e.g: 30 for 30% percentage
	SkipListBufferSize       int   // chan buffer between internal map and skiplist
	DefaultTtlSecs           int64 //default cache item duration in secs
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

func New(user_config *CacheConfig) (*Cache, error) {

	//new a cache with default config

	cache_config := &CacheConfig{
		CacheBytesLimit:          1024 * 1024 * 50, //50Mbytes
		MaxTtlSecs:               7200,             //2 hours
		RecycleCheckIntervalSecs: 30,               //30 secs
		RecycleRatioThreshold:    80,               //80% usage will trigger recycling
		SkipListBufferSize:       20000,
		DefaultTtlSecs:           30, //default cache item duration is 30 secs
	}

	if user_config != nil {

		//
		if user_config.CacheBytesLimit < 0 {
			return nil, errors.New("config CacheBytesLimit error")
		} else if user_config.CacheBytesLimit == 0 {
			//bypass using default value
		} else {
			cache_config.CacheBytesLimit = user_config.CacheBytesLimit
		}

		//
		if user_config.MaxTtlSecs < 0 {
			return nil, errors.New("config MaxTtlSecs error")
		} else if user_config.MaxTtlSecs == 0 {
			//bypass using default value
		} else {
			cache_config.MaxTtlSecs = user_config.MaxTtlSecs
		}

		//
		if user_config.RecycleCheckIntervalSecs < 0 {
			return nil, errors.New("config RecycleIntervalSecs error")
		} else if user_config.RecycleCheckIntervalSecs == 0 {
			//bypass using default value
		} else {
			cache_config.RecycleCheckIntervalSecs = user_config.RecycleCheckIntervalSecs
		}

		//
		if user_config.RecycleRatioThreshold < 0 {
			return nil, errors.New("config RecycleRatioThreshold error")
		} else if user_config.RecycleRatioThreshold == 0 {
			//bypass using default value
		} else {
			cache_config.RecycleRatioThreshold = user_config.RecycleRatioThreshold
		}

		//
		if user_config.SkipListBufferSize <= 0 {
			return nil, errors.New("config SkipListBufferSize error")
		} else if user_config.SkipListBufferSize == 0 {
			//bypass using default value
		} else {
			cache_config.SkipListBufferSize = user_config.SkipListBufferSize
		}

	}

	var config_recycle_bytes_threshold int64 = (cache_config.CacheBytesLimit * int64(cache_config.RecycleRatioThreshold) / 100)
	if config_recycle_bytes_threshold < 1 {
		return nil, errors.New("CacheBytesLimit*RecycleRatioThreshold must >= 1")
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

	//start the recycle go routine
	safeInfiLoop(func() {
		// //remove expired keys
		// cache.s.RemoveByScore(cache.now_unixtime)

		// //check overlimit
		// for cache.s.TotalBytes() >= int64(cache.recycle_bytes_threshold) {
		// 	cache.s.RemoveByRank(0, cache.s.Len()/10+1) //10% recycled and +1 for safety
		// }

	}, nil, int64(cache.cache_config.RecycleCheckIntervalSecs), 30)

	//
	return cache, nil
}

// get current unix time in the cache
func (cache *Cache) GetUnixTime() int64 {
	return cache.now_unixtime
}

// // if not found or timeout => return nil,0
// // if found and not timeout =>return not_nil_pointer,left_secs
// func (cache *Cache) Get(key string) (value CacheItem, ttl int64) {
// 	//check expire
// 	e, exist := lf.s.Get(key)
// 	if !exist {
// 		return nil, 0
// 	}
// 	if e.Score <= lf.now_unixtime {
// 		return nil, 0
// 	}
// 	return e.Value.(CacheItem), e.Score - lf.now_unixtime
// }

// func (cache *Cache) Set(key string, value CacheItem, ttlSecond int64) error {
// 	return cache.set(key, cache_element{

// 	}, ttlSecond)
// }

// todo write docs
func (cache *Cache) Set(key string, value CacheItem, ttlSecond int64) error {

	cache.lock.Lock()
	defer cache.lock.Unlock()

	if value == nil {
		return errors.New("value can not be nil")
	}

	if ttlSecond < 0 {
		return errors.New("ttl error")
	}

	if ttlSecond > cache.cache_config.MaxTtlSecs {
		ttlSecond = cache.cache_config.MaxTtlSecs
	}

	//var expireTime int64
	ttlLeft := cache.cache_config.DefaultTtlSecs
	if ttlSecond == 0 {
		//keep
		prev_ttl_left, exist := cache.ttl(key)
		if exist {
			ttlLeft = prev_ttl_left
		}
	}
	expire_time := cache.now_unixtime + ttlLeft

	//set to map
	cache.sync_map.Store(key, &cache_element{
		Score: expire_time,
		Value: value,
	})

	//dispatch update msg to chan
	cache.sl_channel <- func() {
		cache.skip_list.remove_all_member(key)
		cache.skip_list.insert(key, expire_time)
	}

	return nil
}

// func (lf *Cache) Delete(key string) {
// 	lf.s.Remove(key)
// }

// return the left cached time in secs
func (cache *Cache) ttl(key string) (int64, bool) {

	e, exist := cache.sync_map.Load(key)
	if !exist {
		return 0, false
	}

	c_e := e.(cache_element)
	ttl := c_e.Score - cache.now_unixtime

	if ttl <= 0 {
		return 0, false
	}
	return ttl, true
}

func (cache *Cache) TotalItems() int32 {
	return cache.element_count
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
