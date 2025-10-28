package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/monitor"
	"github.com/xurenlu/aipipe/internal/utils"
)

// monitorCmd 代表监控命令
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "监控日志文件",
	Long: `监控日志文件，实时分析新增的日志内容。

支持两种模式:
1. 自动模式: 从配置文件读取所有监控文件
   aipipe monitor

2. 手动模式: 指定单个文件监控
   aipipe monitor --file /var/log/app.log --format nginx

示例:
  aipipe monitor                                    # 监控所有配置的文件
  aipipe monitor --file /var/log/app.log           # 监控指定文件
  aipipe monitor --file /var/log/nginx/access.log --format nginx`,
	Run: func(cmd *cobra.Command, args []string) {
		// 创建文件监控器
		fileMonitor, err := monitor.NewFileMonitor()
		if err != nil {
			fmt.Printf("❌ 创建文件监控器失败: %v\n", err)
			return
		}
		defer fileMonitor.Stop()

		// 如果指定了文件，使用手动模式
		if filePath != "" {
			startManualMonitor(fileMonitor, filePath, logFormat)
		} else {
			// 否则使用自动模式，从配置文件读取
			startAutoMonitor(fileMonitor)
		}
	},
}

// 手动监控模式
func startManualMonitor(fileMonitor *monitor.FileMonitor, filePath, format string) {
	fmt.Printf("🚀 AIPipe 监控模式 - 监控文件: %s\n", filePath)
	fmt.Printf("📋 日志格式: %s\n", format)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 添加文件监控
	err := fileMonitor.AddFile(filePath, func(filePath, line string) {
		// 处理新日志行
		processLogLine(line, format)
	})

	if err != nil {
		fmt.Printf("❌ 添加文件监控失败: %v\n", err)
		return
	}

	fmt.Println("✅ 文件监控已启动，按 Ctrl+C 停止")
	waitForInterrupt()
}

// 自动监控模式
func startAutoMonitor(fileMonitor *monitor.FileMonitor) {
	// 加载监控配置
	monitorConfig, err := loadMonitorConfigFromFile()
	if err != nil {
		fmt.Printf("❌ 加载监控配置失败: %v\n", err)
		return
	}

	if len(monitorConfig.Files) == 0 {
		fmt.Println("❌ 没有配置任何监控文件")
		fmt.Println("💡 使用 'aipipe dashboard add' 添加监控文件")
		return
	}

	// 按优先级排序
	sort.Slice(monitorConfig.Files, func(i, j int) bool {
		return monitorConfig.Files[i].Priority < monitorConfig.Files[j].Priority
	})

	fmt.Printf("🚀 AIPipe 监控模式 - 监控 %d 个文件\n", len(monitorConfig.Files))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 添加所有启用的文件监控
	enabledCount := 0
	for _, file := range monitorConfig.Files {
		if !file.Enabled {
			continue
		}

		// 检查文件是否存在
		if _, err := os.Stat(file.Path); os.IsNotExist(err) {
			fmt.Printf("⚠️  文件不存在，跳过: %s\n", file.Path)
			continue
		}

		// 添加文件监控
		err := fileMonitor.AddFile(file.Path, func(filePath, line string) {
			// 找到对应的格式
			var format string
			for _, f := range monitorConfig.Files {
				if f.Path == filePath {
					format = f.Format
					break
				}
			}
			processLogLine(line, format)
		})

		if err != nil {
			fmt.Printf("❌ 添加文件监控失败: %s - %v\n", file.Path, err)
			continue
		}

		fmt.Printf("✅ 已添加监控: %s (%s, 优先级: %d)\n", file.Path, file.Format, file.Priority)
		enabledCount++
	}

	if enabledCount == 0 {
		fmt.Println("❌ 没有可用的监控文件")
		return
	}

	fmt.Printf("✅ 文件监控已启动，监控 %d 个文件，按 Ctrl+C 停止\n", enabledCount)
	waitForInterrupt()
}

// 加载监控配置
func loadMonitorConfigFromFile() (MonitorConfig, error) {
	var config MonitorConfig
	configPath := filepath.Join(os.Getenv("HOME"), ".aipipe-monitor.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return MonitorConfig{Files: []MonitorFile{}}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("读取监控配置文件失败: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("解析监控配置文件失败: %w", err)
	}

	return config, nil
}

// 等待中断信号
func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\n🛑 监控已停止")
}

// 处理日志行
func processLogLine(line, format string) {
	// 使用 AI 分析日志
	analysis, err := utils.AnalyzeLog(line, format, globalConfig)
	if err != nil {
		fmt.Printf("❌ 分析失败: %v\n", err)
		return
	}

	if analysis.Important {
		fmt.Printf("⚠️  [重要] %s\n", line)
		fmt.Printf("   📝 摘要: %s\n", analysis.Summary)
	} else {
		if showNotImportant {
			fmt.Printf("🔇 [过滤] %s\n", line)
		}
	}
}

func init() {
	rootCmd.AddCommand(monitorCmd)
}
