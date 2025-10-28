package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	AIEndpoint   string         `json:"ai_endpoint"`
	Token        string         `json:"token"`
	Model        string         `json:"model"`
	CustomPrompt string         `json:"custom_prompt"`
	PromptFile   string         `json:"prompt_file"`   // 提示词文件路径
	MaxRetries   int            `json:"max_retries"`
	Timeout      int            `json:"timeout"`
	RateLimit    int            `json:"rate_limit"`
	LocalFilter  bool           `json:"local_filter"`
	OutputFormat OutputFormat   `json:"output_format"`
	LogLevel     LogLevelConfig `json:"log_level"`
	Notifiers    NotifierConfig `json:"notifiers"`
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
