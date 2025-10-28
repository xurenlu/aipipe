package concurrency

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// 并发统计
type ConcurrencyStats struct {
	TotalJobs        int64         `json:"total_jobs"`
	ProcessedJobs    int64         `json:"processed_jobs"`
	ActiveWorkers    int           `json:"active_workers"`
	BlockedJobs      int64         `json:"blocked_jobs"`
	RejectedJobs     int64         `json:"rejected_jobs"`
	AverageLatency   time.Duration `json:"average_latency"`
	Throughput       float64       `json:"throughput"`
	ErrorRate        float64       `json:"error_rate"`
	BackpressureRate float64       `json:"backpressure_rate"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// 并发控制器
type ConcurrencyController struct {
	config         ConcurrencyConfig
	backpressure   *BackpressureController
	loadBalancer   *LoadBalancer
	adaptiveScaler *AdaptiveScaler
	stats          ConcurrencyStats
	mutex          sync.RWMutex
	stopChan       chan bool
}

// 并发控制器方法

// 创建新的并发控制器
func NewConcurrencyController(config ConcurrencyConfig) *ConcurrencyController {
	cc := &ConcurrencyController{
		config:   config,
		stopChan: make(chan bool),
	}

	// 创建背压控制器
	cc.backpressure = &BackpressureController{
		threshold: config.BackpressureThreshold,
		callbacks: make([]func(int64), 0),
	}

	// 创建负载均衡器
	cc.loadBalancer = &LoadBalancer{
		strategy:    config.LoadBalanceStrategy,
		workers:     make([]*Worker, 0),
		workerStats: make(map[int]*WorkerStats),
	}

	// 创建自适应扩缩容器
	cc.adaptiveScaler = &AdaptiveScaler{
		config:      config,
		workerStats: make(map[int]*WorkerStats),
	}

	// 启动自适应扩缩容
	if config.Enabled && config.AdaptiveScaling {
		go cc.startAdaptiveScaling()
	}

	return cc
}

// 启动自适应扩缩容
func (cc *ConcurrencyController) startAdaptiveScaling() {
	ticker := time.NewTicker(cc.config.ScalingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cc.checkAndScale()
		case <-cc.stopChan:
			return
		}
	}
}

// 检查并执行扩缩容
func (cc *ConcurrencyController) checkAndScale() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	// 计算当前负载
	currentLoad := cc.calculateCurrentLoad()

	// 检查是否需要扩容
	if currentLoad > cc.config.ScaleUpThreshold && cc.adaptiveScaler.currentWorkers < cc.config.MaxWorkers {
		cc.scaleUp()
	}

	// 检查是否需要缩容
	if currentLoad < cc.config.ScaleDownThreshold && cc.adaptiveScaler.currentWorkers > cc.config.MinWorkers {
		cc.scaleDown()
	}
}

// 计算当前负载
func (cc *ConcurrencyController) calculateCurrentLoad() float64 {
	if cc.adaptiveScaler.currentWorkers == 0 {
		return 0
	}

	totalLoad := int64(0)
	for _, stats := range cc.adaptiveScaler.workerStats {
		totalLoad += stats.CurrentLoad
	}

	return float64(totalLoad) / float64(cc.adaptiveScaler.currentWorkers)
}

// 扩容
func (cc *ConcurrencyController) scaleUp() {
	if cc.adaptiveScaler.currentWorkers >= cc.config.MaxWorkers {
		return
	}

	// 创建新的工作协程
	newWorker := NewWorker(cc.adaptiveScaler.currentWorkers, workerPool)
	cc.loadBalancer.workers = append(cc.loadBalancer.workers, newWorker)
	cc.adaptiveScaler.currentWorkers++

	// 启动工作协程
	newWorker.Start()

	// 更新统计
	cc.adaptiveScaler.workerStats[newWorker.ID] = &WorkerStats{
		ID:           newWorker.ID,
		LastActivity: time.Now(),
		IsHealthy:    true,
	}

	cc.adaptiveScaler.lastScaleTime = time.Now()
}

// 缩容
func (cc *ConcurrencyController) scaleDown() {
	if cc.adaptiveScaler.currentWorkers <= cc.config.MinWorkers {
		return
	}

	// 找到负载最低的工作协程
	var targetWorker *Worker
	minLoad := int64(^uint64(0) >> 1)

	for _, worker := range cc.loadBalancer.workers {
		if stats, exists := cc.adaptiveScaler.workerStats[worker.ID]; exists {
			if stats.CurrentLoad < minLoad {
				minLoad = stats.CurrentLoad
				targetWorker = worker
			}
		}
	}

	if targetWorker != nil {
		// 停止工作协程
		targetWorker.Stop()

		// 从负载均衡器中移除
		for i, worker := range cc.loadBalancer.workers {
			if worker.ID == targetWorker.ID {
				cc.loadBalancer.workers = append(cc.loadBalancer.workers[:i], cc.loadBalancer.workers[i+1:]...)
				break
			}
		}

		// 更新统计
		delete(cc.adaptiveScaler.workerStats, targetWorker.ID)
		cc.adaptiveScaler.currentWorkers--
		cc.adaptiveScaler.lastScaleTime = time.Now()
	}
}

// 并发控制命令处理函数

// 显示并发控制统计信息
func handleConcurrencyStats() {
	fmt.Println("⚡ 并发控制统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := concurrencyController.stats
	fmt.Printf("总任务数: %d\n", stats.TotalJobs)
	fmt.Printf("已处理任务数: %d\n", stats.ProcessedJobs)
	fmt.Printf("活跃工作协程数: %d\n", stats.ActiveWorkers)
	fmt.Printf("阻塞任务数: %d\n", stats.BlockedJobs)
	fmt.Printf("拒绝任务数: %d\n", stats.RejectedJobs)
	fmt.Printf("平均延迟: %v\n", stats.AverageLatency)
	fmt.Printf("吞吐量: %.2f 任务/秒\n", stats.Throughput)
	fmt.Printf("错误率: %.2f%%\n", stats.ErrorRate)
	fmt.Printf("背压率: %.2f%%\n", stats.BackpressureRate)

	// 显示配置信息
	fmt.Println("\n并发控制配置:")
	fmt.Printf("  最大并发数: %d\n", globalConfig.Concurrency.MaxConcurrency)
	fmt.Printf("  背压阈值: %d\n", globalConfig.Concurrency.BackpressureThreshold)
	fmt.Printf("  负载均衡策略: %s\n", globalConfig.Concurrency.LoadBalanceStrategy)
	fmt.Printf("  自适应扩缩容: %t\n", globalConfig.Concurrency.AdaptiveScaling)
	fmt.Printf("  扩容阈值: %.2f\n", globalConfig.Concurrency.ScaleUpThreshold)
	fmt.Printf("  缩容阈值: %.2f\n", globalConfig.Concurrency.ScaleDownThreshold)
	fmt.Printf("  最小工作协程数: %d\n", globalConfig.Concurrency.MinWorkers)
	fmt.Printf("  最大工作协程数: %d\n", globalConfig.Concurrency.MaxWorkers)
	fmt.Printf("  扩缩容检查间隔: %v\n", globalConfig.Concurrency.ScalingInterval)
	fmt.Printf("  启用状态: %t\n", globalConfig.Concurrency.Enabled)
}

// 测试并发控制功能
func handleConcurrencyTest() {
	fmt.Println("🧪 测试并发控制功能...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 测试负载均衡器
	fmt.Println("1. 测试负载均衡器...")
	loadBalancer := NewLoadBalancer("round_robin")

	// 创建测试工作协程
	testWorkers := make([]*Worker, 3)
	for i := 0; i < 3; i++ {
		worker := NewWorker(i, workerPool)
		testWorkers[i] = worker
		loadBalancer.workers = append(loadBalancer.workers, worker)
	}

	// 测试轮询选择
	for i := 0; i < 6; i++ {
		worker := loadBalancer.SelectWorker()
		if worker != nil {
			fmt.Printf("   ✅ 轮询选择工作协程 %d\n", worker.ID)
		} else {
			fmt.Println("   ❌ 轮询选择失败")
		}
	}

	// 测试优先级队列
	fmt.Println("2. 测试优先级队列...")
	priorityQueue := NewPriorityQueue()

	// 添加不同优先级的任务
	jobs := []ProcessingJob{
		{ID: "job1", Lines: []string{"test1"}, Priority: 1},
		{ID: "job2", Lines: []string{"test2"}, Priority: 3},
		{ID: "job3", Lines: []string{"test3"}, Priority: 2},
	}

	for i, job := range jobs {
		priority := TaskPriority(i + 1)
		priorityQueue.AddJob(job, priority)
		fmt.Printf("   ✅ 添加任务 %s (优先级 %d)\n", job.ID, priority)
	}

	// 按优先级获取任务
	for i := 0; i < 3; i++ {
		job := priorityQueue.GetNextJob()
		if job != nil {
			fmt.Printf("   ✅ 获取任务 %s\n", job.ID)
		} else {
			fmt.Println("   ❌ 获取任务失败")
		}
	}

	// 测试任务调度器
	fmt.Println("3. 测试任务调度器...")
	scheduler := NewTaskScheduler(testWorkers, loadBalancer)

	// 提交任务
	testJob := ProcessingJob{
		ID:     "test_job",
		Lines:  []string{"test line"},
		Format: "java",
	}

	if err := scheduler.SubmitTask(testJob, PriorityHigh); err != nil {
		fmt.Printf("   ❌ 任务提交失败: %v\n", err)
	} else {
		fmt.Println("   ✅ 任务提交成功")
	}

	// 显示统计
	stats := scheduler.GetStats()
	fmt.Printf("  总任务数: %d\n", stats.TotalJobs)

	fmt.Println("\n✅ 并发控制功能测试完成")
}
