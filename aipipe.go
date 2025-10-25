package main

import (
	"bufio"
	"bytes"
	"context"
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

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
	"github.com/BurntSushi/toml"
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

// 配置文件结构
type Config struct {
	AIEndpoint   string         `json:"ai_endpoint"`
	Token        string         `json:"token"`
	Model        string         `json:"model"`
	CustomPrompt string         `json:"custom_prompt"`
	Notifiers    NotifierConfig `json:"notifiers"`
}

// 默认配置
var defaultConfig = Config{
	AIEndpoint:   "https://your-ai-server.com/api/v1/chat/completions",
	Token:        "your-api-token-here",
	Model:        "gpt-4",
	CustomPrompt: "",
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
}

// 全局配置变量
var globalConfig Config

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

func main() {
	flag.Parse()

	// 检查是否使用多源监控
	if *multiSource != "" {
		processMultiSource()
		return
	}

	// 加载配置文件
	if *configFile != "" {
		// 使用指定的配置文件
		if err := loadConfigWithFormat(*configFile); err != nil {
			log.Fatalf("❌ 加载指定配置文件失败: %v", err)
		}
	} else {
		// 使用默认配置文件
		if err := loadConfig(); err != nil {
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
	} else if *logFormat == "journald" && (*journalServices != "" || *journalPriority != "" || *journalSince != "" || *journalUser != "" || *journalBoot || *journalKernel) {
		// journalctl模式
		fmt.Println("📰 使用journalctl监控系统日志...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		processJournalctl()
	} else {
		// 标准输入模式
		fmt.Println("📥 从标准输入读取日志...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if *noBatch {
			processStdin()
		} else {
			processStdinWithBatch()
		}
	}
}

// 加载配置文件
func loadConfig() error {
	configPath := filepath.Join(os.Getenv("HOME"), ".config", "aipipe.json")

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，创建默认配置文件
		return createDefaultConfig(configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置文件
	if err := json.Unmarshal(data, &globalConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
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

// 处理多源监控
func processMultiSource() {
	// 加载多源配置文件
	config, err := loadMultiSourceConfig(*multiSource)
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

	// 监控文件
	watchFileWithContext(ctx, source.Path, merger, batcher)
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
	// 实现带上下文的文件监控逻辑
	// 这里可以复用现有的watchFile逻辑，但需要支持context取消
	// 为了简化，这里先使用基本的文件监控
	watchFile(filePath)
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
	// 构建请求，使用 system 和 user 两条消息
	reqBody := ChatRequest{
		Model: globalConfig.Model,
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
		fmt.Printf("URL: %s\n", globalConfig.AIEndpoint)
		fmt.Printf("Method: POST\n")
		fmt.Printf("Headers:\n")
		fmt.Printf("  Content-Type: application/json\n")
		fmt.Printf("  api-key: %s...%s\n", globalConfig.Token[:10], globalConfig.Token[len(globalConfig.Token)-10:])
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
	req, err := http.NewRequest("POST", globalConfig.AIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", globalConfig.Token)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if *debug {
		fmt.Printf("⏳ 发送请求中...\n")
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(startTime)

	if err != nil {
		if *debug {
			fmt.Printf("❌ 请求失败: %v\n", err)
			fmt.Println(strings.Repeat("=", 80) + "\n")
		}
		return "", err
	}
	defer resp.Body.Close()

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
		return "", fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
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
