package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
)

// 邮件配置
type EmailConfig struct {
	Enabled   bool     `json:"enabled"`
	Provider  string   `json:"provider"`   // "smtp" 或 "resend"
	Host      string   `json:"host"`       // SMTP服务器地址
	Port      int      `json:"port"`       // SMTP端口
	Username  string   `json:"username"`   // 用户名
	Password  string   `json:"password"`   // 密码或API密钥
	FromEmail string   `json:"from_email"` // 发件人邮箱
	ToEmails  []string `json:"to_emails"`  // 收件人邮箱列表
}

// Webhook配置
type WebhookConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Secret  string `json:"secret,omitempty"` // 可选的签名密钥
}

// 通知器配置
type NotifierConfig struct {
	Email          EmailConfig     `json:"email"`
	DingTalk       WebhookConfig   `json:"dingtalk"`
	WeChat         WebhookConfig   `json:"wechat"`
	Feishu         WebhookConfig   `json:"feishu"`
	Slack          WebhookConfig   `json:"slack"`
	CustomWebhooks []WebhookConfig `json:"custom_webhooks,omitempty"`
}

// AI 服务配置
type AIService struct {
	Name     string `json:"name"`     // 服务名称
	Endpoint string `json:"endpoint"` // API 端点
	Token    string `json:"token"`    // API Token
	Model    string `json:"model"`    // 模型名称
	Priority int    `json:"priority"` // 优先级（数字越小优先级越高）
	Enabled  bool   `json:"enabled"`  // 是否启用
}

// AI 服务管理器
type AIServiceManager struct {
	services    []AIService
	current     int
	fallback    bool
	rateLimiter map[string]time.Time
	mutex       sync.RWMutex
}

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

// 过滤结果
type FilterResult struct {
	Action          string `json:"action"`           // 动作
	RuleID          string `json:"rule_id"`          // 匹配的规则ID
	RuleName        string `json:"rule_name"`        // 规则名称
	Category        string `json:"category"`         // 分类
	Color           string `json:"color"`            // 颜色
	ShouldProcess   bool   `json:"should_process"`   // 是否应该处理
	ShouldAlert     bool   `json:"should_alert"`     // 是否应该告警
	ShouldIgnore    bool   `json:"should_ignore"`    // 是否应该忽略
	ShouldHighlight bool   `json:"should_highlight"` // 是否应该高亮
}

// 缓存项
type CacheItem struct {
	Key        string      `json:"key"`
	Value      interface{} `json:"value"`
	ExpiresAt  time.Time   `json:"expires_at"`
	CreatedAt  time.Time   `json:"created_at"`
	AccessCount int        `json:"access_count"`
	Size       int64       `json:"size"`
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
	LogHash   string    `json:"log_hash"`
	RuleID    string    `json:"rule_id"`
	Matched   bool      `json:"matched"`
	Result    *FilterResult `json:"result"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// 缓存统计
type CacheStats struct {
	TotalItems     int     `json:"total_items"`
	HitCount       int64   `json:"hit_count"`
	MissCount      int64   `json:"miss_count"`
	EvictionCount  int64   `json:"eviction_count"`
	MemoryUsage    int64   `json:"memory_usage"`
	HitRate        float64 `json:"hit_rate"`
	ExpiredItems   int     `json:"expired_items"`
}

// 缓存管理器
type CacheManager struct {
	aiCache      map[string]*AIAnalysisCache
	ruleCache    map[string]*RuleMatchCache
	configCache  map[string]*CacheItem
	stats        CacheStats
	mutex        sync.RWMutex
	maxSize      int64
	maxItems     int
	cleanupInterval time.Duration
	stopCleanup  chan bool
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

// 配置文件结构
type Config struct {
	AIEndpoint   string         `json:"ai_endpoint"` // 向后兼容
	Token        string         `json:"token"`       // 向后兼容
	Model        string         `json:"model"`       // 向后兼容
	CustomPrompt string         `json:"custom_prompt"`
	Notifiers    NotifierConfig `json:"notifiers"`

	// 新增配置项
	MaxRetries  int  `json:"max_retries"`  // API 重试次数
	Timeout     int  `json:"timeout"`      // 请求超时时间（秒）
	RateLimit   int  `json:"rate_limit"`   // 请求频率限制（每分钟）
	LocalFilter bool `json:"local_filter"` // 是否启用本地过滤

	// 多AI服务支持
	AIServices []AIService `json:"ai_services"` // AI 服务列表
	DefaultAI  string      `json:"default_ai"`  // 默认AI服务名称

	// 规则引擎配置
	Rules []FilterRule `json:"rules"` // 过滤规则列表
	
	// 缓存配置
	Cache CacheConfig `json:"cache"` // 缓存配置
}

// 错误级别
type ErrorLevel int

const (
	ErrorLevelInfo ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelError
	ErrorLevelCritical
)

// 错误分类
type ErrorCategory string

const (
	ErrorCategoryConfig     ErrorCategory = "config"
	ErrorCategoryNetwork    ErrorCategory = "network"
	ErrorCategoryAI         ErrorCategory = "ai"
	ErrorCategoryProcessing ErrorCategory = "processing"
	ErrorCategoryOutput     ErrorCategory = "output"
	ErrorCategoryFile       ErrorCategory = "file"
)

// AIPipe 错误结构
type AIPipeError struct {
	Code       string                 `json:"code"`
	Category   ErrorCategory          `json:"category"`
	Level      ErrorLevel             `json:"level"`
	Message    string                 `json:"message"`
	Suggestion string                 `json:"suggestion"`
	Context    map[string]interface{} `json:"context"`
	Timestamp  time.Time              `json:"timestamp"`
	StackTrace string                 `json:"stack_trace"`
}

func (e *AIPipeError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Category, e.Code, e.Message)
}

// 配置验证错误
type ConfigValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

func (e *ConfigValidationError) Error() string {
	return fmt.Sprintf("配置验证失败 [%s]: %s (当前值: %s)", e.Field, e.Message, e.Value)
}

// 错误恢复策略
type ErrorRecovery struct {
	strategies map[ErrorCategory]RecoveryStrategy
	maxRetries int
	backoff    time.Duration
}

type RecoveryStrategy interface {
	CanRecover(err error) bool
	Recover(err error) error
}

// 网络错误恢复策略
type NetworkErrorRecovery struct {
	maxRetries int
	backoff    time.Duration
}

func (ner *NetworkErrorRecovery) CanRecover(err error) bool {
	// 检查是否是网络相关错误
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "network")
}

func (ner *NetworkErrorRecovery) Recover(err error) error {
	// 实现网络错误恢复逻辑
	time.Sleep(ner.backoff)
	return nil
}

// 配置错误恢复策略
type ConfigErrorRecovery struct {
	fallbackConfig *Config
	validator      *ConfigValidator
}

func (cer *ConfigErrorRecovery) CanRecover(err error) bool {
	return strings.Contains(err.Error(), "config") || strings.Contains(err.Error(), "配置文件")
}

func (cer *ConfigErrorRecovery) Recover(err error) error {
	// 使用默认配置
	globalConfig = *cer.fallbackConfig
	return nil
}

// 错误处理器
type ErrorHandler struct {
	recovery *ErrorRecovery
	logger   *log.Logger
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		recovery: &ErrorRecovery{
			strategies: make(map[ErrorCategory]RecoveryStrategy),
			maxRetries: 3,
			backoff:    time.Second * 2,
		},
		logger: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}

func (eh *ErrorHandler) RegisterStrategy(category ErrorCategory, strategy RecoveryStrategy) {
	eh.recovery.strategies[category] = strategy
}

func (eh *ErrorHandler) Handle(err error, context map[string]interface{}) error {
	// 创建 AIPipe 错误
	aipipeErr := &AIPipeError{
		Code:       "UNKNOWN_ERROR",
		Category:   ErrorCategoryProcessing,
		Level:      ErrorLevelError,
		Message:    err.Error(),
		Context:    context,
		Timestamp:  time.Now(),
		StackTrace: getStackTrace(),
	}

	// 根据错误类型设置分类和级别
	eh.classifyError(aipipeErr)

	// 记录错误
	eh.logError(aipipeErr)

	// 尝试恢复
	if strategy, exists := eh.recovery.strategies[aipipeErr.Category]; exists {
		if strategy.CanRecover(err) {
			if recoverErr := strategy.Recover(err); recoverErr == nil {
				if eh.logger != nil {
					eh.logger.Printf("错误已恢复: %s", aipipeErr.Message)
				}
				return nil
			}
		}
	}

	return aipipeErr
}

func (eh *ErrorHandler) classifyError(err *AIPipeError) {
	errMsg := strings.ToLower(err.Message)

	// 网络错误
	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "connection") {
		err.Category = ErrorCategoryNetwork
		err.Level = ErrorLevelWarning
		err.Code = "NETWORK_ERROR"
		err.Suggestion = "检查网络连接和服务器状态"
	}

	// AI 服务错误
	if strings.Contains(errMsg, "api") || strings.Contains(errMsg, "ai") {
		err.Category = ErrorCategoryAI
		err.Level = ErrorLevelError
		err.Code = "AI_SERVICE_ERROR"
		err.Suggestion = "检查 AI 服务配置和 Token 有效性"
	}

	// 配置错误
	if strings.Contains(errMsg, "config") || strings.Contains(errMsg, "配置文件") {
		err.Category = ErrorCategoryConfig
		err.Level = ErrorLevelCritical
		err.Code = "CONFIG_ERROR"
		err.Suggestion = "检查配置文件格式和内容"
	}

	// 文件错误
	if strings.Contains(errMsg, "file") || strings.Contains(errMsg, "文件") {
		err.Category = ErrorCategoryFile
		err.Level = ErrorLevelError
		err.Code = "FILE_ERROR"
		err.Suggestion = "检查文件路径和权限"
	}
}

func (eh *ErrorHandler) logError(err *AIPipeError) {
	if eh.logger == nil {
		return // 如果 logger 为 nil，不输出日志
	}

	levelStr := []string{"INFO", "WARNING", "ERROR", "CRITICAL"}[err.Level]
	eh.logger.Printf("[%s] %s: %s", levelStr, err.Category, err.Message)

	if err.Suggestion != "" {
		eh.logger.Printf("建议: %s", err.Suggestion)
	}

	if *debug {
		eh.logger.Printf("上下文: %+v", err.Context)
		eh.logger.Printf("堆栈跟踪: %s", err.StackTrace)
	}
}

func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// AI 服务管理器方法

// 创建新的AI服务管理器
func NewAIServiceManager(services []AIService) *AIServiceManager {
	// 按优先级排序
	sortedServices := make([]AIService, len(services))
	copy(sortedServices, services)

	// 简单的冒泡排序按优先级排序
	for i := 0; i < len(sortedServices)-1; i++ {
		for j := 0; j < len(sortedServices)-i-1; j++ {
			if sortedServices[j].Priority > sortedServices[j+1].Priority {
				sortedServices[j], sortedServices[j+1] = sortedServices[j+1], sortedServices[j]
			}
		}
	}

	return &AIServiceManager{
		services:    sortedServices,
		current:     0,
		fallback:    false,
		rateLimiter: make(map[string]time.Time),
	}
}

// 获取下一个可用的AI服务
func (asm *AIServiceManager) GetNextService() (*AIService, error) {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()

	// 查找启用的服务
	for i := 0; i < len(asm.services); i++ {
		service := &asm.services[asm.current]
		if service.Enabled {
			// 检查频率限制
			if asm.isRateLimited(service.Name) {
				asm.current = (asm.current + 1) % len(asm.services)
				continue
			}

			// 更新当前索引
			asm.current = (asm.current + 1) % len(asm.services)
			return service, nil
		}
		asm.current = (asm.current + 1) % len(asm.services)
	}

	return nil, fmt.Errorf("没有可用的AI服务")
}

// 检查服务是否被频率限制
func (asm *AIServiceManager) isRateLimited(serviceName string) bool {
	if lastCall, exists := asm.rateLimiter[serviceName]; exists {
		// 检查是否在限制时间内
		if time.Since(lastCall) < time.Minute/time.Duration(globalConfig.RateLimit) {
			return true
		}
	}
	return false
}

// 记录服务调用时间
func (asm *AIServiceManager) RecordCall(serviceName string) {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	asm.rateLimiter[serviceName] = time.Now()
}

// 获取服务统计信息
func (asm *AIServiceManager) GetStats() map[string]interface{} {
	asm.mutex.RLock()
	defer asm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_services":   len(asm.services),
		"enabled_services": 0,
		"current_index":    asm.current,
		"fallback_mode":    asm.fallback,
	}

	for _, service := range asm.services {
		if service.Enabled {
			stats["enabled_services"] = stats["enabled_services"].(int) + 1
		}
	}

	return stats
}

// 启用/禁用服务
func (asm *AIServiceManager) SetServiceEnabled(serviceName string, enabled bool) error {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()

	for i := range asm.services {
		if asm.services[i].Name == serviceName {
			asm.services[i].Enabled = enabled
			return nil
		}
	}

	return fmt.Errorf("服务 %s 不存在", serviceName)
}

// 获取服务列表
func (asm *AIServiceManager) GetServices() []AIService {
	asm.mutex.RLock()
	defer asm.mutex.RUnlock()

	services := make([]AIService, len(asm.services))
	copy(services, asm.services)
	return services
}

// 规则引擎方法

// 创建新的规则引擎
func NewRuleEngine(rules []FilterRule) *RuleEngine {
	// 按优先级排序规则
	sortedRules := make([]FilterRule, len(rules))
	copy(sortedRules, rules)

	// 简单的冒泡排序按优先级排序
	for i := 0; i < len(sortedRules)-1; i++ {
		for j := 0; j < len(sortedRules)-i-1; j++ {
			if sortedRules[j].Priority > sortedRules[j+1].Priority {
				sortedRules[j], sortedRules[j+1] = sortedRules[j+1], sortedRules[j]
			}
		}
	}

	// 编译正则表达式
	compiledRules := make(map[string]*regexp.Regexp)
	for _, rule := range sortedRules {
		if rule.Enabled && rule.Pattern != "" {
			if compiled, err := regexp.Compile(rule.Pattern); err == nil {
				compiledRules[rule.ID] = compiled
			}
		}
	}

	// 统计启用的规则
	enabledCount := 0
	for _, rule := range sortedRules {
		if rule.Enabled {
			enabledCount++
		}
	}

	return &RuleEngine{
		rules:         sortedRules,
		compiledRules: compiledRules,
		cache:         make(map[string]bool),
		stats: RuleStats{
			TotalRules:   len(sortedRules),
			EnabledRules: enabledCount,
		},
	}
}

// 过滤日志行
func (re *RuleEngine) Filter(line string) *FilterResult {
	re.mutex.RLock()
	defer re.mutex.RUnlock()

	// 检查缓存
	if cached, exists := re.cache[line]; exists {
		re.stats.CacheHits++
		if cached {
			return &FilterResult{
				Action:       "ignore",
				ShouldIgnore: true,
			}
		}
	} else {
		re.stats.CacheMisses++
	}

	// 遍历规则（按优先级顺序）
	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}

		// 检查是否匹配
		if compiled, exists := re.compiledRules[rule.ID]; exists {
			if compiled.MatchString(line) {
				// 更新统计
				re.updateStats(rule.Action)

				// 缓存结果
				re.cache[line] = (rule.Action == "ignore")

				return re.createFilterResult(rule)
			}
		}
	}

	// 没有匹配的规则，默认处理
	return &FilterResult{
		Action:        "process",
		ShouldProcess: true,
	}
}

// 更新统计信息
func (re *RuleEngine) updateStats(action string) {
	switch action {
	case "filter":
		re.stats.FilteredLines++
	case "alert":
		re.stats.AlertedLines++
	case "ignore":
		re.stats.IgnoredLines++
	case "highlight":
		re.stats.HighlightedLines++
	}
}

// 创建过滤结果
func (re *RuleEngine) createFilterResult(rule FilterRule) *FilterResult {
	result := &FilterResult{
		Action:   rule.Action,
		RuleID:   rule.ID,
		RuleName: rule.Name,
		Category: rule.Category,
		Color:    rule.Color,
	}

	// 设置动作标志
	switch rule.Action {
	case "filter":
		result.ShouldProcess = false
	case "alert":
		result.ShouldProcess = true
		result.ShouldAlert = true
	case "ignore":
		result.ShouldIgnore = true
	case "highlight":
		result.ShouldProcess = true
		result.ShouldHighlight = true
	default:
		result.ShouldProcess = true
	}

	return result
}

// 添加规则
func (re *RuleEngine) AddRule(rule FilterRule) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	// 检查ID是否已存在
	for _, existingRule := range re.rules {
		if existingRule.ID == rule.ID {
			return fmt.Errorf("规则ID %s 已存在", rule.ID)
		}
	}

	// 编译正则表达式
	if rule.Pattern != "" {
		compiled, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("正则表达式编译失败: %w", err)
		}
		re.compiledRules[rule.ID] = compiled
	}

	// 添加到规则列表
	re.rules = append(re.rules, rule)

	// 重新排序
	re.sortRules()

	// 更新统计
	re.stats.TotalRules++
	if rule.Enabled {
		re.stats.EnabledRules++
	}

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

			// 删除编译的正则表达式
			delete(re.compiledRules, ruleID)

			// 更新统计
			re.stats.TotalRules--
			if rule.Enabled {
				re.stats.EnabledRules--
			}

			return nil
		}
	}

	return fmt.Errorf("规则ID %s 不存在", ruleID)
}

// 启用/禁用规则
func (re *RuleEngine) SetRuleEnabled(ruleID string, enabled bool) error {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	for i, rule := range re.rules {
		if rule.ID == ruleID {
			oldEnabled := rule.Enabled
			re.rules[i].Enabled = enabled

			// 更新统计
			if oldEnabled && !enabled {
				re.stats.EnabledRules--
			} else if !oldEnabled && enabled {
				re.stats.EnabledRules++
			}

			return nil
		}
	}

	return fmt.Errorf("规则ID %s 不存在", ruleID)
}

// 排序规则
func (re *RuleEngine) sortRules() {
	for i := 0; i < len(re.rules)-1; i++ {
		for j := 0; j < len(re.rules)-i-1; j++ {
			if re.rules[j].Priority > re.rules[j+1].Priority {
				re.rules[j], re.rules[j+1] = re.rules[j+1], re.rules[j]
			}
		}
	}
}

// 获取规则列表
func (re *RuleEngine) GetRules() []FilterRule {
	re.mutex.RLock()
	defer re.mutex.RUnlock()

	rules := make([]FilterRule, len(re.rules))
	copy(rules, re.rules)
	return rules
}

// 获取统计信息
func (re *RuleEngine) GetStats() RuleStats {
	re.mutex.RLock()
	defer re.mutex.RUnlock()

	return re.stats
}

// 清空缓存
func (re *RuleEngine) ClearCache() {
	re.mutex.Lock()
	defer re.mutex.Unlock()

	re.cache = make(map[string]bool)
	re.stats.CacheHits = 0
	re.stats.CacheMisses = 0
}

// 测试规则
func (re *RuleEngine) TestRule(ruleID, testLine string) (bool, error) {
	re.mutex.RLock()
	defer re.mutex.RUnlock()

	compiled, exists := re.compiledRules[ruleID]
	if !exists {
		return false, fmt.Errorf("规则ID %s 不存在或未编译", ruleID)
	}

	return compiled.MatchString(testLine), nil
}

// 配置验证器
type ConfigValidator struct {
	errors []ConfigValidationError
}

func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make([]ConfigValidationError, 0),
	}
}

func (cv *ConfigValidator) Validate(config *Config) error {
	cv.errors = cv.errors[:0] // 清空之前的错误

	// 验证必填字段
	cv.validateRequired("ai_endpoint", config.AIEndpoint)
	cv.validateRequired("token", config.Token)
	cv.validateRequired("model", config.Model)

	// 验证 URL 格式
	cv.validateURL("ai_endpoint", config.AIEndpoint)

	// 验证数值范围
	cv.validateRange("max_retries", config.MaxRetries, 0, 10)
	cv.validateRange("timeout", config.Timeout, 5, 300)
	cv.validateRange("rate_limit", config.RateLimit, 1, 1000)

	// 验证 Token 长度
	cv.validateMinLength("token", config.Token, 10)

	if len(cv.errors) > 0 {
		return fmt.Errorf("配置验证失败，发现 %d 个错误", len(cv.errors))
	}

	return nil
}

func (cv *ConfigValidator) validateRequired(field, value string) {
	if strings.TrimSpace(value) == "" {
		cv.errors = append(cv.errors, ConfigValidationError{
			Field:   field,
			Message: "此字段为必填项",
			Value:   value,
		})
	}
}

func (cv *ConfigValidator) validateURL(field, value string) {
	if value == "" {
		return
	}

	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		cv.errors = append(cv.errors, ConfigValidationError{
			Field:   field,
			Message: "必须是有效的 URL 格式",
			Value:   value,
		})
	}
}

func (cv *ConfigValidator) validateRange(field string, value, min, max int) {
	if value < min || value > max {
		cv.errors = append(cv.errors, ConfigValidationError{
			Field:   field,
			Message: fmt.Sprintf("值必须在 %d 到 %d 之间", min, max),
			Value:   fmt.Sprintf("%d", value),
		})
	}
}

func (cv *ConfigValidator) validateMinLength(field, value string, minLen int) {
	if len(value) < minLen {
		cv.errors = append(cv.errors, ConfigValidationError{
			Field:   field,
			Message: fmt.Sprintf("长度至少为 %d 个字符", minLen),
			Value:   fmt.Sprintf("%d", len(value)),
		})
	}
}

func (cv *ConfigValidator) GetErrors() []ConfigValidationError {
	return cv.errors
}

// 默认配置
var defaultConfig = Config{
	AIEndpoint:   "https://your-ai-server.com/api/v1/chat/completions",
	Token:        "your-api-token-here",
	Model:        "gpt-4",
	CustomPrompt: "",
	MaxRetries:   3,
	Timeout:      30,
	RateLimit:    100,
	LocalFilter:  true,
	Notifiers: NotifierConfig{
		Email: EmailConfig{
			Enabled:   false,
			Provider:  "smtp",
			Host:      "smtp.gmail.com",
			Port:      587,
			Username:  "",
			Password:  "",
			FromEmail: "",
			ToEmails:  []string{},
		},
		DingTalk: WebhookConfig{
			Enabled: false,
			URL:     "",
		},
		WeChat: WebhookConfig{
			Enabled: false,
			URL:     "",
		},
		Feishu: WebhookConfig{
			Enabled: false,
			URL:     "",
		},
		Slack: WebhookConfig{
			Enabled: false,
			URL:     "",
		},
		CustomWebhooks: []WebhookConfig{},
	},
	Cache: CacheConfig{
		MaxSize:         100 * 1024 * 1024, // 100MB
		MaxItems:        1000,
		DefaultTTL:      1 * time.Hour,
		AITTL:           24 * time.Hour,
		RuleTTL:         1 * time.Hour,
		ConfigTTL:       30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
		Enabled:         true,
	},
}

// 全局配置变量
var globalConfig Config

// 全局错误处理器
var errorHandler *ErrorHandler

// 全局AI服务管理器
var aiServiceManager *AIServiceManager

// 全局规则引擎
var ruleEngine *RuleEngine

// 全局缓存管理器
var cacheManager *CacheManager

// 批处理配置
const (
	BATCH_MAX_SIZE  = 10              // 批处理最大行数
	BATCH_WAIT_TIME = 3 * time.Second // 批处理等待时间
)

// OpenAI API 请求和响应结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// 日志分析结果（单条）
type LogAnalysis struct {
	ShouldFilter bool   `json:"should_filter"`
	Summary      string `json:"summary"`
	Reason       string `json:"reason"`
}

// 批量日志分析结果
type BatchLogAnalysis struct {
	Results        []LogAnalysis `json:"results"`         // 每行日志的分析结果
	OverallSummary string        `json:"overall_summary"` // 整体摘要
	ImportantCount int           `json:"important_count"` // 重要日志数量
}

// 日志批处理器
type LogBatcher struct {
	lines     []string
	timer     *time.Timer
	mu        sync.Mutex
	processor func([]string)
}

// 文件状态（用于记住读取位置）
type FileState struct {
	Path   string    `json:"path"`
	Offset int64     `json:"offset"`
	Inode  uint64    `json:"inode"`
	Time   time.Time `json:"time"`
}

// 日志行合并器（用于合并多行日志，如 Java 堆栈跟踪）
type LogLineMerger struct {
	format      string
	buffer      string
	hasBuffered bool
}

// 多源监控配置
type MultiSourceConfig struct {
	Sources []SourceConfig `json:"sources"`
}

type SourceConfig struct {
	Name        string         `json:"name"`        // 源名称
	Type        string         `json:"type"`        // 源类型: file, journalctl, stdin
	Path        string         `json:"path"`        // 文件路径（type=file时）
	Format      string         `json:"format"`      // 日志格式
	Journal     *JournalConfig `json:"journal"`     // journalctl配置（type=journalctl时）
	Enabled     bool           `json:"enabled"`     // 是否启用
	Priority    int            `json:"priority"`    // 优先级（数字越小优先级越高）
	Description string         `json:"description"` // 描述
}

type JournalConfig struct {
	Services []string `json:"services"` // 监控的服务
	Priority string   `json:"priority"` // 日志级别
	Since    string   `json:"since"`    // 开始时间
	Until    string   `json:"until"`    // 结束时间
	User     string   `json:"user"`     // 用户过滤
	Boot     bool     `json:"boot"`     // 当前启动
	Kernel   bool     `json:"kernel"`   // 内核消息
}

var (
	logFormat        = flag.String("format", "java", "日志格式 (java, php, nginx, ruby, fastapi, python, go, rust, csharp, kotlin, nodejs, typescript, docker, kubernetes, postgresql, mysql, redis, elasticsearch, git, jenkins, github, journald, macos-console, syslog)")
	verbose          = flag.Bool("verbose", false, "显示详细输出")
	filePath         = flag.String("f", "", "要监控的日志文件路径（类似 tail -f）")
	debug            = flag.Bool("debug", false, "调试模式，打印 HTTP 请求和响应详情")
	noBatch          = flag.Bool("no-batch", false, "禁用批处理，逐行分析（增加 API 调用）")
	batchSize        = flag.Int("batch-size", BATCH_MAX_SIZE, "批处理最大行数")
	batchWait        = flag.Duration("batch-wait", BATCH_WAIT_TIME, "批处理等待时间")
	showNotImportant = flag.Bool("show-not-important", false, "显示被过滤的日志（默认不显示）")
	contextLines     = flag.Int("context", 3, "重要日志显示的上下文行数（前后各N行）")

	// 新增配置管理命令
	configTest     = flag.Bool("config-test", false, "测试配置文件")
	configValidate = flag.Bool("config-validate", false, "验证配置文件")
	configShow     = flag.Bool("config-show", false, "显示当前配置")

	// AI服务管理命令
	aiList  = flag.Bool("ai-list", false, "列出所有AI服务")
	aiTest  = flag.Bool("ai-test", false, "测试所有AI服务")
	aiStats = flag.Bool("ai-stats", false, "显示AI服务统计信息")

	// 规则管理命令
	ruleList    = flag.Bool("rule-list", false, "列出所有过滤规则")
	ruleTest    = flag.String("rule-test", "", "测试规则 (格式: rule_id,test_line)")
	ruleStats   = flag.Bool("rule-stats", false, "显示规则引擎统计信息")
	ruleAdd     = flag.String("rule-add", "", "添加规则 (JSON格式)")
	ruleRemove  = flag.String("rule-remove", "", "删除规则 (规则ID)")
	ruleEnable  = flag.String("rule-enable", "", "启用规则 (规则ID)")
	ruleDisable = flag.String("rule-disable", "", "禁用规则 (规则ID)")
	
	// 缓存管理命令
	cacheStats      = flag.Bool("cache-stats", false, "显示缓存统计信息")
	cacheClear      = flag.Bool("cache-clear", false, "清空所有缓存")
	cacheTest       = flag.Bool("cache-test", false, "测试缓存功能")

	// journalctl 特定配置
	journalServices = flag.String("journal-services", "", "监控的systemd服务列表，逗号分隔 (如: nginx,docker,postgresql)")
	journalPriority = flag.String("journal-priority", "", "监控的日志级别 (emerg,alert,crit,err,warning,notice,info,debug)")
	journalSince    = flag.String("journal-since", "", "监控开始时间 (如: '1 hour ago', '2023-10-17 10:00:00')")
	journalUntil    = flag.String("journal-until", "", "监控结束时间 (如: 'now', '2023-10-17 18:00:00')")
	journalUser     = flag.String("journal-user", "", "监控特定用户的日志")
	journalBoot     = flag.Bool("journal-boot", false, "只监控当前启动的日志")
	journalKernel   = flag.Bool("journal-kernel", false, "只监控内核消息")

	// 多源监控配置
	multiSource = flag.String("multi-source", "", "多源监控配置文件路径")
	configFile  = flag.String("config", "", "指定配置文件路径")

	// 全局变量：当前监控的日志文件路径（用于通知）
	currentLogFile = "stdin"
)

// 构建journalctl命令
func buildJournalctlCommand() []string {
	args := []string{"journalctl", "-f", "--no-pager"}

	// 添加服务过滤
	if *journalServices != "" {
		services := strings.Split(*journalServices, ",")
		for _, service := range services {
			service = strings.TrimSpace(service)
			if service != "" {
				args = append(args, "-u", service)
			}
		}
	}

	// 添加优先级过滤
	if *journalPriority != "" {
		args = append(args, "-p", *journalPriority)
	}

	// 添加时间范围
	if *journalSince != "" {
		args = append(args, "--since", *journalSince)
	}
	if *journalUntil != "" {
		args = append(args, "--until", *journalUntil)
	}

	// 添加用户过滤
	if *journalUser != "" {
		args = append(args, "_UID="+*journalUser)
	}

	// 添加启动过滤
	if *journalBoot {
		args = append(args, "-b")
	}

	// 添加内核过滤
	if *journalKernel {
		args = append(args, "-k")
	}

	return args
}

// 检查是否应该使用多源监控
func shouldUseMultiSource() bool {
	// 如果指定了多源配置文件，使用多源监控
	if *multiSource != "" {
		return true
	}

	// 检查是否存在多源配置文件
	configPath, err := findMultiSourceConfig()
	if err != nil {
		return false
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); err == nil {
		if *verbose {
			log.Printf("🔍 自动检测到多源配置文件: %s", configPath)
		}
		return true
	}

	return false
}

func main() {
	flag.Parse()

	// 初始化错误处理器
	errorHandler = NewErrorHandler()
	errorHandler.RegisterStrategy(ErrorCategoryNetwork, &NetworkErrorRecovery{
		maxRetries: 3,
		backoff:    time.Second * 2,
	})
	errorHandler.RegisterStrategy(ErrorCategoryConfig, &ConfigErrorRecovery{
		fallbackConfig: &defaultConfig,
		validator:      NewConfigValidator(),
	})

	// 处理配置管理命令
	if *configTest {
		handleConfigTest()
		return
	}

	if *configValidate {
		handleConfigValidate()
		return
	}

	if *configShow {
		handleConfigShow()
		return
	}

	if *aiList {
		handleAIList()
		return
	}

	if *aiTest {
		handleAITest()
		return
	}

	if *aiStats {
		handleAIStats()
		return
	}

	if *ruleList {
		handleRuleList()
		return
	}

	if *ruleTest != "" {
		handleRuleTest()
		return
	}

	if *ruleStats {
		handleRuleStats()
		return
	}

	if *ruleAdd != "" {
		handleRuleAdd()
		return
	}

	if *ruleRemove != "" {
		handleRuleRemove()
		return
	}

	if *ruleEnable != "" {
		handleRuleEnable()
		return
	}

	if *ruleDisable != "" {
		handleRuleDisable()
		return
	}
	
	if *cacheStats {
		handleCacheStats()
		return
	}
	
	if *cacheClear {
		handleCacheClear()
		return
	}
	
	if *cacheTest {
		handleCacheTest()
		return
	}

	// 加载配置文件
	if err := loadConfig(); err != nil {
		if handledErr := errorHandler.Handle(err, map[string]interface{}{
			"operation":   "load_config",
			"config_path": "~/.config/aipipe.json",
		}); handledErr != nil {
			log.Printf("⚠️  加载配置文件失败，使用默认配置: %v", err)
			globalConfig = defaultConfig
		}
	}

	fmt.Printf("🚀 AIPipe 启动 - 监控 %s 格式日志\n", *logFormat)

	// 显示模式提示
	if !*showNotImportant {
		fmt.Println("💡 只显示重要日志（过滤的日志不显示）")
		if !*verbose {
			fmt.Println("   使用 --show-not-important 显示所有日志")
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if *filePath != "" {
		// 文件监控模式
		fmt.Printf("📁 监控文件: %s\n", *filePath)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if err := watchFile(*filePath); err != nil {
			log.Fatalf("❌ 监控文件失败: %v", err)
		}
	} else {
		// 标准输入模式
		fmt.Println("📥 从标准输入读取日志...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		processStdin()
	}
}

// 配置管理命令处理函数

// 测试配置文件
func handleConfigTest() {
	fmt.Println("🧪 测试配置文件...")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 测试 AI 服务连接
	fmt.Println("🔗 测试 AI 服务连接...")
	if err := testAIConnection(); err != nil {
		fmt.Printf("❌ AI 服务连接失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 配置文件测试通过！")
}

// 验证配置文件
func handleConfigValidate() {
	fmt.Println("🔍 验证配置文件...")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置验证失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 配置文件验证通过！")
}

// 显示当前配置
func handleConfigShow() {
	fmt.Println("📋 当前配置:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 显示配置信息（隐藏敏感信息）
	fmt.Printf("AI 端点: %s\n", globalConfig.AIEndpoint)
	fmt.Printf("模型: %s\n", globalConfig.Model)
	fmt.Printf("Token: %s...%s\n", globalConfig.Token[:min(8, len(globalConfig.Token))], globalConfig.Token[max(0, len(globalConfig.Token)-8):])
	fmt.Printf("最大重试次数: %d\n", globalConfig.MaxRetries)
	fmt.Printf("超时时间: %d 秒\n", globalConfig.Timeout)
	fmt.Printf("频率限制: %d 次/分钟\n", globalConfig.RateLimit)
	fmt.Printf("本地过滤: %t\n", globalConfig.LocalFilter)

	if globalConfig.CustomPrompt != "" {
		fmt.Printf("自定义提示词: %s\n", globalConfig.CustomPrompt)
	}
}

// AI服务管理命令处理函数

// 列出所有AI服务
func handleAIList() {
	fmt.Println("🤖 AI 服务列表:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	services := aiServiceManager.GetServices()
	if len(services) == 0 {
		fmt.Println("没有配置AI服务")
		return
	}

	for i, service := range services {
		status := "❌ 禁用"
		if service.Enabled {
			status = "✅ 启用"
		}

		fmt.Printf("%d. %s %s\n", i+1, status, service.Name)
		fmt.Printf("   端点: %s\n", service.Endpoint)
		fmt.Printf("   模型: %s\n", service.Model)
		fmt.Printf("   Token: %s...%s\n", service.Token[:min(8, len(service.Token))], service.Token[max(0, len(service.Token)-8):])
		fmt.Printf("   优先级: %d\n", service.Priority)
		fmt.Println()
	}
}

// 测试所有AI服务
func handleAITest() {
	fmt.Println("🧪 测试所有AI服务...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	services := aiServiceManager.GetServices()
	if len(services) == 0 {
		fmt.Println("没有配置AI服务")
		return
	}

	successCount := 0
	for _, service := range services {
		if !service.Enabled {
			fmt.Printf("⏭️  跳过禁用的服务: %s\n", service.Name)
			continue
		}

		fmt.Printf("🔗 测试服务: %s...", service.Name)

		// 创建测试请求
		testPrompt := "请回复 'OK' 表示连接正常"
		reqBody := ChatRequest{
			Model: service.Model,
			Messages: []ChatMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			fmt.Printf(" ❌ 构建请求失败\n")
			continue
		}

		// 创建HTTP请求
		req, err := http.NewRequest("POST", service.Endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf(" ❌ 创建请求失败\n")
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-key", service.Token)

		// 发送请求
		client := &http.Client{
			Timeout: time.Duration(globalConfig.Timeout) * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf(" ❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf(" ❌ API错误 %d: %s\n", resp.StatusCode, string(body))
			continue
		}

		fmt.Printf(" ✅ 成功\n")
		successCount++
	}

	fmt.Printf("\n📊 测试结果: %d/%d 服务可用\n", successCount, len(services))
	if successCount == 0 {
		os.Exit(1)
	}
}

// 显示AI服务统计信息
func handleAIStats() {
	fmt.Println("📊 AI 服务统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := aiServiceManager.GetStats()
	fmt.Printf("总服务数: %d\n", stats["total_services"])
	fmt.Printf("启用服务数: %d\n", stats["enabled_services"])
	fmt.Printf("当前索引: %d\n", stats["current_index"])
	fmt.Printf("故障转移模式: %t\n", stats["fallback_mode"])

	// 显示服务详情
	services := aiServiceManager.GetServices()
	if len(services) > 0 {
		fmt.Println("\n服务详情:")
		for _, service := range services {
			status := "❌ 禁用"
			if service.Enabled {
				status = "✅ 启用"
			}
			fmt.Printf("  %s %s (优先级: %d)\n", status, service.Name, service.Priority)
		}
	}
}

// 测试 AI 服务连接
func testAIConnection() error {
	// 创建一个简单的测试请求
	testPrompt := "请回复 'OK' 表示连接正常"

	// 构建请求
	reqBody := ChatRequest{
		Model: globalConfig.Model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: testPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", globalConfig.AIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", globalConfig.Token)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(globalConfig.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 加载配置文件
// 自动检测默认配置文件
func findDefaultConfig() (string, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")

	// 按优先级顺序检查配置文件
	configFiles := []string{
		"aipipe.json",
		"aipipe.yaml",
		"aipipe.yml",
		"aipipe.toml",
	}

	for _, filename := range configFiles {
		configPath := filepath.Join(configDir, filename)
		if _, err := os.Stat(configPath); err == nil {
			if *verbose {
				log.Printf("🔍 找到默认配置文件: %s", configPath)
			}
			return configPath, nil
		}
	}

	// 没有找到任何配置文件，返回默认路径
	defaultPath := filepath.Join(configDir, "aipipe.json")
	return defaultPath, nil
}

func loadConfig() error {
	var configPath string
	var err error

	// 如果指定了配置文件路径，使用指定的路径
	if *configFile != "" {
		configPath = *configFile
	} else {
		// 否则查找默认配置文件
		configPath, err = findDefaultConfig()
		if err != nil {
			return fmt.Errorf("查找默认配置文件失败: %v", err)
		}
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置文件
		return createDefaultConfig(configPath)
	}

	// 使用多格式加载
	if err := loadConfigWithFormat(configPath); err != nil {
		return err
	}

	// 验证必要的配置项
	if globalConfig.AIEndpoint == "" {
		globalConfig.AIEndpoint = defaultConfig.AIEndpoint
	}
	if globalConfig.Token == "" {
		globalConfig.Token = defaultConfig.Token
	}
	if globalConfig.Model == "" {
		globalConfig.Model = defaultConfig.Model
	}

	// 设置默认值
	if globalConfig.MaxRetries == 0 {
		globalConfig.MaxRetries = defaultConfig.MaxRetries
	}
	if globalConfig.Timeout == 0 {
		globalConfig.Timeout = defaultConfig.Timeout
	}
	if globalConfig.RateLimit == 0 {
		globalConfig.RateLimit = defaultConfig.RateLimit
	}

	// 初始化AI服务管理器
	if len(globalConfig.AIServices) > 0 {
		// 使用新的多AI服务配置
		aiServiceManager = NewAIServiceManager(globalConfig.AIServices)
	} else {
		// 向后兼容：使用旧的单服务配置
		legacyService := AIService{
			Name:     "default",
			Endpoint: globalConfig.AIEndpoint,
			Token:    globalConfig.Token,
			Model:    globalConfig.Model,
			Priority: 1,
			Enabled:  true,
		}
		aiServiceManager = NewAIServiceManager([]AIService{legacyService})
	}

	// 初始化规则引擎
	ruleEngine = NewRuleEngine(globalConfig.Rules)

	// 初始化缓存管理器
	cacheManager = NewCacheManager(globalConfig.Cache)

	// 验证配置
	validator := NewConfigValidator()
	if err := validator.Validate(&globalConfig); err != nil {
		// 显示详细的验证错误
		fmt.Printf("❌ 配置验证失败:\n")
		for _, validationErr := range validator.GetErrors() {
			fmt.Printf("  • %s: %s (当前值: %s)\n", validationErr.Field, validationErr.Message, validationErr.Value)
		}
		return fmt.Errorf("配置验证失败: %w", err)
	}

	if *verbose {
		fmt.Printf("✅ 已加载配置文件: %s\n", configPath)
		fmt.Printf("   AI 端点: %s\n", globalConfig.AIEndpoint)
		fmt.Printf("   模型: %s\n", globalConfig.Model)
		if globalConfig.CustomPrompt != "" {
			fmt.Printf("   自定义提示词: %s\n", globalConfig.CustomPrompt)
		}
	}

	return nil
}

// 创建默认配置文件
func createDefaultConfig(configPath string) error {
	// 确保配置目录存在
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 创建默认配置文件
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化默认配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %w", err)
	}

	fmt.Printf("📝 已创建默认配置文件: %s\n", configPath)
	fmt.Println("   请编辑配置文件设置您的 AI 服务器端点和 Token")

	globalConfig = defaultConfig
	return nil
}

// 从标准输入处理日志
func processStdin() {
	if *noBatch {
		// 禁用批处理，逐行处理
		processStdinLineByLine()
		return
	}

	// 使用批处理模式
	processStdinWithBatch()
}

// 逐行处理（原始方式）
func processStdinLineByLine() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineCount := 0
	filteredCount := 0
	alertCount := 0

	// 创建日志行合并器
	merger := NewLogLineMerger(*logFormat)

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if strings.TrimSpace(line) == "" {
			continue
		}

		// 尝试合并多行日志
		completeLine, hasComplete := merger.Add(line)
		if hasComplete {
			filtered, alerted := processLogLine(completeLine)
			if filtered {
				filteredCount++
			}
			if alerted {
				alertCount++
			}
		}
	}

	// 刷新最后的缓冲
	if lastLine, hasLast := merger.Flush(); hasLast {
		filtered, alerted := processLogLine(lastLine)
		if filtered {
			filteredCount++
		}
		if alerted {
			alertCount++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("读取输入失败: %v", err)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 统计: 总计 %d 行, 过滤 %d 行, 告警 %d 次\n", lineCount, filteredCount, alertCount)
}

// 处理journalctl命令
func processJournalctl() {
	// 构建journalctl命令
	args := buildJournalctlCommand()

	// 显示使用的命令
	fmt.Printf("🔧 执行命令: %s\n", strings.Join(args, " "))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 创建命令
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 创建管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("❌ 创建管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		log.Fatalf("❌ 启动journalctl失败: %v", err)
	}

	// 处理输出
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineCount := 0
	filteredCount := 0
	alertCount := 0
	batchCount := 0

	// 创建批处理器
	batcher := NewLogBatcher(func(lines []string) {
		batchCount++
		if *verbose || *debug {
			log.Printf("📦 批次 #%d: 处理 %d 行日志", batchCount, len(lines))
		}

		filtered, alerted := processBatch(lines)
		filteredCount += filtered
		alertCount += alerted
	})

	// 创建日志行合并器
	merger := NewLogLineMerger(*logFormat)

	// 读取日志行
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if strings.TrimSpace(line) == "" {
			continue
		}

		// 尝试合并多行日志
		completeLine, hasComplete := merger.Add(line)
		if hasComplete {
			// 添加到批处理器
			batcher.Add(completeLine)
		}
	}

	// 刷新最后的缓冲
	if lastLine, hasLast := merger.Flush(); hasLast {
		batcher.Add(lastLine)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("❌ 读取journalctl输出失败: %v", err)
	}

	// 刷新剩余的日志
	batcher.Flush()

	// 等待命令结束
	cmd.Wait()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 统计: 总计 %d 行, 过滤 %d 行, 告警 %d 次, 批次 %d 个\n", lineCount, filteredCount, alertCount, batchCount)
}

// 自动检测多源配置文件
func findMultiSourceConfig() (string, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")

	// 按优先级顺序检查多源配置文件
	configFiles := []string{
		"aipipe-sources.json",
		"aipipe-sources.yaml",
		"aipipe-sources.yml",
		"aipipe-sources.toml",
		"aipipe-multi.json",
		"aipipe-multi.yaml",
		"aipipe-multi.yml",
		"aipipe-multi.toml",
	}

	for _, filename := range configFiles {
		configPath := filepath.Join(configDir, filename)
		if _, err := os.Stat(configPath); err == nil {
			if *verbose {
				log.Printf("🔍 找到多源配置文件: %s", configPath)
			}
			return configPath, nil
		}
	}

	// 没有找到任何配置文件，返回默认路径
	defaultPath := filepath.Join(configDir, "aipipe-sources.json")
	return defaultPath, nil
}

// 处理多源监控
func processMultiSource() {
	var configPath string
	var err error

	if *multiSource != "" {
		// 使用指定的配置文件
		configPath = *multiSource
	} else {
		// 自动检测多源配置文件
		configPath, err = findMultiSourceConfig()
		if err != nil {
			log.Fatalf("❌ 查找多源配置文件失败: %v", err)
		}
	}

	// 加载多源配置文件
	config, err := loadMultiSourceConfig(configPath)
	if err != nil {
		log.Fatalf("❌ 加载多源配置文件失败: %v", err)
	}

	// 加载主配置文件
	if err := loadConfig(); err != nil {
		log.Printf("⚠️  加载主配置文件失败，使用默认配置: %v", err)
		globalConfig = defaultConfig
	}

	fmt.Printf("🚀 AIPipe 多源监控启动 - 监控 %d 个源\n", len(config.Sources))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 显示启用的源
	enabledSources := 0
	for _, source := range config.Sources {
		if source.Enabled {
			enabledSources++
			fmt.Printf("📡 源: %s (%s) - %s\n", source.Name, source.Type, source.Description)
		}
	}

	if enabledSources == 0 {
		log.Fatalf("❌ 没有启用的监控源")
	}

	fmt.Printf("✅ 启用 %d 个监控源\n", enabledSources)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 创建等待组
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动每个监控源
	for _, source := range config.Sources {
		if !source.Enabled {
			continue
		}

		wg.Add(1)
		go func(src SourceConfig) {
			defer wg.Done()
			monitorSource(ctx, src)
		}(source)
	}

	// 等待所有监控源完成
	wg.Wait()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 多源监控完成")
}

// 监控单个源
func monitorSource(ctx context.Context, source SourceConfig) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ 源 %s 监控panic恢复: %v", source.Name, r)
		}
	}()

	fmt.Printf("🔍 启动监控源: %s (%s)\n", source.Name, source.Type)

	switch source.Type {
	case "file":
		monitorFileSource(ctx, source)
	case "journalctl":
		monitorJournalSource(ctx, source)
	case "stdin":
		monitorStdinSource(ctx, source)
	default:
		log.Printf("❌ 不支持的源类型: %s", source.Type)
	}
}

// 监控文件源
func monitorFileSource(ctx context.Context, source SourceConfig) {
	if source.Path == "" {
		log.Printf("❌ 源 %s 缺少文件路径", source.Name)
		return
	}

	// 设置当前日志文件路径
	currentLogFile = source.Path

	// 创建日志行合并器
	merger := NewLogLineMerger(source.Format)

	// 创建批处理器
	batcher := NewLogBatcher(func(lines []string) {
		processBatch(lines)
	})

	// 启动文件监控（非阻塞）
	watchFileWithContext(ctx, source.Path, merger, batcher)

	// 等待context取消，保持goroutine运行
	<-ctx.Done()
	log.Printf("🔍 监控源已停止: %s", source.Name)
}

// 监控journalctl源
func monitorJournalSource(ctx context.Context, source SourceConfig) {
	if source.Journal == nil {
		log.Printf("❌ 源 %s 缺少journalctl配置", source.Name)
		return
	}

	// 构建journalctl命令
	args := buildJournalctlCommandFromConfig(source.Journal)

	// 创建命令
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)

	// 创建管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("❌ 源 %s 创建管道失败: %v", source.Name, err)
		return
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		log.Printf("❌ 源 %s 启动journalctl失败: %v", source.Name, err)
		return
	}

	// 创建日志行合并器
	merger := NewLogLineMerger(source.Format)

	// 创建批处理器
	batcher := NewLogBatcher(func(lines []string) {
		processBatch(lines)
	})

	// 处理输出
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 尝试合并多行日志
		completeLine, hasComplete := merger.Add(line)
		if hasComplete {
			batcher.Add(completeLine)
		}
	}

	// 刷新最后的缓冲
	if lastLine, hasLast := merger.Flush(); hasLast {
		batcher.Add(lastLine)
	}

	// 刷新剩余的日志
	batcher.Flush()

	// 等待命令结束
	cmd.Wait()
}

// 监控stdin源
func monitorStdinSource(ctx context.Context, source SourceConfig) {
	// 创建日志行合并器
	merger := NewLogLineMerger(source.Format)

	// 创建批处理器
	batcher := NewLogBatcher(func(lines []string) {
		processBatch(lines)
	})

	// 处理标准输入
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 尝试合并多行日志
		completeLine, hasComplete := merger.Add(line)
		if hasComplete {
			batcher.Add(completeLine)
		}
	}

	// 刷新最后的缓冲
	if lastLine, hasLast := merger.Flush(); hasLast {
		batcher.Add(lastLine)
	}

	// 刷新剩余的日志
	batcher.Flush()
}

// 从配置构建journalctl命令
func buildJournalctlCommandFromConfig(journal *JournalConfig) []string {
	args := []string{"journalctl", "-f", "--no-pager"}

	// 添加服务过滤
	if len(journal.Services) > 0 {
		for _, service := range journal.Services {
			service = strings.TrimSpace(service)
			if service != "" {
				args = append(args, "-u", service)
			}
		}
	}

	// 添加优先级过滤
	if journal.Priority != "" {
		args = append(args, "-p", journal.Priority)
	}

	// 添加时间范围
	if journal.Since != "" {
		args = append(args, "--since", journal.Since)
	}
	if journal.Until != "" {
		args = append(args, "--until", journal.Until)
	}

	// 添加用户过滤
	if journal.User != "" {
		args = append(args, "_UID="+journal.User)
	}

	// 添加启动过滤
	if journal.Boot {
		args = append(args, "-b")
	}

	// 添加内核过滤
	if journal.Kernel {
		args = append(args, "-k")
	}

	return args
}

// 配置文件格式检测
func detectConfigFormat(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	default:
		// 尝试读取文件内容来检测格式
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "json" // 默认格式
		}

		// 检测JSON格式
		var jsonTest interface{}
		if json.Unmarshal(data, &jsonTest) == nil {
			return "json"
		}

		// 检测YAML格式
		var yamlTest interface{}
		if yaml.Unmarshal(data, &yamlTest) == nil {
			return "yaml"
		}

		// 检测TOML格式
		var tomlTest interface{}
		if _, err := toml.Decode(string(data), &tomlTest); err == nil {
			return "toml"
		}

		return "json" // 默认格式
	}
}

// 解析配置文件
func parseConfigFile(data []byte, format string, target interface{}) error {
	switch format {
	case "json":
		return json.Unmarshal(data, target)
	case "yaml":
		return yaml.Unmarshal(data, target)
	case "toml":
		_, err := toml.Decode(string(data), target)
		return err
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", format)
	}
}

// 加载多源配置文件
func loadMultiSourceConfig(configPath string) (*MultiSourceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 自动检测配置文件格式
	format := detectConfigFormat(configPath)
	if *verbose {
		log.Printf("🔍 检测到配置文件格式: %s", format)
	}

	var config MultiSourceConfig
	if err := parseConfigFile(data, format, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败 (%s格式): %v", format, err)
	}

	return &config, nil
}

// 加载主配置文件（支持多种格式）
func loadConfigWithFormat(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 自动检测配置文件格式
	format := detectConfigFormat(configPath)
	if *verbose {
		log.Printf("🔍 检测到主配置文件格式: %s", format)
	}

	if err := parseConfigFile(data, format, &globalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败 (%s格式): %v", format, err)
	}

	return nil
}

// 带上下文的文件监控
func watchFileWithContext(ctx context.Context, filePath string, merger *LogLineMerger, batcher *LogBatcher) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("⚠️  文件不存在，等待创建: %s", filePath)
		// 等待文件创建，每5秒检查一次
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := os.Stat(filePath); err == nil {
					log.Printf("✅ 文件已创建: %s", filePath)
					break
				}
			}
		}
	}

	// 启动文件监控goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ 文件监控panic恢复: %v", r)
			}
		}()

		// 使用fsnotify监控文件
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("❌ 创建文件监控器失败: %v", err)
			return
		}
		defer watcher.Close()

		// 监控文件目录
		dir := filepath.Dir(filePath)
		if err := watcher.Add(dir); err != nil {
			log.Printf("❌ 添加目录监控失败: %v", err)
			return
		}

		// 读取初始文件内容
		file, err := os.Open(filePath)
		if err != nil {
			log.Printf("❌ 打开文件失败: %v", err)
			return
		}
		defer file.Close()

		// 定位到文件末尾
		file.Seek(0, io.SeekEnd)

		// 读取文件内容
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}

			// 尝试合并多行日志
			completeLine, hasComplete := merger.Add(line)
			if hasComplete {
				batcher.Add(completeLine)
			}
		}

		// 刷新最后的缓冲
		if lastLine, hasLast := merger.Flush(); hasLast {
			batcher.Add(lastLine)
		}

		// 监控文件变化
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// 文件写入事件
				if event.Op&fsnotify.Write == fsnotify.Write {
					if event.Name == filePath {
						// 读取新内容
						file, err := os.Open(filePath)
						if err != nil {
							continue
						}

						// 定位到文件末尾
						file.Seek(0, io.SeekEnd)

						scanner := bufio.NewScanner(file)
						scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

						for scanner.Scan() {
							line := scanner.Text()
							if strings.TrimSpace(line) == "" {
								continue
							}

							completeLine, hasComplete := merger.Add(line)
							if hasComplete {
								batcher.Add(completeLine)
							}
						}

						file.Close()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("⚠️  文件监控错误: %v", err)
			}
		}
	}()

	// 函数立即返回，goroutine继续在后台运行
}

// 批处理模式处理标准输入
func processStdinWithBatch() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineCount := 0
	filteredCount := 0
	alertCount := 0
	batchCount := 0

	// 创建批处理器
	batcher := NewLogBatcher(func(lines []string) {
		batchCount++
		if *verbose || *debug {
			log.Printf("📦 批次 #%d: 处理 %d 行日志", batchCount, len(lines))
		}

		filtered, alerted := processBatch(lines)
		filteredCount += filtered
		alertCount += alerted
	})

	// 创建日志行合并器
	merger := NewLogLineMerger(*logFormat)

	// 读取日志行
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if strings.TrimSpace(line) == "" {
			continue
		}

		// 尝试合并多行日志
		completeLine, hasComplete := merger.Add(line)
		if hasComplete {
			// 添加到批处理器
			batcher.Add(completeLine)
		}
	}

	// 刷新最后的缓冲
	if lastLine, hasLast := merger.Flush(); hasLast {
		batcher.Add(lastLine)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("读取输入失败: %v", err)
	}

	// 刷新剩余的日志
	batcher.Flush()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 统计: 总计 %d 行, 过滤 %d 行, 告警 %d 次, 批次 %d 个\n", lineCount, filteredCount, alertCount, batchCount)
}

// 创建日志批处理器
func NewLogBatcher(processor func([]string)) *LogBatcher {
	batcher := &LogBatcher{
		lines:     make([]string, 0, *batchSize),
		processor: processor,
	}
	return batcher
}

// 添加日志到批处理器
func (b *LogBatcher) Add(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lines = append(b.lines, line)

	// 如果达到批处理大小，立即处理
	if len(b.lines) >= *batchSize {
		b.flush()
		return
	}

	// 重置定时器
	if b.timer != nil {
		b.timer.Stop()
	}
	b.timer = time.AfterFunc(*batchWait, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.flush()
	})
}

// 刷新批处理器（内部方法，不加锁）
func (b *LogBatcher) flush() {
	if len(b.lines) == 0 {
		return
	}

	// 处理当前批次
	b.processor(b.lines)

	// 清空批次
	b.lines = make([]string, 0, *batchSize)
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
}

// 刷新批处理器（公共方法，加锁）
func (b *LogBatcher) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flush()
}

// 处理一批日志
func processBatch(lines []string) (filtered int, alerted int) {
	if len(lines) == 0 {
		return 0, 0
	}

	// 先进行本地预过滤
	needAIAnalysis := make([]string, 0)
	localFiltered := make(map[int]*LogAnalysis) // 索引 -> 本地分析结果

	for i, line := range lines {
		if localAnalysis := tryLocalFilter(line); localAnalysis != nil {
			localFiltered[i] = localAnalysis
			filtered++
		} else {
			needAIAnalysis = append(needAIAnalysis, line)
		}
	}

	// 如果有需要 AI 分析的日志，批量调用
	var aiResults map[int]*LogAnalysis
	if len(needAIAnalysis) > 0 {
		batchAnalysis, err := analyzeBatchLogs(needAIAnalysis, *logFormat)
		if err != nil {
			if *verbose {
				log.Printf("❌ 批量分析失败: %v", err)
			}
			// 失败时逐行处理
			for _, line := range needAIAnalysis {
				f, a := processLogLine(line)
				if f {
					filtered++
				}
				if a {
					alerted++
				}
			}
			return filtered, alerted
		}

		// 构建 AI 结果映射
		aiResults = make(map[int]*LogAnalysis)
		aiIndex := 0
		for i := range lines {
			if _, isLocal := localFiltered[i]; !isLocal {
				if aiIndex < len(batchAnalysis.Results) {
					aiResults[i] = &batchAnalysis.Results[aiIndex]
					aiIndex++
				}
			}
		}

		// 显示整体摘要
		if batchAnalysis.ImportantCount > 0 {
			fmt.Printf("\n📋 批次摘要: %s (重要日志: %d 条)\n\n",
				batchAnalysis.OverallSummary, batchAnalysis.ImportantCount)
		}
	}

	// 第一步：标记重要日志的索引
	importantIndices := make(map[int]bool)
	importantLogs := make([]string, 0)

	for i, line := range lines {
		var analysis *LogAnalysis
		if localResult, ok := localFiltered[i]; ok {
			analysis = localResult
		} else if aiResult, ok := aiResults[i]; ok {
			analysis = aiResult
		} else {
			analysis = &LogAnalysis{
				ShouldFilter: true,
				Summary:      "无分析结果",
				Reason:       "批量分析失败或结果缺失",
			}
		}

		if !analysis.ShouldFilter {
			importantIndices[i] = true
			importantLogs = append(importantLogs, line)
			alerted++
		} else {
			filtered++
		}
	}

	// 第二步：计算需要显示的行（重要行 + 上下文）
	shouldDisplay := make(map[int]bool)
	for i := range importantIndices {
		// 显示重要行本身
		shouldDisplay[i] = true

		// 显示前面的上下文
		for j := i - *contextLines; j < i; j++ {
			if j >= 0 {
				shouldDisplay[j] = true
			}
		}

		// 显示后面的上下文
		for j := i + 1; j <= i+*contextLines; j++ {
			if j < len(lines) {
				shouldDisplay[j] = true
			}
		}
	}

	// 第三步：输出日志（带上下文）
	lastDisplayed := -10 // 上次显示的行号
	for i, line := range lines {
		var analysis *LogAnalysis
		if localResult, ok := localFiltered[i]; ok {
			analysis = localResult
		} else if aiResult, ok := aiResults[i]; ok {
			analysis = aiResult
		} else {
			analysis = &LogAnalysis{
				ShouldFilter: true,
				Summary:      "无分析结果",
			}
		}

		// 判断是否应该显示这行
		if !shouldDisplay[i] && !*showNotImportant {
			continue // 不显示
		}

		// 如果距离上次显示的行较远，插入分隔符
		if i > lastDisplayed+1 && lastDisplayed >= 0 {
			fmt.Println("   ...")
		}

		// 显示日志
		isImportant := importantIndices[i]
		isContext := shouldDisplay[i] && !isImportant

		if isImportant {
			fmt.Printf("⚠️  [重要] %s\n", line)
		} else if isContext {
			fmt.Printf("   │ %s\n", line) // 上下文行用 │ 标记
		} else if *showNotImportant {
			fmt.Printf("🔇 [过滤] %s\n", line)
			if *verbose && analysis.Reason != "" {
				fmt.Printf("   原因: %s\n", analysis.Reason)
			}
		}

		lastDisplayed = i
	}

	// 如果有重要日志，发送一次批量通知
	if len(importantLogs) > 0 {
		// 收集所有重要日志的摘要
		summaries := make([]string, 0)
		for _, result := range aiResults {
			if result != nil && !result.ShouldFilter && result.Summary != "" {
				summaries = append(summaries, result.Summary)
			}
		}

		// 构建批量通知摘要
		var notifySummary string
		if len(summaries) > 0 {
			if len(summaries) == 1 {
				notifySummary = summaries[0]
			} else if len(summaries) <= 3 {
				notifySummary = strings.Join(summaries, "、")
			} else {
				notifySummary = fmt.Sprintf("%s 等 %d 个问题", strings.Join(summaries[:2], "、"), len(summaries))
			}
		} else {
			notifySummary = fmt.Sprintf("发现 %d 条重要日志", len(importantLogs))
		}

		// 构建通知内容（提供更详细的上下文）
		notifyContent := ""
		if len(importantLogs) == 1 {
			// 单条日志，显示完整内容
			notifyContent = importantLogs[0]
		} else if len(importantLogs) <= 5 {
			// 5条以内，显示所有日志（截断长行）
			formattedLogs := make([]string, len(importantLogs))
			for i, line := range importantLogs {
				if len(line) > 200 {
					formattedLogs[i] = line[:200] + "..."
				} else {
					formattedLogs[i] = line
				}
			}
			notifyContent = strings.Join(formattedLogs, "\n\n")
		} else {
			// 超过5条，显示前3条和统计信息
			formattedLogs := make([]string, 0, 4)
			for i, line := range importantLogs {
				if i >= 3 {
					break
				}
				if len(line) > 150 {
					formattedLogs = append(formattedLogs, line[:150]+"...")
				} else {
					formattedLogs = append(formattedLogs, line)
				}
			}
			formattedLogs = append(formattedLogs, fmt.Sprintf("... 还有 %d 条重要日志", len(importantLogs)-3))
			notifyContent = strings.Join(formattedLogs, "\n\n")
		}

		// 发送一次通知
		go sendNotification(notifySummary, notifyContent)
	}

	return filtered, alerted
}

// 处理单行日志
func processLogLine(line string) (filtered bool, alerted bool) {
	// 分析日志
	analysis, err := analyzeLog(line, *logFormat)
	if err != nil {
		if *verbose {
			log.Printf("❌ 分析日志失败: %v", err)
		}
		// 出错时默认显示日志
		fmt.Println(line)
		return false, false
	}

	if analysis.ShouldFilter {
		// 过滤掉的日志 - 默认不显示，除非开启 --show-not-important
		if *showNotImportant {
			fmt.Printf("🔇 [过滤] %s\n", line)
			if *verbose && analysis.Reason != "" {
				fmt.Printf("   原因: %s\n", analysis.Reason)
			}
		}
		return true, false
	} else {
		// 重要日志，需要通知用户
		fmt.Printf("⚠️  [重要] %s\n", line)
		fmt.Printf("   📝 摘要: %s\n", analysis.Summary)

		// 发送 macOS 通知
		go sendNotification(analysis.Summary, line)
		return false, true
	}
}

// 监控文件
func watchFile(path string) error {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 设置全局变量，用于通知
	currentLogFile = absPath

	// 加载上次的状态
	state := loadFileState(absPath)

	// 打开文件
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	currentInode := getInode(fileInfo)

	// 如果是同一个文件且有保存的位置，从上次位置开始读取
	if state != nil && state.Inode == currentInode && state.Offset > 0 {
		fmt.Printf("📌 从上次位置继续读取 (偏移: %d 字节)\n", state.Offset)
		if _, err := file.Seek(state.Offset, 0); err != nil {
			fmt.Printf("⚠️  无法定位到上次位置，从文件末尾开始: %v\n", err)
			file.Seek(0, 2) // 定位到文件末尾
		}
	} else {
		// 新文件或轮转后的文件，从末尾开始
		fmt.Println("📌 从文件末尾开始监控新内容")
		file.Seek(0, 2)
	}

	// 创建文件监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控器失败: %w", err)
	}
	defer watcher.Close()

	// 监控文件
	if err := watcher.Add(absPath); err != nil {
		return fmt.Errorf("添加文件监控失败: %w", err)
	}

	reader := bufio.NewReader(file)
	lineCount := 0
	filteredCount := 0
	alertCount := 0
	batchCount := 0

	// 创建批处理器（如果未禁用批处理）
	var batcher *LogBatcher
	if !*noBatch {
		batcher = NewLogBatcher(func(lines []string) {
			batchCount++
			if *verbose || *debug {
				log.Printf("📦 批次 #%d: 处理 %d 行日志", batchCount, len(lines))
			}

			filtered, alerted := processBatch(lines)
			filteredCount += filtered
			alertCount += alerted

			// 批处理完成后保存文件位置
			offset, _ := file.Seek(0, 1)
			saveFileState(absPath, offset, currentInode)
		})
	}

	// 创建日志行合并器
	merger := NewLogLineMerger(*logFormat)

	// 立即读取当前位置到文件末尾的内容
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("读取文件失败: %v", err)
			}
			break
		}

		line = strings.TrimSuffix(line, "\n")
		if strings.TrimSpace(line) == "" {
			continue
		}

		lineCount++

		// 尝试合并多行日志
		completeLine, hasComplete := merger.Add(line)
		if !hasComplete {
			continue
		}

		if *noBatch {
			// 逐行处理模式
			filtered, alerted := processLogLine(completeLine)
			if filtered {
				filteredCount++
			}
			if alerted {
				alertCount++
			}
			// 保存当前位置
			offset, _ := file.Seek(0, 1)
			saveFileState(absPath, offset, currentInode)
		} else {
			// 批处理模式
			batcher.Add(completeLine)
		}
	}

	// 刷新合并器的最后缓冲
	if lastLine, hasLast := merger.Flush(); hasLast {
		if *noBatch {
			filtered, alerted := processLogLine(lastLine)
			if filtered {
				filteredCount++
			}
			if alerted {
				alertCount++
			}
			offset, _ := file.Seek(0, 1)
			saveFileState(absPath, offset, currentInode)
		} else {
			batcher.Add(lastLine)
		}
	}

	// 刷新初始批次
	if batcher != nil {
		batcher.Flush()
	}

	fmt.Println("⏳ 等待新日志...")

	// 监控文件变化
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				// 文件有新内容
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err != io.EOF {
							log.Printf("读取文件失败: %v", err)
						}
						break
					}

					line = strings.TrimSuffix(line, "\n")
					if strings.TrimSpace(line) == "" {
						continue
					}

					lineCount++

					// 尝试合并多行日志
					completeLine, hasComplete := merger.Add(line)
					if !hasComplete {
						continue
					}

					if *noBatch {
						// 逐行处理模式
						filtered, alerted := processLogLine(completeLine)
						if filtered {
							filteredCount++
						}
						if alerted {
							alertCount++
						}
						// 保存当前位置
						offset, _ := file.Seek(0, 1)
						saveFileState(absPath, offset, currentInode)
					} else {
						// 批处理模式
						batcher.Add(completeLine)
					}
				}
			}

			// 检测文件轮转（删除或重命名）
			if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
				fmt.Println("🔄 检测到日志轮转，等待新文件...")

				// 刷新合并器缓冲区（处理未完成的日志）
				if lastLine, hasLast := merger.Flush(); hasLast {
					if *noBatch {
						filtered, alerted := processLogLine(lastLine)
						if filtered {
							filteredCount++
						}
						if alerted {
							alertCount++
						}
					} else {
						batcher.Add(lastLine)
					}
				}

				// 等待新文件出现
				time.Sleep(1 * time.Second)

				// 尝试重新打开文件
				newFile, err := os.Open(absPath)
				if err != nil {
					fmt.Printf("⚠️  等待新文件创建: %v\n", err)
					continue
				}

				// 关闭旧文件
				file.Close()
				file = newFile
				reader = bufio.NewReader(file)

				// 重新创建合并器（新文件）
				merger = NewLogLineMerger(*logFormat)

				// 获取新文件信息
				fileInfo, err := file.Stat()
				if err == nil {
					currentInode = getInode(fileInfo)
					fmt.Println("✅ 已切换到新文件")
					// 重置偏移量
					saveFileState(absPath, 0, currentInode)
				}
			}

		case <-ticker.C:
			// 定期检查文件是否被轮转（大小变小）
			currentInfo, err := os.Stat(absPath)
			if err != nil {
				continue
			}

			currentSize := currentInfo.Size()
			currentOffset, _ := file.Seek(0, 1)

			// 如果文件大小小于当前偏移量，说明文件被截断或轮转
			if currentSize < currentOffset {
				fmt.Println("🔄 检测到文件截断或轮转，重新打开文件...")

				// 刷新合并器缓冲区（处理未完成的日志）
				if lastLine, hasLast := merger.Flush(); hasLast {
					if *noBatch {
						filtered, alerted := processLogLine(lastLine)
						if filtered {
							filteredCount++
						}
						if alerted {
							alertCount++
						}
					} else {
						batcher.Add(lastLine)
					}
				}

				// 重新打开文件
				file.Close()
				newFile, err := os.Open(absPath)
				if err != nil {
					log.Printf("重新打开文件失败: %v", err)
					continue
				}

				file = newFile
				reader = bufio.NewReader(file)

				// 重新创建合并器（新文件）
				merger = NewLogLineMerger(*logFormat)

				// 获取新文件信息
				fileInfo, _ := file.Stat()
				currentInode = getInode(fileInfo)

				// 从头开始读取
				saveFileState(absPath, 0, currentInode)
				fmt.Println("✅ 已重新打开文件，从头开始读取")
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("监控错误: %v", err)
		}
	}
}

// 获取文件状态路径
func getStateFilePath(logPath string) string {
	dir := filepath.Dir(logPath)
	base := filepath.Base(logPath)
	return filepath.Join(dir, ".aipipe_"+base+".state")
}

// 加载文件状态
func loadFileState(path string) *FileState {
	stateFile := getStateFilePath(path)
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil
	}

	var state FileState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil
	}

	return &state
}

// 保存文件状态
func saveFileState(path string, offset int64, inode uint64) {
	state := FileState{
		Path:   path,
		Offset: offset,
		Inode:  inode,
		Time:   time.Now(),
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}

	stateFile := getStateFilePath(path)
	os.WriteFile(stateFile, data, 0644)
}

// 分析日志内容（单条）
func analyzeLog(logLine string, format string) (*LogAnalysis, error) {
	// 本地预过滤：对于明确的低级别日志，直接过滤，不调用 AI
	if localAnalysis := tryLocalFilter(logLine); localAnalysis != nil {
		return localAnalysis, nil
	}

	// 构建系统提示词和用户提示词
	systemPrompt := buildSystemPrompt(format)
	userPrompt := buildUserPrompt(logLine)

	// 调用 AI API
	response, err := callAIAPI(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("调用 AI API 失败: %w", err)
	}

	// 解析响应
	analysis, err := parseAnalysisResponse(response)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 后处理：保守策略，当 AI 无法确定时，默认过滤
	analysis = applyConservativeFilter(analysis)

	return analysis, nil
}

// 批量分析日志
func analyzeBatchLogs(logLines []string, format string) (*BatchLogAnalysis, error) {
	if len(logLines) == 0 {
		return &BatchLogAnalysis{}, nil
	}

	// 构建系统提示词
	systemPrompt := buildSystemPrompt(format)

	// 构建批量分析的用户提示词
	userPrompt := buildBatchUserPrompt(logLines)

	// 调用 AI API
	response, err := callAIAPI(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("调用 AI API 失败: %w", err)
	}

	// 解析批量响应
	batchAnalysis, err := parseBatchAnalysisResponse(response, len(logLines))
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 应用保守策略到每一条结果
	for i := range batchAnalysis.Results {
		batchAnalysis.Results[i] = *applyConservativeFilter(&batchAnalysis.Results[i])
		if !batchAnalysis.Results[i].ShouldFilter {
			batchAnalysis.ImportantCount++
		}
	}

	return batchAnalysis, nil
}

// 本地预过滤：对于明确的低级别日志，直接过滤，不调用 AI
// 返回 nil 表示无法本地判断，需要调用 AI
func tryLocalFilter(logLine string) *LogAnalysis {
	// 转换为大写以便匹配
	upperLine := strings.ToUpper(logLine)

	// 定义低级别日志的正则模式
	// 匹配常见的日志级别格式：[DEBUG]、DEBUG、 DEBUG 、[D] 等
	lowLevelPatterns := []struct {
		level   string
		pattern string
		summary string
	}{
		{"TRACE", `\b(TRACE|TRC)\b`, "TRACE 级别日志"},
		{"DEBUG", `\b(DEBUG|DBG|D)\b`, "DEBUG 级别日志"},
		{"INFO", `\b(INFO|INF|I)\b`, "INFO 级别日志"},
		{"VERBOSE", `\bVERBOSE\b`, "VERBOSE 级别日志"},
	}

	for _, pattern := range lowLevelPatterns {
		// 使用正则表达式匹配
		matched, err := regexp.MatchString(pattern.pattern, upperLine)
		if err == nil && matched {
			// 额外检查：确保不包含明显的错误关键词
			// 有时候 INFO 日志也可能包含 error 等关键词，需要进一步判断
			hasErrorKeywords := strings.Contains(upperLine, "ERROR") ||
				strings.Contains(upperLine, "EXCEPTION") ||
				strings.Contains(upperLine, "FATAL") ||
				strings.Contains(upperLine, "CRITICAL") ||
				strings.Contains(upperLine, "FAILED") ||
				strings.Contains(upperLine, "FAILURE")

			// 如果日志级别是低级别，但包含错误关键词，还是交给 AI 判断
			if hasErrorKeywords {
				continue
			}

			if *verbose || *debug {
				log.Printf("⚡ 本地过滤: 检测到 %s 级别，直接过滤（不调用 AI）", pattern.level)
			}

			return &LogAnalysis{
				ShouldFilter: true,
				Summary:      pattern.summary,
				Reason:       fmt.Sprintf("本地过滤：%s 级别的日志通常无需关注", pattern.level),
			}
		}
	}

	// 无法本地判断，返回 nil，需要调用 AI
	return nil
}

// 应用保守过滤策略
// 当 AI 无法判断或返回不确定结果时，默认过滤掉，避免误报
func applyConservativeFilter(analysis *LogAnalysis) *LogAnalysis {
	// 检查的关键词列表（表示 AI 无法确定或日志异常）
	uncertainKeywords := []string{
		"日志内容异常",
		"日志内容不完整",
		"无法判断",
		"日志格式异常",
		"日志内容不符合预期",
		"无法确定",
		"不确定",
		"无法识别",
		"格式不正确",
		"内容异常",
		"无法解析",
	}

	// 检查 summary 和 reason 字段
	checkText := strings.ToLower(analysis.Summary + " " + analysis.Reason)

	for _, keyword := range uncertainKeywords {
		if strings.Contains(checkText, strings.ToLower(keyword)) {
			// 发现不确定的关键词，强制过滤
			if *verbose || *debug {
				log.Printf("🔍 检测到不确定关键词「%s」，采用保守策略：过滤此日志", keyword)
			}
			analysis.ShouldFilter = true
			if analysis.Reason == "" {
				analysis.Reason = "AI 无法确定日志重要性，采用保守策略过滤"
			} else {
				analysis.Reason = analysis.Reason + "（保守策略：过滤）"
			}
			break
		}
	}

	return analysis
}

// 获取格式特定的示例
func getFormatSpecificExamples(format string) string {
	switch format {
	case "java":
		return `Java 特定示例：
   - "INFO com.example.service.UserService - User created successfully"
   - "ERROR com.example.dao.DatabaseDAO - Connection pool exhausted"
   - "WARN com.example.controller.AuthController - Invalid JWT token"`

	case "php":
		return `PHP 特定示例：
   - "PHP Notice: Undefined variable $user in /app/index.php"
   - "PHP Fatal error: Call to undefined function mysql_connect()"
   - "PHP Warning: file_get_contents() failed to open stream"`

	case "nginx":
		return `Nginx 特定示例：
   - "127.0.0.1 - - [13/Oct/2025:10:00:01 +0000] \"GET /api/health HTTP/1.1\" 200"
   - "upstream server temporarily disabled while connecting to upstream"
   - "connect() failed (111: Connection refused) while connecting to upstream"`

	case "go":
		return `Go 特定示例：
   - "INFO: Starting server on :8080"
   - "ERROR: database connection failed: dial tcp: connection refused"
   - "WARN: goroutine leak detected"`

	case "rust":
		return `Rust 特定示例：
   - "INFO: Server listening on 127.0.0.1:8080"
   - "ERROR: thread 'main' panicked at 'index out of bounds'"
   - "WARN: memory usage high: 512MB"`

	case "csharp":
		return `C# 特定示例：
   - "INFO: Application started"
   - "ERROR: System.Exception: Database connection timeout"
   - "WARN: Memory pressure detected"`

	case "nodejs":
		return `Node.js 特定示例：
   - "info: Server running on port 3000"
   - "error: Error: ENOENT: no such file or directory"
   - "warn: DeprecationWarning: Buffer() is deprecated"`

	case "docker":
		return `Docker 特定示例：
   - "Container started successfully"
   - "ERROR: failed to start container: port already in use"
   - "WARN: container running out of memory"`

	case "kubernetes":
		return `Kubernetes 特定示例：
   - "Pod started successfully"
   - "ERROR: Failed to pull image: ImagePullBackOff"
   - "WARN: Pod evicted due to memory pressure"`

	case "postgresql":
		return `PostgreSQL 特定示例：
   - "LOG: database system is ready to accept connections"
   - "ERROR: relation \"users\" does not exist"
   - "WARN: checkpoint request timed out"`

	case "mysql":
		return `MySQL 特定示例：
   - "InnoDB: Database was not shut down normally"
   - "ERROR 1045: Access denied for user 'root'@'localhost'"
   - "Warning: Aborted connection to db"`

	case "redis":
		return `Redis 特定示例：
   - "Redis server version 6.2.6, bits=64"
   - "ERROR: OOM command not allowed when used memory > 'maxmemory'"
   - "WARN: overcommit_memory is set to 0"`

	case "journald":
		return `Linux journald 特定示例：
   - "Oct 17 10:00:01 systemd[1]: Started Network Manager Script Dispatcher Service"
   - "Oct 17 10:00:02 kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500 or sacrifice child"
   - "Oct 17 10:00:03 sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2"`

	case "macos-console":
		return `macOS Console 特定示例：
   - "2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: (AppleH11ANEInterface) ANE0: EnableMemoryUnwireTimer: ERROR: Cannot enable Memory Unwire Timer"
   - "2025-10-17 10:00:02.234567+0800 0x1f11722 Error 0x185174d 386 0 locationd: (TCC) [com.apple.TCC:access] send_message_with_reply_sync(): XPC_ERROR_CONNECTION_INVALID"
   - "2025-10-17 10:00:03.345678+0800 0x1f11e95 Error 0x1851731 558 0 searchpartyd: (TCC) [com.apple.TCC:access] send_message_with_reply_sync(): XPC_ERROR_CONNECTION_INVALID"`

	case "syslog":
		return `Syslog 特定示例：
   - "Oct 17 10:00:01 hostname systemd[1]: Started Network Manager Script Dispatcher Service"
   - "Oct 17 10:00:02 hostname kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500"
   - "Oct 17 10:00:03 hostname sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2"`

	default:
		return ""
	}
}

// 构建系统提示词（定义角色和判断标准）
func buildSystemPrompt(format string) string {
	formatExamples := getFormatSpecificExamples(format)

	basePrompt := fmt.Sprintf(`你是一个专业的日志分析助手，专门分析 %s 格式的日志。

你的任务是判断日志是否需要关注，并以 JSON 格式返回分析结果。

返回格式：
{
  "should_filter": true/false,  // true 表示应该过滤（不重要），false 表示需要关注
  "summary": "简短摘要（20字内）",
  "reason": "判断原因"
}

判断标准和示例：

【应该过滤的日志】(should_filter=true) - 正常运行状态，无需告警：
1. 健康检查和心跳
   - "Health check endpoint called"
   - "Heartbeat received from client"
   - "/health returned 200"
   
2. 应用启动和配置加载
   - "Application started successfully"
   - "Configuration loaded from config.yml"
   - "Server listening on port 8080"
   
3. 正常的业务操作（INFO/DEBUG）
   - "User logged in: john@example.com"
   - "Retrieved 20 records from database"
   - "Cache hit for key: user_123"
   - "Request processed in 50ms"
   
4. 定时任务正常执行
   - "Scheduled task completed successfully"
   - "Cleanup job finished, removed 10 items"
   
5. 静态资源请求
   - "GET /static/css/style.css 200"
   - "Serving static file: logo.png"

6. 常规数据库操作
   - "Query executed successfully in 10ms"
   - "Transaction committed"
   
7. 正常的API请求响应
   - "GET /api/users 200 OK"
   - "POST /api/data returned 201"

【需要关注的日志】(should_filter=false) - 异常情况，需要告警：
1. 错误和异常（ERROR级别）
   - "ERROR: Database connection failed"
   - "NullPointerException at line 123"
   - "Failed to connect to Redis"
   - 任何包含 Exception, Error, Failed 的错误信息
   
2. 数据库问题
   - "Database connection timeout"
   - "Deadlock detected"
   - "Slow query: 5000ms"
   - "Connection pool exhausted"
   
3. 认证和授权问题
   - "Authentication failed for user admin"
   - "Invalid token: access denied"
   - "Permission denied: insufficient privileges"
   - "Multiple failed login attempts from 192.168.1.100"
   
4. 性能问题（WARN级别或慢响应）
   - "Request timeout after 30s"
   - "Response time exceeded threshold: 5000ms"
   - "Memory usage high: 85%%"
   - "Thread pool near capacity: 95/100"
   
5. 资源耗尽
   - "Out of memory error"
   - "Disk space low: 95%% used"
   - "Too many open files"
   
6. 外部服务调用失败
   - "Payment gateway timeout"
   - "Failed to call external API: 500"
   - "Third-party service unavailable"
   
7. 业务异常
   - "Order processing failed: insufficient balance"
   - "Payment declined: invalid card"
   - "Data validation failed"
   
8. 安全问题
   - "SQL injection attempt detected"
   - "Suspicious activity from IP"
   - "Rate limit exceeded"
   - "Invalid CSRF token"
   
9. 数据一致性问题
   - "Data mismatch detected"
   - "Inconsistent state in transaction"
   
10. 服务降级和熔断
    - "Circuit breaker opened"
    - "Service degraded mode activated"`, format)

	// 添加格式特定的示例
	if formatExamples != "" {
		basePrompt += "\n\n" + formatExamples
	}

	basePrompt += `

注意：
- 如果日志级别是 ERROR 或包含 Exception/Error，通常需要关注
- 如果包含 "failed", "timeout", "unable", "cannot" 等负面词汇，需要仔细判断
- 如果是 WARN 级别，需要根据具体内容判断严重程度
- 健康检查、心跳、正常的 INFO 日志通常可以过滤

重要原则（保守策略）：
- 如果日志内容不完整、格式异常或无法确定重要性，请设置 should_filter=true
- 在 summary 或 reason 中明确说明"日志内容异常"、"无法判断"等原因
- 我们采取保守策略：只提示确认重要的信息，不确定的一律过滤

只返回 JSON，不要其他内容。`

	// 如果有自定义提示词，添加到系统提示词中
	if globalConfig.CustomPrompt != "" {
		basePrompt += "\n\n" + globalConfig.CustomPrompt
	}

	return basePrompt
}

// 构建用户提示词（实际要分析的日志）
func buildUserPrompt(logLine string) string {
	return fmt.Sprintf("请分析以下日志：\n\n%s", logLine)
}

// 构建批量用户提示词
func buildBatchUserPrompt(logLines []string) string {
	var sb strings.Builder
	sb.WriteString("请批量分析以下日志，对每一行给出判断：\n\n")

	for i, line := range logLines {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, line))
	}

	sb.WriteString("\n请返回 JSON 格式：\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"results\": [\n")
	sb.WriteString("    {\"should_filter\": true/false, \"summary\": \"摘要\", \"reason\": \"原因\"},\n")
	sb.WriteString("    ...\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"overall_summary\": \"这批日志的整体摘要（20字内）\",\n")
	sb.WriteString(fmt.Sprintf("  \"important_count\": 0  // 重要日志数量（%d 条中有几条）\n", len(logLines)))
	sb.WriteString("}\n")
	sb.WriteString("\n注意：results 数组必须包含 " + fmt.Sprintf("%d", len(logLines)) + " 个元素，按顺序对应每一行日志。")

	return sb.String()
}

// 调用 AI API
func callAIAPI(systemPrompt, userPrompt string) (string, error) {
	// 获取AI服务
	service, err := aiServiceManager.GetNextService()
	if err != nil {
		return "", fmt.Errorf("获取AI服务失败: %w", err)
	}

	// 记录服务调用
	aiServiceManager.RecordCall(service.Name)

	// 构建请求，使用 system 和 user 两条消息
	reqBody := ChatRequest{
		Model: service.Model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Debug: 打印请求信息
	if *debug {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("🔍 DEBUG: HTTP 请求详情")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("服务: %s\n", service.Name)
		fmt.Printf("URL: %s\n", service.Endpoint)
		fmt.Printf("Method: POST\n")
		fmt.Printf("Headers:\n")
		fmt.Printf("  Content-Type: application/json\n")
		fmt.Printf("  api-key: %s...%s\n", service.Token[:min(10, len(service.Token))], service.Token[max(0, len(service.Token)-10):])
		fmt.Printf("\nRequest Body:\n")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonData, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(jsonData))
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", service.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", service.Token)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(globalConfig.Timeout) * time.Second,
	}

	if *debug {
		fmt.Printf("⏳ 发送请求中...\n")
	}

	startTime := time.Now()
	var resp *http.Response
	var httpErr error

	// 重试机制
	for i := 0; i < globalConfig.MaxRetries; i++ {
		resp, httpErr = client.Do(req)
		if httpErr == nil {
			break
		}

		// 使用错误处理器处理网络错误
		if handledErr := errorHandler.Handle(httpErr, map[string]interface{}{
			"operation":   "ai_api_call",
			"service":     service.Name,
			"endpoint":    service.Endpoint,
			"retry":       i + 1,
			"max_retries": globalConfig.MaxRetries,
		}); handledErr != nil {
			if i == globalConfig.MaxRetries-1 {
				if *debug {
					fmt.Printf("❌ 请求失败 (重试 %d/%d): %v\n", i+1, globalConfig.MaxRetries, handledErr)
					fmt.Println(strings.Repeat("=", 80) + "\n")
				}
				return "", handledErr
			}
			time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
		} else {
			// 错误已恢复，重试
			continue
		}
	}

	if httpErr != nil {
		if *debug {
			fmt.Printf("❌ 请求失败: %v\n", httpErr)
			fmt.Println(strings.Repeat("=", 80) + "\n")
		}
		return "", httpErr
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Debug: 打印响应信息
	if *debug {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("🔍 DEBUG: HTTP 响应详情")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Status Code: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("Response Time: %v\n", elapsed)
		fmt.Printf("Content-Length: %d bytes\n", len(body))
		fmt.Printf("\nResponse Headers:\n")
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
		fmt.Printf("\nResponse Body:\n")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(body))
		}
		fmt.Println(strings.Repeat("=", 80) + "\n")
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))

		// 使用错误处理器处理 API 错误
		if handledErr := errorHandler.Handle(apiErr, map[string]interface{}{
			"operation":     "ai_api_response",
			"service":       service.Name,
			"status_code":   resp.StatusCode,
			"endpoint":      service.Endpoint,
			"response_body": string(body),
		}); handledErr != nil {
			return "", handledErr
		}
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API 响应中没有内容")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// 解析批量 AI 响应
func parseBatchAnalysisResponse(response string, expectedCount int) (*BatchLogAnalysis, error) {
	// 提取 JSON（处理 markdown 代码块）
	jsonStr := extractJSON(response)

	var batchAnalysis BatchLogAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &batchAnalysis); err != nil {
		return nil, fmt.Errorf("解析批量 JSON 失败: %w\n原始响应: %s\n提取的JSON: %s", err, response, jsonStr)
	}

	// 验证结果数量
	if len(batchAnalysis.Results) != expectedCount {
		if *verbose || *debug {
			log.Printf("⚠️  批量分析结果数量不匹配：期望 %d 条，实际 %d 条", expectedCount, len(batchAnalysis.Results))
		}

		// 如果结果少于预期，补充默认结果（过滤）
		for len(batchAnalysis.Results) < expectedCount {
			batchAnalysis.Results = append(batchAnalysis.Results, LogAnalysis{
				ShouldFilter: true,
				Summary:      "结果缺失",
				Reason:       "批量分析返回结果数量不足",
			})
		}
	}

	return &batchAnalysis, nil
}

// 提取 JSON（从可能包含 markdown 代码块的响应中）
func extractJSON(response string) string {
	jsonStr := response

	// 处理 ```json ... ``` 格式
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json")
		if start != -1 {
			start += 7
			remaining := response[start:]
			end := strings.Index(remaining, "```")
			if end != -1 {
				jsonStr = remaining[:end]
			}
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```")
		if start != -1 {
			start += 3
			remaining := response[start:]
			end := strings.Index(remaining, "```")
			if end != -1 {
				jsonStr = remaining[:end]
			}
		}
	}

	// 清理字符串
	jsonStr = strings.TrimSpace(jsonStr)

	// 智能定位 JSON 起始和结束
	if len(jsonStr) > 0 && jsonStr[0] != '{' && jsonStr[0] != '[' {
		startBrace := strings.Index(jsonStr, "{")
		startBracket := strings.Index(jsonStr, "[")

		start := -1
		if startBrace != -1 && (startBracket == -1 || startBrace < startBracket) {
			start = startBrace
		} else if startBracket != -1 {
			start = startBracket
		}

		if start != -1 {
			jsonStr = jsonStr[start:]
		}
	}

	if len(jsonStr) > 0 && jsonStr[len(jsonStr)-1] != '}' && jsonStr[len(jsonStr)-1] != ']' {
		endBrace := strings.LastIndex(jsonStr, "}")
		endBracket := strings.LastIndex(jsonStr, "]")

		end := -1
		if endBrace != -1 && endBrace > endBracket {
			end = endBrace
		} else if endBracket != -1 {
			end = endBracket
		}

		if end != -1 {
			jsonStr = jsonStr[:end+1]
		}
	}

	return jsonStr
}

// 解析 AI 响应（单条）
func parseAnalysisResponse(response string) (*LogAnalysis, error) {
	jsonStr := extractJSON(response)

	var analysis LogAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w\n原始响应: %s\n提取的JSON: %s", err, response, jsonStr)
	}

	return &analysis, nil
}

// 发送通知（支持多种方式）
func sendNotification(summary, logLine string) {
	// 截断日志内容，避免通知太长
	displayLog := logLine
	if len(displayLog) > 100 {
		displayLog = displayLog[:100] + "..."
	}

	// 发送系统通知
	sendSystemNotification(summary, displayLog)

	// 发送邮件通知
	if globalConfig.Notifiers.Email.Enabled {
		go safeSendEmailNotification(summary, logLine)
	}

	// 发送webhook通知
	if globalConfig.Notifiers.DingTalk.Enabled {
		go safeSendWebhookNotification(globalConfig.Notifiers.DingTalk, summary, logLine, "dingtalk")
	}
	if globalConfig.Notifiers.WeChat.Enabled {
		go safeSendWebhookNotification(globalConfig.Notifiers.WeChat, summary, logLine, "wechat")
	}
	if globalConfig.Notifiers.Feishu.Enabled {
		go safeSendWebhookNotification(globalConfig.Notifiers.Feishu, summary, logLine, "feishu")
	}
	if globalConfig.Notifiers.Slack.Enabled {
		go safeSendWebhookNotification(globalConfig.Notifiers.Slack, summary, logLine, "slack")
	}

	// 发送自定义webhook通知
	for _, webhook := range globalConfig.Notifiers.CustomWebhooks {
		if webhook.Enabled {
			go safeSendWebhookNotification(webhook, summary, logLine, "custom")
		}
	}
}

// 发送系统通知
func sendSystemNotification(summary, displayLog string) {
	// 检测操作系统并发送相应的通知
	if isMacOS() {
		sendMacOSNotification(summary, displayLog)
	} else if isLinux() {
		sendLinuxNotification(summary, displayLog)
	} else {
		if *verbose || *debug {
			log.Printf("⚠️  不支持的操作系统，跳过系统通知")
		}
		return
	}

	// 播放系统声音
	go playSystemSound()
}

// 检测是否为 macOS
func isMacOS() bool {
	return strings.Contains(strings.ToLower(runtime.GOOS), "darwin")
}

// 检测是否为 Linux
func isLinux() bool {
	return strings.Contains(strings.ToLower(runtime.GOOS), "linux")
}

// 发送 macOS 通知
func sendMacOSNotification(summary, displayLog string) {
	// 使用 osascript 通过标准输入发送通知（更好地支持 UTF-8 中文）
	script := fmt.Sprintf(`display notification "%s" with title "⚠️ 重要日志告警" subtitle "%s"`,
		escapeForAppleScript(displayLog),
		escapeForAppleScript(summary))

	cmd := exec.Command("osascript", "-")
	cmd.Stdin = strings.NewReader(script)

	// 设置环境变量确保使用 UTF-8
	cmd.Env = append(os.Environ(), "LANG=zh_CN.UTF-8")

	err := cmd.Run()

	if err != nil {
		if *verbose || *debug {
			log.Printf("⚠️  发送 macOS 通知失败: %v", err)
			log.Printf("💡 请检查通知权限：系统设置 > 通知 > 终端")
		}
	} else {
		if *verbose || *debug {
			log.Printf("✅ macOS 通知已发送: %s", summary)
		}
	}
}

// 发送 Linux 通知
func sendLinuxNotification(summary, displayLog string) {
	// 尝试使用 notify-send (需要安装 libnotify-bin)
	cmd := exec.Command("notify-send",
		"⚠️ 重要日志告警",
		fmt.Sprintf("%s\n%s", summary, displayLog),
		"--urgency=critical",
		"--expire-time=10000")

	err := cmd.Run()

	if err != nil {
		// 如果 notify-send 失败，尝试使用其他方式
		if *verbose || *debug {
			log.Printf("⚠️  notify-send 失败，尝试其他通知方式: %v", err)
		}

		// 可以在这里添加其他 Linux 通知方式，比如：
		// - 写入到系统日志
		// - 发送到桌面通知服务
		// - 等等

		if *verbose || *debug {
			log.Printf("⚠️  Linux 系统通知发送失败")
		}
		return
	}

	if *verbose || *debug {
		log.Printf("✅ Linux 通知已发送: %s", summary)
	}
}

// 播放系统提示音
func playSystemSound() {
	if isMacOS() {
		playMacOSSound()
	} else if isLinux() {
		playLinuxSound()
	}
	// 其他平台不播放声音，静默失败
}

// 播放 macOS 系统声音
func playMacOSSound() {
	// 使用 afplay 播放系统声音文件（经验证此方式可靠）
	soundPaths := []string{
		"/System/Library/Sounds/Glass.aiff",
		"/System/Library/Sounds/Ping.aiff",
		"/System/Library/Sounds/Pop.aiff",
		"/System/Library/Sounds/Purr.aiff",
		"/System/Library/Sounds/Bottle.aiff",
		"/System/Library/Sounds/Funk.aiff",
	}

	for _, soundPath := range soundPaths {
		cmd := exec.Command("afplay", soundPath)
		if err := cmd.Run(); err == nil {
			if *verbose || *debug {
				log.Printf("🔊 播放 macOS 声音: %s", soundPath)
			}
			return // 播放成功
		}
	}

	// 如果所有声音文件都失败，使用 beep 作为最后保障
	if *verbose || *debug {
		log.Printf("⚠️  macOS 声音文件不可用，使用 beep")
	}
	cmd := exec.Command("osascript", "-e", "beep 1")
	cmd.Run()
}

// 播放 Linux 系统声音
func playLinuxSound() {
	// 尝试使用 paplay (PulseAudio)
	cmd := exec.Command("paplay", "/usr/share/sounds/alsa/Front_Left.wav")
	if err := cmd.Run(); err == nil {
		if *verbose || *debug {
			log.Printf("🔊 播放 Linux 声音: PulseAudio")
		}
		return
	}

	// 尝试使用 aplay (ALSA)
	cmd = exec.Command("aplay", "/usr/share/sounds/alsa/Front_Left.wav")
	if err := cmd.Run(); err == nil {
		if *verbose || *debug {
			log.Printf("🔊 播放 Linux 声音: ALSA")
		}
		return
	}

	// 尝试使用 speaker-test (生成测试音)
	cmd = exec.Command("speaker-test", "-t", "sine", "-f", "1000", "-l", "1")
	if err := cmd.Run(); err == nil {
		if *verbose || *debug {
			log.Printf("🔊 播放 Linux 声音: speaker-test")
		}
		return
	}

	// 如果所有方式都失败，静默失败
	if *verbose || *debug {
		log.Printf("⚠️  Linux 声音播放失败")
	}
}

// 转义 AppleScript 字符串
func escapeForAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}

// 获取文件 inode（用于检测文件轮转）
func getInode(info os.FileInfo) uint64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Ino
	}
	return 0
}

// 创建日志行合并器
func NewLogLineMerger(format string) *LogLineMerger {
	return &LogLineMerger{
		format:      format,
		buffer:      "",
		hasBuffered: false,
	}
}

// 判断一行是否是新日志条目的开始
func isNewLogLine(line string, format string) bool {
	// 空行不是新日志
	if strings.TrimSpace(line) == "" {
		return false
	}

	switch format {
	case "java":
		// Java 日志通常以时间戳或日志级别开头
		// 常见格式：
		// - 2024-01-01 12:00:00
		// - [2024-01-01 12:00:00]
		// - 2024-01-01T12:00:00.000Z
		// - INFO: ...
		// - [INFO] ...
		// 堆栈跟踪行通常是：
		// - 以空格或制表符开头
		// - "at " 开头
		// - "Caused by:" 开头
		// - "..." 开头（省略的堆栈）
		// - 异常类名开头（如 java.lang.NullPointerException）

		// 如果以空白字符开头，通常是续行
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			return false
		}

		// 堆栈跟踪特征
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "at ") ||
			strings.HasPrefix(trimmed, "Caused by:") ||
			strings.HasPrefix(trimmed, "Suppressed:") ||
			strings.HasPrefix(trimmed, "...") {
			return false
		}

		// 检查是否是异常类名（通常包含包名和异常类型）
		// 例如：java.lang.NullPointerException, com.example.CustomException
		// 但要排除以时间戳或日志级别开头的情况
		if strings.Contains(trimmed, "Exception") ||
			strings.Contains(trimmed, "Error") ||
			strings.Contains(trimmed, "Throwable") {
			// 如果包含异常关键词，但不以时间戳开头，认为是续行
			if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}|^\[|^\d{2}:\d{2}:\d{2}`).MatchString(line) {
				return false
			}
		}

		// 时间戳正则：匹配常见的时间格式
		timestampPatterns := []string{
			`^\d{4}-\d{2}-\d{2}`,                     // 2024-01-01
			`^\[\d{4}-\d{2}-\d{2}`,                   // [2024-01-01
			`^\d{2}:\d{2}:\d{2}`,                     // 12:00:00
			`^(INFO|DEBUG|WARN|ERROR|TRACE|FATAL)`,   // 日志级别开头
			`^\[(INFO|DEBUG|WARN|ERROR|TRACE|FATAL)`, // [INFO]
		}

		for _, pattern := range timestampPatterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				return true
			}
		}

		// 默认：如果不匹配新行特征，认为是续行（保守策略）
		return false

	case "python", "fastapi":
		// Python 日志格式类似 Java
		// 如果以空白字符开头，通常是续行
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			return false
		}

		trimmed := strings.TrimSpace(line)

		// Python 堆栈跟踪特征
		if strings.HasPrefix(trimmed, "Traceback") ||
			strings.HasPrefix(trimmed, "File ") ||
			strings.HasPrefix(trimmed, "During handling") {
			return false
		}

		// Python 异常类名（类似 Java）
		// 例如：ValueError, KeyError, sqlalchemy.exc.OperationalError
		if (strings.Contains(trimmed, "Error:") ||
			strings.Contains(trimmed, "Exception:") ||
			strings.Contains(trimmed, "Warning:")) &&
			!regexp.MustCompile(`^\d{4}-\d{2}-\d{2}|^\[`).MatchString(line) {
			return false
		}

		// 时间戳检查
		timestampPatterns := []string{
			`^\d{4}-\d{2}-\d{2}`,                     // 2024-01-01
			`^\[\d{4}-\d{2}-\d{2}`,                   // [2024-01-01
			`^\d{2}:\d{2}:\d{2}`,                     // 12:00:00
			`^(INFO|DEBUG|WARNING|ERROR|CRITICAL)`,   // 日志级别开头
			`^\[(INFO|DEBUG|WARNING|ERROR|CRITICAL)`, // [INFO]
		}

		for _, pattern := range timestampPatterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				return true
			}
		}

		// 默认：如果不匹配新行特征，认为是续行
		return false

	case "php":
		// PHP 日志通常以 [日期] 开头
		// [01-Jan-2024 12:00:00] PHP Error: ...
		if matched, _ := regexp.MatchString(`^\[[\d-]+.*?\]`, line); matched {
			return true
		}

		// 续行通常不以 [ 开头
		if len(line) > 0 && line[0] != '[' {
			return false
		}

		return true

	case "nginx":
		// Nginx 访问日志通常以 IP 地址开头
		// 192.168.1.1 - - [01/Jan/2024:12:00:00 +0000] ...
		if matched, _ := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`, line); matched {
			return true
		}

		// Nginx 错误日志以时间戳开头
		// 2024/01/01 12:00:00 [error] ...
		if matched, _ := regexp.MatchString(`^\d{4}/\d{2}/\d{2}`, line); matched {
			return true
		}

		return true

	case "ruby":
		// Ruby 日志格式类似其他语言
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			return false
		}

		// Ruby 堆栈跟踪
		if strings.Contains(line, ".rb:") && !strings.Contains(line, "[") {
			return false
		}

		if matched, _ := regexp.MatchString(`^\[|\d{4}-\d{2}-\d{2}`, line); matched {
			return true
		}

		return true

	default:
		// 默认：以时间戳或日志级别开头的认为是新行
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}|^\[|^(INFO|DEBUG|WARN|ERROR)`, line); matched {
			return true
		}
		return true
	}
}

// 添加一行到合并器
// 返回值：完整的日志行（如果有），是否有完整行
func (m *LogLineMerger) Add(line string) (string, bool) {
	// 判断这一行是否是新日志的开始
	if isNewLogLine(line, m.format) {
		// 如果缓冲区有内容，先返回缓冲区的内容
		if m.hasBuffered {
			oldBuffer := m.buffer
			m.buffer = line
			m.hasBuffered = true
			return oldBuffer, true
		} else {
			// 缓冲区为空，直接缓存这一行
			m.buffer = line
			m.hasBuffered = true
			return "", false
		}
	} else {
		// 这一行是续行，拼接到缓冲区
		if m.hasBuffered {
			m.buffer = m.buffer + "\n" + line
		} else {
			// 没有缓冲，这种情况理论上不应该发生（第一行就是续行）
			// 但为了健壮性，还是缓存它
			m.buffer = line
			m.hasBuffered = true
		}
		return "", false
	}
}

// 刷新合并器，返回缓冲区中的内容
func (m *LogLineMerger) Flush() (string, bool) {
	if m.hasBuffered {
		result := m.buffer
		m.buffer = ""
		m.hasBuffered = false
		return result, true
	}
	return "", false
}

// 安全发送邮件通知（带panic恢复和超时控制）
func safeSendEmailNotification(summary, logLine string) {
	defer func() {
		if r := recover(); r != nil {
			if *verbose || *debug {
				log.Printf("❌ 邮件通知panic恢复: %v", r)
			}
		}
	}()

	// 使用context控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用channel控制并发
	done := make(chan error, 1)
	go func() {
		done <- sendEmailNotificationWithContext(ctx, summary, logLine)
	}()

	select {
	case err := <-done:
		if err != nil && (*verbose || *debug) {
			log.Printf("❌ 邮件发送失败: %v", err)
		}
	case <-ctx.Done():
		if *verbose || *debug {
			log.Printf("❌ 邮件发送超时: %v", ctx.Err())
		}
	}
}

// 带context的邮件发送函数
func sendEmailNotificationWithContext(ctx context.Context, summary, logLine string) error {
	emailConfig := globalConfig.Notifiers.Email

	if !emailConfig.Enabled || len(emailConfig.ToEmails) == 0 {
		return nil
	}

	subject := fmt.Sprintf("⚠️ 重要日志告警: %s", summary)
	body := fmt.Sprintf(`
重要日志告警

摘要: %s

日志内容:
%s

文件: %s

时间: %s
来源: AIPipe 日志监控系统
`, summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	var err error
	if emailConfig.Provider == "resend" {
		err = sendResendEmailWithContext(ctx, emailConfig, subject, body)
	} else {
		err = sendSMTPEmailWithContext(ctx, emailConfig, subject, body)
	}

	if err != nil {
		return fmt.Errorf("邮件发送失败: %w", err)
	}

	if *verbose || *debug {
		log.Printf("✅ 邮件已发送: %s", subject)
	}
	return nil
}

// 发送邮件通知（兼容旧接口）
func sendEmailNotification(summary, logLine string) {
	ctx := context.Background()
	if err := sendEmailNotificationWithContext(ctx, summary, logLine); err != nil {
		if *verbose || *debug {
			log.Printf("❌ 邮件发送失败: %v", err)
		}
	}
}

// 带context的SMTP邮件发送
func sendSMTPEmailWithContext(ctx context.Context, config EmailConfig, subject, body string) error {
	// 检查context是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 构建邮件内容
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		config.FromEmail, strings.Join(config.ToEmails, ","), subject, body)

	// 构建SMTP地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 创建TLS配置
	tlsConfig := &tls.Config{
		ServerName: config.Host,
	}

	// 建立连接
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Quit()

	// 认证
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	// 发送邮件
	if err := client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	for _, to := range config.ToEmails {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据写入器失败: %w", err)
	}
	defer writer.Close()

	if _, err := writer.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	return nil
}

// 带context的Resend邮件发送
func sendResendEmailWithContext(ctx context.Context, config EmailConfig, subject, body string) error {
	// 检查context是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 构建请求
	payload := map[string]interface{}{
		"from":    config.FromEmail,
		"to":      config.ToEmails,
		"subject": subject,
		"html":    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Password) // 使用password字段存储API key

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API错误 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 通过SMTP发送邮件
func sendSMTPEmail(config EmailConfig, subject, body string) error {
	// 构建邮件内容
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		config.FromEmail, strings.Join(config.ToEmails, ","), subject, body)

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// 使用统一的SMTP发送方式
	err := smtp.SendMail(addr, auth, config.FromEmail, config.ToEmails, []byte(message))

	return err
}

// SSL邮件发送

// 通过Resend API发送邮件
func sendResendEmail(config EmailConfig, subject, body string) error {
	type ResendRequest struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
	}

	reqBody := ResendRequest{
		From:    config.FromEmail,
		To:      config.ToEmails,
		Subject: subject,
		Text:    body,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error: %s", string(body))
	}

	return nil
}

// 安全发送webhook通知（带panic恢复和超时控制）
func safeSendWebhookNotification(config WebhookConfig, summary, logLine, webhookType string) {
	defer func() {
		if r := recover(); r != nil {
			if *verbose || *debug {
				log.Printf("❌ %s webhook通知panic恢复: %v", webhookType, r)
			}
		}
	}()

	// 使用context控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用channel控制并发
	done := make(chan error, 1)
	go func() {
		done <- sendWebhookNotificationWithContext(ctx, config, summary, logLine, webhookType)
	}()

	select {
	case err := <-done:
		if err != nil && (*verbose || *debug) {
			log.Printf("❌ %s webhook发送失败: %v", webhookType, err)
		}
	case <-ctx.Done():
		if *verbose || *debug {
			log.Printf("❌ %s webhook发送超时: %v", webhookType, ctx.Err())
		}
	}
}

// 带context的webhook发送函数
func sendWebhookNotificationWithContext(ctx context.Context, config WebhookConfig, summary, logLine, webhookType string) error {
	if !config.Enabled || config.URL == "" {
		return nil
	}

	var payload interface{}

	// 根据webhook类型构建不同的payload
	switch webhookType {
	case "dingtalk":
		payload = buildDingTalkPayload(summary, logLine)
	case "wechat":
		payload = buildWeChatPayload(summary, logLine)
	case "feishu":
		payload = buildFeishuPayload(summary, logLine)
	case "slack":
		payload = buildSlackPayload(summary, logLine)
	default:
		payload = buildGenericPayload(summary, logLine)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建webhook payload失败: %w", err)
	}

	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建webhook请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 如果配置了签名密钥，添加签名
	if config.Secret != "" {
		addWebhookSignature(req, jsonData, config.Secret, webhookType)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("发送webhook失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook响应错误 %d: %s", resp.StatusCode, string(body))
	}

	if *verbose || *debug {
		log.Printf("✅ %s webhook已发送: %s", webhookType, summary)
	}
	return nil
}

// 发送webhook通知（兼容旧接口）
func sendWebhookNotification(config WebhookConfig, summary, logLine, webhookType string) {
	ctx := context.Background()
	if err := sendWebhookNotificationWithContext(ctx, config, summary, logLine, webhookType); err != nil {
		if *verbose || *debug {
			log.Printf("❌ %s webhook发送失败: %v", webhookType, err)
		}
	}
}

// 构建钉钉webhook payload
func buildDingTalkPayload(summary, logLine string) map[string]interface{} {
	content := fmt.Sprintf("⚠️ 重要日志告警\n\n📋 摘要: %s\n\n📝 日志内容:\n%s\n\n📁 文件: %s\n\n⏰ 时间: %s",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
}

// 构建企业微信webhook payload
func buildWeChatPayload(summary, logLine string) map[string]interface{} {
	content := fmt.Sprintf("⚠️ 重要日志告警\n\n📋 摘要: %s\n\n📝 日志内容:\n%s\n\n📁 文件: %s\n\n⏰ 时间: %s",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
}

// 构建飞书webhook payload
func buildFeishuPayload(summary, logLine string) map[string]interface{} {
	// 构建更详细的飞书通知内容
	content := fmt.Sprintf("⚠️ 重要日志告警\n\n📋 摘要: %s\n\n📝 日志内容:\n%s\n\n📁 文件: %s\n\n⏰ 时间: %s\n\n🔍 来源: AIPipe 日志监控系统",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
		},
	}
}

// 构建Slack webhook payload
func buildSlackPayload(summary, logLine string) map[string]interface{} {
	text := fmt.Sprintf("⚠️ 重要日志告警\n\n*摘要:* %s\n\n*日志内容:*\n```\n%s\n```\n\n*文件:* `%s`\n\n*时间:* %s",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"text":       text,
		"username":   "AIPipe",
		"icon_emoji": ":warning:",
	}
}

// 构建通用webhook payload
func buildGenericPayload(summary, logLine string) map[string]interface{} {
	return map[string]interface{}{
		"summary":   summary,
		"log_line":  logLine,
		"log_file":  currentLogFile,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"source":    "AIPipe",
		"level":     "warning",
	}
}

// 添加webhook签名
func addWebhookSignature(req *http.Request, body []byte, secret, webhookType string) {
	// 这里可以实现不同webhook平台的签名算法
	// 目前只是占位符实现
	switch webhookType {
	case "dingtalk":
		// 钉钉签名实现
		// req.Header.Set("X-DingTalk-Signature", signature)
	case "wechat":
		// 企业微信签名实现
		// req.Header.Set("X-WeChat-Signature", signature)
	case "feishu":
		// 飞书签名实现
		// req.Header.Set("X-Feishu-Signature", signature)
	case "slack":
		// Slack签名实现
		// req.Header.Set("X-Slack-Signature", signature)
	default:
		// 通用签名
		// req.Header.Set("X-Webhook-Signature", signature)
	}
}

// 智能识别webhook类型
func detectWebhookType(webhookURL string) string {
	u, err := url.Parse(webhookURL)
	if err != nil {
		return "custom"
	}

	host := strings.ToLower(u.Host)
	path := strings.ToLower(u.Path)

	// 钉钉
	if strings.Contains(host, "dingtalk") || strings.Contains(path, "dingtalk") {
		return "dingtalk"
	}

	// 企业微信
	if strings.Contains(host, "qyapi.weixin.qq.com") || strings.Contains(path, "wechat") {
		return "wechat"
	}

	// 飞书
	if strings.Contains(host, "feishu") || strings.Contains(path, "feishu") {
		return "feishu"
	}

	// Slack
	if strings.Contains(host, "slack.com") || strings.Contains(path, "slack") {
		return "slack"
	}

	return "custom"
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

// 保存规则到配置文件
func saveRulesToConfig() error {
	// 获取当前规则
	rules := ruleEngine.GetRules()

	// 更新全局配置
	globalConfig.Rules = rules

	// 获取配置文件路径
	configPath, err := findDefaultConfig()
	if err != nil {
		return fmt.Errorf("查找配置文件失败: %w", err)
	}

	// 读取现有配置
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析现有配置
	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 更新规则
	config["rules"] = rules

	// 保存配置
	updatedData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
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

// 获取AI分析结果
func (cm *CacheManager) GetAIAnalysis(logHash string) (*AIAnalysisCache, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	item, exists := cm.aiCache[logHash]
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

// 设置AI分析结果
func (cm *CacheManager) SetAIAnalysis(logHash string, result *AIAnalysisCache) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	// 检查是否需要清理空间
	if cm.needsEviction() {
		cm.evictOldest()
	}
	
	cm.aiCache[logHash] = result
	cm.updateStats()
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
		Key:        key,
		Value:      value,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
		AccessCount: 0,
		Size:       cm.calculateItemSize(value),
	}
	
	cm.configCache[key] = item
	cm.updateStats()
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

// 生成日志哈希
func generateLogHash(logLine string) string {
	hash := sha256.Sum256([]byte(logLine))
	return fmt.Sprintf("%x", hash)
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
		"test": "value",
		"number": 123,
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
