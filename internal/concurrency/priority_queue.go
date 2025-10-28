package concurrency

import "sort"

// 创建优先级队列
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		jobs:       make([]ProcessingJob, 0),
		priorities: make(map[string]TaskPriority),
	}
}

// 添加任务
func (pq *PriorityQueue) AddJob(job ProcessingJob, priority TaskPriority) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.priorities[job.ID] = priority
	pq.jobs = append(pq.jobs, job)

	// 按优先级排序
	pq.sortByPriority()
}

// 获取下一个任务
func (pq *PriorityQueue) GetNextJob() *ProcessingJob {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.jobs) == 0 {
		return nil
	}

	job := pq.jobs[0]
	pq.jobs = pq.jobs[1:]
	delete(pq.priorities, job.ID)

	return &job
}

// 按优先级排序
func (pq *PriorityQueue) sortByPriority() {
	sort.Slice(pq.jobs, func(i, j int) bool {
		priorityI := pq.priorities[pq.jobs[i].ID]
		priorityJ := pq.priorities[pq.jobs[j].ID]
		return priorityI > priorityJ // 高优先级在前
	})
}

// 获取队列长度
func (pq *PriorityQueue) Length() int {
	pq.mutex.RLock()
	defer pq.mutex.RUnlock()

	return len(pq.jobs)
}
