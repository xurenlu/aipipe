package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// 缓存项
type CacheItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	ExpiresAt   time.Time   `json:"expires_at"`
	CreatedAt   time.Time   `json:"created_at"`
	AccessCount int         `json:"access_count"`
	Size        int64       `json:"size"`
}

// AI分析结果缓存
type AIAnalysisCache struct {
	LogHash    string    `json:"log_hash"`
	Result     string    `json:"result"`
	Confidence float64   `json:"confidence"`
	Model      string    `json:"model"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// 规则匹配缓存
type RuleMatchCache struct {
	LogHash   string        `json:"log_hash"`
	RuleID    string        `json:"rule_id"`
	Matched   bool          `json:"matched"`
	Result    *FilterResult `json:"result"`
	CreatedAt time.Time     `json:"created_at"`
	ExpiresAt time.Time     `json:"expires_at"`
}

// 缓存统计
type CacheStats struct {
	TotalItems    int     `json:"total_items"`
	HitCount      int64   `json:"hit_count"`
	MissCount     int64   `json:"miss_count"`
	EvictionCount int64   `json:"eviction_count"`
	MemoryUsage   int64   `json:"memory_usage"`
	HitRate       float64 `json:"hit_rate"`
	ExpiredItems  int     `json:"expired_items"`
}

// 缓存管理器
type CacheManager struct {
	aiCache         map[string]*AIAnalysisCache
	ruleCache       map[string]*RuleMatchCache
	configCache     map[string]*CacheItem
	stats           CacheStats
	mutex           sync.RWMutex
	maxSize         int64
	maxItems        int
	cleanupInterval time.Duration
	stopCleanup     chan bool
}

// 缓存配置
type CacheConfig struct {
	MaxSize         int64         `json:"max_size"`         // 最大内存使用量（字节）
	MaxItems        int           `json:"max_items"`        // 最大缓存项数
	DefaultTTL      time.Duration `json:"default_ttl"`      // 默认过期时间
	AITTL           time.Duration `json:"ai_ttl"`           // AI分析结果过期时间
	RuleTTL         time.Duration `json:"rule_ttl"`         // 规则匹配过期时间
	ConfigTTL       time.Duration `json:"config_ttl"`       // 配置缓存过期时间
	CleanupInterval time.Duration `json:"cleanup_interval"` // 清理间隔
	Enabled         bool          `json:"enabled"`          // 是否启用缓存
}

// 缓存管理器方法

// 创建新的缓存管理器
func NewCacheManager(config CacheConfig) *CacheManager {
	cm := &CacheManager{
		aiCache:         make(map[string]*AIAnalysisCache),
		ruleCache:       make(map[string]*RuleMatchCache),
		configCache:     make(map[string]*CacheItem),
		maxSize:         config.MaxSize,
		maxItems:        config.MaxItems,
		cleanupInterval: config.CleanupInterval,
		stopCleanup:     make(chan bool),
	}

	// 启动清理协程
	if config.Enabled {
		go cm.startCleanup()
	}

	return cm
}

// 启动定期清理
func (cm *CacheManager) startCleanup() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.cleanup()
		case <-cm.stopCleanup:
			return
		}
	}
}

// 清理过期缓存
func (cm *CacheManager) cleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	// 清理AI分析缓存
	for key, item := range cm.aiCache {
		if now.After(item.ExpiresAt) {
			delete(cm.aiCache, key)
			expiredCount++
		}
	}

	// 清理规则匹配缓存
	for key, item := range cm.ruleCache {
		if now.After(item.ExpiresAt) {
			delete(cm.ruleCache, key)
			expiredCount++
		}
	}

	// 清理配置缓存
	for key, item := range cm.configCache {
		if now.After(item.ExpiresAt) {
			delete(cm.configCache, key)
			expiredCount++
		}
	}

	cm.stats.ExpiredItems = expiredCount
	cm.updateStats()
}

// 获取配置缓存
func (cm *CacheManager) GetConfig(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	item, exists := cm.configCache[key]
	if !exists {
		cm.stats.MissCount++
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		cm.stats.MissCount++
		return nil, false
	}

	item.AccessCount++
	cm.stats.HitCount++
	return item.Value, true
}

// 设置配置缓存
func (cm *CacheManager) SetConfig(key string, value interface{}, ttl time.Duration) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查是否需要清理空间
	if cm.needsEviction() {
		cm.evictOldest()
	}

	now := time.Now()
	item := &CacheItem{
		Key:         key,
		Value:       value,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		AccessCount: 0,
		Size:        cm.calculateItemSize(value),
	}

	cm.configCache[key] = item
	cm.updateStats()
}

// 更新统计信息
func (cm *CacheManager) updateStats() {
	cm.stats.TotalItems = len(cm.aiCache) + len(cm.ruleCache) + len(cm.configCache)

	// 计算命中率
	total := cm.stats.HitCount + cm.stats.MissCount
	if total > 0 {
		cm.stats.HitRate = float64(cm.stats.HitCount) / float64(total) * 100
	}

	// 计算内存使用量
	cm.stats.MemoryUsage = cm.calculateMemoryUsage()
}

// 计算内存使用量
func (cm *CacheManager) calculateMemoryUsage() int64 {
	var total int64

	for _, item := range cm.aiCache {
		total += int64(len(item.LogHash) + len(item.Result) + len(item.Model))
	}

	for _, item := range cm.ruleCache {
		total += int64(len(item.LogHash) + len(item.RuleID))
		if item.Result != nil {
			total += int64(len(item.Result.Action) + len(item.Result.RuleID))
		}
	}

	for _, item := range cm.configCache {
		total += int64(len(item.Key)) + item.Size
	}

	return total
}

// 计算项目大小
func (cm *CacheManager) calculateItemSize(value interface{}) int64 {
	data, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return int64(len(data))
}

// 检查是否需要清理
func (cm *CacheManager) needsEviction() bool {
	return cm.stats.MemoryUsage > cm.maxSize || cm.stats.TotalItems > cm.maxItems
}

// 清理最旧的项
func (cm *CacheManager) evictOldest() {
	// 简单的LRU策略：清理访问次数最少的项
	var oldestKey string
	var oldestAccess int = int(^uint(0) >> 1) // 最大int值

	for key, item := range cm.configCache {
		if item.AccessCount < oldestAccess {
			oldestAccess = item.AccessCount
			oldestKey = key
		}
	}

	if oldestKey != "" {
		delete(cm.configCache, oldestKey)
		cm.stats.EvictionCount++
	}
}

// 清空所有缓存
func (cm *CacheManager) Clear() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.aiCache = make(map[string]*AIAnalysisCache)
	cm.ruleCache = make(map[string]*RuleMatchCache)
	cm.configCache = make(map[string]*CacheItem)
	cm.stats = CacheStats{}
}

// 获取统计信息
func (cm *CacheManager) GetStats() CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	cm.updateStats()
	return cm.stats
}

// 停止缓存管理器
func (cm *CacheManager) Stop() {
	close(cm.stopCleanup)
}

// 缓存管理命令处理函数

// 显示缓存统计信息
func handleCacheStats() {
	fmt.Println("📊 缓存统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := cacheManager.GetStats()
	fmt.Printf("总缓存项数: %d\n", stats.TotalItems)
	fmt.Printf("缓存命中次数: %d\n", stats.HitCount)
	fmt.Printf("缓存未命中次数: %d\n", stats.MissCount)
	fmt.Printf("缓存命中率: %.2f%%\n", stats.HitRate)
	fmt.Printf("内存使用量: %d 字节 (%.2f MB)\n", stats.MemoryUsage, float64(stats.MemoryUsage)/(1024*1024))
	fmt.Printf("清理次数: %d\n", stats.EvictionCount)
	fmt.Printf("过期项数: %d\n", stats.ExpiredItems)

	// 显示各类型缓存详情
	fmt.Println("\n缓存类型详情:")
	fmt.Printf("  AI分析缓存: %d 项\n", len(cacheManager.aiCache))
	fmt.Printf("  规则匹配缓存: %d 项\n", len(cacheManager.ruleCache))
	fmt.Printf("  配置缓存: %d 项\n", len(cacheManager.configCache))
}

// 清空所有缓存
func handleCacheClear() {
	fmt.Println("🗑️  清空所有缓存...")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	cacheManager.Clear()
	fmt.Println("✅ 所有缓存已清空")
}

// 测试缓存功能
func handleCacheTest() {
	fmt.Println("🧪 测试缓存功能...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 测试配置缓存
	testKey := "test_config"
	testValue := map[string]interface{}{
		"test":    "value",
		"number":  123,
		"enabled": true,
	}

	fmt.Println("1. 测试配置缓存...")
	cacheManager.SetConfig(testKey, testValue, 1*time.Minute)

	if cached, found := cacheManager.GetConfig(testKey); found {
		fmt.Printf("   ✅ 配置缓存测试成功: %v\n", cached)
	} else {
		fmt.Println("   ❌ 配置缓存测试失败")
	}

	// 测试AI分析缓存
	testLogHash := generateLogHash("test log line")
	aiResult := &AIAnalysisCache{
		LogHash:    testLogHash,
		Result:     "This is a test log",
		Confidence: 0.95,
		Model:      "gpt-4",
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	fmt.Println("2. 测试AI分析缓存...")
	cacheManager.SetAIAnalysis(testLogHash, aiResult)

	if cached, found := cacheManager.GetAIAnalysis(testLogHash); found {
		fmt.Printf("   ✅ AI分析缓存测试成功: %s\n", cached.Result)
	} else {
		fmt.Println("   ❌ AI分析缓存测试失败")
	}

	// 测试规则匹配缓存
	testRuleID := "test_rule"
	ruleResult := &RuleMatchCache{
		LogHash:   testLogHash,
		RuleID:    testRuleID,
		Matched:   true,
		Result:    &FilterResult{Action: "highlight", RuleID: testRuleID},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	fmt.Println("3. 测试规则匹配缓存...")
	cacheManager.SetRuleMatch(testLogHash, testRuleID, ruleResult)

	if cached, found := cacheManager.GetRuleMatch(testLogHash, testRuleID); found {
		fmt.Printf("   ✅ 规则匹配缓存测试成功: %s\n", cached.Result.Action)
	} else {
		fmt.Println("   ❌ 规则匹配缓存测试失败")
	}

	// 显示最终统计
	fmt.Println("\n最终缓存统计:")
	stats := cacheManager.GetStats()
	fmt.Printf("  总缓存项数: %d\n", stats.TotalItems)
	fmt.Printf("  缓存命中率: %.2f%%\n", stats.HitRate)
	fmt.Printf("  内存使用量: %.2f KB\n", float64(stats.MemoryUsage)/1024)

	fmt.Println("\n✅ 缓存功能测试完成")
}
