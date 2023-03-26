package cache

import (
	"fmt"
	"sync"
	"time"
)

type Entry struct {
	value      any
	expiration time.Time
	read       chan struct{}
}

func (e Entry) isExpired() bool {
	return e.expiration.Before(time.Now())
}

type Cache struct {
	entries map[string]Entry
	lock    *sync.RWMutex
}

func InitCache() Cache {
	entries := make(map[string]Entry)
	var lock = sync.RWMutex{}
	return Cache{entries, &lock}
}

func (c Cache) Add(key string, value any, ttlSecs int) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	expiration := time.Now().Add(time.Second * time.Duration(ttlSecs))
	c.entries[key] = Entry{
		value:      value,
		expiration: expiration,
	}
	return nil
}

func (c Cache) Get(key string, f func(key string) (value any, ttlSecs int, err error)) (any, error) {
	v, ok := c.entries[key]
	if !ok || (v.value != nil && v.isExpired()) {
		// open channel
		read := make(chan struct{})
		c.entries[key] = Entry{
			read: read,
		}

		value, ttl, err := f(key)
		if err != nil {
			return nil, err
		}
		err = c.Add(key, value, ttl)
		if err != nil {
			return nil, err
		}

		// close channel
		close(read)
		return c.Get(key, nil)
	}
	if v.read != nil {
		for {
			select {
			case <-v.read:
				fmt.Println("read from waiting channel")
				return c.Get(key, nil)
			}
		}
	}
	return v.value, nil
}

func (c Cache) Delete(key string) {
	delete(c.entries, key)
}
