# 16. API 参考

> 完整 API 文档和接口说明

## 🎯 概述

本文档提供 AIPipe 的完整 API 参考，包括命令行接口、配置接口和编程接口。

## 🖥️ 命令行接口

### 1. 主命令

```bash
aipipe [flags] [command]
```

**全局标志**:
- `--file string`: 要监控的日志文件路径
- `--format string`: 日志格式 (默认: "java")
- `--show-not-important`: 显示被过滤的日志
- `--verbose`: 显示详细输出
- `--help`: 显示帮助信息

### 2. 子命令

#### analyze - 分析日志

```bash
aipipe analyze [flags]
```

**功能**: 分析标准输入的日志内容

**标志**:
- `--format string`: 日志格式 (默认: "java")
- `--prompt-file string`: 自定义提示词文件
- `--show-not-important`: 显示被过滤的日志
- `--verbose`: 显示详细输出

**示例**:
```bash
# 分析单行日志
echo "ERROR Database connection failed" | aipipe analyze --format java

# 分析文件内容
cat app.log | aipipe analyze --format java

# 使用自定义提示词
echo "ERROR Database connection failed" | aipipe analyze --format java --prompt-file prompts/custom.txt
```

#### monitor - 监控文件

```bash
aipipe monitor [flags]
```

**功能**: 监控日志文件，实时分析新增内容

**标志**:
- `--file string`: 要监控的日志文件路径
- `--format string`: 日志格式 (默认: "java")
- `--from-beginning`: 从头开始读取文件
- `--show-not-important`: 显示被过滤的日志
- `--verbose`: 显示详细输出

**示例**:
```bash
# 监控单个文件
aipipe monitor --file /var/log/app.log --format java

# 监控所有配置的文件
aipipe monitor

# 从头开始监控
aipipe monitor --file /var/log/app.log --from-beginning
```

#### config - 配置管理

```bash
aipipe config [command]
```

**子命令**:
- `init`: 初始化配置文件
- `show`: 显示当前配置
- `set`: 设置配置值
- `validate`: 验证配置文件
- `template`: 生成配置模板

**示例**:
```bash
# 初始化配置
aipipe config init

# 显示配置
aipipe config show

# 设置配置
# 编辑配置文件 ~/.aipipe/config.json
# 修改 "ai_model" 字段的值为 "gpt-4"

# 验证配置
aipipe config validate
```

#### rules - 规则管理

```bash
aipipe rules [command]
```

**子命令**:
- `add`: 添加规则
- `list`: 列出规则
- `remove`: 删除规则
- `enable`: 启用规则
- `disable`: 禁用规则
- `test`: 测试规则
- `stats`: 显示规则统计

**示例**:
```bash
# 添加规则
aipipe rules add --pattern "DEBUG" --action "ignore"

# 列出规则
aipipe rules list

# 测试规则
aipipe rules test --pattern "ERROR Database connection failed"
```

#### notify - 通知管理

```bash
aipipe notify [command]
```

**子命令**:
- `test`: 测试通知
- `status`: 显示通知状态
- `send`: 发送通知
- `enable`: 启用通知
- `disable`: 禁用通知

**示例**:
```bash
# 测试所有通知
aipipe notify test

# 测试邮件通知
aipipe notify test --email

# 发送测试通知
aipipe notify send --message "测试通知"
```

#### cache - 缓存管理

```bash
aipipe cache [command]
```

**子命令**:
- `stats`: 显示缓存统计
- `clear`: 清空缓存
- `status`: 显示缓存状态
- `warmup`: 预热缓存

**示例**:
```bash
# 显示缓存统计
aipipe cache stats

# 清空缓存
aipipe cache clear

# 预热缓存
aipipe cache warmup
```

#### ai - AI 服务管理

```bash
aipipe ai [command]
```

**子命令**:
- `list`: 列出 AI 服务
- `add`: 添加 AI 服务
- `remove`: 删除 AI 服务
- `enable`: 启用 AI 服务
- `disable`: 禁用 AI 服务
- `test`: 测试 AI 服务
- `stats`: 显示 AI 服务统计

**示例**:
```bash
# 列出 AI 服务
aipipe ai list

# 添加 AI 服务
aipipe ai add --name "openai" --endpoint "https://api.openai.com/v1/chat/completions" --api-key "sk-key"

# 测试 AI 服务
aipipe ai test
```

#### dashboard - 系统面板

```bash
aipipe dashboard [command]
```

**子命令**:
- `show`: 显示系统状态
- `add`: 添加监控文件
- `list`: 列出监控文件
- `remove`: 删除监控文件
- `enable`: 启用监控文件
- `disable`: 禁用监控文件

**示例**:
```bash
# 显示系统状态
aipipe dashboard show

# 添加监控文件
aipipe dashboard add

# 列出监控文件
aipipe dashboard list
```

## ⚙️ 配置接口

### 1. 配置文件格式

**位置**: `~/.aipipe/config.json`

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-api-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60,
  "local_filter": true,
  "show_not_important": false,
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "to": "admin@example.com"
    },
    "system": {
      "enabled": true,
      "sound": true
    }
  },
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  }
}
```

### 2. 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `OPENAI_API_KEY` | OpenAI API 密钥 | - |
| `AIPIPE_AI_ENDPOINT` | AI 服务端点 | - |
| `AIPIPE_AI_MODEL` | AI 模型 | gpt-3.5-turbo |
| `AIPIPE_CONFIG_FILE` | 配置文件路径 | ~/.aipipe/config.json |
| `AIPIPE_LOG_LEVEL` | 日志级别 | info |
| `AIPIPE_DEBUG` | 调试模式 | false |

### 3. 配置验证

```bash
# 验证配置文件
aipipe config validate

# 验证特定配置
aipipe config validate --key "ai_endpoint"

# 显示验证结果
aipipe config validate --verbose
```

## 🔌 编程接口

### 1. Go 接口

#### 配置接口

```go
// Config 表示 AIPipe 配置
type Config struct {
    AIEndpoint        string            `json:"ai_endpoint"`
    AIAPIKey          string            `json:"ai_api_key"`
    AIModel           string            `json:"ai_model"`
    MaxRetries        int               `json:"max_retries"`
    Timeout           int               `json:"timeout"`
    RateLimit         int               `json:"rate_limit"`
    LocalFilter       bool              `json:"local_filter"`
    ShowNotImportant  bool              `json:"show_not_important"`
    Notifications     NotificationConfig `json:"notifications"`
    Cache             CacheConfig       `json:"cache"`
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error)

// SaveConfig 保存配置文件
func SaveConfig(config *Config) error

// ValidateConfig 验证配置文件
func ValidateConfig(config *Config) error
```

#### AI 服务接口

```go
// AIService 表示 AI 服务
type AIService struct {
    endpoint string
    apiKey   string
    client   *http.Client
}

// NewAIService 创建新的 AI 服务
func NewAIService(endpoint, apiKey string) *AIService

// Analyze 分析日志
func (s *AIService) Analyze(ctx context.Context, logLine string) (*Result, error)

// Result 表示分析结果
type Result struct {
    Important  bool     `json:"important"`
    Summary    string   `json:"summary"`
    Keywords   []string `json:"keywords"`
    Confidence float64  `json:"confidence"`
}
```

#### 文件监控接口

```go
// FileMonitor 表示文件监控器
type FileMonitor struct {
    files map[string]*MonitoredFile
    watcher *fsnotify.Watcher
}

// NewFileMonitor 创建新的文件监控器
func NewFileMonitor() (*FileMonitor, error)

// AddFile 添加监控文件
func (m *FileMonitor) AddFile(path string, callback func(string, string)) error

// RemoveFile 删除监控文件
func (m *FileMonitor) RemoveFile(path string) error

// Start 启动监控
func (m *FileMonitor) Start() error

// Stop 停止监控
func (m *FileMonitor) Stop() error
```

#### 缓存接口

```go
// Cache 表示缓存接口
type Cache interface {
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Clear() error
    Stats() CacheStats
}

// CacheStats 表示缓存统计
type CacheStats struct {
    Hits       int64   `json:"hits"`
    Misses     int64   `json:"misses"`
    HitRate    float64 `json:"hit_rate"`
    Size       int64   `json:"size"`
    MemoryUsage int64  `json:"memory_usage"`
}
```

### 2. HTTP 接口

#### 分析接口

```http
POST /api/v1/analyze
Content-Type: application/json

{
  "log_line": "ERROR Database connection failed",
  "format": "java"
}
```

**响应**:
```json
{
  "important": true,
  "summary": "数据库连接失败，需要立即处理",
  "keywords": ["database", "connection", "failed"],
  "confidence": 0.95
}
```

#### 监控接口

```http
POST /api/v1/monitor
Content-Type: application/json

{
  "path": "/var/log/app.log",
  "format": "java",
  "enabled": true
}
```

**响应**:
```json
{
  "success": true,
  "message": "监控文件添加成功"
}
```

#### 状态接口

```http
GET /api/v1/status
```

**响应**:
```json
{
  "status": "running",
  "uptime": "1h30m",
  "files_monitored": 3,
  "cache_hit_rate": 0.85,
  "ai_services": 2
}
```

## 📊 返回码

### 1. 成功码

| 码 | 描述 |
|----|------|
| 0 | 成功 |

### 2. 错误码

| 码 | 描述 | 解决方案 |
|----|------|----------|
| 1 | 通用错误 | 查看详细错误信息 |
| 2 | 配置错误 | 检查配置文件 |
| 3 | 网络错误 | 检查网络连接 |
| 4 | 权限错误 | 检查文件权限 |
| 5 | 资源错误 | 检查系统资源 |

## 🔍 错误处理

### 1. 错误格式

```json
{
  "error": {
    "code": 1001,
    "message": "配置文件错误",
    "details": "无法解析 JSON 格式",
    "timestamp": "2024-01-01T10:00:00Z"
  }
}
```

### 2. 错误处理示例

```go
result, err := service.Analyze(ctx, logLine)
if err != nil {
    switch {
    case errors.Is(err, ErrConfigInvalid):
        log.Error("配置错误:", err)
    case errors.Is(err, ErrNetworkError):
        log.Error("网络错误:", err)
    case errors.Is(err, ErrPermissionDenied):
        log.Error("权限错误:", err)
    default:
        log.Error("未知错误:", err)
    }
    return
}
```

## 📋 最佳实践

### 1. 使用建议

- 使用适当的超时设置
- 启用本地过滤减少 API 调用
- 配置合理的缓存策略
- 监控资源使用情况

### 2. 性能优化

- 使用批处理处理大量日志
- 启用缓存提高响应速度
- 调整并发参数
- 定期清理缓存

### 3. 安全考虑

- 保护 API 密钥
- 使用 HTTPS 通信
- 设置适当的权限
- 定期更新依赖

## 🎉 总结

AIPipe 的 API 参考提供了：

- **完整的命令行接口**: 所有命令和选项
- **详细的配置接口**: 配置格式和环境变量
- **丰富的编程接口**: Go 接口和 HTTP 接口
- **清晰的错误处理**: 错误码和处理方法
- **实用的最佳实践**: 使用建议和性能优化

---

*继续阅读: [17. 支持格式](17-supported-formats.md)*
