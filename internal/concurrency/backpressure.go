package concurrency

import (
	"fmt"
	"os"
)

// 创建背压控制器
func NewBackpressureController(threshold int) *BackpressureController {
	return &BackpressureController{
		threshold: threshold,
		callbacks: make([]func(int64), 0),
	}
}

// 检查背压
func (bc *BackpressureController) CheckBackpressure() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return int(bc.currentLoad) >= bc.threshold
}

// 增加负载
func (bc *BackpressureController) AddLoad(load int64) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.currentLoad += load

	// 检查是否触发背压
	if int(bc.currentLoad) >= bc.threshold {
		bc.blockedCount++
		// 触发回调
		for _, callback := range bc.callbacks {
			callback(bc.currentLoad)
		}
	}
}

// 减少负载
func (bc *BackpressureController) RemoveLoad(load int64) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.currentLoad -= load
	if bc.currentLoad < 0 {
		bc.currentLoad = 0
	}
}

// 拒绝任务
func (bc *BackpressureController) RejectTask() {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.rejectedCount++
}

// 添加背压回调
func (bc *BackpressureController) AddCallback(callback func(int64)) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	bc.callbacks = append(bc.callbacks, callback)
}

// 测试背压控制功能
func handleBackpressureTest() {
	fmt.Println("🔄 测试背压控制功能...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	// 创建背压控制器
	backpressure := NewBackpressureController(5)

	// 添加回调
	backpressure.AddCallback(func(load int64) {
		fmt.Printf("   ⚠️  背压触发，当前负载: %d\n", load)
	})

	// 测试正常负载
	fmt.Println("1. 测试正常负载...")
	for i := 0; i < 3; i++ {
		backpressure.AddLoad(1)
		fmt.Printf("   ✅ 添加负载 %d，当前负载: %d\n", i+1, backpressure.currentLoad)
	}

	// 测试背压触发
	fmt.Println("2. 测试背压触发...")
	for i := 0; i < 5; i++ {
		backpressure.AddLoad(1)
		fmt.Printf("   📊 添加负载 %d，当前负载: %d，背压状态: %t\n",
			i+4, backpressure.currentLoad, backpressure.CheckBackpressure())
	}

	// 测试负载减少
	fmt.Println("3. 测试负载减少...")
	for i := 0; i < 3; i++ {
		backpressure.RemoveLoad(1)
		fmt.Printf("   ✅ 减少负载 %d，当前负载: %d，背压状态: %t\n",
			i+1, backpressure.currentLoad, backpressure.CheckBackpressure())
	}

	// 测试任务拒绝
	fmt.Println("4. 测试任务拒绝...")
	for i := 0; i < 3; i++ {
		backpressure.RejectTask()
		fmt.Printf("   ❌ 拒绝任务 %d\n", i+1)
	}

	fmt.Println("\n✅ 背压控制功能测试完成")
}
