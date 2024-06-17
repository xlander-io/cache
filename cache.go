package cache

import (
	"errors"
	"time"
)

type CacheItem interface {
	CacheBytes() int
}

type CacheConfig struct {
	CacheBytesLimit       int64
	MaxTtlSecs            int
	RecycleIntervalSecs   int
	RecycleRatioThreshold int //1-100, e.g: 30 for 30% percentage
}

var DefaultCacheConfig = CacheConfig{
	CacheBytesLimit:       1024 * 1024 * 50, //50Mbytes
	MaxTtlSecs:            7200,             //2 hours
	RecycleIntervalSecs:   30,               //30 secs
	RecycleRatioThreshold: 80,               //80% usage will trigger recycling
}

type Cache struct {
	cache_bytes_limit       int64
	max_ttl_secs            int
	recycle_interval_secs   int
	recycle_ratio_threshold int
	recycle_bytes_threshold int64
	s                       *SortedSet
	now_unixtime            int64
}

func New(config_ *CacheConfig) (*Cache, error) {

	//config
	config := &CacheConfig{
		CacheBytesLimit:       DefaultCacheConfig.CacheBytesLimit,
		MaxTtlSecs:            DefaultCacheConfig.MaxTtlSecs,
		RecycleIntervalSecs:   DefaultCacheConfig.RecycleIntervalSecs,
		RecycleRatioThreshold: DefaultCacheConfig.RecycleRatioThreshold,
	}

	if config_ != nil {
		//
		if config_.CacheBytesLimit < 0 {
			return nil, errors.New("config CacheBytesLimit error")
		} else if config_.CacheBytesLimit == 0 {
			//bypass using default value
		} else {
			config.CacheBytesLimit = config_.CacheBytesLimit
		}
		//
		if config_.MaxTtlSecs < 0 {
			return nil, errors.New("config MaxTtlSecs error")
		} else if config_.MaxTtlSecs == 0 {
			//bypass using default value
		} else {
			config.MaxTtlSecs = config_.MaxTtlSecs
		}
		//
		if config_.RecycleIntervalSecs < 0 {
			return nil, errors.New("config RecycleIntervalSecs error")
		} else if config_.RecycleIntervalSecs == 0 {
			//bypass using default value
		} else {
			config.RecycleIntervalSecs = config_.RecycleIntervalSecs
		}
		//
		if config_.RecycleRatioThreshold < 0 {
			return nil, errors.New("config RecycleRatioThreshold error")
		} else if config_.RecycleRatioThreshold == 0 {
			//bypass using default value
		} else {
			config.RecycleRatioThreshold = config_.RecycleRatioThreshold
		}

	}

	var config_recycle_bytes_threshold int64 = (config.CacheBytesLimit * int64(config.RecycleRatioThreshold) / 100)
	if config_recycle_bytes_threshold < 1 {
		return nil, errors.New("CacheBytesLimit*RecycleRatioThreshold must >= 1")
	}
	///
	cache := &Cache{
		s:                       NewSortedSet(),
		now_unixtime:            time.Now().Unix(),
		cache_bytes_limit:       config.CacheBytesLimit,
		max_ttl_secs:            config.MaxTtlSecs,
		recycle_interval_secs:   config.RecycleIntervalSecs,
		recycle_ratio_threshold: config.RecycleRatioThreshold,
		recycle_bytes_threshold: config_recycle_bytes_threshold,
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
		//remove expired keys
		cache.s.RemoveByScore(cache.now_unixtime)

		//check overlimit
		for cache.s.TotalBytes() >= int64(cache.recycle_bytes_threshold) {
			cache.s.RemoveByRank(0, cache.s.Len()/10+1) //10% recycled and +1 for safety
		}

	}, nil, int64(cache.recycle_interval_secs), 30)
	//
	return cache, nil
}

// RecycleOverLimitRatio of records will be recycled if the number of total keys exceeds this limit
// func (lf *Cache) SetMaxRecords(limit int64) {
// 	if limit < MinRecords {
// 		limit = MinRecords
// 	}
// 	lf.limit = limit
// }

// get current unix time in the cache
func (lf *Cache) GetUnixTime() int64 {
	return lf.now_unixtime
}

// if not found or timeout => return nil,0
// if found and not timeout =>return not_nil_pointer,left_secs
func (lf *Cache) Get(key string) (value CacheItem, ttl int64) {
	//check expire
	e, exist := lf.s.Get(key)
	if !exist {
		return nil, 0
	}
	if e.Score <= lf.now_unixtime {
		return nil, 0
	}
	return e.Value.(CacheItem), e.Score - lf.now_unixtime
}

// if ttl < 0 just return and nothing changes
// ttl is set to MaxTTLSecs if ttl > MaxTTLSecs
// if record exist , "0" ttl changes nothing
// if record not exist, "0" ttl is equal to "30" seconds
func (lf *Cache) Set(key string, value CacheItem, ttlSecond int64) error {
	if value == nil {
		return errors.New("value can not be nil")
	}
	if ttlSecond < 0 {
		return errors.New("ttl error")
	}

	// t := reflect.TypeOf(value).Kind()
	// if t != reflect.Ptr && t != reflect.Slice && t != reflect.Map {
	// 	return errors.New("value only support Pointer Slice and Map")
	// }

	if ttlSecond > MaxTTLSecs {
		ttlSecond = MaxTTLSecs
	}
	var expireTime int64

	if ttlSecond == 0 {
		//keep
		ttlLeft, exist := lf.ttl(key)
		if !exist {
			ttlLeft = 30
		}
		expireTime = lf.now_unixtime + ttlLeft
	} else {
		//new expire
		expireTime = lf.now_unixtime + ttlSecond
	}
	lf.s.Add(key, expireTime, value)
	return nil
}

func (lf *Cache) Delete(key string) {
	lf.s.Remove(key)
}

// get ttl of a key in seconds
func (lf *Cache) ttl(key string) (int64, bool) {
	e, exist := lf.s.Get(key)
	if !exist {
		return 0, false
	}
	ttl := e.Score - lf.now_unixtime
	if ttl <= 0 {
		return 0, false
	}
	return ttl, true
}

func (lf *Cache) TotalItems() int64 {
	return int64(lf.s.Len())
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
