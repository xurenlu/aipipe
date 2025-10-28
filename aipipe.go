package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// 主配置结构
type Config struct {
	AIEndpoint   string `json:"ai_endpoint"`
	Token        string `json:"token"`
	Model        string `json:"model"`
	CustomPrompt string `json:"custom_prompt"`
	MaxRetries   int    `json:"max_retries"`
	Timeout      int    `json:"timeout"`
	RateLimit    int    `json:"rate_limit"`
	LocalFilter  bool   `json:"local_filter"`
}

// 工作协程
type Worker struct {
	ID            int
	ProcessedJobs int64
	TotalTime     time.Duration
	AverageTime   time.Duration
	ErrorCount    int64
	LastActivity  time.Time
	CurrentLoad   int64
	IsHealthy     bool
}

// 处理任务
type ProcessingJob struct {
	ID       string
	Data     string
	Priority TaskPriority
	Created  time.Time
}

// I/O统计
type IOStats struct {
	ReadBytes   int64
	WriteBytes  int64
	ReadOps     int64
	WriteOps    int64
	ReadTime    time.Duration
	WriteTime   time.Duration
	CacheHits   int64
	CacheMisses int64
}

// 并发统计
type ConcurrencyStats struct {
	ActiveWorkers   int
	TotalJobs       int64
	CompletedJobs   int64
	FailedJobs      int64
	AverageWaitTime time.Duration
	AverageProcTime time.Duration
	Throughput      float64
	ErrorRate       float64
}

// 配置验证器
type ConfigValidator struct {
	errors []ConfigValidationError
}

// 配置验证错误
type ConfigValidationError struct {
	Field   string
	Message string
}

// AI服务
type AIService struct {
	Name     string
	Endpoint string
	Token    string
	Model    string
	Enabled  bool
	Priority int
}

// AI服务管理器
type AIServiceManager struct {
	services    []AIService
	current     int
	fallback    bool
	rateLimiter map[string]time.Time
	mutex       sync.RWMutex
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

// 性能指标
type PerformanceMetrics struct {
	ProcessedLines int64     `json:"processed_lines"`
	FilteredLines  int64     `json:"filtered_lines"`
	AlertedLines   int64     `json:"alerted_lines"`
	APICalls       int64     `json:"api_calls"`
	ProcessingTime int64     `json:"processing_time_ms"`
	ErrorCount     int64     `json:"error_count"`
	CacheHits      int64     `json:"cache_hits"`
	CacheMisses    int64     `json:"cache_misses"`
	MemoryUsage    int64     `json:"memory_usage_bytes"`
	LastUpdated    time.Time `json:"last_updated"`
	Throughput     float64   `json:"throughput"`      // 每秒处理行数
	AverageLatency float64   `json:"average_latency"` // 平均延迟(ms)
	ErrorRate      float64   `json:"error_rate"`      // 错误率
	CacheHitRate   float64   `json:"cache_hit_rate"`  // 缓存命中率
}

// 内存优化相关结构

// 内存配置
type MemoryConfig struct {
	MaxMemoryUsage      int64         `json:"max_memory_usage"`      // 最大内存使用量（字节）
	GCThreshold         int64         `json:"gc_threshold"`          // 垃圾回收阈值
	StreamBufferSize    int           `json:"stream_buffer_size"`    // 流式处理缓冲区大小
	ChunkSize           int           `json:"chunk_size"`            // 分块处理大小
	MemoryCheckInterval time.Duration `json:"memory_check_interval"` // 内存检查间隔
	AutoGC              bool          `json:"auto_gc"`               // 自动垃圾回收
	MemoryLimit         int64         `json:"memory_limit"`          // 内存限制
	Enabled             bool          `json:"enabled"`               // 是否启用内存优化
}

// 内存统计
type MemoryStats struct {
	CurrentUsage   int64     `json:"current_usage"`   // 当前内存使用量
	PeakUsage      int64     `json:"peak_usage"`      // 峰值内存使用量
	GCCount        int64     `json:"gc_count"`        // 垃圾回收次数
	GCTime         int64     `json:"gc_time"`         // 垃圾回收时间（纳秒）
	AllocCount     int64     `json:"alloc_count"`     // 分配次数
	FreeCount      int64     `json:"free_count"`      // 释放次数
	HeapSize       int64     `json:"heap_size"`       // 堆大小
	StackSize      int64     `json:"stack_size"`      // 栈大小
	LastGC         time.Time `json:"last_gc"`         // 上次垃圾回收时间
	MemoryPressure float64   `json:"memory_pressure"` // 内存压力（0-1）
}

// 流式处理器
type StreamProcessor struct {
	BufferSize     int
	ChunkSize      int
	ProcessFunc    func([]string) error
	Buffer         []string
	TotalProcessed int64
	mutex          sync.Mutex
}

// 并发处理相关结构

// 并发控制配置
type ConcurrencyConfig struct {
	MaxConcurrency        int           `json:"max_concurrency"`        // 最大并发数
	BackpressureThreshold int           `json:"backpressure_threshold"` // 背压阈值
	LoadBalanceStrategy   string        `json:"load_balance_strategy"`  // 负载均衡策略
	AdaptiveScaling       bool          `json:"adaptive_scaling"`       // 自适应扩缩容
	ScaleUpThreshold      float64       `json:"scale_up_threshold"`     // 扩容阈值
	ScaleDownThreshold    float64       `json:"scale_down_threshold"`   // 缩容阈值
	MinWorkers            int           `json:"min_workers"`            // 最小工作协程数
	MaxWorkers            int           `json:"max_workers"`            // 最大工作协程数
	ScalingInterval       time.Duration `json:"scaling_interval"`       // 扩缩容检查间隔
	Enabled               bool          `json:"enabled"`                // 是否启用并发控制
}

// 背压控制器
type BackpressureController struct {
	threshold     int
	currentLoad   int64
	blockedCount  int64
	rejectedCount int64
	mutex         sync.RWMutex
	callbacks     []func(int64)
}

// 负载均衡器
type LoadBalancer struct {
	strategy     string
	workers      []*Worker
	currentIndex int
	workerStats  map[int]*WorkerStats
	mutex        sync.RWMutex
}

// 工作协程统计
type WorkerStats struct {
	ID            int           `json:"id"`
	ProcessedJobs int64         `json:"processed_jobs"`
	TotalTime     time.Duration `json:"total_time"`
	AverageTime   time.Duration `json:"average_time"`
	ErrorCount    int64         `json:"error_count"`
	LastActivity  time.Time     `json:"last_activity"`
	CurrentLoad   int64         `json:"current_load"`
	IsHealthy     bool          `json:"is_healthy"`
}

// 自适应扩缩容器
type AdaptiveScaler struct {
	config         ConcurrencyConfig
	currentWorkers int
	workerStats    map[int]*WorkerStats
	lastScaleTime  time.Time
	mutex          sync.RWMutex
}

// 任务优先级
type TaskPriority int

const (
	PriorityLow      TaskPriority = 1
	PriorityNormal   TaskPriority = 2
	PriorityHigh     TaskPriority = 3
	PriorityCritical TaskPriority = 4
)

// 优先级队列
type PriorityQueue struct {
	jobs       []ProcessingJob
	priorities map[string]TaskPriority
	mutex      sync.RWMutex
}

// I/O优化相关结构

// I/O配置
type IOConfig struct {
	BufferSize       int           `json:"buffer_size"`       // 缓冲区大小
	BatchSize        int           `json:"batch_size"`        // 批处理大小
	FlushInterval    time.Duration `json:"flush_interval"`    // 刷新间隔
	AsyncIO          bool          `json:"async_io"`          // 异步I/O
	ReadAhead        int           `json:"read_ahead"`        // 预读大小
	WriteBehind      bool          `json:"write_behind"`      // 写后置
	Compression      bool          `json:"compression"`       // 压缩
	CompressionLevel int           `json:"compression_level"` // 压缩级别
	CacheSize        int64         `json:"cache_size"`        // 缓存大小
	Enabled          bool          `json:"enabled"`           // 是否启用I/O优化
}

// 文件监控器
type FileMonitor struct {
	filePath  string
	lastSize  int64
	lastMod   time.Time
	watcher   *fsnotify.Watcher
	callbacks []func(string, []byte)
	mutex     sync.RWMutex
	stopChan  chan bool
}

// 压缩器
type Compressor struct {
	level      int
	algorithm  string
	compressed map[string][]byte
	mutex      sync.RWMutex
}

// 缓存管理器
type IOCacheManager struct {
	cache       map[string][]byte
	maxSize     int64
	currentSize int64
	stats       IOStats
	mutex       sync.RWMutex
}

// 任务调度器
type TaskScheduler struct {
	priorityQueue *PriorityQueue
	workers       []*Worker
	loadBalancer  *LoadBalancer
	stats         ConcurrencyStats
	mutex         sync.RWMutex
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

// 全局工作池管理器
var workerPool *WorkerPool

// 全局内存管理器
var memoryManager *MemoryManager

// 全局并发控制器
var concurrencyController *ConcurrencyController

// 全局I/O优化器
var ioOptimizer *IOOptimizer

// 全局配置向导
var configWizard *ConfigWizard

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
	Line         string  `json:"line"`      // 日志行内容
	Important    bool    `json:"important"` // 是否重要
	ShouldFilter bool    `json:"should_filter"`
	Summary      string  `json:"summary"`
	Reason       string  `json:"reason"`
	Confidence   float64 `json:"confidence"` // 置信度
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
	cacheStats = flag.Bool("cache-stats", false, "显示缓存统计信息")
	cacheClear = flag.Bool("cache-clear", false, "清空所有缓存")
	cacheTest  = flag.Bool("cache-test", false, "测试缓存功能")

	// 工作池管理命令
	workerStats      = flag.Bool("worker-stats", false, "显示工作池统计信息")
	workerTest       = flag.Bool("worker-test", false, "测试工作池功能")
	performanceStats = flag.Bool("perf-stats", false, "显示性能指标")

	// 内存管理命令
	memoryStats = flag.Bool("memory-stats", false, "显示内存统计信息")
	memoryTest  = flag.Bool("memory-test", false, "测试内存管理功能")
	memoryGC    = flag.Bool("memory-gc", false, "强制垃圾回收")

	// 并发控制命令
	concurrencyStats = flag.Bool("concurrency-stats", false, "显示并发控制统计信息")
	concurrencyTest  = flag.Bool("concurrency-test", false, "测试并发控制功能")
	backpressureTest = flag.Bool("backpressure-test", false, "测试背压控制功能")

	// I/O管理命令
	ioStats = flag.Bool("io-stats", false, "显示I/O统计信息")
	ioTest  = flag.Bool("io-test", false, "测试I/O优化功能")
	ioFlush = flag.Bool("io-flush", false, "强制刷新I/O缓冲区")

	// 用户体验命令
	configInit     = flag.Bool("config-init", false, "启动配置向导")
	configTemplate = flag.Bool("config-template", false, "显示配置模板")
	outputFormat   = flag.String("output-format", "", "输出格式 (json, csv, table, custom)")
	outputColor    = flag.Bool("output-color", true, "启用颜色输出")
	logLevel       = flag.String("log-level", "", "日志级别 (debug, info, warn, error, fatal)")

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

	if *workerStats {
		handleWorkerStats()
		return
	}

	if *workerTest {
		handleWorkerTest()
		return
	}

	if *performanceStats {
		handlePerformanceStats()
		return
	}

	if *memoryStats {
		handleMemoryStats()
		return
	}

	if *memoryTest {
		handleMemoryTest()
		return
	}

	if *memoryGC {
		handleMemoryGC()
		return
	}

	if *concurrencyStats {
		handleConcurrencyStats()
		return
	}

	if *concurrencyTest {
		handleConcurrencyTest()
		return
	}

	if *backpressureTest {
		handleBackpressureTest()
		return
	}

	if *ioStats {
		handleIOStats()
		return
	}

	if *ioTest {
		handleIOTest()
		return
	}

	if *ioFlush {
		handleIOFlush()
		return
	}

	if *configInit {
		handleConfigInit()
		return
	}

	if *configTemplate {
		handleConfigTemplate()
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

// 生成日志哈希
func generateLogHash(logLine string) string {
	hash := sha256.Sum256([]byte(logLine))
	return fmt.Sprintf("%x", hash)
}

// 分析单行日志（工作池使用）
func analyzeLogLine(line, format string) (*LogAnalysis, error) {
	analysis, err := analyzeLog(line, format)
	if err != nil {
		return nil, err
	}

	// 设置行内容
	analysis.Line = line

	// 根据ShouldFilter设置Important
	analysis.Important = !analysis.ShouldFilter

	// 设置默认置信度
	if analysis.Confidence == 0 {
		analysis.Confidence = 0.8
	}

	return analysis, nil
}

// 显示性能指标
func handlePerformanceStats() {
	fmt.Println("📈 性能指标:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 获取缓存统计
	cacheStats := cacheManager.GetStats()

	// 获取工作池统计
	workerStats := workerPool.GetStats()

	// 计算内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("处理统计:")
	fmt.Printf("  总处理行数: %d\n", workerStats.TotalLines)
	fmt.Printf("  完成任务数: %d\n", workerStats.CompletedJobs)
	fmt.Printf("  失败任务数: %d\n", workerStats.FailedJobs)
	fmt.Printf("  错误率: %.2f%%\n", workerStats.ErrorRate)

	fmt.Println("\n性能指标:")
	fmt.Printf("  吞吐量: %.2f 行/秒\n", workerStats.Throughput)
	fmt.Printf("  平均处理时间: %v\n", workerStats.AverageTime)
	fmt.Printf("  活跃工作协程: %d\n", workerStats.ActiveWorkers)

	fmt.Println("\n缓存统计:")
	fmt.Printf("  缓存命中次数: %d\n", cacheStats.HitCount)
	fmt.Printf("  缓存未命中次数: %d\n", cacheStats.MissCount)
	fmt.Printf("  缓存命中率: %.2f%%\n", cacheStats.HitRate)
	fmt.Printf("  总缓存项数: %d\n", cacheStats.TotalItems)

	fmt.Println("\n内存使用:")
	fmt.Printf("  当前内存使用: %.2f MB\n", float64(m.Alloc)/(1024*1024))
	fmt.Printf("  系统内存使用: %.2f MB\n", float64(m.Sys)/(1024*1024))
	fmt.Printf("  垃圾回收次数: %d\n", m.NumGC)
	fmt.Printf("  垃圾回收时间: %v\n", time.Duration(m.PauseTotalNs))

	fmt.Println("\n系统信息:")
	fmt.Printf("  Go版本: %s\n", runtime.Version())
	fmt.Printf("  CPU核心数: %d\n", runtime.NumCPU())
	fmt.Printf("  Goroutine数: %d\n", runtime.NumGoroutine())
}

// 用户体验相关结构

// 配置向导
type ConfigWizard struct {
	steps       []WizardStep
	currentStep int
	config      Config
	responses   map[string]interface{}
	mutex       sync.RWMutex
}

// 向导步骤
type WizardStep struct {
	ID          string                  `json:"id"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Type        string                  `json:"type"` // input, select, confirm, file
	Options     []WizardOption          `json:"options"`
	Required    bool                    `json:"required"`
	Default     interface{}             `json:"default"`
	Validation  func(interface{}) error `json:"-"`
}

// 向导选项
type WizardOption struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// 配置模板
type ConfigTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Config      Config `json:"config"`
	Category    string `json:"category"`
}

// 颜色支持
type ColorSupport struct {
	Enabled bool
	Colors  map[string]string
}

// 交互式提示
type InteractivePrompt struct {
	message   string
	options   []string
	validator func(string) error
}

// 配置向导实现

// 创建新的配置向导
func NewConfigWizard() *ConfigWizard {
	wizard := &ConfigWizard{
		steps:       make([]WizardStep, 0),
		currentStep: 0,
		config:      defaultConfig,
		responses:   make(map[string]interface{}),
	}

	// 初始化向导步骤
	wizard.initSteps()

	return wizard
}

// 初始化向导步骤
func (w *ConfigWizard) initSteps() {
	w.steps = []WizardStep{
		{
			ID:          "ai_endpoint",
			Title:       "AI服务端点配置",
			Description: "请输入AI服务的API端点URL",
			Type:        "input",
			Required:    true,
			Default:     "https://your-ai-server.com/api/v1/chat/completions",
			Validation:  validateURL,
		},
		{
			ID:          "ai_token",
			Title:       "API Token配置",
			Description: "请输入AI服务的API Token",
			Type:        "input",
			Required:    true,
			Default:     "your-api-token-here",
			Validation:  validateToken,
		},
		{
			ID:          "ai_model",
			Title:       "AI模型选择",
			Description: "请选择要使用的AI模型",
			Type:        "select",
			Required:    true,
			Default:     "gpt-4",
			Options: []WizardOption{
				{Value: "gpt-4", Label: "GPT-4", Description: "OpenAI GPT-4模型"},
				{Value: "gpt-3.5-turbo", Label: "GPT-3.5 Turbo", Description: "OpenAI GPT-3.5 Turbo模型"},
				{Value: "claude-3", Label: "Claude 3", Description: "Anthropic Claude 3模型"},
				{Value: "gemini-pro", Label: "Gemini Pro", Description: "Google Gemini Pro模型"},
			},
		},
		{
			ID:          "output_format",
			Title:       "输出格式选择",
			Description: "请选择日志输出格式",
			Type:        "select",
			Required:    true,
			Default:     "table",
			Options: []WizardOption{
				{Value: "table", Label: "表格格式", Description: "易读的表格格式"},
				{Value: "json", Label: "JSON格式", Description: "机器可读的JSON格式"},
				{Value: "csv", Label: "CSV格式", Description: "逗号分隔值格式"},
				{Value: "custom", Label: "自定义格式", Description: "使用自定义模板"},
			},
		},
		{
			ID:          "log_level",
			Title:       "日志级别配置",
			Description: "请选择要监控的日志级别",
			Type:        "select",
			Required:    true,
			Default:     "info",
			Options: []WizardOption{
				{Value: "debug", Label: "DEBUG", Description: "显示所有日志级别"},
				{Value: "info", Label: "INFO", Description: "显示信息级别及以上"},
				{Value: "warn", Label: "WARN", Description: "显示警告级别及以上"},
				{Value: "error", Label: "ERROR", Description: "只显示错误级别"},
				{Value: "fatal", Label: "FATAL", Description: "只显示致命错误"},
			},
		},
		{
			ID:          "enable_features",
			Title:       "功能启用配置",
			Description: "请选择要启用的功能",
			Type:        "select",
			Required:    false,
			Default:     "basic",
			Options: []WizardOption{
				{Value: "basic", Label: "基础功能", Description: "只启用基本日志分析功能"},
				{Value: "advanced", Label: "高级功能", Description: "启用所有高级功能"},
				{Value: "enterprise", Label: "企业功能", Description: "启用企业级功能"},
			},
		},
		{
			ID:          "confirm_config",
			Title:       "配置确认",
			Description: "请确认配置是否正确",
			Type:        "confirm",
			Required:    true,
			Default:     true,
		},
	}
}

// 启动配置向导
func (w *ConfigWizard) Start() error {
	fmt.Println("🎯 AIPipe 配置向导")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("欢迎使用AIPipe配置向导！")
	fmt.Println("我们将引导您完成基本配置。")
	fmt.Println()

	for w.currentStep < len(w.steps) {
		step := w.steps[w.currentStep]

		fmt.Printf("步骤 %d/%d: %s\n", w.currentStep+1, len(w.steps), step.Title)
		fmt.Printf("描述: %s\n", step.Description)
		fmt.Println()

		response, err := w.promptStep(step)
		if err != nil {
			return fmt.Errorf("步骤 %d 输入错误: %v", w.currentStep+1, err)
		}

		// 验证输入
		if step.Validation != nil {
			if err := step.Validation(response); err != nil {
				fmt.Printf("❌ 验证失败: %v\n", err)
				fmt.Println("请重新输入。")
				continue
			}
		}

		// 保存响应
		w.responses[step.ID] = response

		// 更新配置
		w.updateConfig(step.ID, response)

		fmt.Println("✅ 配置已保存")
		fmt.Println()

		w.currentStep++
	}

	// 保存配置文件
	if err := w.saveConfig(); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	fmt.Println("🎉 配置向导完成！")
	fmt.Println("配置文件已保存到 ~/.config/aipipe.json")
	fmt.Println("您现在可以使用 AIPipe 了！")

	return nil
}

// 提示用户输入
func (w *ConfigWizard) promptStep(step WizardStep) (interface{}, error) {
	switch step.Type {
	case "input":
		return w.promptInput(step)
	case "select":
		return w.promptSelect(step)
	case "confirm":
		return w.promptConfirm(step)
	case "file":
		return w.promptFile(step)
	default:
		return nil, fmt.Errorf("不支持的步骤类型: %s", step.Type)
	}
}

// 输入提示
func (w *ConfigWizard) promptInput(step WizardStep) (string, error) {
	prompt := fmt.Sprintf("请输入 %s", step.Title)
	if step.Default != nil {
		prompt += fmt.Sprintf(" (默认: %v)", step.Default)
	}
	prompt += ": "

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" && step.Default != nil {
		return step.Default.(string), nil
	}

	return input, nil
}

// 选择提示
func (w *ConfigWizard) promptSelect(step WizardStep) (string, error) {
	fmt.Println("请选择:")
	for i, option := range step.Options {
		fmt.Printf("  %d. %s - %s\n", i+1, option.Label, option.Description)
	}

	prompt := "请输入选项编号"
	if step.Default != nil {
		prompt += fmt.Sprintf(" (默认: %v)", step.Default)
	}
	prompt += ": "

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" && step.Default != nil {
		return step.Default.(string), nil
	}

	// 解析选择
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(step.Options) {
		return "", fmt.Errorf("无效的选择: %s", input)
	}

	return step.Options[choice-1].Value, nil
}

// 确认提示
func (w *ConfigWizard) promptConfirm(step WizardStep) (bool, error) {
	prompt := "是否确认"
	if step.Default != nil {
		prompt += fmt.Sprintf(" (默认: %v)", step.Default)
	}
	prompt += " [y/N]: "

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" && step.Default != nil {
		return step.Default.(bool), nil
	}

	return input == "y" || input == "yes", nil
}

// 文件提示
func (w *ConfigWizard) promptFile(step WizardStep) (string, error) {
	prompt := fmt.Sprintf("请输入文件路径")
	if step.Default != nil {
		prompt += fmt.Sprintf(" (默认: %v)", step.Default)
	}
	prompt += ": "

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" && step.Default != nil {
		return step.Default.(string), nil
	}

	// 验证文件是否存在
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", input)
	}

	return input, nil
}

// 更新配置
func (w *ConfigWizard) updateConfig(stepID string, response interface{}) {
	switch stepID {
	case "ai_endpoint":
		w.config.AIEndpoint = response.(string)
	case "ai_token":
		w.config.Token = response.(string)
	case "ai_model":
		w.config.Model = response.(string)
	case "output_format":
		w.config.OutputFormat.Type = response.(string)
	case "log_level":
		w.config.LogLevel.Level = response.(string)
		w.updateLogLevelConfig(response.(string))
	case "enable_features":
		w.updateFeatureConfig(response.(string))
	}
}

// 更新日志级别配置
func (w *ConfigWizard) updateLogLevelConfig(level string) {
	switch level {
	case "debug":
		w.config.LogLevel.ShowDebug = true
		w.config.LogLevel.ShowInfo = true
		w.config.LogLevel.ShowWarn = true
		w.config.LogLevel.ShowError = true
		w.config.LogLevel.ShowFatal = true
		w.config.LogLevel.MinLevel = "debug"
	case "info":
		w.config.LogLevel.ShowDebug = false
		w.config.LogLevel.ShowInfo = true
		w.config.LogLevel.ShowWarn = true
		w.config.LogLevel.ShowError = true
		w.config.LogLevel.ShowFatal = true
		w.config.LogLevel.MinLevel = "info"
	case "warn":
		w.config.LogLevel.ShowDebug = false
		w.config.LogLevel.ShowInfo = false
		w.config.LogLevel.ShowWarn = true
		w.config.LogLevel.ShowError = true
		w.config.LogLevel.ShowFatal = true
		w.config.LogLevel.MinLevel = "warn"
	case "error":
		w.config.LogLevel.ShowDebug = false
		w.config.LogLevel.ShowInfo = false
		w.config.LogLevel.ShowWarn = false
		w.config.LogLevel.ShowError = true
		w.config.LogLevel.ShowFatal = true
		w.config.LogLevel.MinLevel = "error"
	case "fatal":
		w.config.LogLevel.ShowDebug = false
		w.config.LogLevel.ShowInfo = false
		w.config.LogLevel.ShowWarn = false
		w.config.LogLevel.ShowError = false
		w.config.LogLevel.ShowFatal = true
		w.config.LogLevel.MinLevel = "fatal"
	}
}

// 更新功能配置
func (w *ConfigWizard) updateFeatureConfig(feature string) {
	switch feature {
	case "basic":
		// 基础功能：只启用基本配置
		w.config.WorkerPool.Enabled = false
		w.config.Memory.Enabled = false
		w.config.Concurrency.Enabled = false
		w.config.IO.Enabled = false
	case "advanced":
		// 高级功能：启用所有功能
		w.config.WorkerPool.Enabled = true
		w.config.Memory.Enabled = true
		w.config.Concurrency.Enabled = true
		w.config.IO.Enabled = true
	case "enterprise":
		// 企业功能：启用所有功能并优化配置
		w.config.WorkerPool.Enabled = true
		w.config.Memory.Enabled = true
		w.config.Concurrency.Enabled = true
		w.config.IO.Enabled = true
		// 优化企业级配置
		w.config.WorkerPool.MaxWorkers = 8
		w.config.Memory.MaxMemoryUsage = 2 * 1024 * 1024 * 1024 // 2GB
		w.config.Concurrency.MaxConcurrency = 200
		w.config.IO.BufferSize = 128 * 1024 // 128KB
	}
}

// 保存配置文件
func (w *ConfigWizard) saveConfig() error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "aipipe.json")

	data, err := json.MarshalIndent(w.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// 验证URL函数
func validateURL(value interface{}) error {
	url, ok := value.(string)
	if !ok {
		return fmt.Errorf("URL必须是字符串")
	}

	if url == "" {
		return fmt.Errorf("URL不能为空")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL必须以http://或https://开头")
	}

	return nil
}

// 验证Token函数
func validateToken(value interface{}) error {
	token, ok := value.(string)
	if !ok {
		return fmt.Errorf("Token必须是字符串")
	}

	if token == "" {
		return fmt.Errorf("Token不能为空")
	}

	if len(token) < 10 {
		return fmt.Errorf("Token长度至少10个字符")
	}

	return nil
}
