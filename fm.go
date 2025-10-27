package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

// 创建文件监控器
func NewFileMonitor(filePath string) *FileMonitor {
	return &FileMonitor{
		filePath:  filePath,
		callbacks: make([]func(string, []byte), 0),
		stopChan:  make(chan bool),
	}
}

// 添加文件变化回调
func (fm *FileMonitor) AddCallback(callback func(string, []byte)) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fm.callbacks = append(fm.callbacks, callback)
}

// 启动文件监控
func (fm *FileMonitor) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	fm.watcher = watcher

	// 添加文件监控
	if err := watcher.Add(fm.filePath); err != nil {
		return err
	}

	// 启动监控协程
	go fm.monitor()

	return nil
}

// 监控文件变化
func (fm *FileMonitor) monitor() {
	for {
		select {
		case event := <-fm.watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				fm.handleFileChange()
			}
		case err := <-fm.watcher.Errors:
			if err != nil {
				fmt.Printf("文件监控错误: %v\n", err)
			}
		case <-fm.stopChan:
			return
		}
	}
}

// 处理文件变化
func (fm *FileMonitor) handleFileChange() {
	// 读取文件内容
	data, err := os.ReadFile(fm.filePath)
	if err != nil {
		return
	}

	// 调用所有回调
	fm.mutex.RLock()
	callbacks := make([]func(string, []byte), len(fm.callbacks))
	copy(callbacks, fm.callbacks)
	fm.mutex.RUnlock()

	for _, callback := range callbacks {
		callback(fm.filePath, data)
	}
}

// 停止文件监控
func (fm *FileMonitor) Stop() {
	close(fm.stopChan)
	if fm.watcher != nil {
		fm.watcher.Close()
	}
}
