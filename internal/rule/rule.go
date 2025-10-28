package rule

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 过滤规则
type FilterRule struct {
	ID          string `json:"id"`          // 规则ID
	Name        string `json:"name"`        // 规则名称
	Pattern     string `json:"pattern"`     // 正则表达式模式
	Action      string `json:"action"`      // 动作: filter, alert, ignore, highlight
	Priority    int    `json:"priority"`    // 优先级（数字越小优先级越高）
	Description string `json:"description"` // 规则描述
	Enabled     bool   `json:"enabled"`     // 是否启用
	Category    string `json:"category"`    // 规则分类
	Color       string `json:"color"`       // 高亮颜色
}

// 规则引擎
type RuleEngine struct {
	rules         []FilterRule
	compiledRules map[string]*regexp.Regexp
	cache         map[string]bool
	mutex         sync.RWMutex
	stats         RuleStats
}

// 规则统计
type RuleStats struct {
	TotalRules       int `json:"total_rules"`
	EnabledRules     int `json:"enabled_rules"`
	CacheHits        int `json:"cache_hits"`
	CacheMisses      int `json:"cache_misses"`
	FilteredLines    int `json:"filtered_lines"`
	AlertedLines     int `json:"alerted_lines"`
	IgnoredLines     int `json:"ignored_lines"`
	HighlightedLines int `json:"highlighted_lines"`
}

// 规则管理命令处理函数

// 列出所有过滤规则
func handleRuleList() {
	fmt.Println("📋 过滤规则列表:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	rules := ruleEngine.GetRules()
	if len(rules) == 0 {
		fmt.Println("没有配置过滤规则")
		return
	}

	for i, rule := range rules {
		status := "❌ 禁用"
		if rule.Enabled {
			status = "✅ 启用"
		}

		fmt.Printf("%d. %s %s\n", i+1, status, rule.Name)
		fmt.Printf("   ID: %s\n", rule.ID)
		fmt.Printf("   模式: %s\n", rule.Pattern)
		fmt.Printf("   动作: %s\n", rule.Action)
		fmt.Printf("   优先级: %d\n", rule.Priority)
		fmt.Printf("   分类: %s\n", rule.Category)
		if rule.Description != "" {
			fmt.Printf("   描述: %s\n", rule.Description)
		}
		if rule.Color != "" {
			fmt.Printf("   颜色: %s\n", rule.Color)
		}
		fmt.Println()
	}
}

// 测试规则
func handleRuleTest() {
	// 解析参数
	parts := strings.SplitN(*ruleTest, ",", 2)
	if len(parts) != 2 {
		fmt.Printf("❌ 参数格式错误，应为: rule_id,test_line\n")
		os.Exit(1)
	}

	ruleID := parts[0]
	testLine := parts[1]

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🧪 测试规则: %s\n", ruleID)
	fmt.Printf("测试行: %s\n", testLine)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	matched, err := ruleEngine.TestRule(ruleID, testLine)
	if err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
		os.Exit(1)
	}

	if matched {
		fmt.Printf("✅ 匹配成功\n")
	} else {
		fmt.Printf("❌ 不匹配\n")
	}
}

// 显示规则引擎统计信息
func handleRuleStats() {
	fmt.Println("📊 规则引擎统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := ruleEngine.GetStats()
	fmt.Printf("总规则数: %d\n", stats.TotalRules)
	fmt.Printf("启用规则数: %d\n", stats.EnabledRules)
	fmt.Printf("缓存命中: %d\n", stats.CacheHits)
	fmt.Printf("缓存未命中: %d\n", stats.CacheMisses)
	fmt.Printf("过滤行数: %d\n", stats.FilteredLines)
	fmt.Printf("告警行数: %d\n", stats.AlertedLines)
	fmt.Printf("忽略行数: %d\n", stats.IgnoredLines)
	fmt.Printf("高亮行数: %d\n", stats.HighlightedLines)

	// 计算缓存命中率
	totalCache := stats.CacheHits + stats.CacheMisses
	if totalCache > 0 {
		hitRate := float64(stats.CacheHits) / float64(totalCache) * 100
		fmt.Printf("缓存命中率: %.2f%%\n", hitRate)
	}
}

// 添加规则
func handleRuleAdd() {
	fmt.Println("➕ 添加过滤规则...")

	// 解析JSON
	var rule FilterRule
	if err := json.Unmarshal([]byte(*ruleAdd), &rule); err != nil {
		fmt.Printf("❌ JSON解析失败: %v\n", err)
		os.Exit(1)
	}

	// 验证必填字段
	if rule.ID == "" {
		fmt.Printf("❌ 规则ID不能为空\n")
		os.Exit(1)
	}
	if rule.Pattern == "" {
		fmt.Printf("❌ 规则模式不能为空\n")
		os.Exit(1)
	}
	if rule.Action == "" {
		fmt.Printf("❌ 规则动作不能为空\n")
		os.Exit(1)
	}

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 添加规则
	if err := ruleEngine.AddRule(rule); err != nil {
		fmt.Printf("❌ 添加规则失败: %v\n", err)
		os.Exit(1)
	}

	// 保存规则到配置文件
	if err := saveRulesToConfig(); err != nil {
		fmt.Printf("⚠️  规则添加成功，但保存到配置文件失败: %v\n", err)
	} else {
		fmt.Printf("✅ 规则 %s 添加并保存成功\n", rule.ID)
	}
}

// 删除规则
func handleRuleRemove() {
	ruleID := *ruleRemove

	fmt.Printf("🗑️  删除规则: %s\n", ruleID)

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 删除规则
	if err := ruleEngine.RemoveRule(ruleID); err != nil {
		fmt.Printf("❌ 删除规则失败: %v\n", err)
		os.Exit(1)
	}

	// 保存规则到配置文件
	if err := saveRulesToConfig(); err != nil {
		fmt.Printf("⚠️  规则删除成功，但保存到配置文件失败: %v\n", err)
	} else {
		fmt.Printf("✅ 规则 %s 删除并保存成功\n", ruleID)
	}
}

// 启用规则
func handleRuleEnable() {
	ruleID := *ruleEnable

	fmt.Printf("✅ 启用规则: %s\n", ruleID)

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 启用规则
	if err := ruleEngine.SetRuleEnabled(ruleID, true); err != nil {
		fmt.Printf("❌ 启用规则失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 规则 %s 启用成功\n", ruleID)
}

// 禁用规则
func handleRuleDisable() {
	ruleID := *ruleDisable

	fmt.Printf("❌ 禁用规则: %s\n", ruleID)

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 禁用规则
	if err := ruleEngine.SetRuleEnabled(ruleID, false); err != nil {
		fmt.Printf("❌ 禁用规则失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 规则 %s 禁用成功\n", ruleID)
}

// 获取规则匹配结果
func (cm *CacheManager) GetRuleMatch(logHash, ruleID string) (*RuleMatchCache, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	key := logHash + ":" + ruleID
	item, exists := cm.ruleCache[key]
	if !exists {
		cm.stats.MissCount++
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		cm.stats.MissCount++
		return nil, false
	}

	cm.stats.HitCount++
	return item, true
}

// 设置规则匹配结果
func (cm *CacheManager) SetRuleMatch(logHash, ruleID string, result *RuleMatchCache) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查是否需要清理空间
	if cm.needsEviction() {
		cm.evictOldest()
	}

	key := logHash + ":" + ruleID
	cm.ruleCache[key] = result
	cm.updateStats()
}
