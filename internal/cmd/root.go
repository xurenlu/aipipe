package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/config"
)

var (
	// 全局配置
	globalConfig *config.Config

	// 全局标志
	verbose          bool
	showNotImportant bool
	logFormat        string
	filePath         string
)

// rootCmd 代表基础命令
var rootCmd = &cobra.Command{
	Use:   "aipipe",
	Short: "AIPipe - 智能日志分析工具",
	Long: `AIPipe 是一个基于 AI 的智能日志分析工具，能够实时监控和分析各种格式的日志文件。

主要功能:
- 智能日志分析: 使用 AI 判断日志重要性
- 实时监控: 支持标准输入和文件监控
- 多格式支持: 支持 20+ 种日志格式
- 通知告警: 多渠道实时通知
- 规则引擎: 灵活的正则表达式过滤
- 缓存优化: 提高性能和减少 API 调用

使用示例:
  aipipe monitor --file /var/log/app.log
  tail -f app.log | aipipe analyze
  aipipe config init
  aipipe rules add --pattern "ERROR" --action alert`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 加载配置
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("⚠️  加载配置文件失败，使用默认配置: %v\n", err)
			globalConfig = &config.DefaultConfig
		} else {
			globalConfig = cfg
		}
	},
}

// Execute 添加所有子命令到根命令并设置适当的标志
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "显示详细输出")
	rootCmd.PersistentFlags().BoolVar(&showNotImportant, "show-not-important", false, "显示被过滤的日志")
	rootCmd.PersistentFlags().StringVarP(&logFormat, "format", "f", "java", "日志格式")
	rootCmd.PersistentFlags().StringVar(&filePath, "file", "", "要监控的日志文件路径")
}
