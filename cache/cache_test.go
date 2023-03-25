package cache

import (
	"sync"
	"testing"
)

func Test_BasicCache(t *testing.T) {
	cache := InitCache()
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		cache.Add("string", 3, 3)
	}()
	go func() {
		wg.Add(1)
		defer wg.Done()
		cache.Add("string1", 4, 3)
	}()
	wg.Wait()

	v, err := cache.Get("string", nil)
	if err != nil {
		t.Error("something wrong happened: %w", err)
	}
	if v.(int) != 3 {
		t.Error("not tje same")
	}
}
