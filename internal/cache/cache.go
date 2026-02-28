package cache

import (
	"container/list"
	"log"

	"test-task/pkg/models"
)

type Cache struct {
	capacity  int
	cacheMap  map[string]*list.Element
	cacheList *list.List
}

func CreateCache(capacity int) *Cache {
	return &Cache{
		capacity:  capacity,
		cacheMap:  make(map[string]*list.Element),
		cacheList: list.New(),
	}
}

func (cache *Cache) Add(order *models.Order) {

	if existingElement, exist := cache.cacheMap[order.OrderUID]; exist {
		existingElement.Value = order
		cache.cacheList.MoveToFront(existingElement)
		return
	}

	element := cache.cacheList.PushFront(order)
	cache.cacheMap[order.OrderUID] = element
	log.Printf("Add order into cache: %v", order.OrderUID)

	if len(cache.cacheMap) > cache.capacity {
		log.Printf("Remove oldest orders")
		cache.removeOldest()
	}
}

func (cache *Cache) Get(orderUid string) (order *models.Order, exist bool, err error) {
	element, exist := cache.cacheMap[orderUid]
	if !exist {
		return nil, false, err
	}

	cache.cacheList.MoveToFront(element)
	return element.Value.(*models.Order), true, nil
}

func (cache *Cache) removeOldest() {

	oldestElement := cache.cacheList.Back()
	if oldestElement != nil {
		delete(cache.cacheMap, oldestElement.Value.(*models.Order).OrderUID)
		cache.cacheList.Remove(oldestElement)
	}
}
