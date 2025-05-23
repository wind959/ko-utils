package maputil

import (
	"fmt"
	"sync"
)

const defaultShardCount = 32

// ConcurrentMap ConcurrentMap协程安全的map结构
type ConcurrentMap[K comparable, V any] struct {
	shardCount uint64
	locks      []sync.RWMutex
	maps       []map[K]V
}

// NewConcurrentMap create a ConcurrentMap with specific shard count.
func NewConcurrentMap[K comparable, V any](shardCount int) *ConcurrentMap[K, V] {
	if shardCount <= 0 {
		shardCount = defaultShardCount
	}

	cm := &ConcurrentMap[K, V]{
		shardCount: uint64(shardCount),
		locks:      make([]sync.RWMutex, shardCount),
		maps:       make([]map[K]V, shardCount),
	}

	for i := range cm.maps {
		cm.maps[i] = make(map[K]V)
	}

	return cm
}

// Set 在map中设置key和value
func (cm *ConcurrentMap[K, V]) Set(key K, value V) {
	shard := cm.getShard(key)

	cm.locks[shard].Lock()
	cm.maps[shard][key] = value

	cm.locks[shard].Unlock()
}

// Get 根据key获取value, 如果不存在key,返回零值
func (cm *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	shard := cm.getShard(key)

	cm.locks[shard].RLock()
	value, ok := cm.maps[shard][key]
	cm.locks[shard].RUnlock()

	return value, ok
}

// GetOrSet 返回键的现有值（如果存在），否则，设置key并返回给定值
func (cm *ConcurrentMap[K, V]) GetOrSet(key K, value V) (actual V, ok bool) {
	shard := cm.getShard(key)

	cm.locks[shard].RLock()
	if actual, ok := cm.maps[shard][key]; ok {
		cm.locks[shard].RUnlock()
		return actual, ok
	}
	cm.locks[shard].RUnlock()

	// lock again
	cm.locks[shard].Lock()
	if actual, ok = cm.maps[shard][key]; ok {
		cm.locks[shard].Unlock()
		return
	}

	cm.maps[shard][key] = value
	cm.locks[shard].Unlock()

	return value, ok
}

// Delete 删除key
func (cm *ConcurrentMap[K, V]) Delete(key K) {
	shard := cm.getShard(key)

	cm.locks[shard].Lock()
	delete(cm.maps[shard], key)
	cm.locks[shard].Unlock()
}

// GetAndDelete 获取key，然后删除
func (cm *ConcurrentMap[K, V]) GetAndDelete(key K) (actual V, ok bool) {
	shard := cm.getShard(key)

	cm.locks[shard].RLock()
	if actual, ok = cm.maps[shard][key]; ok {
		cm.locks[shard].RUnlock()
		cm.Delete(key)
		return
	}
	cm.locks[shard].RUnlock()

	return actual, false
}

// Has 验证是否包含key
func (cm *ConcurrentMap[K, V]) Has(key K) bool {
	_, ok := cm.Get(key)
	return ok
}

// Range 为map中每个键和值顺序调用迭代器。 如果iterator返回false，则停止迭代
func (cm *ConcurrentMap[K, V]) Range(iterator func(key K, value V) bool) {
	for shard := range cm.locks {
		cm.locks[shard].RLock()

		for k, v := range cm.maps[shard] {
			if !iterator(k, v) {
				cm.locks[shard].RUnlock()
				return
			}
		}
		cm.locks[shard].RUnlock()
	}
}

// getShard get shard by a key.
func (cm *ConcurrentMap[K, V]) getShard(key K) uint64 {
	hash := fnv32(fmt.Sprintf("%v", key))
	return uint64(hash) % cm.shardCount
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
