package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

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
			Type:     "table",
			Color:    true,
			Width:    120,
			Headers:  true,
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
