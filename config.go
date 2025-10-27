package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
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

// 输出格式配置
type OutputFormat struct {
	Type     string `json:"type"`     // json, csv, table, custom
	Template string `json:"template"` // 自定义模板
	Color    bool   `json:"color"`    // 颜色支持
	Filter   string `json:"filter"`   // 输出过滤
	Width    int    `json:"width"`    // 表格宽度
	Headers  bool   `json:"headers"`  // 显示表头
}

// 日志级别配置
type LogLevelConfig struct {
	Level     string `json:"level"`      // debug, info, warn, error, fatal
	ShowDebug bool   `json:"show_debug"` // 显示调试日志
	ShowInfo  bool   `json:"show_info"`  // 显示信息日志
	ShowWarn  bool   `json:"show_warn"`  // 显示警告日志
	ShowError bool   `json:"show_error"` // 显示错误日志
	ShowFatal bool   `json:"show_fatal"` // 显示致命日志
	MinLevel  string `json:"min_level"`  // 最小日志级别
	MaxLevel  string `json:"max_level"`  // 最大日志级别
	Enabled   bool   `json:"enabled"`    // 是否启用日志级别过滤
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

	// 工作池配置
	WorkerPool WorkerPoolConfig `json:"worker_pool"` // 工作池配置

	// 内存优化配置
	Memory MemoryConfig `json:"memory"` // 内存优化配置

	// 并发控制配置
	Concurrency ConcurrencyConfig `json:"concurrency"` // 并发控制配置

	// I/O优化配置
	IO IOConfig `json:"io"`

	// 用户体验配置
	OutputFormat OutputFormat   `json:"output_format"`
	LogLevel     LogLevelConfig `json:"log_level"`
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

// 处理配置向导
func handleConfigInit() {
	fmt.Println("🎯 启动配置向导...")
	wizard := NewConfigWizard()
	if err := wizard.Start(); err != nil {
		fmt.Printf("❌ 配置向导失败: %v\n", err)
		os.Exit(1)
	}
}

// 处理配置模板
func handleConfigTemplate() {
	fmt.Println("📋 配置模板:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 显示示例配置
	template := Config{
		AIEndpoint:   "https://your-ai-server.com/api/v1/chat/completions",
		Token:        "your-api-token-here",
		Model:        "gpt-4",
		CustomPrompt: "你的自定义提示词",
		MaxRetries:   3,
		Timeout:      30,
		RateLimit:    100,
		LocalFilter:  true,
		OutputFormat: OutputFormat{
			Type:    "table",
			Color:   true,
			Width:   120,
			Headers: true,
		},
		LogLevel: LogLevelConfig{
			Level:     "info",
			ShowInfo:  true,
			ShowWarn:  true,
			ShowError: true,
			ShowFatal: true,
			MinLevel:  "info",
			MaxLevel:  "fatal",
			Enabled:   true,
		},
	}

	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		fmt.Printf("❌ 生成模板失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
	fmt.Println("\n💡 提示:")
	fmt.Println("1. 将上述配置保存到 ~/.config/aipipe.json")
	fmt.Println("2. 修改 AIEndpoint、Token 和 Model 为你的实际值")
	fmt.Println("3. 使用 --config-init 启动交互式配置向导")
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

// 处理配置测试
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

// 查找多源配置文件
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
	if len(globalConfig.Token) > 16 {
		fmt.Printf("Token: %s...%s\n", globalConfig.Token[:8], globalConfig.Token[len(globalConfig.Token)-8:])
	} else {
		fmt.Printf("Token: %s\n", strings.Repeat("*", len(globalConfig.Token)))
	}
	fmt.Printf("最大重试次数: %d\n", globalConfig.MaxRetries)
	fmt.Printf("超时时间: %d 秒\n", globalConfig.Timeout)
	fmt.Printf("频率限制: %d 次/分钟\n", globalConfig.RateLimit)
	fmt.Printf("本地过滤: %t\n", globalConfig.LocalFilter)

	if globalConfig.CustomPrompt != "" {
		fmt.Printf("自定义提示词: %s\n", globalConfig.CustomPrompt)
	}
}

// 默认配置变量
var defaultConfig Config

// 初始化默认配置
func init() {
	defaultConfig = Config{
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
		WorkerPool: WorkerPoolConfig{
			MaxWorkers:   4,
			QueueSize:    100,
			BatchSize:    10,
			Timeout:      30 * time.Second,
			RetryCount:   3,
			BackoffDelay: 1 * time.Second,
			Enabled:      true,
		},
		Memory: MemoryConfig{
			MaxMemoryUsage:      512 * 1024 * 1024, // 512MB
			GCThreshold:         256 * 1024 * 1024, // 256MB
			StreamBufferSize:    1000,
			ChunkSize:           100,
			MemoryCheckInterval: 5 * time.Second,
			AutoGC:              true,
			MemoryLimit:         1024 * 1024 * 1024, // 1GB
			Enabled:             true,
		},
		Concurrency: ConcurrencyConfig{
			MaxConcurrency:        100,
			BackpressureThreshold: 80,
			LoadBalanceStrategy:   "round_robin",
			AdaptiveScaling:       true,
			ScaleUpThreshold:      0.8,
			ScaleDownThreshold:    0.3,
			MinWorkers:            2,
			MaxWorkers:            20,
			ScalingInterval:       30 * time.Second,
			Enabled:               true,
		},
		IO: IOConfig{
			BufferSize:       64 * 1024, // 64KB
			BatchSize:        1000,
			FlushInterval:    5 * time.Second,
			AsyncIO:          true,
			ReadAhead:        32 * 1024, // 32KB
			WriteBehind:      true,
			Compression:      false,
			CompressionLevel: 6,
			CacheSize:        10 * 1024 * 1024, // 10MB
			Enabled:          true,
		},
		OutputFormat: OutputFormat{
			Type:     "table",
			Template: "",
			Color:    true,
			Filter:   "",
			Width:    120,
			Headers:  true,
		},
		LogLevel: LogLevelConfig{
			Level:     "info",
			ShowDebug: false,
			ShowInfo:  true,
			ShowWarn:  true,
			ShowError: true,
			ShowFatal: true,
			MinLevel:  "info",
			MaxLevel:  "fatal",
			Enabled:   true,
		},
	}
}

// 简化版 loadConfig 函数
func loadConfig() error {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "aipipe.json")

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		globalConfig = defaultConfig
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON配置
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 合并默认配置
	globalConfig = mergeConfig(defaultConfig, config)

	return nil
}

// 合并配置
func mergeConfig(defaultConfig, userConfig Config) Config {
	merged := defaultConfig

	// 合并基本配置
	if userConfig.AIEndpoint != "" {
		merged.AIEndpoint = userConfig.AIEndpoint
	}
	if userConfig.Token != "" {
		merged.Token = userConfig.Token
	}
	if userConfig.Model != "" {
		merged.Model = userConfig.Model
	}
	if userConfig.CustomPrompt != "" {
		merged.CustomPrompt = userConfig.CustomPrompt
	}

	// 合并其他配置项
	if userConfig.MaxRetries > 0 {
		merged.MaxRetries = userConfig.MaxRetries
	}
	if userConfig.Timeout > 0 {
		merged.Timeout = userConfig.Timeout
	}
	if userConfig.RateLimit > 0 {
		merged.RateLimit = userConfig.RateLimit
	}
	merged.LocalFilter = userConfig.LocalFilter

	// 合并AI服务列表
	if len(userConfig.AIServices) > 0 {
		merged.AIServices = userConfig.AIServices
	}
	if userConfig.DefaultAI != "" {
		merged.DefaultAI = userConfig.DefaultAI
	}

	// 合并规则列表
	if len(userConfig.Rules) > 0 {
		merged.Rules = userConfig.Rules
	}

	// 合并其他子配置
	merged.Cache = userConfig.Cache
	merged.WorkerPool = userConfig.WorkerPool
	merged.Memory = userConfig.Memory
	merged.Concurrency = userConfig.Concurrency
	merged.IO = userConfig.IO
	merged.OutputFormat = userConfig.OutputFormat
	merged.LogLevel = userConfig.LogLevel

	return merged
}

// 加载多源配置文件
func loadMultiSourceConfig(configPath string) (*MultiSourceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config MultiSourceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// 查找默认配置文件
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

	// 没有找到，返回默认路径
	return filepath.Join(configDir, "aipipe.json"), nil
}

// 检测配置文件格式
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
		return "json" // 默认格式
	}
}

// 解析配置文件
func parseConfigFile(data []byte, format string, target interface{}) error {
	switch format {
	case "json":
		return json.Unmarshal(data, target)
	case "yaml":
		// 如果有 yaml 包，使用它
		// 否则只支持 JSON
		return json.Unmarshal(data, target)
	case "toml":
		// 如果有 toml 包，使用它
		// 否则只支持 JSON
		return json.Unmarshal(data, target)
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", format)
	}
}

// 输出格式化器结构
type OutputFormatter struct {
	format   OutputFormat
	template *template.Template
	mutex    sync.RWMutex
}

// 创建输出格式化器
func NewOutputFormatter(format OutputFormat) *OutputFormatter {
	formatter := &OutputFormatter{
		format: format,
	}

	// 如果有自定义模板，解析它
	if format.Template != "" {
		tmpl, err := template.New("output").Parse(format.Template)
		if err == nil {
			formatter.template = tmpl
		}
	}

	return formatter
}

// 格式化输出
func (of *OutputFormatter) Format(data interface{}) (string, error) {
	of.mutex.RLock()
	defer of.mutex.RUnlock()

	switch of.format.Type {
	case "json":
		return of.formatJSON(data)
	case "csv":
		return of.formatCSV(data)
	case "table":
		return of.formatTable(data)
	case "custom":
		return of.formatCustom(data)
	default:
		return fmt.Sprintf("%+v", data), nil
	}
}

// JSON格式
func (of *OutputFormatter) formatJSON(data interface{}) (string, error) {
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// CSV格式
func (of *OutputFormatter) formatCSV(data interface{}) (string, error) {
	// 简化版CSV格式化
	return fmt.Sprintf("%+v", data), nil
}

// Table格式
func (of *OutputFormatter) formatTable(data interface{}) (string, error) {
	// 简化版表格格式化
	return fmt.Sprintf("%+v", data), nil
}

// 自定义格式
func (of *OutputFormatter) formatCustom(data interface{}) (string, error) {
	if of.template == nil {
		return fmt.Sprintf("%+v", data), nil
	}

	var buf bytes.Buffer
	if err := of.template.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
