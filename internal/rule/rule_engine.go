package rule

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/xurenlu/aipipe/internal/config"
)

// 过滤结果
type FilterResult struct {
	Rule        config.FilterRule `json:"rule"`
	Matched     bool              `json:"matched"`
	Action      string            `json:"action"`
	Priority    int               `json:"priority"`
	Description string            `json:"description"`
	Color       string            `json:"color"`
	Timestamp   time.Time         `json:"timestamp"`
}

// 规则统计
type RuleStats struct {
	TotalRules     int                    `json:"total_rules"`
	EnabledRules   int                    `json:"enabled_rules"`
	DisabledRules  int                    `json:"disabled_rules"`
	MatchCounts    map[string]int64       `json:"match_counts"`
	ActionCounts   map[string]int64       `json:"action_counts"`
	CategoryCounts map[string]int64       `json:"category_counts"`
	LastUpdated    time.Time              `json:"last_updated"`
	Performance    map[string]interface{} `json:"performance"`
}

// 规则引擎
type RuleEngine struct {
	rules         []config.FilterRule
	compiledRules map[string]*regexp.Regexp
	stats         RuleStats
	mutex         sync.RWMutex
}

// 创建新的规则引擎
func NewRuleEngine(rules []config.FilterRule) *RuleEngine {
	re := &RuleEngine{
		rules:         rules,
		compiledRules: make(map[string]*regexp.Regexp),
		stats: RuleStats{
			MatchCounts:    make(map[string]int64),
			ActionCounts:   make(map[string]int64),
			CategoryCounts: make(map[string]int64),
			LastUpdated:    time.Now(),
			Performance:    make(map[string]interface{}),
		},
	}
	
	// 编译规则
	re.compileRules()
	re.updateStats()
	
	return re
}

// 编译所有规则
func (re *RuleEngine) compileRules() {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	re.compiledRules = make(map[string]*regexp.Regexp)
	
	for _, rule := range re.rules {
		if rule.Enabled && rule.Pattern != "" {
			if compiled, err := regexp.Compile(rule.Pattern); err == nil {
				re.compiledRules[rule.ID] = compiled
			}
		}
	}
}

// 过滤日志行
func (re *RuleEngine) Filter(line string) *FilterResult {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	
	// 按优先级排序规则
	sortedRules := make([]config.FilterRule, len(re.rules))
	copy(sortedRules, re.rules)
	sort.Slice(sortedRules, func(i, j int) bool {
		return sortedRules[i].Priority < sortedRules[j].Priority
	})
	
	// 遍历规则，找到第一个匹配的
	for _, rule := range sortedRules {
		if !rule.Enabled {
			continue
		}
		
		if compiled, exists := re.compiledRules[rule.ID]; exists {
			if compiled.MatchString(line) {
				// 更新统计
				re.updateStats()
				re.stats.MatchCounts[rule.ID]++
				re.stats.ActionCounts[rule.Action]++
				re.stats.CategoryCounts[rule.Category]++
				
				return re.createFilterResult(rule)
			}
		}
	}
	
	return nil
}

// 创建过滤结果
func (re *RuleEngine) createFilterResult(rule config.FilterRule) *FilterResult {
	return &FilterResult{
		Rule:        rule,
		Matched:     true,
		Action:      rule.Action,
		Priority:    rule.Priority,
		Description: rule.Description,
		Color:       rule.Color,
		Timestamp:   time.Now(),
	}
}

// 更新统计信息
func (re *RuleEngine) updateStats() {
	re.stats.TotalRules = len(re.rules)
	re.stats.EnabledRules = 0
	re.stats.DisabledRules = 0
	
	for _, rule := range re.rules {
		if rule.Enabled {
			re.stats.EnabledRules++
		} else {
			re.stats.DisabledRules++
		}
	}
	
	re.stats.LastUpdated = time.Now()
}

// 添加规则
func (re *RuleEngine) AddRule(rule config.FilterRule) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	// 检查规则ID是否已存在
	for _, existingRule := range re.rules {
		if existingRule.ID == rule.ID {
			return fmt.Errorf("规则ID已存在: %s", rule.ID)
		}
	}
	
	// 编译规则
	if rule.Pattern != "" {
		if compiled, err := regexp.Compile(rule.Pattern); err != nil {
			return fmt.Errorf("编译规则失败: %w", err)
		} else {
			re.compiledRules[rule.ID] = compiled
		}
	}
	
	// 添加规则
	re.rules = append(re.rules, rule)
	
	// 重新排序
	re.sortRules()
	re.updateStats()
	
	return nil
}

// 删除规则
func (re *RuleEngine) RemoveRule(ruleID string) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	for i, rule := range re.rules {
		if rule.ID == ruleID {
			// 删除规则
			re.rules = append(re.rules[:i], re.rules[i+1:]...)
			
			// 删除编译的规则
			delete(re.compiledRules, ruleID)
			
			// 更新统计
			re.updateStats()
			
			return nil
		}
	}
	
	return fmt.Errorf("未找到规则: %s", ruleID)
}

// 设置规则启用状态
func (re *RuleEngine) SetRuleEnabled(ruleID string, enabled bool) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	for i := range re.rules {
		if re.rules[i].ID == ruleID {
			re.rules[i].Enabled = enabled
			
			// 如果启用，编译规则
			if enabled && re.rules[i].Pattern != "" {
				if compiled, err := regexp.Compile(re.rules[i].Pattern); err == nil {
					re.compiledRules[ruleID] = compiled
				} else {
					return fmt.Errorf("编译规则失败: %w", err)
				}
			} else if !enabled {
				// 如果禁用，删除编译的规则
				delete(re.compiledRules, ruleID)
			}
			
			re.updateStats()
			return nil
		}
	}
	
	return fmt.Errorf("未找到规则: %s", ruleID)
}

// 排序规则
func (re *RuleEngine) sortRules() {
	sort.Slice(re.rules, func(i, j int) bool {
		return re.rules[i].Priority < re.rules[j].Priority
	})
}

// 获取所有规则
func (re *RuleEngine) GetRules() []config.FilterRule {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	
	rules := make([]config.FilterRule, len(re.rules))
	copy(rules, re.rules)
	return rules
}

// 获取统计信息
func (re *RuleEngine) GetStats() RuleStats {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	
	return re.stats
}

// 清除缓存
func (re *RuleEngine) ClearCache() {
	re.mutex.Lock()
	defer re.mutex.Unlock()
	
	re.compiledRules = make(map[string]*regexp.Regexp)
	re.compileRules()
}

// 测试规则
func (re *RuleEngine) TestRule(ruleID, testLine string) (bool, error) {
	re.mutex.RLock()
	defer re.mutex.RUnlock()
	
	if compiled, exists := re.compiledRules[ruleID]; exists {
		return compiled.MatchString(testLine), nil
	}
	
	return false, fmt.Errorf("规则未找到或未编译: %s", ruleID)
}
