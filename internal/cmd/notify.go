package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/notification"
)

// notifyCmd 代表通知命令
var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "通知管理",
	Long: `管理通知系统，包括测试通知、配置通知器和发送测试消息。

子命令:
  test      - 发送测试通知
  status    - 显示通知状态
  enable    - 启用通知器
  disable   - 禁用通知器`,
}

// notifyTestCmd 代表测试通知命令
var notifyTestCmd = &cobra.Command{
	Use:   "test",
	Short: "发送测试通知",
	Long:  "发送测试通知到所有启用的通知器",
	Run: func(cmd *cobra.Command, args []string) {
		notificationManager := notification.NewNotificationManager(globalConfig)

		message := &notification.NotificationMessage{
			Title:    "AIPipe 测试通知",
			Content:  "这是一条测试通知，用于验证通知系统是否正常工作。",
			Level:    "info",
			Source:   "AIPipe",
			Metadata: make(map[string]string),
		}

		err := notificationManager.Send(message)
		if err != nil {
			fmt.Printf("❌ 发送测试通知失败: %v\n", err)
			return
		}

		fmt.Println("✅ 测试通知发送成功")
	},
}

// notifyStatusCmd 代表通知状态命令
var notifyStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示通知状态",
	Long:  "显示所有通知器的状态信息",
	Run: func(cmd *cobra.Command, args []string) {
		notificationManager := notification.NewNotificationManager(globalConfig)

		enabledCount := notificationManager.GetEnabledCount()

		fmt.Println("📢 通知系统状态:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("启用的通知器: %d 个\n", enabledCount)

		// 显示各种通知器的状态
		fmt.Println("\n通知器详情:")

		// 邮件通知
		if globalConfig.Notifiers.Email.Enabled {
			fmt.Println("  ✅ 邮件通知: 已启用")
			fmt.Printf("    提供商: %s\n", globalConfig.Notifiers.Email.Provider)
			fmt.Printf("    收件人: %v\n", globalConfig.Notifiers.Email.ToEmails)
		} else {
			fmt.Println("  ❌ 邮件通知: 已禁用")
		}

		// 钉钉通知
		if globalConfig.Notifiers.DingTalk.Enabled {
			fmt.Println("  ✅ 钉钉通知: 已启用")
		} else {
			fmt.Println("  ❌ 钉钉通知: 已禁用")
		}

		// 企业微信通知
		if globalConfig.Notifiers.WeChat.Enabled {
			fmt.Println("  ✅ 企业微信通知: 已启用")
		} else {
			fmt.Println("  ❌ 企业微信通知: 已禁用")
		}

		// 飞书通知
		if globalConfig.Notifiers.Feishu.Enabled {
			fmt.Println("  ✅ 飞书通知: 已启用")
		} else {
			fmt.Println("  ❌ 飞书通知: 已禁用")
		}

		// Slack通知
		if globalConfig.Notifiers.Slack.Enabled {
			fmt.Println("  ✅ Slack通知: 已启用")
		} else {
			fmt.Println("  ❌ Slack通知: 已禁用")
		}

		// 自定义Webhook
		customCount := 0
		for _, webhook := range globalConfig.Notifiers.CustomWebhooks {
			if webhook.Enabled {
				customCount++
			}
		}
		if customCount > 0 {
			fmt.Printf("  ✅ 自定义Webhook: %d 个已启用\n", customCount)
		} else {
			fmt.Println("  ❌ 自定义Webhook: 无启用")
		}

		// 系统通知
		fmt.Println("  ✅ 系统通知: 已启用")
	},
}

// notifySendCmd 代表发送通知命令
var notifySendCmd = &cobra.Command{
	Use:   "send <title> <content>",
	Short: "发送自定义通知",
	Long:  "发送自定义标题和内容的通知",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		title := args[0]
		content := args[1]

		notificationManager := notification.NewNotificationManager(globalConfig)

		err := notificationManager.SendSimple(title, content, "info")
		if err != nil {
			fmt.Printf("❌ 发送通知失败: %v\n", err)
			return
		}

		fmt.Println("✅ 通知发送成功")
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)

	// 添加通知子命令
	notifyCmd.AddCommand(notifyTestCmd)
	notifyCmd.AddCommand(notifyStatusCmd)
	notifyCmd.AddCommand(notifySendCmd)
}
