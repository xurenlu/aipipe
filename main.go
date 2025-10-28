package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xurenlu/aipipe/internal/config"
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
	globalConfig, err := config.LoadConfig()
	if err != nil {
		log.Printf("⚠️  加载配置文件失败，使用默认配置: %v", err)
		globalConfig = &config.DefaultConfig
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
