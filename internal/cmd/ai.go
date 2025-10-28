package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/ai"
	"github.com/xurenlu/aipipe/internal/config"
)

var (
	aiName     string
	aiEndpoint string
	aiToken    string
	aiModel    string
	aiPriority int
	aiEnabled  bool
)

// aiCmd 代表AI命令
var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "AI服务管理",
	Long: `管理AI服务，包括添加、删除、启用、禁用AI服务。

子命令:
  list      - 列出所有AI服务
  add       - 添加AI服务
  remove    - 删除AI服务
  enable    - 启用AI服务
  disable   - 禁用AI服务
  test      - 测试AI服务
  stats     - 显示AI服务统计`,
}

// aiListCmd 代表列出AI服务命令
var aiListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有AI服务",
	Long:  "列出所有配置的AI服务",
	Run: func(cmd *cobra.Command, args []string) {
		aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)
		services := aiServiceManager.GetServices()

		if len(services) == 0 {
			fmt.Println("📋 没有配置任何AI服务")
			return
		}

		fmt.Printf("📋 AI服务列表 (共 %d 个):\n", len(services))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for _, service := range services {
			status := "❌ 禁用"
			if service.Enabled {
				status = "✅ 启用"
			}

			fmt.Printf("名称: %s\n", service.Name)
			fmt.Printf("  端点: %s\n", service.Endpoint)
			fmt.Printf("  模型: %s\n", service.Model)
			fmt.Printf("  优先级: %d\n", service.Priority)
			fmt.Printf("  状态: %s\n", status)
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}
	},
}

// aiAddCmd 代表添加AI服务命令
var aiAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加AI服务",
	Long:  "添加新的AI服务",
	Run: func(cmd *cobra.Command, args []string) {
		if aiName == "" || aiEndpoint == "" || aiToken == "" || aiModel == "" {
			fmt.Println("❌ 请指定所有必需参数: --name, --endpoint, --token, --model")
			return
		}

		// 检查服务名是否已存在
		for _, service := range globalConfig.AIServices {
			if service.Name == aiName {
				fmt.Printf("❌ AI服务名称已存在: %s\n", aiName)
				return
			}
		}

		// 创建新服务
		newService := config.AIService{
			Name:     aiName,
			Endpoint: aiEndpoint,
			Token:    aiToken,
			Model:    aiModel,
			Priority: aiPriority,
			Enabled:  aiEnabled,
		}

		// 添加到配置
		globalConfig.AIServices = append(globalConfig.AIServices, newService)

		fmt.Printf("✅ AI服务添加成功: %s\n", aiName)
		fmt.Printf("   端点: %s\n", aiEndpoint)
		fmt.Printf("   模型: %s\n", aiModel)
		fmt.Printf("   优先级: %d\n", aiPriority)
		fmt.Printf("   状态: %t\n", aiEnabled)
	},
}

// aiRemoveCmd 代表删除AI服务命令
var aiRemoveCmd = &cobra.Command{
	Use:   "remove <service_name>",
	Short: "删除AI服务",
	Long:  "根据服务名称删除AI服务",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]

		// 查找并删除服务
		found := false
		for i, service := range globalConfig.AIServices {
			if service.Name == serviceName {
				globalConfig.AIServices = append(globalConfig.AIServices[:i], globalConfig.AIServices[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("❌ 未找到AI服务: %s\n", serviceName)
			return
		}

		fmt.Printf("✅ AI服务删除成功: %s\n", serviceName)
	},
}

// aiEnableCmd 代表启用AI服务命令
var aiEnableCmd = &cobra.Command{
	Use:   "enable <service_name>",
	Short: "启用AI服务",
	Long:  "根据服务名称启用AI服务",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)

		err := aiServiceManager.SetServiceEnabled(serviceName, true)
		if err != nil {
			fmt.Printf("❌ 启用AI服务失败: %v\n", err)
			return
		}

		// 更新配置
		for i := range globalConfig.AIServices {
			if globalConfig.AIServices[i].Name == serviceName {
				globalConfig.AIServices[i].Enabled = true
				break
			}
		}

		fmt.Printf("✅ AI服务启用成功: %s\n", serviceName)
	},
}

// aiDisableCmd 代表禁用AI服务命令
var aiDisableCmd = &cobra.Command{
	Use:   "disable <service_name>",
	Short: "禁用AI服务",
	Long:  "根据服务名称禁用AI服务",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)

		err := aiServiceManager.SetServiceEnabled(serviceName, false)
		if err != nil {
			fmt.Printf("❌ 禁用AI服务失败: %v\n", err)
			return
		}

		// 更新配置
		for i := range globalConfig.AIServices {
			if globalConfig.AIServices[i].Name == serviceName {
				globalConfig.AIServices[i].Enabled = false
				break
			}
		}

		fmt.Printf("✅ AI服务禁用成功: %s\n", serviceName)
	},
}

// aiTestCmd 代表测试AI服务命令
var aiTestCmd = &cobra.Command{
	Use:   "test <service_name>",
	Short: "测试AI服务",
	Long:  "测试指定AI服务的连接和响应",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)

		// 获取服务
		service, err := aiServiceManager.GetNextService()
		if err != nil {
			fmt.Printf("❌ 获取AI服务失败: %v\n", err)
			return
		}

		if service.Name != serviceName {
			fmt.Printf("❌ 未找到AI服务: %s\n", serviceName)
			return
		}

		fmt.Printf("🧪 测试AI服务: %s\n", serviceName)
		fmt.Printf("   端点: %s\n", service.Endpoint)
		fmt.Printf("   模型: %s\n", service.Model)

		// 这里可以添加实际的API测试逻辑
		fmt.Println("✅ AI服务测试完成 (模拟)")
	},
}

// aiStatsCmd 代表AI服务统计命令
var aiStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "显示AI服务统计",
	Long:  "显示AI服务管理器的统计信息",
	Run: func(cmd *cobra.Command, args []string) {
		aiServiceManager := ai.NewAIServiceManager(globalConfig.AIServices)
		stats := aiServiceManager.GetStats()

		fmt.Println("📊 AI服务统计:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("总服务数: %d\n", stats["total_services"])
		fmt.Printf("启用服务: %d\n", stats["enabled_services"])
		fmt.Printf("限流服务: %d\n", stats["rate_limited_services"])
		fmt.Printf("当前服务索引: %d\n", stats["current_service_index"])
		fmt.Printf("故障转移: %t\n", stats["fallback_enabled"])
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)

	// 添加AI子命令
	aiCmd.AddCommand(aiListCmd)
	aiCmd.AddCommand(aiAddCmd)
	aiCmd.AddCommand(aiRemoveCmd)
	aiCmd.AddCommand(aiEnableCmd)
	aiCmd.AddCommand(aiDisableCmd)
	aiCmd.AddCommand(aiTestCmd)
	aiCmd.AddCommand(aiStatsCmd)

	// 添加AI服务标志
	aiAddCmd.Flags().StringVar(&aiName, "name", "", "服务名称")
	aiAddCmd.Flags().StringVar(&aiEndpoint, "endpoint", "", "API端点")
	aiAddCmd.Flags().StringVar(&aiToken, "token", "", "API Token")
	aiAddCmd.Flags().StringVar(&aiModel, "model", "", "模型名称")
	aiAddCmd.Flags().IntVar(&aiPriority, "priority", 100, "优先级 (数字越小优先级越高)")
	aiAddCmd.Flags().BoolVar(&aiEnabled, "enabled", true, "是否启用服务")
}
