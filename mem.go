package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

// 内存管理器
type MemoryManager struct {
	config          MemoryConfig
	stats           MemoryStats
	streamProcessor *StreamProcessor
	mutex           sync.RWMutex
	lastGC          time.Time
	allocations     map[uintptr]int64
}

// 内存监控器
type MemoryMonitor struct {
	enabled       bool
	checkInterval time.Duration
	threshold     int64
	callbacks     []func(MemoryStats)
	mutex         sync.RWMutex
	stopChan      chan bool
}

// 内存池
type MemoryPool struct {
	pool          sync.Pool
	chunkSize     int
	maxChunks     int
	currentChunks int
	allocations   map[uintptr]int64
	mutex         sync.Mutex
}

// 内存分配器
type MemoryAllocator struct {
	pool           *MemoryPool
	allocations    map[uintptr]int64
	totalAllocated int64
	mutex          sync.RWMutex
}

// 内存管理器方法

// 创建新的内存管理器
func NewMemoryManager(config MemoryConfig) *MemoryManager {
	mm := &MemoryManager{
		config:      config,
		allocations: make(map[uintptr]int64),
		lastGC:      time.Now(),
	}

	// 创建流式处理器
	mm.streamProcessor = &StreamProcessor{
		BufferSize: config.StreamBufferSize,
		ChunkSize:  config.ChunkSize,
		Buffer:     make([]string, 0, config.StreamBufferSize),
	}

	// 启动内存监控
	if config.Enabled {
		go mm.startMemoryMonitoring()
	}

	return mm
}

// 启动内存监控
func (mm *MemoryManager) startMemoryMonitoring() {
	ticker := time.NewTicker(mm.config.MemoryCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		mm.checkMemoryUsage()
	}
}

// 检查内存使用情况
func (mm *MemoryManager) checkMemoryUsage() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 更新统计信息
	mm.stats.CurrentUsage = int64(m.Alloc)
	mm.stats.HeapSize = int64(m.HeapSys)
	mm.stats.StackSize = int64(m.StackSys)
	mm.stats.GCCount = int64(m.NumGC)
	mm.stats.GCTime = int64(m.PauseTotalNs)
	mm.stats.AllocCount = int64(m.Mallocs)
	mm.stats.FreeCount = int64(m.Frees)
	mm.stats.LastGC = time.Unix(0, int64(m.LastGC))

	// 更新峰值使用量
	if mm.stats.CurrentUsage > mm.stats.PeakUsage {
		mm.stats.PeakUsage = mm.stats.CurrentUsage
	}

	// 计算内存压力
	if mm.config.MemoryLimit > 0 {
		mm.stats.MemoryPressure = float64(mm.stats.CurrentUsage) / float64(mm.config.MemoryLimit)
	}

	// 检查是否需要垃圾回收
	if mm.config.AutoGC && mm.stats.CurrentUsage > mm.config.GCThreshold {
		mm.forceGC()
	}
}

// 强制垃圾回收
func (mm *MemoryManager) forceGC() {
	start := time.Now()
	runtime.GC()
	mm.lastGC = time.Now()

	// 更新统计
	mm.stats.GCCount++
	mm.stats.GCTime += int64(time.Since(start).Nanoseconds())
}

// 获取内存统计信息
func (mm *MemoryManager) GetStats() MemoryStats {
	// 更新当前统计
	mm.checkMemoryUsage()

	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	return mm.stats
}

// 分配内存
func (mm *MemoryManager) Allocate(size int64) uintptr {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// 检查内存限制
	if mm.config.MemoryLimit > 0 && mm.stats.CurrentUsage+size > mm.config.MemoryLimit {
		// 触发垃圾回收
		mm.forceGC()

		// 如果仍然超限，返回0
		if mm.stats.CurrentUsage+size > mm.config.MemoryLimit {
			return 0
		}
	}

	// 分配内存（这里简化处理，实际应该使用内存池）
	ptr := uintptr(0) // 简化实现
	mm.allocations[ptr] = size
	mm.stats.AllocCount++

	return ptr
}

// 释放内存
func (mm *MemoryManager) Free(ptr uintptr) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	if size, exists := mm.allocations[ptr]; exists {
		delete(mm.allocations, ptr)
		mm.stats.FreeCount++
		mm.stats.CurrentUsage -= size
	}
}

// 流式处理日志
func (mm *MemoryManager) ProcessStream(lines []string, processFunc func([]string) error) error {
	if !mm.config.Enabled {
		return processFunc(lines)
	}

	mm.streamProcessor.mutex.Lock()
	defer mm.streamProcessor.mutex.Unlock()

	// 添加到缓冲区
	mm.streamProcessor.Buffer = append(mm.streamProcessor.Buffer, lines...)

	// 检查是否需要处理
	if len(mm.streamProcessor.Buffer) >= mm.streamProcessor.ChunkSize {
		// 处理当前块
		chunk := make([]string, mm.streamProcessor.ChunkSize)
		copy(chunk, mm.streamProcessor.Buffer[:mm.streamProcessor.ChunkSize])

		// 移除已处理的部分
		mm.streamProcessor.Buffer = mm.streamProcessor.Buffer[mm.streamProcessor.ChunkSize:]

		// 处理块
		if err := processFunc(chunk); err != nil {
			return err
		}

		mm.streamProcessor.TotalProcessed += int64(len(chunk))
	}

	return nil
}

// 刷新缓冲区
func (mm *MemoryManager) FlushBuffer() error {
	mm.streamProcessor.mutex.Lock()
	defer mm.streamProcessor.mutex.Unlock()

	if len(mm.streamProcessor.Buffer) > 0 {
		// 处理剩余数据
		if err := mm.streamProcessor.ProcessFunc(mm.streamProcessor.Buffer); err != nil {
			return err
		}

		mm.streamProcessor.TotalProcessed += int64(len(mm.streamProcessor.Buffer))
		mm.streamProcessor.Buffer = mm.streamProcessor.Buffer[:0] // 清空缓冲区
	}

	return nil
}

// 创建内存池
func NewMemoryPool(chunkSize, maxChunks int) *MemoryPool {
	mp := &MemoryPool{
		chunkSize:   chunkSize,
		maxChunks:   maxChunks,
		allocations: make(map[uintptr]int64),
	}

	// 初始化池
	mp.pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, chunkSize)
		},
	}

	return mp
}

// 从池中获取内存块
func (mp *MemoryPool) Get() []byte {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.currentChunks >= mp.maxChunks {
		return nil // 池已满
	}

	chunk := mp.pool.Get().([]byte)
	mp.currentChunks++
	return chunk
}

// 将内存块返回到池中
func (mp *MemoryPool) Put(chunk []byte) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.currentChunks > 0 {
		mp.pool.Put(chunk)
		mp.currentChunks--
	}
}

// 创建内存分配器
func NewMemoryAllocator(pool *MemoryPool) *MemoryAllocator {
	return &MemoryAllocator{
		pool:        pool,
		allocations: make(map[uintptr]int64),
	}
}

// 分配内存
func (ma *MemoryAllocator) Allocate(size int64) []byte {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	// 尝试从池中获取
	if size <= int64(ma.pool.chunkSize) {
		chunk := ma.pool.Get()
		if chunk != nil {
			ptr := uintptr(unsafe.Pointer(&chunk[0]))
			ma.allocations[ptr] = size
			ma.totalAllocated += size
			return chunk[:size]
		}
	}

	// 池中无法获取，直接分配
	chunk := make([]byte, size)
	ptr := uintptr(unsafe.Pointer(&chunk[0]))
	ma.allocations[ptr] = size
	ma.totalAllocated += size

	return chunk
}

// 释放内存
func (ma *MemoryAllocator) Free(chunk []byte) {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	ptr := uintptr(unsafe.Pointer(&chunk[0]))
	if size, exists := ma.allocations[ptr]; exists {
		delete(ma.allocations, ptr)
		ma.totalAllocated -= size

		// 尝试返回到池中
		ma.pool.Put(chunk)
	}
}

// 获取分配统计
func (ma *MemoryAllocator) GetStats() map[string]int64 {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	return map[string]int64{
		"total_allocated":    ma.totalAllocated,
		"active_allocations": int64(len(ma.allocations)),
	}
}

// 内存管理命令处理函数

// 显示内存统计信息
func handleMemoryStats() {
	fmt.Println("🧠 内存统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := memoryManager.GetStats()
	fmt.Printf("当前内存使用: %.2f MB\n", float64(stats.CurrentUsage)/(1024*1024))
	fmt.Printf("峰值内存使用: %.2f MB\n", float64(stats.PeakUsage)/(1024*1024))
	fmt.Printf("堆大小: %.2f MB\n", float64(stats.HeapSize)/(1024*1024))
	fmt.Printf("栈大小: %.2f MB\n", float64(stats.StackSize)/(1024*1024))
	fmt.Printf("垃圾回收次数: %d\n", stats.GCCount)
	fmt.Printf("垃圾回收时间: %v\n", time.Duration(stats.GCTime))
	fmt.Printf("分配次数: %d\n", stats.AllocCount)
	fmt.Printf("释放次数: %d\n", stats.FreeCount)
	fmt.Printf("上次垃圾回收: %v\n", stats.LastGC.Format("2006-01-02 15:04:05"))
	fmt.Printf("内存压力: %.2f%%\n", stats.MemoryPressure*100)

	// 显示配置信息
	fmt.Println("\n内存配置:")
	fmt.Printf("  最大内存使用: %.2f MB\n", float64(globalConfig.Memory.MaxMemoryUsage)/(1024*1024))
	fmt.Printf("  GC阈值: %.2f MB\n", float64(globalConfig.Memory.GCThreshold)/(1024*1024))
	fmt.Printf("  流式缓冲区大小: %d\n", globalConfig.Memory.StreamBufferSize)
	fmt.Printf("  分块大小: %d\n", globalConfig.Memory.ChunkSize)
	fmt.Printf("  内存检查间隔: %v\n", globalConfig.Memory.MemoryCheckInterval)
	fmt.Printf("  自动GC: %t\n", globalConfig.Memory.AutoGC)
	fmt.Printf("  内存限制: %.2f MB\n", float64(globalConfig.Memory.MemoryLimit)/(1024*1024))
	fmt.Printf("  启用状态: %t\n", globalConfig.Memory.Enabled)
}

// 测试内存管理功能
func handleMemoryTest() {
	fmt.Println("🧪 测试内存管理功能...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 测试内存分配
	fmt.Println("1. 测试内存分配...")
	ptr1 := memoryManager.Allocate(1024 * 1024) // 1MB
	if ptr1 != 0 {
		fmt.Println("   ✅ 1MB内存分配成功")
	} else {
		fmt.Println("   ❌ 1MB内存分配失败")
	}

	ptr2 := memoryManager.Allocate(2 * 1024 * 1024) // 2MB
	if ptr2 != 0 {
		fmt.Println("   ✅ 2MB内存分配成功")
	} else {
		fmt.Println("   ❌ 2MB内存分配失败")
	}

	// 测试流式处理
	fmt.Println("2. 测试流式处理...")
	testLines := []string{
		"2024-01-01 10:00:00 [INFO] Test log line 1",
		"2024-01-01 10:00:01 [ERROR] Test log line 2",
		"2024-01-01 10:00:02 [WARN] Test log line 3",
	}

	processFunc := func(lines []string) error {
		fmt.Printf("   📝 处理了 %d 行日志\n", len(lines))
		return nil
	}

	if err := memoryManager.ProcessStream(testLines, processFunc); err != nil {
		fmt.Printf("   ❌ 流式处理失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 流式处理成功")
	}

	// 测试内存池
	fmt.Println("3. 测试内存池...")
	pool := NewMemoryPool(1024, 10)
	chunk1 := pool.Get()
	if chunk1 != nil {
		fmt.Println("   ✅ 从内存池获取内存块成功")
		pool.Put(chunk1)
		fmt.Println("   ✅ 将内存块返回到池中成功")
	} else {
		fmt.Println("   ❌ 从内存池获取内存块失败")
	}

	// 测试内存分配器
	fmt.Println("4. 测试内存分配器...")
	allocator := NewMemoryAllocator(pool)
	chunk2 := allocator.Allocate(512)
	if chunk2 != nil {
		fmt.Println("   ✅ 内存分配器分配成功")
		allocator.Free(chunk2)
		fmt.Println("   ✅ 内存分配器释放成功")
	} else {
		fmt.Println("   ❌ 内存分配器分配失败")
	}

	// 释放测试内存
	if ptr1 != 0 {
		memoryManager.Free(ptr1)
	}
	if ptr2 != 0 {
		memoryManager.Free(ptr2)
	}

	// 显示最终统计
	fmt.Println("\n最终内存统计:")
	stats := memoryManager.GetStats()
	fmt.Printf("  当前内存使用: %.2f MB\n", float64(stats.CurrentUsage)/(1024*1024))
	fmt.Printf("  分配次数: %d\n", stats.AllocCount)
	fmt.Printf("  释放次数: %d\n", stats.FreeCount)
	fmt.Printf("  内存压力: %.2f%%\n", stats.MemoryPressure*100)

	fmt.Println("\n✅ 内存管理功能测试完成")
}

// 强制垃圾回收
func handleMemoryGC() {
	fmt.Println("🗑️  强制垃圾回收...")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 获取回收前统计
	statsBefore := memoryManager.GetStats()
	fmt.Printf("回收前内存使用: %.2f MB\n", float64(statsBefore.CurrentUsage)/(1024*1024))

	// 强制垃圾回收
	start := time.Now()
	runtime.GC()
	runtime.GC() // 执行两次确保完全回收
	elapsed := time.Since(start)

	// 获取回收后统计
	statsAfter := memoryManager.GetStats()
	fmt.Printf("回收后内存使用: %.2f MB\n", float64(statsAfter.CurrentUsage)/(1024*1024))
	fmt.Printf("回收时间: %v\n", elapsed)
	fmt.Printf("释放内存: %.2f MB\n", float64(statsBefore.CurrentUsage-statsAfter.CurrentUsage)/(1024*1024))

	fmt.Println("✅ 垃圾回收完成")
}
