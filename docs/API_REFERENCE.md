# AIPipe API 参考文档 📚

**版本**: v1.1.0  
**更新时间**: 2024年10月28日  
**状态**: 生产就绪

## 📋 目录

1. [命令行接口](#命令行接口)
2. [配置API](#配置api)
3. [AI服务API](#ai服务api)
4. [缓存API](#缓存api)
5. [规则引擎API](#规则引擎api)
6. [工作池API](#工作池api)
7. [内存管理API](#内存管理api)
8. [并发控制API](#并发控制api)
9. [I/O优化API](#io优化api)
10. [通知API](#通知api)

## 🖥️ 命令行接口

### 基本用法

```bash
aipipe [选项] [文件...]
```

### 全局选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--config` | string | `~/.config/aipipe.json` | 配置文件路径 |
| `--help` | bool | false | 显示帮助信息 |
| `--version` | bool | false | 显示版本信息 |
| `--verbose` | bool | false | 详细输出模式 |
| `--debug` | bool | false | 调试模式 |

### 配置管理选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `--config-init` | bool | 启动配置向导 |
| `--config-test` | bool | 测试配置 |
| `--config-validate` | bool | 验证配置 |
| `--config-show` | bool | 显示当前配置 |
| `--config-template` | string | 生成配置模板 |

### 输出格式选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--output-format` | string | `table` | 输出格式 (json/csv/table/custom) |
| `--output-color` | bool | true | 启用颜色输出 |
| `--log-level` | string | `info` | 日志级别过滤 |

### 性能选项

| 选项 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `--workers` | int | 4 | 工作线程数 |
| `--batch-size` | int | 10 | 批处理大小 |
| `--cache-size` | int | 1000 | 缓存大小 |
| `--memory-limit` | string | `512MB` | 内存限制 |

### 示例

```bash
# 基本使用
aipipe /var/log/app.log

# 使用自定义配置
aipipe --config /path/to/config.json /var/log/app.log

# 启动配置向导
aipipe --config-init

# 测试配置
aipipe --config-test

# 生成配置模板
aipipe --config-template yaml > config.yaml

# 调试模式
aipipe --debug --verbose /var/log/app.log

# 自定义输出格式
aipipe --output-format json --log-level error /var/log/app.log
```

## ⚙️ 配置API

### Config 结构

```go
type Config struct {
    AI          AIConfig          `json:"ai"`
    Cache       CacheConfig       `json:"cache"`
    Worker      WorkerConfig      `json:"worker"`
    Memory      MemoryConfig      `json:"memory"`
    Concurrency ConcurrencyConfig `json:"concurrency"`
    IO          IOConfig          `json:"io"`
    OutputFormat OutputFormat     `json:"output_format"`
    LogLevel    LogLevelConfig    `json:"log_level"`
    Notifications NotificationsConfig `json:"notifications"`
}
```

### 配置加载函数

```go
// 加载配置文件
func loadConfig(configPath string) (*Config, error)

// 加载多源配置
func loadMultiSourceConfig() (*Config, error)

// 保存配置
func saveConfig(config *Config, path string) error

// 验证配置
func validateConfig(config *Config) error
```

### 配置向导

```go
// 创建配置向导
func NewConfigWizard() *ConfigWizard

// 启动向导
func (w *ConfigWizard) Start() (*Config, error)

// 更新配置
func (w *ConfigWizard) updateConfig(step *WizardStep, value string) error
```

### 配置验证

```go
// 创建配置验证器
func NewConfigValidator() *ConfigValidator

// 验证配置
func (v *ConfigValidator) Validate(config *Config) []ValidationError

// 获取错误
func (v *ConfigValidator) GetErrors() []ValidationError
```

## 🤖 AI服务API

### AIServiceManager 结构

```go
type AIServiceManager struct {
    services    []*AIService
    loadBalancer LoadBalancer
    healthChecker HealthChecker
    mutex       sync.RWMutex
}
```

### AI服务管理

```go
// 创建AI服务管理器
func NewAIServiceManager(config AIConfig) *AIServiceManager

// 添加AI服务
func (m *AIServiceManager) AddService(service *AIService) error

// 移除AI服务
func (m *AIServiceManager) RemoveService(id string) error

// 获取可用服务
func (m *AIServiceManager) GetAvailableService() (*AIService, error)

// 分析日志
func (m *AIServiceManager) AnalyzeLogs(logs []string) (*LogAnalysis, error)
```

### 负载均衡

```go
// 负载均衡策略
type LoadBalancer interface {
    SelectService(services []*AIService) *AIService
}

// 轮询负载均衡
type RoundRobinLoadBalancer struct{}

// 最少连接负载均衡
type LeastConnLoadBalancer struct{}

// 随机负载均衡
type RandomLoadBalancer struct{}
```

### 健康检查

```go
// 健康检查器
type HealthChecker struct {
    interval time.Duration
    timeout  time.Duration
}

// 开始健康检查
func (h *HealthChecker) Start(services []*AIService)

// 检查服务健康
func (h *HealthChecker) CheckHealth(service *AIService) bool
```

## 💾 缓存API

### CacheManager 结构

```go
type CacheManager struct {
    aiCache     *AICache
    ruleCache   *RuleCache
    configCache *ConfigCache
    stats       *CacheStats
}
```

### 缓存管理

```go
// 创建缓存管理器
func NewCacheManager(config CacheConfig) *CacheManager

// 获取缓存
func (m *CacheManager) Get(key string) (interface{}, bool)

// 设置缓存
func (m *CacheManager) Set(key string, value interface{}, ttl time.Duration) error

// 删除缓存
func (m *CacheManager) Delete(key string) error

// 清空缓存
func (m *CacheManager) Clear() error

// 获取统计信息
func (m *CacheManager) GetStats() *CacheStats
```

### 缓存类型

```go
// AI分析缓存
type AICache struct {
    cache map[string]*CachedAnalysis
    mutex sync.RWMutex
}

// 规则匹配缓存
type RuleCache struct {
    cache map[string]*CachedMatch
    mutex sync.RWMutex
}

// 配置缓存
type ConfigCache struct {
    cache map[string]*CachedConfig
    mutex sync.RWMutex
}
```

## 🔧 规则引擎API

### RuleEngine 结构

```go
type RuleEngine struct {
    rules       []*Rule
    compiled    map[string]*regexp.Regexp
    stats       *RuleStats
    mutex       sync.RWMutex
}
```

### 规则管理

```go
// 创建规则引擎
func NewRuleEngine() *RuleEngine

// 添加规则
func (e *RuleEngine) AddRule(rule *Rule) error

// 移除规则
func (e *RuleEngine) RemoveRule(id string) error

// 启用规则
func (e *RuleEngine) EnableRule(id string) error

// 禁用规则
func (e *RuleEngine) DisableRule(id string) error

// 测试规则
func (e *RuleEngine) TestRule(rule *Rule, text string) (bool, error)

// 匹配规则
func (e *RuleEngine) MatchRules(text string) []*Rule
```

### 规则结构

```go
type Rule struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Pattern     string    `json:"pattern"`
    Action      string    `json:"action"`
    Priority    int       `json:"priority"`
    Enabled     bool      `json:"enabled"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## 👷 工作池API

### WorkerPool 结构

```go
type WorkerPool struct {
    workers     []*Worker
    jobQueue    chan *Job
    resultQueue chan *Result
    config      WorkerConfig
}
```

### 工作池管理

```go
// 创建工作池
func NewWorkerPool(config WorkerConfig) *WorkerPool

// 启动工作池
func (p *WorkerPool) Start() error

// 停止工作池
func (p *WorkerPool) Stop() error

// 提交任务
func (p *WorkerPool) SubmitJob(job *Job) error

// 获取结果
func (p *WorkerPool) GetResult() *Result

// 获取统计信息
func (p *WorkerPool) GetStats() *WorkerStats
```

### 任务和结果

```go
// 任务结构
type Job struct {
    ID      string
    Type    string
    Data    interface{}
    Result  chan *Result
}

// 结果结构
type Result struct {
    JobID   string
    Success bool
    Data    interface{}
    Error   error
}
```

## 🧠 内存管理API

### MemoryManager 结构

```go
type MemoryManager struct {
    pool        *MemoryPool
    monitor     *MemoryMonitor
    gc          *GarbageCollector
    config      MemoryConfig
}
```

### 内存管理

```go
// 创建内存管理器
func NewMemoryManager(config MemoryConfig) *MemoryManager

// 分配内存
func (m *MemoryManager) Allocate(size int) ([]byte, error)

// 释放内存
func (m *MemoryManager) Free(data []byte) error

// 获取内存使用情况
func (m *MemoryManager) GetUsage() *MemoryUsage

// 强制垃圾回收
func (m *MemoryManager) ForceGC() error

// 获取统计信息
func (m *MemoryManager) GetStats() *MemoryStats
```

### 内存池

```go
// 内存池
type MemoryPool struct {
    pools map[int]*sync.Pool
    mutex sync.RWMutex
}

// 获取内存块
func (p *MemoryPool) Get(size int) []byte

// 归还内存块
func (p *MemoryPool) Put(data []byte)
```

## ⚡ 并发控制API

### ConcurrencyController 结构

```go
type ConcurrencyController struct {
    loadBalancer LoadBalancer
    backpressure BackpressureController
    priorityQueue PriorityQueue
    config       ConcurrencyConfig
}
```

### 并发控制

```go
// 创建并发控制器
func NewConcurrencyController(config ConcurrencyConfig) *ConcurrencyController

// 提交任务
func (c *ConcurrencyController) SubmitTask(task *Task) error

// 获取任务
func (c *ConcurrencyController) GetTask() *Task

// 完成任务
func (c *ConcurrencyController) CompleteTask(task *Task) error

// 获取统计信息
func (c *ConcurrencyController) GetStats() *ConcurrencyStats
```

### 背压控制

```go
// 背压控制器
type BackpressureController struct {
    maxQueueSize int
    currentSize  int
    mutex        sync.RWMutex
}

// 检查是否可以接受新任务
func (b *BackpressureController) CanAccept() bool

// 增加队列大小
func (b *BackpressureController) Increment() error

// 减少队列大小
func (b *BackpressureController) Decrement()
```

## 📁 I/O优化API

### IOOptimizer 结构

```go
type IOOptimizer struct {
    asyncIO     *AsyncIOProcessor
    batchIO     *BatchIOProcessor
    fileMonitor *FileMonitor
    config      IOConfig
}
```

### I/O优化

```go
// 创建I/O优化器
func NewIOOptimizer(config IOConfig) *IOOptimizer

// 异步读取
func (o *IOOptimizer) AsyncRead(filePath string) (*AsyncIOOperation, error)

// 异步写入
func (o *IOOptimizer) AsyncWrite(filePath string, data []byte) (*AsyncIOOperation, error)

// 刷新所有缓冲区
func (o *IOOptimizer) FlushAll() error

// 获取统计信息
func (o *IOOptimizer) GetStats() *IOStats
```

### 文件监控

```go
// 文件监控器
type FileMonitor struct {
    watcher *fsnotify.Watcher
    callbacks map[string][]FileChangeCallback
    mutex    sync.RWMutex
}

// 添加文件监控
func (m *FileMonitor) AddCallback(filePath string, callback FileChangeCallback) error

// 开始监控
func (m *FileMonitor) Start() error

// 停止监控
func (m *FileMonitor) Stop() error
```

## 📢 通知API

### 通知配置

```go
type NotificationsConfig struct {
    Email    EmailConfig    `json:"email"`
    Webhook  WebhookConfig  `json:"webhook"`
    DingTalk DingTalkConfig `json:"dingtalk"`
    WeChat   WeChatConfig   `json:"wechat"`
    Feishu   FeishuConfig   `json:"feishu"`
    Slack    SlackConfig    `json:"slack"`
}
```

### 通知发送

```go
// 发送通知
func sendNotification(analysis *LogAnalysis) error

// 发送邮件
func sendEmail(analysis *LogAnalysis) error

// 发送Webhook
func sendWebhook(analysis *LogAnalysis) error

// 发送钉钉
func sendDingTalk(analysis *LogAnalysis) error

// 发送微信
func sendWeChat(analysis *LogAnalysis) error

// 发送飞书
func sendFeishu(analysis *LogAnalysis) error

// 发送Slack
func sendSlack(analysis *LogAnalysis) error
```

## 📊 统计和监控API

### 统计信息

```go
// 获取所有统计信息
func GetStats() *SystemStats

// 获取AI服务统计
func GetAIServiceStats() *AIServiceStats

// 获取缓存统计
func GetCacheStats() *CacheStats

// 获取工作池统计
func GetWorkerStats() *WorkerStats

// 获取内存统计
func GetMemoryStats() *MemoryStats

// 获取并发统计
func GetConcurrencyStats() *ConcurrencyStats

// 获取I/O统计
func GetIOStats() *IOStats
```

### 性能监控

```go
// 性能指标
type PerformanceMetrics struct {
    ProcessingRate    float64 `json:"processing_rate"`
    MemoryUsage       int64   `json:"memory_usage"`
    CPUUsage          float64 `json:"cpu_usage"`
    CacheHitRate      float64 `json:"cache_hit_rate"`
    ErrorRate         float64 `json:"error_rate"`
    AverageLatency    int64   `json:"average_latency"`
}
```

## 🔧 工具函数

### 配置工具

```go
// 创建默认配置
func createDefaultConfig() *Config

// 测试AI连接
func testAIConnection(config AIConfig) error

// 验证URL
func validateURL(url string) error

// 验证Token
func validateToken(token string) error
```

### 日志工具

```go
// 获取日志上下文
func getLogContext(filePath string, lineNumber int, contextLines int) ([]string, error)

// 获取日志行
func getLogLines(filePath string, startLine, endLine int) ([]string, error)

// 获取日志行数
func getLogLineCount(filePath string) (int, error)
```

### 文件工具

```go
// 获取文件状态
func getLogFileState(filePath string) (*FileState, error)

// 保存文件状态
func saveLogFileState(filePath string, state *FileState) error

// 加载文件状态
func loadLogFileState(filePath string) (*FileState, error)
```

## 📝 错误处理

### 错误类型

```go
// 配置错误
type ConfigError struct {
    Field   string
    Message string
}

// 验证错误
type ValidationError struct {
    Field   string
    Message string
    Value   interface{}
}

// 系统错误
type SystemError struct {
    Component string
    Message   string
    Cause     error
}
```

### 错误处理

```go
// 错误处理器
type ErrorHandler struct {
    errors []error
    mutex  sync.RWMutex
}

// 添加错误
func (h *ErrorHandler) AddError(err error)

// 获取错误
func (h *ErrorHandler) GetErrors() []error

// 清除错误
func (h *ErrorHandler) ClearErrors()
```

---

**API状态**: ✅ 完整  
**版本**: v1.1.0  
**文档状态**: ✅ 最新  
**维护状态**: ✅ 活跃  

*最后更新: 2024年10月28日*
