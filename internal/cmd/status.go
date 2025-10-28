package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/ai"
	"github.com/xurenlu/aipipe/internal/cache"
	"github.com/xurenlu/aipipe/internal/monitor"
	"github.com/xurenlu/aipipe/internal/notification"
	"github.com/xurenlu/aipipe/internal/rule"
)

// dashboardCmd 代表仪表板命令
var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "系统仪表板",
	Long: `AIPipe 系统仪表板，显示当前状态并提供交互式管理功能。

功能:
- 显示当前监听的文件和格式
- 显示配置信息和各模块状态
- 交互式添加监控文件源
- 管理监听配置

子命令:
  show      - 显示系统状态
  add       - 交互式添加监控文件
  list      - 列出所有监控文件
  remove    - 移除监控文件`,
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
	
	// 加载监控配置
	if err := loadMonitorConfig(); err != nil {
		fmt.Printf("  ❌ 加载监控配置失败: %v\n", err)
		return
	}
	
	if len(monitorConfig.Files) == 0 {
		fmt.Println("  📥 标准输入模式 (未监听文件)")
		fmt.Printf("  📝 日志格式: %s\n", logFormat)
	} else {
		fmt.Printf("  📋 已配置 %d 个监控文件:\n", len(monitorConfig.Files))
		
		for i, file := range monitorConfig.Files {
			status := "❌ 禁用"
			if file.Enabled {
				status = "✅ 启用"
			}
			
			// 检查文件是否存在
			if _, err := os.Stat(file.Path); err == nil {
				if info, err := os.Stat(file.Path); err == nil {
					fmt.Printf("    %d. %s\n", i+1, file.Path)
					fmt.Printf("       格式: %s | 优先级: %d | 状态: %s\n", file.Format, file.Priority, status)
					fmt.Printf("       大小: %d 字节 | 修改: %s\n", info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
				}
			} else {
				fmt.Printf("    %d. %s (文件不存在)\n", i+1, file.Path)
				fmt.Printf("       格式: %s | 优先级: %d | 状态: %s\n", file.Format, file.Priority, status)
			}
		}
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

// 监控文件配置
type MonitorConfig struct {
	Files []MonitorFile `json:"files"`
}

type MonitorFile struct {
	Path     string `json:"path"`
	Format   string `json:"format"`
	Enabled  bool   `json:"enabled"`
	Priority int    `json:"priority"`
}

// 全局监控配置
var monitorConfig MonitorConfig

// 监控配置文件路径
const monitorConfigFile = ".aipipe-monitor.json"

// 加载监控配置
func loadMonitorConfig() error {
	configPath := filepath.Join(os.Getenv("HOME"), monitorConfigFile)
	
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		monitorConfig = MonitorConfig{Files: []MonitorFile{}}
		return nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取监控配置文件失败: %w", err)
	}
	
	if err := json.Unmarshal(data, &monitorConfig); err != nil {
		return fmt.Errorf("解析监控配置文件失败: %w", err)
	}
	
	return nil
}

// 保存监控配置
func saveMonitorConfig() error {
	configPath := filepath.Join(os.Getenv("HOME"), monitorConfigFile)
	
	data, err := json.MarshalIndent(monitorConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化监控配置失败: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存监控配置文件失败: %w", err)
	}
	
	return nil
}

// dashboardShowCmd 代表显示状态命令
var dashboardShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示系统状态",
	Long:  "显示 AIPipe 的当前状态，包括监听的文件、配置信息、服务状态等",
	Run: func(cmd *cobra.Command, args []string) {
		showSystemStatus()
	},
}

// dashboardAddCmd 代表添加监控文件命令
var dashboardAddCmd = &cobra.Command{
	Use:   "add",
	Short: "交互式添加监控文件",
	Long:  "通过交互式界面添加新的监控文件源",
	Run: func(cmd *cobra.Command, args []string) {
		addMonitorFileInteractive()
	},
}

// dashboardListCmd 代表列出监控文件命令
var dashboardListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有监控文件",
	Long:  "列出所有已配置的监控文件",
	Run: func(cmd *cobra.Command, args []string) {
		listMonitorFiles()
	},
}

// dashboardRemoveCmd 代表移除监控文件命令
var dashboardRemoveCmd = &cobra.Command{
	Use:   "remove <file_path>",
	Short: "移除监控文件",
	Long:  "根据文件路径移除监控文件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		removeMonitorFile(args[0])
	},
}

// 交互式添加监控文件
func addMonitorFileInteractive() {
	fmt.Println("🔧 交互式添加监控文件")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	reader := bufio.NewReader(os.Stdin)

	// 获取文件路径
	fmt.Print("📁 请输入文件路径: ")
	filePath, _ := reader.ReadString('\n')
	filePath = strings.TrimSpace(filePath)

	if filePath == "" {
		fmt.Println("❌ 文件路径不能为空")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("❌ 文件不存在: %s\n", filePath)
		return
	}

	// 选择日志格式
	fmt.Println("\n📝 请选择日志格式:")
	formats := []string{"java", "nginx", "php", "python", "go", "rust", "docker", "kubernetes", "syslog", "journald", "mysql", "postgresql", "redis", "elasticsearch", "git", "jenkins", "github", "macos-console", "custom"}

	for i, format := range formats {
		fmt.Printf("  %d. %s\n", i+1, format)
	}

	fmt.Print("请选择格式 (1-19): ")
	formatInput, _ := reader.ReadString('\n')
	formatInput = strings.TrimSpace(formatInput)

	formatIndex, err := strconv.Atoi(formatInput)
	if err != nil || formatIndex < 1 || formatIndex > len(formats) {
		fmt.Println("❌ 无效的选择")
		return
	}

	selectedFormat := formats[formatIndex-1]

	// 如果是自定义格式，让用户输入
	if selectedFormat == "custom" {
		fmt.Print("请输入自定义格式: ")
		customFormat, _ := reader.ReadString('\n')
		selectedFormat = strings.TrimSpace(customFormat)
	}

	// 设置优先级
	fmt.Print("🎯 请输入优先级 (1-100, 数字越小优先级越高): ")
	priorityInput, _ := reader.ReadString('\n')
	priorityInput = strings.TrimSpace(priorityInput)

	priority := 50 // 默认优先级
	if p, err := strconv.Atoi(priorityInput); err == nil && p >= 1 && p <= 100 {
		priority = p
	}

	// 确认添加
	fmt.Printf("\n📋 确认添加监控文件:\n")
	fmt.Printf("  文件路径: %s\n", filePath)
	fmt.Printf("  日志格式: %s\n", selectedFormat)
	fmt.Printf("  优先级: %d\n", priority)

	fmt.Print("\n是否确认添加? (y/N): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("❌ 已取消添加")
		return
	}

	// 加载现有配置
	if err := loadMonitorConfig(); err != nil {
		fmt.Printf("❌ 加载监控配置失败: %v\n", err)
		return
	}
	
	// 检查文件是否已存在
	for _, existingFile := range monitorConfig.Files {
		if existingFile.Path == filePath {
			fmt.Printf("❌ 文件已存在: %s\n", filePath)
			return
		}
	}
	
	// 添加新文件到配置
	newFile := MonitorFile{
		Path:     filePath,
		Format:   selectedFormat,
		Enabled:  true,
		Priority: priority,
	}
	
	monitorConfig.Files = append(monitorConfig.Files, newFile)
	
	// 保存配置
	if err := saveMonitorConfig(); err != nil {
		fmt.Printf("❌ 保存监控配置失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 监控文件添加成功: %s (%s)\n", filePath, selectedFormat)
	fmt.Println("💡 使用 'aipipe dashboard show' 查看当前状态")
}

// 列出监控文件
func listMonitorFiles() {
	fmt.Println("📋 监控文件列表")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// 加载监控配置
	if err := loadMonitorConfig(); err != nil {
		fmt.Printf("❌ 加载监控配置失败: %v\n", err)
		return
	}
	
	if len(monitorConfig.Files) == 0 {
		fmt.Println("📥 当前使用标准输入模式")
		fmt.Printf("  格式: %s\n", logFormat)
	} else {
		for i, file := range monitorConfig.Files {
			status := "❌ 禁用"
			if file.Enabled {
				status = "✅ 启用"
			}
			
			// 检查文件是否存在
			if _, err := os.Stat(file.Path); err == nil {
				absPath, _ := filepath.Abs(file.Path)
				fmt.Printf("%d. ✅ %s\n", i+1, absPath)
				fmt.Printf("   格式: %s | 优先级: %d | 状态: %s\n", file.Format, file.Priority, status)
			} else {
				fmt.Printf("%d. ❌ %s (文件不存在)\n", i+1, file.Path)
				fmt.Printf("   格式: %s | 优先级: %d | 状态: %s\n", file.Format, file.Priority, status)
			}
		}
	}
	
	fmt.Println()
}

// 移除监控文件
func removeMonitorFile(path string) {
	// 加载监控配置
	if err := loadMonitorConfig(); err != nil {
		fmt.Printf("❌ 加载监控配置失败: %v\n", err)
		return
	}
	
	// 查找并移除文件
	found := false
	for i, file := range monitorConfig.Files {
		if file.Path == path {
			monitorConfig.Files = append(monitorConfig.Files[:i], monitorConfig.Files[i+1:]...)
			found = true
			break
		}
	}
	
	if !found {
		fmt.Printf("❌ 未找到监控文件: %s\n", path)
		return
	}
	
	// 保存配置
	if err := saveMonitorConfig(); err != nil {
		fmt.Printf("❌ 保存监控配置失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 已移除监控文件: %s\n", path)
}

func init() {
	rootCmd.AddCommand(dashboardCmd)

	// 添加仪表板子命令
	dashboardCmd.AddCommand(dashboardShowCmd)
	dashboardCmd.AddCommand(dashboardAddCmd)
	dashboardCmd.AddCommand(dashboardListCmd)
	dashboardCmd.AddCommand(dashboardRemoveCmd)
}
