package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xurenlu/aipipe/internal/ai"
	"github.com/xurenlu/aipipe/internal/cache"
	"github.com/xurenlu/aipipe/internal/config"
	"github.com/xurenlu/aipipe/internal/monitor"
	"github.com/xurenlu/aipipe/internal/notification"
	"github.com/xurenlu/aipipe/internal/rule"
	"github.com/xurenlu/aipipe/internal/utils"
)

// 全局变量
var (
	// 命令行参数
	logFormat        = flag.String("format", "java", "日志格式 (java, php, nginx, ruby, fastapi, python, go, rust, csharp, kotlin, nodejs, typescript, docker, kubernetes, postgresql, mysql, redis, elasticsearch, git, jenkins, github, journald, macos-console, syslog)")
	verbose          = flag.Bool("verbose", false, "显示详细输出")
	filePath         = flag.String("f", "", "要监控的日志文件路径（类似 tail -f）")
	debug            = flag.Bool("debug", false, "调试模式，打印 HTTP 请求和响应详情")
	noBatch          = flag.Bool("no-batch", false, "禁用批处理，逐行分析（增加 API 调用）")
	batchSize        = flag.Int("batch-size", 10, "批处理最大行数")
	batchWait        = flag.Duration("batch-wait", 3*time.Second, "批处理等待时间")
	showNotImportant = flag.Bool("show-not-important", false, "显示被过滤的日志（默认不显示）")
	contextLines     = flag.Int("context", 3, "重要日志显示的上下文行数（前后各N行）")

	// 新增配置管理命令
	configTest     = flag.Bool("config-test", false, "测试配置文件")
	configValidate = flag.Bool("config-validate", false, "验证配置文件")
	configShow     = flag.Bool("config-show", false, "显示当前配置")

	// 用户体验命令
	configInit     = flag.Bool("config-init", false, "启动配置向导")
	configTemplate = flag.Bool("config-template", false, "显示配置模板")
	outputFormat   = flag.String("output-format", "", "输出格式 (json, csv, table, custom)")
	outputColor    = flag.Bool("output-color", true, "启用颜色输出")
	logLevel       = flag.String("log-level", "", "日志级别 (debug, info, warn, error, fatal)")

	// 多源监控配置
	multiSource = flag.String("multi-source", "", "多源监控配置文件路径")
	configFile  = flag.String("config", "", "指定配置文件路径")
)

// 全局配置变量
var globalConfig *config.Config

// 全局管理器
var (
	cacheManager      *cache.CacheManager
	notificationManager *notification.NotificationManager
	ruleEngine        *rule.RuleEngine
	aiServiceManager  *ai.AIServiceManager
	fileMonitor       *monitor.FileMonitor
)

func main() {
	flag.Parse()

	// 处理配置管理命令
	if *configTest {
		config.HandleConfigTest()
		return
	}

	if *configValidate {
		config.HandleConfigValidate()
		return
	}

	if *configShow {
		config.HandleConfigShow()
		return
	}

	if *configInit {
		config.HandleConfigInit()
		return
	}

	if *configTemplate {
		config.HandleConfigTemplate()
		return
	}

	// 加载配置文件
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("⚠️  加载配置文件失败，使用默认配置: %v", err)
		globalConfig = &config.DefaultConfig
	} else {
		globalConfig = cfg
	}

	// 初始化管理器
	initializeManagers()

	fmt.Printf("🚀 AIPipe 启动 - 监控 %s 格式日志\n", *logFormat)

	// 显示功能状态
	showFeatureStatus()

	// 显示模式提示
	if !*showNotImportant {
		fmt.Println("💡 只显示重要日志（过滤的日志不显示）")
		if !*verbose {
			fmt.Println("   使用 --show-not-important 显示所有日志")
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 显示配置信息
	if *verbose {
		fmt.Printf("AI 端点: %s\n", globalConfig.AIEndpoint)
		fmt.Printf("模型: %s\n", globalConfig.Model)
		fmt.Printf("最大重试次数: %d\n", globalConfig.MaxRetries)
		fmt.Printf("超时时间: %d 秒\n", globalConfig.Timeout)
		fmt.Printf("频率限制: %d 次/分钟\n", globalConfig.RateLimit)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	// 根据参数选择运行模式
	if *filePath != "" {
		// 文件监控模式
		fmt.Printf("📁 监控文件: %s\n", *filePath)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("⚠️  文件监控功能正在开发中...")
	} else {
		// 标准输入模式
		fmt.Println("📥 从标准输入读取日志...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		utils.ProcessStdin(globalConfig, *showNotImportant)
	}
}

// 初始化所有管理器
func initializeManagers() {
	if globalConfig == nil {
		log.Printf("⚠️  配置未加载，跳过管理器初始化")
		return
	}
	
	// 初始化缓存管理器
	cacheManager = cache.NewCacheManager(globalConfig.Cache)
	
	// 初始化通知管理器
	notificationManager = notification.NewNotificationManager(globalConfig)
	
	// 初始化规则引擎
	ruleEngine = rule.NewRuleEngine(globalConfig.Rules)
	
	// 初始化 AI 服务管理器
	aiServiceManager = ai.NewAIServiceManager(globalConfig.AIServices)
	
	// 初始化文件监控器
	var err error
	fileMonitor, err = monitor.NewFileMonitor()
	if err != nil {
		log.Printf("⚠️  文件监控器初始化失败: %v", err)
	}
}

// 显示功能状态
func showFeatureStatus() {
	fmt.Println("🔧 功能状态:")
	
	if globalConfig == nil {
		fmt.Println("   ❌ 配置未加载")
		return
	}
	
	// 缓存状态
	if globalConfig.Cache.Enabled {
		fmt.Println("   ✅ 缓存系统: 已启用")
	} else {
		fmt.Println("   ❌ 缓存系统: 已禁用")
	}
	
	// 通知状态
	if notificationManager != nil {
		enabledNotifiers := notificationManager.GetEnabledCount()
		fmt.Printf("   📢 通知系统: %d 个通知器已启用\n", enabledNotifiers)
	} else {
		fmt.Println("   ❌ 通知系统: 未初始化")
	}
	
	// 规则引擎状态
	if ruleEngine != nil {
		stats := ruleEngine.GetStats()
		fmt.Printf("   🔍 规则引擎: %d 个规则已启用\n", stats.EnabledRules)
	} else {
		fmt.Println("   ❌ 规则引擎: 未初始化")
	}
	
	// AI 服务状态
	if aiServiceManager != nil {
		aiStats := aiServiceManager.GetStats()
		fmt.Printf("   🤖 AI 服务: %d 个服务已启用\n", aiStats["enabled_services"])
	} else {
		fmt.Println("   ❌ AI 服务: 未初始化")
	}
	
	// 文件监控状态
	if fileMonitor != nil {
		monitorStatus := fileMonitor.GetStatus()
		fmt.Printf("   📁 文件监控: %d 个文件已监控\n", monitorStatus["active_files"])
	} else {
		fmt.Println("   ❌ 文件监控: 未启用")
	}
}
