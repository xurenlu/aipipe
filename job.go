package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// 工作池相关结构

// 工作池配置
type WorkerPoolConfig struct {
	MaxWorkers   int           `json:"max_workers"`   // 最大工作协程数
	QueueSize    int           `json:"queue_size"`    // 队列大小
	BatchSize    int           `json:"batch_size"`    // 批处理大小
	Timeout      time.Duration `json:"timeout"`       // 超时时间
	RetryCount   int           `json:"retry_count"`   // 重试次数
	BackoffDelay time.Duration `json:"backoff_delay"` // 退避延迟
	Enabled      bool          `json:"enabled"`       // 是否启用
}

// 处理任务
type ProcessingJob struct {
	ID        string                 `json:"id"`
	Lines     []string               `json:"lines"`
	Format    string                 `json:"format"`
	Priority  int                    `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// 处理结果
type ProcessingResult struct {
	JobID          string                 `json:"job_id"`
	ProcessedLines int                    `json:"processed_lines"`
	FilteredLines  int                    `json:"filtered_lines"`
	AlertedLines   int                    `json:"alerted_lines"`
	ErrorCount     int                    `json:"error_count"`
	ProcessingTime time.Duration          `json:"processing_time"`
	CreatedAt      time.Time              `json:"created_at"`
	Results        []LogAnalysis          `json:"results"`
	Errors         []string               `json:"errors"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// 工作池统计
type WorkerPoolStats struct {
	TotalJobs     int64         `json:"total_jobs"`
	CompletedJobs int64         `json:"completed_jobs"`
	FailedJobs    int64         `json:"failed_jobs"`
	ActiveWorkers int           `json:"active_workers"`
	QueueLength   int           `json:"queue_length"`
	AverageTime   time.Duration `json:"average_time"`
	TotalLines    int64         `json:"total_lines"`
	ErrorRate     float64       `json:"error_rate"`
	Throughput    float64       `json:"throughput"` // 每秒处理行数
}

// 工作池
type WorkerPool struct {
	config     WorkerPoolConfig
	jobQueue   chan ProcessingJob
	resultChan chan ProcessingResult
	workerPool chan chan ProcessingJob
	workers    []*Worker
	quit       chan bool
	stats      WorkerPoolStats
	mutex      sync.RWMutex
	startTime  time.Time
}

// 工作协程
type Worker struct {
	ID            int
	WorkerPool    chan chan ProcessingJob
	JobChannel    chan ProcessingJob
	Quit          chan bool
	WorkerPoolRef *WorkerPool
}

// 工作池方法

// 创建新的工作池
func NewWorkerPool(config WorkerPoolConfig) *WorkerPool {
	wp := &WorkerPool{
		config:     config,
		jobQueue:   make(chan ProcessingJob, config.QueueSize),
		resultChan: make(chan ProcessingResult, config.QueueSize),
		workerPool: make(chan chan ProcessingJob, config.MaxWorkers),
		workers:    make([]*Worker, 0, config.MaxWorkers),
		quit:       make(chan bool),
		startTime:  time.Now(),
	}

	// 创建工作协程
	for i := 0; i < config.MaxWorkers; i++ {
		worker := NewWorker(i, wp)
		wp.workers = append(wp.workers, worker)
		worker.Start()
	}

	// 启动调度器
	go wp.dispatch()

	return wp
}

// 创建新的工作协程
func NewWorker(id int, wp *WorkerPool) *Worker {
	return &Worker{
		ID:            id,
		WorkerPool:    wp.workerPool,
		JobChannel:    make(chan ProcessingJob),
		Quit:          make(chan bool),
		WorkerPoolRef: wp,
	}
}

// 启动工作协程
func (w *Worker) Start() {
	go func() {
		for {
			// 将工作协程的通道注册到工作池
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// 处理任务
				w.processJob(job)
			case <-w.Quit:
				return
			}
		}
	}()
}

// 停止工作协程
func (w *Worker) Stop() {
	go func() {
		w.Quit <- true
	}()
}

// 处理任务
func (w *Worker) processJob(job ProcessingJob) {
	startTime := time.Now()

	// 更新统计
	w.WorkerPoolRef.mutex.Lock()
	w.WorkerPoolRef.stats.ActiveWorkers++
	w.WorkerPoolRef.mutex.Unlock()

	defer func() {
		w.WorkerPoolRef.mutex.Lock()
		w.WorkerPoolRef.stats.ActiveWorkers--
		w.WorkerPoolRef.mutex.Unlock()
	}()

	result := ProcessingResult{
		JobID:          job.ID,
		ProcessedLines: len(job.Lines),
		CreatedAt:      time.Now(),
		Results:        make([]LogAnalysis, 0),
		Errors:         make([]string, 0),
		Metadata:       make(map[string]interface{}),
	}

	// 处理每一行日志
	for _, line := range job.Lines {
		// 检查缓存
		logHash := generateLogHash(line)
		if cached, found := cacheManager.GetAIAnalysis(logHash); found {
			// 使用缓存结果
			result.Results = append(result.Results, LogAnalysis{
				Line:       line,
				Important:  true,
				Reason:     cached.Result,
				Confidence: cached.Confidence,
			})
			result.FilteredLines++
			continue
		}

		// 应用规则过滤
		if globalConfig.LocalFilter && ruleEngine != nil {
			filterResult := ruleEngine.Filter(line)
			if filterResult.ShouldIgnore {
				continue
			}
			if filterResult.ShouldProcess {
				// 需要AI分析
				analysis, err := analyzeLogLine(line, job.Format)
				if err != nil {
					result.ErrorCount++
					result.Errors = append(result.Errors, fmt.Sprintf("分析失败: %v", err))
					continue
				}

				// 缓存结果
				cacheResult := &AIAnalysisCache{
					LogHash:    logHash,
					Result:     analysis.Reason,
					Confidence: analysis.Confidence,
					Model:      globalConfig.Model,
					CreatedAt:  time.Now(),
					ExpiresAt:  time.Now().Add(globalConfig.Cache.AITTL),
				}
				cacheManager.SetAIAnalysis(logHash, cacheResult)

				result.Results = append(result.Results, *analysis)
				if analysis.Important {
					result.AlertedLines++
				}
			}
		} else {
			// 直接AI分析
			analysis, err := analyzeLogLine(line, job.Format)
			if err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, fmt.Sprintf("分析失败: %v", err))
				continue
			}

			// 缓存结果
			cacheResult := &AIAnalysisCache{
				LogHash:    logHash,
				Result:     analysis.Reason,
				Confidence: analysis.Confidence,
				Model:      globalConfig.Model,
				CreatedAt:  time.Now(),
				ExpiresAt:  time.Now().Add(globalConfig.Cache.AITTL),
			}
			cacheManager.SetAIAnalysis(logHash, cacheResult)

			result.Results = append(result.Results, *analysis)
			if analysis.Important {
				result.AlertedLines++
			}
		}
	}

	result.ProcessingTime = time.Since(startTime)

	// 更新统计
	w.WorkerPoolRef.mutex.Lock()
	w.WorkerPoolRef.stats.CompletedJobs++
	w.WorkerPoolRef.stats.TotalLines += int64(result.ProcessedLines)
	w.WorkerPoolRef.mutex.Unlock()

	// 发送结果
	w.WorkerPoolRef.resultChan <- result
}

// 调度器
func (wp *WorkerPool) dispatch() {
	for {
		select {
		case job := <-wp.jobQueue:
			// 获取可用的工作协程
			worker := <-wp.workerPool
			// 分配任务
			worker <- job

			// 更新统计
			wp.mutex.Lock()
			wp.stats.TotalJobs++
			wp.stats.QueueLength = len(wp.jobQueue)
			wp.mutex.Unlock()

		case <-wp.quit:
			// 停止所有工作协程
			for _, worker := range wp.workers {
				worker.Stop()
			}
			return
		}
	}
}

// 提交任务
func (wp *WorkerPool) SubmitJob(job ProcessingJob) error {
	if !wp.config.Enabled {
		return fmt.Errorf("工作池未启用")
	}

	select {
	case wp.jobQueue <- job:
		return nil
	default:
		return fmt.Errorf("工作队列已满")
	}
}

// 获取结果
func (wp *WorkerPool) GetResult() <-chan ProcessingResult {
	return wp.resultChan
}

// 获取统计信息
func (wp *WorkerPool) GetStats() WorkerPoolStats {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()

	// 计算吞吐量
	elapsed := time.Since(wp.startTime)
	if elapsed > 0 {
		wp.stats.Throughput = float64(wp.stats.TotalLines) / elapsed.Seconds()
	}

	// 计算错误率
	if wp.stats.TotalJobs > 0 {
		wp.stats.ErrorRate = float64(wp.stats.FailedJobs) / float64(wp.stats.TotalJobs) * 100
	}

	return wp.stats
}

// 停止工作池
func (wp *WorkerPool) Stop() {
	close(wp.quit)
}

// 工作池管理命令处理函数

// 显示工作池统计信息
func handleWorkerStats() {
	fmt.Println("📊 工作池统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := workerPool.GetStats()
	fmt.Printf("总任务数: %d\n", stats.TotalJobs)
	fmt.Printf("完成任务数: %d\n", stats.CompletedJobs)
	fmt.Printf("失败任务数: %d\n", stats.FailedJobs)
	fmt.Printf("活跃工作协程数: %d\n", stats.ActiveWorkers)
	fmt.Printf("队列长度: %d\n", stats.QueueLength)
	fmt.Printf("平均处理时间: %v\n", stats.AverageTime)
	fmt.Printf("总处理行数: %d\n", stats.TotalLines)
	fmt.Printf("错误率: %.2f%%\n", stats.ErrorRate)
	fmt.Printf("吞吐量: %.2f 行/秒\n", stats.Throughput)

	// 显示配置信息
	fmt.Println("\n工作池配置:")
	fmt.Printf("  最大工作协程数: %d\n", globalConfig.WorkerPool.MaxWorkers)
	fmt.Printf("  队列大小: %d\n", globalConfig.WorkerPool.QueueSize)
	fmt.Printf("  批处理大小: %d\n", globalConfig.WorkerPool.BatchSize)
	fmt.Printf("  超时时间: %v\n", globalConfig.WorkerPool.Timeout)
	fmt.Printf("  重试次数: %d\n", globalConfig.WorkerPool.RetryCount)
	fmt.Printf("  退避延迟: %v\n", globalConfig.WorkerPool.BackoffDelay)
	fmt.Printf("  启用状态: %t\n", globalConfig.WorkerPool.Enabled)
}

// 测试工作池功能
func handleWorkerTest() {
	fmt.Println("🧪 测试工作池功能...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 创建测试任务
	testLines := []string{
		"2024-01-01 10:00:00 [ERROR] Database connection failed",
		"2024-01-01 10:00:01 [INFO] User login successful",
		"2024-01-01 10:00:02 [WARN] High memory usage detected",
		"2024-01-01 10:00:03 [DEBUG] Processing request",
		"2024-01-01 10:00:04 [ERROR] File not found",
	}

	job := ProcessingJob{
		ID:        "test_job_1",
		Lines:     testLines,
		Format:    "java",
		Priority:  1,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"test": true,
		},
	}

	fmt.Println("1. 提交测试任务...")
	if err := workerPool.SubmitJob(job); err != nil {
		fmt.Printf("   ❌ 任务提交失败: %v\n", err)
		return
	}
	fmt.Println("   ✅ 任务提交成功")

	// 等待结果
	fmt.Println("2. 等待处理结果...")
	timeout := time.After(30 * time.Second)

	select {
	case result := <-workerPool.GetResult():
		fmt.Printf("   ✅ 任务处理完成: %s\n", result.JobID)
		fmt.Printf("   处理行数: %d\n", result.ProcessedLines)
		fmt.Printf("   过滤行数: %d\n", result.FilteredLines)
		fmt.Printf("   告警行数: %d\n", result.AlertedLines)
		fmt.Printf("   错误数: %d\n", result.ErrorCount)
		fmt.Printf("   处理时间: %v\n", result.ProcessingTime)
		fmt.Printf("   结果数: %d\n", len(result.Results))

		if len(result.Errors) > 0 {
			fmt.Println("   错误详情:")
			for i, err := range result.Errors {
				fmt.Printf("     %d. %s\n", i+1, err)
			}
		}

	case <-timeout:
		fmt.Println("   ❌ 任务处理超时")
		return
	}

	// 显示最终统计
	fmt.Println("\n最终工作池统计:")
	stats := workerPool.GetStats()
	fmt.Printf("  总任务数: %d\n", stats.TotalJobs)
	fmt.Printf("  完成任务数: %d\n", stats.CompletedJobs)
	fmt.Printf("  吞吐量: %.2f 行/秒\n", stats.Throughput)

	fmt.Println("\n✅ 工作池功能测试完成")
}

// 创建任务调度器
func NewTaskScheduler(workers []*Worker, loadBalancer *LoadBalancer) *TaskScheduler {
	return &TaskScheduler{
		priorityQueue: NewPriorityQueue(),
		workers:       workers,
		loadBalancer:  loadBalancer,
	}
}

// 提交任务
func (ts *TaskScheduler) SubmitTask(job ProcessingJob, priority TaskPriority) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// 检查是否有可用的工作协程
	if len(ts.workers) == 0 {
		return fmt.Errorf("没有可用的工作协程")
	}

	// 添加到优先级队列
	ts.priorityQueue.AddJob(job, priority)

	// 尝试立即分配任务
	ts.tryAssignTask()

	return nil
}

// 尝试分配任务
func (ts *TaskScheduler) tryAssignTask() {
	// 获取下一个任务
	job := ts.priorityQueue.GetNextJob()
	if job == nil {
		return
	}

	// 选择工作协程
	worker := ts.loadBalancer.SelectWorker()
	if worker == nil {
		// 没有可用工作协程，将任务放回队列
		ts.priorityQueue.AddJob(*job, ts.priorityQueue.priorities[job.ID])
		return
	}

	// 分配任务
	select {
	case worker.JobChannel <- *job:
		// 任务分配成功
		ts.stats.TotalJobs++
	default:
		// 工作协程忙，将任务放回队列
		ts.priorityQueue.AddJob(*job, ts.priorityQueue.priorities[job.ID])
	}
}

// 获取统计信息
func (ts *TaskScheduler) GetStats() ConcurrencyStats {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.stats
}
