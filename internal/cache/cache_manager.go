package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/xurenlu/aipipe/internal/config"
)

// 缓存项
type CacheItem struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
	CreatedAt time.Time   `json:"created_at"`
	AccessCount int64     `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
}

// 缓存统计
type CacheStats struct {
	TotalItems    int64   `json:"total_items"`
	HitCount      int64   `json:"hit_count"`
	MissCount     int64   `json:"miss_count"`
	HitRate       float64 `json:"hit_rate"`
	MemoryUsage   int64   `json:"memory_usage"`
	EvictionCount int64   `json:"eviction_count"`
	LastCleanup   time.Time `json:"last_cleanup"`
}

// 缓存管理器
type CacheManager struct {
	config     config.CacheConfig
	items      map[string]*CacheItem
	stats      CacheStats
	mutex      sync.RWMutex
	cleanupTicker *time.Ticker
	stopChan   chan bool
}

// 创建新的缓存管理器
func NewCacheManager(cfg config.CacheConfig) *CacheManager {
	cm := &CacheManager{
		config: cfg,
		items:  make(map[string]*CacheItem),
		stats: CacheStats{
			LastCleanup: time.Now(),
		},
		stopChan: make(chan bool),
	}

	// 启动清理协程
	if cfg.Enabled {
		cm.startCleanup()
	}

	return cm
}

// 生成缓存键
func (cm *CacheManager) generateKey(prefix string, data interface{}) string {
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%s:%x", prefix, hash[:8])
}

// 获取缓存项
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	item, exists := cm.items[key]
	if !exists {
		cm.stats.MissCount++
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		delete(cm.items, key)
		cm.stats.MissCount++
		return nil, false
	}

	// 更新访问统计
	item.AccessCount++
	item.LastAccess = time.Now()
	cm.stats.HitCount++
	cm.updateHitRate()

	return item.Value, true
}

// 设置缓存项
func (cm *CacheManager) Set(key string, value interface{}, ttl time.Duration) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查缓存大小限制
	if len(cm.items) >= cm.config.MaxItems {
		cm.evictOldest()
	}

	// 创建缓存项
	item := &CacheItem{
		Key:        key,
		Value:      value,
		ExpiresAt:  time.Now().Add(ttl),
		CreatedAt:  time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	cm.items[key] = item
	cm.stats.TotalItems = int64(len(cm.items))

	return nil
}

// 删除缓存项
func (cm *CacheManager) Delete(key string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.items, key)
	cm.stats.TotalItems = int64(len(cm.items))
}

// 清空缓存
func (cm *CacheManager) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.items = make(map[string]*CacheItem)
	cm.stats.TotalItems = 0
}

// 获取缓存统计
func (cm *CacheManager) GetStats() CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cm.stats.MemoryUsage = cm.calculateMemoryUsage()
	return cm.stats
}

// 计算内存使用量
func (cm *CacheManager) calculateMemoryUsage() int64 {
	var totalSize int64
	for _, item := range cm.items {
		// 简单的内存估算
		if jsonData, err := json.Marshal(item); err == nil {
			totalSize += int64(len(jsonData))
		}
	}
	return totalSize
}

// 更新命中率
func (cm *CacheManager) updateHitRate() {
	total := cm.stats.HitCount + cm.stats.MissCount
	if total > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(total)
	}
}

// 驱逐最旧的项
func (cm *CacheManager) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range cm.items {
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
		}
	}

	if oldestKey != "" {
		delete(cm.items, oldestKey)
		cm.stats.EvictionCount++
	}
}

// 启动清理协程
func (cm *CacheManager) startCleanup() {
	cm.cleanupTicker = time.NewTicker(cm.config.CleanupInterval)
	
	go func() {
		for {
			select {
			case <-cm.cleanupTicker.C:
				cm.cleanup()
			case <-cm.stopChan:
				return
			}
		}
	}()
}

// 清理过期项
func (cm *CacheManager) cleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, item := range cm.items {
		if now.After(item.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(cm.items, key)
	}

	cm.stats.TotalItems = int64(len(cm.items))
	cm.stats.LastCleanup = now
}

// 停止缓存管理器
func (cm *CacheManager) Stop() {
	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}
	close(cm.stopChan)
}

// 缓存 AI 分析结果
func (cm *CacheManager) CacheAIAnalysis(logLine string, analysis interface{}) {
	if !cm.config.Enabled {
		return
	}

	key := cm.generateKey("ai", logLine)
	cm.Set(key, analysis, cm.config.AITTL)
}

// 获取 AI 分析结果
func (cm *CacheManager) GetAIAnalysis(logLine string) (interface{}, bool) {
	if !cm.config.Enabled {
		return nil, false
	}

	key := cm.generateKey("ai", logLine)
	return cm.Get(key)
}

// 缓存规则匹配结果
func (cm *CacheManager) CacheRuleMatch(logLine string, result interface{}) {
	if !cm.config.Enabled {
		return
	}

	key := cm.generateKey("rule", logLine)
	cm.Set(key, result, cm.config.RuleTTL)
}

// 获取规则匹配结果
func (cm *CacheManager) GetRuleMatch(logLine string) (interface{}, bool) {
	if !cm.config.Enabled {
		return nil, false
	}

	key := cm.generateKey("rule", logLine)
	return cm.Get(key)
}
