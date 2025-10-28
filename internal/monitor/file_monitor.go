package monitor

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// 文件监控器
type FileMonitor struct {
	watcher   *fsnotify.Watcher
	files     map[string]*MonitoredFile
	callbacks map[string]func(string, string) // filepath -> callback
	mutex     sync.RWMutex
	stopChan  chan bool
}

// 被监控的文件
type MonitoredFile struct {
	Path      string    `json:"path"`
	LastPos   int64     `json:"last_pos"`
	LastMod   time.Time `json:"last_mod"`
	Size      int64     `json:"size"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// 创建新的文件监控器
func NewFileMonitor() (*FileMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %w", err)
	}

	fm := &FileMonitor{
		watcher:   watcher,
		files:     make(map[string]*MonitoredFile),
		callbacks: make(map[string]func(string, string)),
		stopChan:  make(chan bool),
	}

	// 启动监控协程
	go fm.watch()

	return fm, nil
}

// 添加文件监控
func (fm *FileMonitor) AddFile(filePath string, callback func(string, string)) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 获取文件信息
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 添加文件到监控列表
	fm.files[filePath] = &MonitoredFile{
		Path:      filePath,
		LastPos:   0,
		LastMod:   info.ModTime(),
		Size:      info.Size(),
		IsActive:  true,
		CreatedAt: time.Now(),
	}

	// 添加回调函数
	fm.callbacks[filePath] = callback

	// 添加文件到 fsnotify
	dir := filepath.Dir(filePath)
	if err := fm.watcher.Add(dir); err != nil {
		return fmt.Errorf("添加目录监控失败: %w", err)
	}

	return nil
}

// 移除文件监控
func (fm *FileMonitor) RemoveFile(filePath string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 从监控列表移除
	delete(fm.files, filePath)
	delete(fm.callbacks, filePath)

	return nil
}

// 监控协程
func (fm *FileMonitor) watch() {
	for {
		select {
		case event := <-fm.watcher.Events:
			fm.handleEvent(event)
		case err := <-fm.watcher.Errors:
			fmt.Printf("文件监控错误: %v\n", err)
		case <-fm.stopChan:
			return
		}
	}
}

// 处理文件事件
func (fm *FileMonitor) handleEvent(event fsnotify.Event) {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	// 查找匹配的文件
	for filePath, monitoredFile := range fm.files {
		if !monitoredFile.IsActive {
			continue
		}

		// 检查是否是目标文件的事件
		if event.Name == filePath {
			if event.Op&fsnotify.Write == fsnotify.Write {
				fm.handleFileWrite(filePath, monitoredFile)
			}
		}
	}
}

// 处理文件写入事件
func (fm *FileMonitor) handleFileWrite(filePath string, monitoredFile *MonitoredFile) {
	// 获取文件当前大小
	info, err := os.Stat(filePath)
	if err != nil {
		return
	}

	currentSize := info.Size()
	if currentSize <= monitoredFile.LastPos {
		return
	}

	// 读取新增内容
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	// 定位到上次读取位置
	file.Seek(monitoredFile.LastPos, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// 调用回调函数
			if callback, exists := fm.callbacks[filePath]; exists {
				callback(filePath, line)
			}
		}
	}

	// 更新文件状态
	monitoredFile.LastPos = currentSize
	monitoredFile.LastMod = info.ModTime()
	monitoredFile.Size = currentSize
}

// 停止监控
func (fm *FileMonitor) Stop() {
	fm.stopChan <- true
	fm.watcher.Close()
}

// 获取监控状态
func (fm *FileMonitor) GetStatus() map[string]interface{} {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	status := make(map[string]interface{})
	status["total_files"] = len(fm.files)
	status["active_files"] = 0

	for _, file := range fm.files {
		if file.IsActive {
			status["active_files"] = status["active_files"].(int) + 1
		}
	}

	return status
}

// 获取文件列表
func (fm *FileMonitor) GetFiles() []MonitoredFile {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()

	files := make([]MonitoredFile, 0, len(fm.files))
	for _, file := range fm.files {
		files = append(files, *file)
	}

	return files
}

// 暂停文件监控
func (fm *FileMonitor) PauseFile(filePath string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if file, exists := fm.files[filePath]; exists {
		file.IsActive = false
		return nil
	}

	return fmt.Errorf("文件未找到: %s", filePath)
}

// 恢复文件监控
func (fm *FileMonitor) ResumeFile(filePath string) error {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	if file, exists := fm.files[filePath]; exists {
		file.IsActive = true
		return nil
	}

	return fmt.Errorf("文件未找到: %s", filePath)
}
