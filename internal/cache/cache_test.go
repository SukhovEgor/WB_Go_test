package cache

import (
	"test-task/pkg/models"
	"testing"
	"math/rand"
)

func TestCache_BaseFunctionality(t *testing.T) {
	cache := MakeCache(5)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	saved_order := models.createRandomOrder(rng)
	cache.Add(saved_order)

	from_cache, found := cache.Get(saved_order.OrderUID)
	if !found {
		t.Errorf("Cache didn't find added order")
	}
	if from_cache != saved_order {
		t.Errorf("Got different order. Got: %s, wanted: %s", from_cache.OrderUID, saved_order.Entry)
	}
}

func TestCache_SearchInEmptyCache(t *testing.T) {
	cache := CreateCache(2)
	not_existing_id := "13s"
	order, found := cache.Get(not_existing_id)
	if found {
		t.Errorf("Found order in empty cache. got %s, searched for %s", order.OrderUID, not_existing_id)
	}
}

func TestCache_CacheEviction(t *testing.T) {
	cache := CreateCache(2)

	order1 := models.MakeRandomOrder()
	order2 := models.MakeRandomOrder()
	order3 := models.MakeRandomOrder()

	cache.Add(order1)
	cache.Add(order2)
	cache.Add(order3) 

	_, found := cache.Get(order1.OrderUID)
	if found {
		t.Error("Order1 should be evicted")
	}

	if _, found := cache.Get(order2.OrderUID); !found {
		t.Error("Order2 should still be in cache")
	}
	if _, found := cache.Get(order3.OrderUID); !found {
		t.Error("Order3 should still be in cache")
	}
}

func TestCache_LRUOrderCheck(t *testing.T) {
	cache := CreateCache(2)

	order1 := models.MakeRandomOrder()
	order2 := models.MakeRandomOrder()
	order3 := models.MakeRandomOrder()

	cache.Add(order1)
	cache.Add(order2)

	cache.Get(order1.OrderUID)

	cache.Add(order3) 

	if _, found := cache.Get(order2.OrderUID); found {
		t.Error("Order2 should be evicted")
	}

	if _, found := cache.Get(order1.OrderUID); !found {
		t.Error("Order1 should still be in cache")
	}
	if _, found := cache.Get(order3.OrderUID); !found {
		t.Error("Order3 should still be in cache")
	}
}
