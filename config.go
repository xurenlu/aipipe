package main

import (
	"encoding/json"
	"fmt"
	"os"
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
	Level      string `json:"level"`       // debug, info, warn, error, fatal
	ShowDebug  bool   `json:"show_debug"`  // 显示调试日志
	ShowInfo   bool   `json:"show_info"`   // 显示信息日志
	ShowWarn   bool   `json:"show_warn"`   // 显示警告日志
	ShowError  bool   `json:"show_error"`  // 显示错误日志
	ShowFatal  bool   `json:"show_fatal"`  // 显示致命日志
	MinLevel   string `json:"min_level"`   // 最小日志级别
	MaxLevel   string `json:"max_level"`   // 最大日志级别
	Enabled    bool   `json:"enabled"`     // 是否启用日志级别过滤
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
