package cache

import "testing"

func TestCacheGet(t *testing.T) {
	cache := NewLRU(2)
	cache.Put("mew", "test")
	value := cache.Get("pikachu")
	if value != nil {
		t.Errorf("Get('pikachu') want %v received %v", nil, value)
	}

	value = cache.Get("mew")
	if value != "test" {
		t.Errorf("Get('pikachu') want %s received %v", "test", value)
	}
}

func TestCacheUpdate(t *testing.T) {
	cache := NewLRU(1)
	cache.Put("pikachu", 1)
	value := cache.Get("pikachu")
	if value != 1 {
		t.Errorf("Get('pikachu') want %d received %v", 1, value)
	}

	cache.Put("pikachu", 3)
	value = cache.Get("pikachu")
	if value != 3 {
		t.Errorf("Get('pikachu') want %d received %v", 3, value)
	}
}

func TestCacheEviction(t *testing.T) {
	cache := NewLRU(1)
	cache.Put("pikachu", 1)
	cache.Put("Espeon", 4)
	value := cache.Get("pikachu")
	if value != nil {
		t.Errorf("Get('pikachu') want %v received %v", nil, value)
	}
}

func TestCacheEvictionTouched(t *testing.T) {
	cache := NewLRU(2)
	cache.Put("pikachu", 1)
	cache.Put("Espeon", 4)
	cache.Get("pikachu")
	cache.Put("Jolteon", 12)
	value := cache.Get("pikachu")
	if value != 1 {
		t.Errorf("Get('pikachu') want %d received %v", 1, value)
	}
	value = cache.Get("Espeon")
	if value != nil {
		t.Errorf("Get('Espeon') want %v received %v", nil, value)
	}
}

func TestNewLRUPanicsOnInvalidCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for non-positive capacity, got none")
		}
	}()

	NewLRU(0)
}
