package main

import "math/rand"

// 创建负载均衡器
func NewLoadBalancer(strategy string) *LoadBalancer {
	return &LoadBalancer{
		strategy:    strategy,
		workers:     make([]*Worker, 0),
		workerStats: make(map[int]*WorkerStats),
	}
}

// 选择工作协程
func (lb *LoadBalancer) SelectWorker() *Worker {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if len(lb.workers) == 0 {
		return nil
	}

	switch lb.strategy {
	case "round_robin":
		return lb.selectRoundRobin()
	case "least_loaded":
		return lb.selectLeastLoaded()
	case "random":
		return lb.selectRandom()
	default:
		return lb.selectRoundRobin()
	}
}

// 轮询选择
func (lb *LoadBalancer) selectRoundRobin() *Worker {
	if len(lb.workers) == 0 {
		return nil
	}

	worker := lb.workers[lb.currentIndex]
	lb.currentIndex = (lb.currentIndex + 1) % len(lb.workers)
	return worker
}

// 选择负载最低的工作协程
func (lb *LoadBalancer) selectLeastLoaded() *Worker {
	if len(lb.workers) == 0 {
		return nil
	}

	var selectedWorker *Worker
	minLoad := int64(^uint64(0) >> 1)

	for _, worker := range lb.workers {
		if stats, exists := lb.workerStats[worker.ID]; exists {
			if stats.CurrentLoad < minLoad {
				minLoad = stats.CurrentLoad
				selectedWorker = worker
			}
		}
	}

	if selectedWorker == nil {
		return lb.workers[0]
	}

	return selectedWorker
}

// 随机选择
func (lb *LoadBalancer) selectRandom() *Worker {
	if len(lb.workers) == 0 {
		return nil
	}

	index := rand.Intn(len(lb.workers))
	return lb.workers[index]
}

// 更新工作协程统计
func (lb *LoadBalancer) UpdateWorkerStats(workerID int, stats *WorkerStats) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.workerStats[workerID] = stats
}
