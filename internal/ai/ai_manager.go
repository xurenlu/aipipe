package ai

import (
	"fmt"
	"sync"
	"time"

	"github.com/xurenlu/aipipe/internal/config"
)

// AI 服务管理器
type AIServiceManager struct {
	services    []config.AIService
	current     int
	fallback    bool
	rateLimiter map[string]time.Time
	mutex       sync.RWMutex
}

// 创建新的 AI 服务管理器
func NewAIServiceManager(services []config.AIService) *AIServiceManager {
	asm := &AIServiceManager{
		services:    services,
		current:     0,
		fallback:    false,
		rateLimiter: make(map[string]time.Time),
	}
	
	// 按优先级排序服务
	asm.sortServices()
	
	return asm
}

// 按优先级排序服务
func (asm *AIServiceManager) sortServices() {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	
	// 简单的冒泡排序，按优先级升序排列
	for i := 0; i < len(asm.services)-1; i++ {
		for j := 0; j < len(asm.services)-i-1; j++ {
			if asm.services[j].Priority > asm.services[j+1].Priority {
				asm.services[j], asm.services[j+1] = asm.services[j+1], asm.services[j]
			}
		}
	}
}

// 获取下一个可用的 AI 服务
func (asm *AIServiceManager) GetNextService() (*config.AIService, error) {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	
	if len(asm.services) == 0 {
		return nil, fmt.Errorf("没有可用的 AI 服务")
	}
	
	// 尝试找到下一个可用的服务
	for i := 0; i < len(asm.services); i++ {
		service := &asm.services[asm.current]
		
		// 检查服务是否启用且未被限流
		if service.Enabled && !asm.isRateLimited(service.Name) {
			asm.current = (asm.current + 1) % len(asm.services)
			return service, nil
		}
		
		asm.current = (asm.current + 1) % len(asm.services)
	}
	
	// 如果所有服务都被限流，返回第一个服务（强制使用）
	if asm.fallback {
		service := &asm.services[0]
		return service, nil
	}
	
	return nil, fmt.Errorf("所有 AI 服务都不可用或被限流")
}

// 检查服务是否被限流
func (asm *AIServiceManager) isRateLimited(serviceName string) bool {
	if lastCall, exists := asm.rateLimiter[serviceName]; exists {
		// 检查是否在限流窗口内（假设 1 分钟限流）
		return time.Since(lastCall) < time.Minute
	}
	return false
}

// 记录服务调用
func (asm *AIServiceManager) RecordCall(serviceName string) {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	
	asm.rateLimiter[serviceName] = time.Now()
}

// 获取服务统计信息
func (asm *AIServiceManager) GetStats() map[string]interface{} {
	asm.mutex.RLock()
	defer asm.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["total_services"] = len(asm.services)
	stats["enabled_services"] = 0
	stats["rate_limited_services"] = 0
	
	for _, service := range asm.services {
		if service.Enabled {
			stats["enabled_services"] = stats["enabled_services"].(int) + 1
		}
		if asm.isRateLimited(service.Name) {
			stats["rate_limited_services"] = stats["rate_limited_services"].(int) + 1
		}
	}
	
	stats["current_service_index"] = asm.current
	stats["fallback_enabled"] = asm.fallback
	
	return stats
}

// 设置服务启用状态
func (asm *AIServiceManager) SetServiceEnabled(serviceName string, enabled bool) error {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	
	for i := range asm.services {
		if asm.services[i].Name == serviceName {
			asm.services[i].Enabled = enabled
			return nil
		}
	}
	
	return fmt.Errorf("未找到服务: %s", serviceName)
}

// 获取所有服务
func (asm *AIServiceManager) GetServices() []config.AIService {
	asm.mutex.RLock()
	defer asm.mutex.RUnlock()
	
	// 返回服务副本
	services := make([]config.AIService, len(asm.services))
	copy(services, asm.services)
	return services
}

// 设置故障转移模式
func (asm *AIServiceManager) SetFallbackMode(enabled bool) {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	
	asm.fallback = enabled
}

// 清除限流记录
func (asm *AIServiceManager) ClearRateLimit() {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()
	
	asm.rateLimiter = make(map[string]time.Time)
}
