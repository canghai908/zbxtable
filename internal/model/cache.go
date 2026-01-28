package models

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"zbxtable/pkg/utils"
)

var (
	// Cache 全局缓存实例（使用go-cache）
	Cache *cache.Cache
	// QueueStore 队列存储（key -> channel）
	queueStore = make(map[string]*Queue)
	queueMu    sync.RWMutex
	// SortedSetStore 排序集合存储（key -> sorted set）
	sortedSetStore = make(map[string]*SortedSet)
	sortedSetMu    sync.RWMutex
)

// Queue 队列结构
type Queue struct {
	items chan []byte
	mu    sync.Mutex
}

// SortedSet 排序集合结构
type SortedSet struct {
	members map[string]float64
	mu      sync.RWMutex
}

// InitCache 初始化缓存（使用go-cache替代Redis）
func InitCache() {
	// 默认过期时间5分钟，清理间隔10分钟
	Cache = cache.New(5*time.Minute, 10*time.Minute)
	utils.Log.Info("Cache initialized (using go-cache)")
}

// ========== 键值存储操作（兼容Redis接口）==========

// CacheSet 设置键值（兼容Redis Set）
func CacheSet(key string, value string, expiration time.Duration) error {
	if expiration < 0 {
		// 负数表示永不过期，使用cache.NoExpiration
		Cache.Set(key, value, cache.NoExpiration)
	} else if expiration == 0 {
		// 0表示使用默认过期时间
		Cache.Set(key, value, cache.DefaultExpiration)
	} else {
		Cache.Set(key, value, expiration)
	}
	return nil
}

// CacheGet 获取键值（兼容Redis Get）
func CacheGet(key string) (string, error) {
	val, found := Cache.Get(key)
	if !found {
		return "", nil // 键不存在，返回空字符串（模拟Redis行为）
	}
	if str, ok := val.(string); ok {
		return str, nil
	}
	return "", nil
}

// ========== 队列操作（兼容Redis LPush/RPop）==========

// CacheLPush 将值推入队列左侧（兼容Redis LPush）
func CacheLPush(key string, value []byte) error {
	queueMu.Lock()
	queue, exists := queueStore[key]
	if !exists {
		queue = &Queue{
			items: make(chan []byte, 10000), // 缓冲10000个元素
		}
		queueStore[key] = queue
	}
	queueMu.Unlock()

	queue.mu.Lock()
	defer queue.mu.Unlock()

	select {
	case queue.items <- value:
		return nil
	default:
		// 队列满了，尝试非阻塞写入
		select {
		case queue.items <- value:
			return nil
		default:
			utils.Log.Warningf("Queue %s is full, dropping message", key)
			return nil // 不返回错误，只是丢弃消息
		}
	}
}

// CacheRPop 从队列右侧弹出值（兼容Redis RPop）
func CacheRPop(key string) (string, error) {
	queueMu.RLock()
	queue, exists := queueStore[key]
	queueMu.RUnlock()

	if !exists {
		return "", nil // 队列不存在，返回空字符串（模拟redis.Nil）
	}

	select {
	case value := <-queue.items:
		return string(value), nil
	case <-time.After(100 * time.Millisecond):
		// 超时，返回空（模拟redis.Nil）
		return "", nil
	default:
		return "", nil
	}
}

// ========== 排序集合操作（兼容Redis ZAdd/ZRevRangeWithScores）==========

// CacheZAdd 添加成员到排序集合（兼容Redis ZAdd）
func CacheZAdd(key string, member string, score float64) error {
	sortedSetMu.Lock()
	defer sortedSetMu.Unlock()

	sortedSet, exists := sortedSetStore[key]
	if !exists {
		sortedSet = &SortedSet{
			members: make(map[string]float64),
		}
		sortedSetStore[key] = sortedSet
	}

	sortedSet.mu.Lock()
	sortedSet.members[member] = score
	sortedSet.mu.Unlock()

	return nil
}

// ZMember 排序集合成员
type ZMember struct {
	Member string
	Score  float64
}

// CacheZRevRangeWithScores 获取排序集合的成员（按分数降序，兼容Redis ZRevRangeWithScores）
func CacheZRevRangeWithScores(key string, start, stop int64) ([]ZMember, error) {
	sortedSetMu.RLock()
	sortedSet, exists := sortedSetStore[key]
	sortedSetMu.RUnlock()

	if !exists {
		return []ZMember{}, nil
	}

	sortedSet.mu.RLock()
	defer sortedSet.mu.RUnlock()

	// 将map转换为切片并排序
	var members []ZMember
	for member, score := range sortedSet.members {
		members = append(members, ZMember{Member: member, Score: score})
	}

	// 按分数降序排序
	for i := 0; i < len(members)-1; i++ {
		for j := i + 1; j < len(members); j++ {
			if members[i].Score < members[j].Score {
				members[i], members[j] = members[j], members[i]
			}
		}
	}

	// 应用范围
	if start < 0 {
		start = int64(len(members)) + start
	}
	if stop < 0 {
		stop = int64(len(members)) + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= int64(len(members)) {
		stop = int64(len(members)) - 1
	}
	if start > stop {
		return []ZMember{}, nil
	}

	result := make([]ZMember, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		if i < int64(len(members)) {
			result = append(result, members[i])
		}
	}

	return result, nil
}
