package service

import "time"

type MemCache struct {
	Key      string
	Data     interface{}
	Expire   time.Time
	KeepTime time.Duration
}

var MemCacheMap = make(map[string]*MemCache)

func NewMemCache(key string, data interface{}, keepTime time.Duration) *MemCache {
	return &MemCache{
		Key:      key,
		Data:     data,
		Expire:   time.Now().Add(keepTime),
		KeepTime: keepTime,
	}
}

func GetMemCache(key string) *MemCache {
	cache, ok := MemCacheMap[key]
	if !ok {
		return nil
	}
	if cache.Expire.Before(time.Now()) {
		return nil
	}
	return cache
}

func SetMemCache(key string, data interface{}, keepTime time.Duration) {
	cache := NewMemCache(key, data, keepTime)
	MemCacheMap[key] = cache
}
