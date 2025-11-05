package cache

import "testing"

func TestCacheGet(t *testing.T) {
	cache := NewLRU(2)
	cache.Put("mew", "test")
	value, exists := cache.Get("pikachu")
	if exists {
		t.Error("Get('pikachu') should not exist")
	}

	value, exists = cache.Get("mew")
	if !exists {
		t.Error("Get('mew') should exist")
	}
	if value != "test" {
		t.Errorf("Get('mew') want %s received %v", "test", value)
	}
}

func TestCacheUpdate(t *testing.T) {
	cache := NewLRU(1)
	cache.Put("pikachu", 1)
	value, exists := cache.Get("pikachu")
	if !exists {
		t.Error("Get('mew') should exist")
	}
	if value != 1 {
		t.Errorf("Get('pikachu') want %d received %v", 1, value)
	}

	cache.Put("pikachu", 3)
	value, exists = cache.Get("pikachu")
	if !exists {
		t.Error("Get('pikachu') should exist")
	}
	if value != 3 {
		t.Errorf("Get('pikachu') want %d received %v", 3, value)
	}
}

func TestCacheEviction(t *testing.T) {
	cache := NewLRU(1)
	cache.Put("pikachu", 1)
	cache.Put("Espeon", 4)
	_, exists := cache.Get("pikachu")
	if exists {
		t.Error("Get('pikachu') should not exist")
	}
}

func TestCacheEvictionTouched(t *testing.T) {
	cache := NewLRU(2)
	cache.Put("pikachu", 1)
	cache.Put("Espeon", 4)
	cache.Get("pikachu")
	cache.Put("Jolteon", 12)
	value, exists := cache.Get("pikachu")
	if !exists {
		t.Error("Get('pikachu') should exist")
	}
	if value != 1 {
		t.Errorf("Get('pikachu') want %d received %v", 1, value)
	}
	value, exists = cache.Get("Espeon")
	if exists {
		t.Error("Get('Espeon') should not exist")
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
