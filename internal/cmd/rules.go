package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/config"
	"github.com/xurenlu/aipipe/internal/rule"
)

var (
	rulePattern     string
	ruleAction      string
	rulePriority    int
	ruleDescription string
	ruleCategory    string
	ruleColor       string
	ruleEnabled     bool
	ruleID          string
)

// rulesCmd 代表规则命令
var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "规则管理",
	Long: `管理过滤规则，包括添加、删除、启用、禁用和测试规则。

子命令:
  add       - 添加新规则
  list      - 列出所有规则
  remove    - 删除规则
  enable    - 启用规则
  disable   - 禁用规则
  test      - 测试规则
  stats     - 显示规则统计`,
}

// rulesAddCmd 代表添加规则命令
var rulesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "添加新规则",
	Long:  "添加新的过滤规则",
	Run: func(cmd *cobra.Command, args []string) {
		if rulePattern == "" {
			fmt.Println("❌ 请指定规则模式 (--pattern)")
			return
		}

		// 创建规则引擎
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)

		// 生成规则ID
		if ruleID == "" {
			ruleID = fmt.Sprintf("rule_%d", len(globalConfig.Rules)+1)
		}

		// 创建新规则
		newRule := config.FilterRule{
			ID:          ruleID,
			Name:        fmt.Sprintf("规则 %s", ruleID),
			Pattern:     rulePattern,
			Action:      ruleAction,
			Priority:    rulePriority,
			Description: ruleDescription,
			Enabled:     ruleEnabled,
			Category:    ruleCategory,
			Color:       ruleColor,
		}

		// 添加规则
		err := ruleEngine.AddRule(newRule)
		if err != nil {
			fmt.Printf("❌ 添加规则失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 规则添加成功: %s\n", ruleID)
		fmt.Printf("   模式: %s\n", rulePattern)
		fmt.Printf("   动作: %s\n", ruleAction)
		fmt.Printf("   优先级: %d\n", rulePriority)
	},
}

// rulesListCmd 代表列出规则命令
var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有规则",
	Long:  "列出所有配置的过滤规则",
	Run: func(cmd *cobra.Command, args []string) {
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)
		rules := ruleEngine.GetRules()

		if len(rules) == 0 {
			fmt.Println("📋 没有配置任何规则")
			return
		}

		fmt.Printf("📋 规则列表 (共 %d 个):\n", len(rules))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for _, rule := range rules {
			status := "❌ 禁用"
			if rule.Enabled {
				status = "✅ 启用"
			}

			fmt.Printf("ID: %s\n", rule.ID)
			fmt.Printf("  名称: %s\n", rule.Name)
			fmt.Printf("  模式: %s\n", rule.Pattern)
			fmt.Printf("  动作: %s\n", rule.Action)
			fmt.Printf("  优先级: %d\n", rule.Priority)
			fmt.Printf("  状态: %s\n", status)
			fmt.Printf("  分类: %s\n", rule.Category)
			if rule.Description != "" {
				fmt.Printf("  描述: %s\n", rule.Description)
			}
			fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		}
	},
}

// rulesRemoveCmd 代表删除规则命令
var rulesRemoveCmd = &cobra.Command{
	Use:   "remove <rule_id>",
	Short: "删除规则",
	Long:  "根据规则ID删除规则",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ruleID := args[0]
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)

		err := ruleEngine.RemoveRule(ruleID)
		if err != nil {
			fmt.Printf("❌ 删除规则失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 规则删除成功: %s\n", ruleID)
	},
}

// rulesEnableCmd 代表启用规则命令
var rulesEnableCmd = &cobra.Command{
	Use:   "enable <rule_id>",
	Short: "启用规则",
	Long:  "根据规则ID启用规则",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ruleID := args[0]
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)

		err := ruleEngine.SetRuleEnabled(ruleID, true)
		if err != nil {
			fmt.Printf("❌ 启用规则失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 规则启用成功: %s\n", ruleID)
	},
}

// rulesDisableCmd 代表禁用规则命令
var rulesDisableCmd = &cobra.Command{
	Use:   "disable <rule_id>",
	Short: "禁用规则",
	Long:  "根据规则ID禁用规则",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ruleID := args[0]
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)

		err := ruleEngine.SetRuleEnabled(ruleID, false)
		if err != nil {
			fmt.Printf("❌ 禁用规则失败: %v\n", err)
			return
		}

		fmt.Printf("✅ 规则禁用成功: %s\n", ruleID)
	},
}

// rulesTestCmd 代表测试规则命令
var rulesTestCmd = &cobra.Command{
	Use:   "test <rule_id> <test_line>",
	Short: "测试规则",
	Long:  "使用测试日志行测试规则匹配",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ruleID := args[0]
		testLine := args[1]
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)

		matched, err := ruleEngine.TestRule(ruleID, testLine)
		if err != nil {
			fmt.Printf("❌ 测试规则失败: %v\n", err)
			return
		}

		if matched {
			fmt.Printf("✅ 规则匹配成功: %s\n", ruleID)
		} else {
			fmt.Printf("❌ 规则不匹配: %s\n", ruleID)
		}
	},
}

// rulesStatsCmd 代表规则统计命令
var rulesStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "显示规则统计",
	Long:  "显示规则引擎的统计信息",
	Run: func(cmd *cobra.Command, args []string) {
		ruleEngine := rule.NewRuleEngine(globalConfig.Rules)
		stats := ruleEngine.GetStats()

		fmt.Println("📊 规则统计:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("总规则数: %d\n", stats.TotalRules)
		fmt.Printf("启用规则: %d\n", stats.EnabledRules)
		fmt.Printf("禁用规则: %d\n", stats.DisabledRules)
		fmt.Printf("最后更新: %s\n", stats.LastUpdated.Format("2006-01-02 15:04:05"))

		if len(stats.MatchCounts) > 0 {
			fmt.Println("\n匹配统计:")
			for ruleID, count := range stats.MatchCounts {
				fmt.Printf("  %s: %d 次\n", ruleID, count)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(rulesCmd)

	// 添加规则子命令
	rulesCmd.AddCommand(rulesAddCmd)
	rulesCmd.AddCommand(rulesListCmd)
	rulesCmd.AddCommand(rulesRemoveCmd)
	rulesCmd.AddCommand(rulesEnableCmd)
	rulesCmd.AddCommand(rulesDisableCmd)
	rulesCmd.AddCommand(rulesTestCmd)
	rulesCmd.AddCommand(rulesStatsCmd)

	// 添加规则标志
	rulesAddCmd.Flags().StringVar(&rulePattern, "pattern", "", "规则模式 (正则表达式)")
	rulesAddCmd.Flags().StringVar(&ruleAction, "action", "filter", "规则动作 (filter, alert, ignore, highlight)")
	rulesAddCmd.Flags().IntVar(&rulePriority, "priority", 100, "规则优先级 (数字越小优先级越高)")
	rulesAddCmd.Flags().StringVar(&ruleDescription, "description", "", "规则描述")
	rulesAddCmd.Flags().StringVar(&ruleCategory, "category", "default", "规则分类")
	rulesAddCmd.Flags().StringVar(&ruleColor, "color", "", "高亮颜色")
	rulesAddCmd.Flags().BoolVar(&ruleEnabled, "enabled", true, "是否启用规则")
	rulesAddCmd.Flags().StringVar(&ruleID, "id", "", "规则ID (可选)")
}
