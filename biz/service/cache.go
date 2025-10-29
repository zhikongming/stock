package service

import (
	"sync"
	"time"
)

type EMCache struct {
	CookieIndex int
	Timeout     time.Time
	Mutex       sync.Mutex
}

func NewEMCache() *EMCache {
	return &EMCache{
		CookieIndex: 0,
		Timeout:     time.Now().Add(time.Hour * 24 * 356),
	}
}

func (c *EMCache) GetCookieIndex() int {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	if time.Now().After(c.Timeout) {
		c.CookieIndex = 0
		c.Timeout = time.Now().Add(time.Hour)
	}
	return c.CookieIndex
}

func (c *EMCache) SetCookieIndex(idx int) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.CookieIndex = idx % len(NidList)
	c.Timeout = time.Now().Add(time.Hour)
}

var emCache = NewEMCache()
