package solution

import (
	"sync"
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	cache := New()

	cache.Set(1, "value1")

	value, found := cache.Get(1)
	if !found || value != "value1" {
		t.Errorf("expected value1, got %v", value)
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := New(WithCustomLifetime(500 * time.Millisecond))

	cache.Set(1, "value1")
	time.Sleep(600 * time.Millisecond)

	_, found := cache.Get(1)
	if found {
		t.Error("expected value to expire, but it didn't")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := New()

	var wg sync.WaitGroup

	numOperations := 100

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cache.Set(uint64(i), i)
		}(i)
	}

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			value, found := cache.Get(uint64(i))
			if found && value != i {
				t.Errorf("expected value %d, got %v", i, value)
			}
		}(i)
	}
	wg.Wait()

}
