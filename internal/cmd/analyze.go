package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/utils"
)

// analyzeCmd 代表分析命令
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "分析标准输入的日志",
	Long: `分析从标准输入读取的日志内容，使用 AI 判断日志重要性。

示例:
  tail -f app.log | aipipe analyze
  echo "ERROR: Database connection failed" | aipipe analyze
  cat logfile.txt | aipipe analyze --format nginx`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🚀 AIPipe 分析模式 - 监控 %s 格式日志\n", logFormat)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// 从标准输入读取日志
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		lineCount := 0
		filteredCount := 0
		alertCount := 0

		for scanner.Scan() {
			line := scanner.Text()
			lineCount++

			if line == "" {
				continue
			}

			// 使用 AI 分析日志
			analysis, err := utils.AnalyzeLog(line, logFormat, globalConfig)
			if err != nil {
				fmt.Printf("❌ 分析失败: %v\n", err)
				continue
			}

			if analysis.Important {
				fmt.Printf("⚠️  [重要] %s\n", line)
				fmt.Printf("   📝 摘要: %s\n", analysis.Summary)
				alertCount++
			} else {
				if showNotImportant {
					fmt.Printf("🔇 [过滤] %s\n", line)
				}
				filteredCount++
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Printf("❌ 读取输入失败: %v\n", err)
			return
		}

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("📊 统计: 总计 %d 行, 过滤 %d 行, 告警 %d 次\n", lineCount, filteredCount, alertCount)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
