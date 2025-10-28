package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/monitor"
	"github.com/xurenlu/aipipe/internal/utils"
)

// monitorCmd 代表监控命令
var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "监控日志文件",
	Long: `监控指定的日志文件，实时分析新增的日志内容。

示例:
  aipipe monitor --file /var/log/app.log
  aipipe monitor --file /var/log/nginx/access.log --format nginx
  aipipe monitor --file /var/log/system.log --format syslog`,
	Run: func(cmd *cobra.Command, args []string) {
		if filePath == "" {
			fmt.Println("❌ 请指定要监控的文件路径 (--file)")
			return
		}

		fmt.Printf("🚀 AIPipe 监控模式 - 监控文件: %s\n", filePath)
		fmt.Printf("📋 日志格式: %s\n", logFormat)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// 创建文件监控器
		fileMonitor, err := monitor.NewFileMonitor()
		if err != nil {
			fmt.Printf("❌ 创建文件监控器失败: %v\n", err)
			return
		}
		defer fileMonitor.Stop()

		// 添加文件监控
		err = fileMonitor.AddFile(filePath, func(filePath, line string) {
			// 处理新日志行
			processLogLine(line, logFormat)
		})
		
		if err != nil {
			fmt.Printf("❌ 添加文件监控失败: %v\n", err)
			return
		}

		fmt.Println("✅ 文件监控已启动，按 Ctrl+C 停止")
		
		// 等待中断信号
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		
		fmt.Println("\n🛑 监控已停止")
	},
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
