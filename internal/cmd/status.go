package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/ai"
	"github.com/xurenlu/aipipe/internal/cache"
	"github.com/xurenlu/aipipe/internal/monitor"
	"github.com/xurenlu/aipipe/internal/notification"
	"github.com/xurenlu/aipipe/internal/rule"
)

// statusCmd 代表状态命令
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示系统状态",
	Long: `显示 AIPipe 的当前状态，包括监听的文件、配置信息、服务状态等。

显示内容:
- 当前监听的文件和格式
- 配置信息概览
- 各模块状态
- 统计信息`,
	Run: func(cmd *cobra.Command, args []string) {
		showSystemStatus()
	},
}

// 显示系统状态
func showSystemStatus() {
	fmt.Println("🔍 AIPipe 系统状态")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 显示配置信息
	showConfigStatus()

	// 显示监听状态
	showMonitoringStatus()

	// 显示各模块状态
	showModuleStatus()

	// 显示统计信息
	showStatistics()
}

// 显示配置状态
func showConfigStatus() {
	fmt.Println("📋 配置信息:")
	fmt.Printf("  AI端点: %s\n", globalConfig.AIEndpoint)
	fmt.Printf("  模型: %s\n", globalConfig.Model)
	fmt.Printf("  最大重试: %d\n", globalConfig.MaxRetries)
	fmt.Printf("  超时时间: %d秒\n", globalConfig.Timeout)
	fmt.Printf("  频率限制: %d次/分钟\n", globalConfig.RateLimit)
	fmt.Printf("  本地过滤: %t\n", globalConfig.LocalFilter)
	
	if globalConfig.PromptFile != "" {
		fmt.Printf("  提示词文件: %s\n", globalConfig.PromptFile)
	}
	
	fmt.Println()
}

// 显示监听状态
func showMonitoringStatus() {
	fmt.Println("📁 监听状态:")
	
	// 检查是否有文件在监听
	if filePath != "" {
		// 检查文件是否存在
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("  ✅ 正在监听: %s\n", filePath)
			fmt.Printf("  📝 日志格式: %s\n", logFormat)
			
			// 显示文件信息
			if info, err := os.Stat(filePath); err == nil {
				fmt.Printf("  📊 文件大小: %d 字节\n", info.Size())
				fmt.Printf("  🕒 最后修改: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
			}
		} else {
			fmt.Printf("  ❌ 文件不存在: %s\n", filePath)
		}
	} else {
		fmt.Println("  📥 标准输入模式 (未监听文件)")
		fmt.Printf("  📝 日志格式: %s\n", logFormat)
	}
	
	fmt.Println()
}

// 显示模块状态
func showModuleStatus() {
	fmt.Println("🔧 模块状态:")
	
	// 缓存状态
	cacheManager := cache.NewCacheManager(globalConfig.Cache)
	if globalConfig.Cache.Enabled {
		stats := cacheManager.GetStats()
		fmt.Printf("  ✅ 缓存系统: 已启用 (%d 项目, %.2f%% 命中率)\n", stats.TotalItems, stats.HitRate*100)
	} else {
		fmt.Println("  ❌ 缓存系统: 已禁用")
	}
	
	// 通知状态
	notificationManager := notification.NewNotificationManager(globalConfig)
	enabledNotifiers := notificationManager.GetEnabledCount()
	fmt.Printf("  📢 通知系统: %d 个通知器已启用\n", enabledNotifiers)
	
	// 规则引擎状态
	ruleEngine := rule.NewRuleEngine(globalConfig.Rules)
	stats := ruleEngine.GetStats()
	fmt.Printf("  🔍 规则引擎: %d 个规则已启用\n", stats.EnabledRules)
	
	// AI服务状态
	aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)
	aiStats := aiServiceManager.GetStats()
	fmt.Printf("  🤖 AI服务: %d 个服务已启用\n", aiStats["enabled_services"])
	
	// 文件监控状态
	fileMonitor, err := monitor.NewFileMonitor()
	if err == nil {
		monitorStatus := fileMonitor.GetStatus()
		fmt.Printf("  📁 文件监控: %d 个文件已监控\n", monitorStatus["active_files"])
	} else {
		fmt.Println("  ❌ 文件监控: 初始化失败")
	}
	
	fmt.Println()
}

// 显示统计信息
func showStatistics() {
	fmt.Println("📊 统计信息:")
	
	// 规则统计
	ruleEngine := rule.NewRuleEngine(globalConfig.Rules)
	ruleStats := ruleEngine.GetStats()
	fmt.Printf("  规则总数: %d (启用: %d, 禁用: %d)\n", 
		ruleStats.TotalRules, ruleStats.EnabledRules, ruleStats.DisabledRules)
	
	// 缓存统计
	cacheManager := cache.NewCacheManager(globalConfig.Cache)
	cacheStats := cacheManager.GetStats()
	fmt.Printf("  缓存项目: %d (内存: %.2f MB)\n", 
		cacheStats.TotalItems, float64(cacheStats.MemoryUsage)/1024/1024)
	
	// AI服务统计
	aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)
	aiStats := aiServiceManager.GetStats()
	fmt.Printf("  AI服务: %d 个 (启用: %d, 限流: %d)\n", 
		aiStats["total_services"], aiStats["enabled_services"], aiStats["rate_limited_services"])
	
	// 通知统计
	notificationManager := notification.NewNotificationManager(globalConfig)
	notifierCount := notificationManager.GetEnabledCount()
	fmt.Printf("  通知器: %d 个已启用\n", notifierCount)
	
	fmt.Println()
}

// 显示监听文件详情
func showMonitoringDetails() {
	fmt.Println("📁 监听文件详情:")
	
	if filePath != "" {
		// 检查文件是否存在
		if _, err := os.Stat(filePath); err == nil {
			// 显示文件路径
			absPath, _ := filepath.Abs(filePath)
			fmt.Printf("  文件路径: %s\n", absPath)
			
			// 显示文件信息
			if info, err := os.Stat(filePath); err == nil {
				fmt.Printf("  文件大小: %d 字节 (%.2f MB)\n", info.Size(), float64(info.Size())/1024/1024)
				fmt.Printf("  最后修改: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
				fmt.Printf("  文件权限: %s\n", info.Mode().String())
			}
			
			// 显示格式信息
			fmt.Printf("  日志格式: %s\n", logFormat)
			fmt.Printf("  显示过滤日志: %t\n", showNotImportant)
			fmt.Printf("  详细输出: %t\n", verbose)
		} else {
			fmt.Printf("  ❌ 文件不存在: %s\n", filePath)
		}
	} else {
		fmt.Println("  📥 当前使用标准输入模式")
		fmt.Printf("  日志格式: %s\n", logFormat)
		fmt.Printf("  显示过滤日志: %t\n", showNotImportant)
		fmt.Printf("  详细输出: %t\n", verbose)
	}
	
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
