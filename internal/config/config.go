package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// AI 服务配置
type AIService struct {
	Name     string `json:"name"`     // 服务名称
	Endpoint string `json:"endpoint"` // API 端点
	Token    string `json:"token"`    // API Token
	Model    string `json:"model"`    // 模型名称
	Priority int    `json:"priority"` // 优先级（数字越小优先级越高）
	Enabled  bool   `json:"enabled"`  // 是否启用
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

// 工作池配置
type WorkerPoolConfig struct {
	MaxWorkers   int           `json:"max_workers"`   // 最大工作协程数
	QueueSize    int           `json:"queue_size"`    // 队列大小
	BatchSize    int           `json:"batch_size"`    // 批处理大小
	Timeout      time.Duration `json:"timeout"`       // 超时时间
	RetryCount   int           `json:"retry_count"`   // 重试次数
	BackoffDelay time.Duration `json:"backoff_delay"` // 退避延迟
	Enabled      bool          `json:"enabled"`       // 是否启用
}

// 内存配置
type MemoryConfig struct {
	MaxMemoryUsage    int64         `json:"max_memory_usage"`    // 最大内存使用量（字节）
	GCThreshold       int64         `json:"gc_threshold"`        // 垃圾回收阈值
	LeakDetection     bool          `json:"leak_detection"`      // 是否启用内存泄漏检测
	ProfilingInterval time.Duration `json:"profiling_interval"`  // 性能分析间隔
	Enabled           bool          `json:"enabled"`             // 是否启用内存优化
}

// 并发控制配置
type ConcurrencyConfig struct {
	MaxConcurrency    int           `json:"max_concurrency"`    // 最大并发数
	BackpressureLimit int           `json:"backpressure_limit"` // 背压限制
	QueueTimeout      time.Duration `json:"queue_timeout"`      // 队列超时时间
	RetryDelay        time.Duration `json:"retry_delay"`        // 重试延迟
	Enabled           bool          `json:"enabled"`            // 是否启用并发控制
}

// I/O配置
type IOConfig struct {
	BufferSize        int           `json:"buffer_size"`        // 缓冲区大小
	BatchSize         int           `json:"batch_size"`         // 批处理大小
	FlushInterval     time.Duration `json:"flush_interval"`     // 刷新间隔
	AsyncIO           bool          `json:"async_io"`           // 是否启用异步I/O
	Compression       bool          `json:"compression"`        // 是否启用压缩
	Enabled           bool          `json:"enabled"`            // 是否启用I/O优化
}

// 多源配置
type MultiSourceConfig struct {
	Enabled bool           `json:"enabled"` // 是否启用多源支持
	Sources []SourceConfig `json:"sources"` // 数据源列表
}

// 数据源配置
type SourceConfig struct {
	Name     string `json:"name"`     // 数据源名称
	Type     string `json:"type"`     // 数据源类型 (file, journald, syslog)
	Path     string `json:"path"`     // 文件路径或配置
	Format   string `json:"format"`   // 日志格式
	Enabled  bool   `json:"enabled"`  // 是否启用
	Priority int    `json:"priority"` // 优先级
}


// 输出格式配置
type OutputFormat struct {
	Type     string `json:"type"`     // json, csv, table, custom
	Template string `json:"template"` // 自定义模板
	Color    bool   `json:"color"`    // 颜色支持
}

// 日志级别配置
type LogLevelConfig struct {
	Level     string `json:"level"`      // debug, info, warn, error, fatal
	MinLevel  string `json:"min_level"`  // 最小级别
	ShowDebug bool   `json:"show_debug"` // 显示调试日志
	ShowInfo  bool   `json:"show_info"`  // 显示信息日志
	ShowWarn  bool   `json:"show_warn"`  // 显示警告日志
	ShowError bool   `json:"show_error"` // 显示错误日志
	ShowFatal bool   `json:"show_fatal"` // 显示致命错误日志
}

// 主配置结构
type Config struct {
	AIEndpoint   string         `json:"ai_endpoint"` // 向后兼容
	Token        string         `json:"token"`       // 向后兼容
	Model        string         `json:"model"`       // 向后兼容
	CustomPrompt string         `json:"custom_prompt"`
	PromptFile   string         `json:"prompt_file"`   // 提示词文件路径
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
	LogLevel     LogLevelConfig `json:"log_level"` // I/O优化配置

	// 多源支持
	MultiSource MultiSourceConfig `json:"multi_source"`
}

// 默认配置变量
var DefaultConfig Config

// 初始化默认配置
func init() {
	DefaultConfig = Config{
		AIEndpoint:   "https://your-ai-server.com/api/v1/chat/completions",
		Token:        "your-api-token-here",
		Model:        "gpt-4",
		CustomPrompt: "",
		PromptFile:   "prompts/advanced.txt", // 提示词文件路径
		MaxRetries:   3,
		Timeout:      30,
		RateLimit:    60,
		LocalFilter:  true,
		AIServices:   []AIService{},
		DefaultAI:    "",
		Rules:        []FilterRule{},
		Cache: CacheConfig{
			MaxSize:         100 * 1024 * 1024, // 100MB
			MaxItems:        1000,
			DefaultTTL:      5 * time.Minute,
			AITTL:           10 * time.Minute,
			RuleTTL:         30 * time.Minute,
			ConfigTTL:       1 * time.Hour,
			CleanupInterval: 5 * time.Minute,
			Enabled:         true,
		},
		WorkerPool: WorkerPoolConfig{
			MaxWorkers:   4,
			QueueSize:    1000,
			BatchSize:    10,
			Timeout:      30 * time.Second,
			RetryCount:   3,
			BackoffDelay: 1 * time.Second,
			Enabled:      true,
		},
		Memory: MemoryConfig{
			MaxMemoryUsage:    512 * 1024 * 1024, // 512MB
			GCThreshold:       128 * 1024 * 1024, // 128MB
			LeakDetection:     true,
			ProfilingInterval: 30 * time.Second,
			Enabled:           true,
		},
		Concurrency: ConcurrencyConfig{
			MaxConcurrency:    100,
			BackpressureLimit: 1000,
			QueueTimeout:      5 * time.Second,
			RetryDelay:        1 * time.Second,
			Enabled:           true,
		},
		IO: IOConfig{
			BufferSize:    64 * 1024, // 64KB
			BatchSize:     10,
			FlushInterval: 1 * time.Second,
			AsyncIO:       true,
			Compression:   false,
			Enabled:       true,
		},
		MultiSource: MultiSourceConfig{
			Enabled: false,
			Sources: []SourceConfig{},
		},
		OutputFormat: OutputFormat{
			Type:     "table",
			Template: "",
			Color:    true,
		},
		LogLevel: LogLevelConfig{
			Level:     "info",
			MinLevel:  "info",
			ShowDebug: false,
			ShowInfo:  true,
			ShowWarn:  true,
			ShowError: true,
			ShowFatal: true,
		},
		Notifiers: NotifierConfig{
			Email: EmailConfig{
				Enabled:   false,
				Provider:  "smtp",
				Host:      "",
				Port:      587,
				Username:  "",
				Password:  "",
				FromEmail: "",
				ToEmails:  []string{},
			},
			DingTalk: WebhookConfig{
				Enabled: false,
				URL:     "",
				Secret:  "",
			},
			WeChat: WebhookConfig{
				Enabled: false,
				URL:     "",
				Secret:  "",
			},
			Feishu: WebhookConfig{
				Enabled: false,
				URL:     "",
				Secret:  "",
			},
			Slack: WebhookConfig{
				Enabled: false,
				URL:     "",
				Secret:  "",
			},
			CustomWebhooks: []WebhookConfig{},
		},
	}
}

// 加载配置文件
func LoadConfig() (*Config, error) {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "aipipe.json")

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		return &DefaultConfig, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON配置
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 合并默认配置
	mergedConfig := mergeConfig(DefaultConfig, config)

	return &mergedConfig, nil
}

// 合并配置
func mergeConfig(defaultConfig, userConfig Config) Config {
	merged := defaultConfig

	// 合并用户配置
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
	if userConfig.MaxRetries > 0 {
		merged.MaxRetries = userConfig.MaxRetries
	}
	if userConfig.Timeout > 0 {
		merged.Timeout = userConfig.Timeout
	}
	if userConfig.RateLimit > 0 {
		merged.RateLimit = userConfig.RateLimit
	}

	// 合并输出格式
	if userConfig.OutputFormat.Type != "" {
		merged.OutputFormat.Type = userConfig.OutputFormat.Type
	}
	if userConfig.OutputFormat.Template != "" {
		merged.OutputFormat.Template = userConfig.OutputFormat.Template
	}

	// 合并日志级别
	if userConfig.LogLevel.Level != "" {
		merged.LogLevel.Level = userConfig.LogLevel.Level
		merged.LogLevel.MinLevel = userConfig.LogLevel.Level
	}

	// 合并通知器配置
	if userConfig.Notifiers.Email.Enabled {
		merged.Notifiers.Email = userConfig.Notifiers.Email
	}
	if userConfig.Notifiers.DingTalk.Enabled {
		merged.Notifiers.DingTalk = userConfig.Notifiers.DingTalk
	}
	if userConfig.Notifiers.WeChat.Enabled {
		merged.Notifiers.WeChat = userConfig.Notifiers.WeChat
	}
	if userConfig.Notifiers.Feishu.Enabled {
		merged.Notifiers.Feishu = userConfig.Notifiers.Feishu
	}
	if userConfig.Notifiers.Slack.Enabled {
		merged.Notifiers.Slack = userConfig.Notifiers.Slack
	}
	if len(userConfig.Notifiers.CustomWebhooks) > 0 {
		merged.Notifiers.CustomWebhooks = userConfig.Notifiers.CustomWebhooks
	}

	return merged
}

// 处理配置测试
func HandleConfigTest() {
	fmt.Println("🧪 测试配置文件...")

	// 加载配置
	_, err := LoadConfig()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 配置文件测试通过！")
}

// 验证配置文件
func HandleConfigValidate() {
	fmt.Println("🔍 验证配置文件...")

	// 加载配置
	_, err := LoadConfig()
	if err != nil {
		fmt.Printf("❌ 配置验证失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ 配置文件验证通过！")
}

// 显示当前配置
func HandleConfigShow() {
	fmt.Println("📋 当前配置:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 显示配置信息（隐藏敏感信息）
	fmt.Printf("AI 端点: %s\n", config.AIEndpoint)
	fmt.Printf("模型: %s\n", config.Model)
	if len(config.Token) > 16 {
		fmt.Printf("Token: %s...%s\n", config.Token[:8], config.Token[len(config.Token)-8:])
	} else {
		fmt.Printf("Token: %s\n", strings.Repeat("*", len(config.Token)))
	}
	fmt.Printf("最大重试次数: %d\n", config.MaxRetries)
	fmt.Printf("超时时间: %d 秒\n", config.Timeout)
	fmt.Printf("频率限制: %d 次/分钟\n", config.RateLimit)
	fmt.Printf("本地过滤: %t\n", config.LocalFilter)

	if config.CustomPrompt != "" {
		fmt.Printf("自定义提示词: %s\n", config.CustomPrompt)
	}
}

// 处理配置向导
func HandleConfigInit() {
	fmt.Println("🎯 启动配置向导...")
	fmt.Println("⚠️  配置向导功能正在开发中...")
	fmt.Println("💡 请手动编辑 ~/.config/aipipe.json 配置文件")
}

// 处理配置模板
func HandleConfigTemplate() {
	fmt.Println("📋 配置模板:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 显示示例配置
	template := Config{
		AIEndpoint:   "https://your-ai-server.com/api/v1/chat/completions",
		Token:        "your-api-token-here",
		Model:        "gpt-4",
		CustomPrompt: "",
		MaxRetries:   3,
		Timeout:      30,
		RateLimit:    60,
		LocalFilter:  true,
		OutputFormat: OutputFormat{
			Type:     "table",
			Template: "",
			Color:    true,
		},
		LogLevel: LogLevelConfig{
			Level:     "info",
			MinLevel:  "info",
			ShowDebug: false,
			ShowInfo:  true,
			ShowWarn:  true,
			ShowError: true,
			ShowFatal: true,
		},
		Notifiers: NotifierConfig{
			Email: EmailConfig{
				Enabled:   false,
				Provider:  "smtp",
				Host:      "smtp.example.com",
				Port:      587,
				Username:  "user@example.com",
				Password:  "password",
				FromEmail: "user@example.com",
				ToEmails:  []string{"admin@example.com"},
			},
			DingTalk: WebhookConfig{
				Enabled: false,
				URL:     "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN",
				Secret:  "YOUR_SECRET",
			},
		},
	}

	// 输出JSON格式的配置模板
	data, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		fmt.Printf("❌ 生成配置模板失败: %v\n", err)
		return
	}

	fmt.Println(string(data))
	fmt.Println()
	fmt.Println("💡 将上述配置保存到 ~/.config/aipipe.json 文件中")
}
