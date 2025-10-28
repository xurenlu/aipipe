package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xurenlu/aipipe/internal/cache"
)

// cacheCmd 代表缓存命令
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "缓存管理",
	Long: `管理缓存系统，包括查看统计信息、清理缓存和配置缓存。

子命令:
  stats     - 显示缓存统计
  clear     - 清理缓存
  status    - 显示缓存状态`,
}

// cacheStatsCmd 代表缓存统计命令
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "显示缓存统计",
	Long:  "显示缓存系统的详细统计信息",
	Run: func(cmd *cobra.Command, args []string) {
		cacheManager := cache.NewCacheManager(globalConfig.Cache)
		stats := cacheManager.GetStats()

		fmt.Println("📊 缓存统计:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("总项目数: %d\n", stats.TotalItems)
		fmt.Printf("命中次数: %d\n", stats.HitCount)
		fmt.Printf("未命中次数: %d\n", stats.MissCount)
		fmt.Printf("命中率: %.2f%%\n", stats.HitRate*100)
		fmt.Printf("内存使用: %d 字节 (%.2f MB)\n", stats.MemoryUsage, float64(stats.MemoryUsage)/1024/1024)
		fmt.Printf("驱逐次数: %d\n", stats.EvictionCount)
		fmt.Printf("最后清理: %s\n", stats.LastCleanup.Format("2006-01-02 15:04:05"))
	},
}

// cacheClearCmd 代表清理缓存命令
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清理缓存",
	Long:  "清理所有缓存数据",
	Run: func(cmd *cobra.Command, args []string) {
		cacheManager := cache.NewCacheManager(globalConfig.Cache)
		
		// 获取清理前的统计
		statsBefore := cacheManager.GetStats()
		
		// 清理缓存
		cacheManager.Clear()
		
		// 获取清理后的统计
		statsAfter := cacheManager.GetStats()
		
		fmt.Printf("✅ 缓存清理完成\n")
		fmt.Printf("   清理前: %d 个项目\n", statsBefore.TotalItems)
		fmt.Printf("   清理后: %d 个项目\n", statsAfter.TotalItems)
	},
}

// cacheStatusCmd 代表缓存状态命令
var cacheStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "显示缓存状态",
	Long:  "显示缓存系统的配置和状态信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 缓存配置:")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("启用状态: %t\n", globalConfig.Cache.Enabled)
		fmt.Printf("最大大小: %d 字节 (%.2f MB)\n", globalConfig.Cache.MaxSize, float64(globalConfig.Cache.MaxSize)/1024/1024)
		fmt.Printf("最大项目数: %d\n", globalConfig.Cache.MaxItems)
		fmt.Printf("默认TTL: %s\n", globalConfig.Cache.DefaultTTL)
		fmt.Printf("AI TTL: %s\n", globalConfig.Cache.AITTL)
		fmt.Printf("规则TTL: %s\n", globalConfig.Cache.RuleTTL)
		fmt.Printf("配置TTL: %s\n", globalConfig.Cache.ConfigTTL)
		fmt.Printf("清理间隔: %s\n", globalConfig.Cache.CleanupInterval)
		
		if globalConfig.Cache.Enabled {
			cacheManager := cache.NewCacheManager(globalConfig.Cache)
			stats := cacheManager.GetStats()
			
			fmt.Println("\n📊 当前状态:")
			fmt.Printf("缓存项目: %d\n", stats.TotalItems)
			fmt.Printf("内存使用: %d 字节 (%.2f MB)\n", stats.MemoryUsage, float64(stats.MemoryUsage)/1024/1024)
			fmt.Printf("命中率: %.2f%%\n", stats.HitRate*100)
		}
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	
	// 添加缓存子命令
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheStatusCmd)
}
