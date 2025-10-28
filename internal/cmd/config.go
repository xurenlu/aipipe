package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/config"
)

// configCmd 代表配置命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long: `管理 AIPipe 的配置文件，包括初始化、验证、显示和模板生成。

子命令:
  init      - 启动配置向导
  show      - 显示当前配置
  validate  - 验证配置文件
  template  - 生成配置模板
  test      - 测试配置`,
}

// configInitCmd 代表配置初始化命令
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "启动配置向导",
	Long:  "启动交互式配置向导，帮助您创建配置文件",
	Run: func(cmd *cobra.Command, args []string) {
		config.HandleConfigInit()
	},
}

// configShowCmd 代表显示配置命令
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "显示当前配置",
	Long:  "显示当前加载的配置信息",
	Run: func(cmd *cobra.Command, args []string) {
		config.HandleConfigShow()
	},
}

// configValidateCmd 代表验证配置命令
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "验证配置文件",
	Long:  "验证配置文件的格式和内容是否正确",
	Run: func(cmd *cobra.Command, args []string) {
		config.HandleConfigValidate()
	},
}

// configTemplateCmd 代表配置模板命令
var configTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "生成配置模板",
	Long:  "生成配置文件模板，保存到指定位置",
	Run: func(cmd *cobra.Command, args []string) {
		config.HandleConfigTemplate()
	},
}

// configTestCmd 代表测试配置命令
var configTestCmd = &cobra.Command{
	Use:   "test",
	Short: "测试配置",
	Long:  "测试当前配置是否正常工作",
	Run: func(cmd *cobra.Command, args []string) {
		config.HandleConfigTest()
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	
	// 添加配置子命令
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configTemplateCmd)
	configCmd.AddCommand(configTestCmd)
}
