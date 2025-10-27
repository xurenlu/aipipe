package main

import (
	"sync"
	"time"
)

// 性能指标收集器
type MetricsCollector struct {
	metrics PerformanceMetrics
	mutex   sync.RWMutex
}

// 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: PerformanceMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// 更新指标
func (mc *MetricsCollector) UpdateMetrics(processed, filtered, alerted, apiCalls, errors int64, processingTime time.Duration) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics.ProcessedLines += processed
	mc.metrics.FilteredLines += filtered
	mc.metrics.AlertedLines += alerted
	mc.metrics.APICalls += apiCalls
	mc.metrics.ErrorCount += errors
	mc.metrics.ProcessingTime += int64(processingTime.Milliseconds())
	mc.metrics.LastUpdated = time.Now()

	// 计算吞吐量
	elapsed := time.Since(mc.metrics.LastUpdated)
	if elapsed > 0 {
		mc.metrics.Throughput = float64(mc.metrics.ProcessedLines) / elapsed.Seconds()
	}

	// 计算平均延迟
	if mc.metrics.ProcessedLines > 0 {
		mc.metrics.AverageLatency = float64(mc.metrics.ProcessingTime) / float64(mc.metrics.ProcessedLines)
	}

	// 计算错误率
	if mc.metrics.ProcessedLines > 0 {
		mc.metrics.ErrorRate = float64(mc.metrics.ErrorCount) / float64(mc.metrics.ProcessedLines) * 100
	}
}

// 更新缓存指标
func (mc *MetricsCollector) UpdateCacheMetrics(hits, misses int64) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics.CacheHits += hits
	mc.metrics.CacheMisses += misses

	// 计算缓存命中率
	total := mc.metrics.CacheHits + mc.metrics.CacheMisses
	if total > 0 {
		mc.metrics.CacheHitRate = float64(mc.metrics.CacheHits) / float64(total) * 100
	}
}

// 更新内存使用
func (mc *MetricsCollector) UpdateMemoryUsage(usage int64) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.metrics.MemoryUsage = usage
}

// 获取指标
func (mc *MetricsCollector) GetMetrics() PerformanceMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return mc.metrics
}
