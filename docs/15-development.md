# 15. 开发指南

> 开发环境、代码结构和贡献指南

## 🎯 概述

本指南介绍如何参与 AIPipe 的开发，包括环境搭建、代码结构和贡献流程。

## 🛠️ 开发环境

### 1. 环境要求

- **Go**: 1.19 或更高版本
- **Git**: 2.0 或更高版本
- **Make**: 3.0 或更高版本
- **Docker**: 20.0 或更高版本（可选）

### 2. 环境搭建

```bash
# 1. 克隆仓库
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 2. 安装依赖
go mod download

# 3. 安装开发工具
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest

# 4. 安装测试工具
go install github.com/stretchr/testify/assert@latest
```

### 3. 开发工具配置

```bash
# VS Code 配置
cat > .vscode/settings.json << EOF
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.testFlags": ["-v"],
    "go.buildTags": "dev"
}
EOF

# Air 热重载配置
cat > .air.toml << EOF
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF
```

## 📁 代码结构

### 1. 目录结构

```
aipipe/
├── cmd/                    # 命令行工具
│   ├── root.go            # 根命令
│   ├── analyze.go         # 分析命令
│   ├── monitor.go         # 监控命令
│   ├── config.go          # 配置命令
│   ├── rules.go           # 规则命令
│   ├── notify.go          # 通知命令
│   ├── cache.go           # 缓存命令
│   ├── ai.go              # AI 服务命令
│   └── status.go          # 状态命令
├── internal/              # 内部包
│   ├── ai/                # AI 服务管理
│   ├── cache/             # 缓存系统
│   ├── cmd/               # 命令实现
│   ├── config/            # 配置管理
│   ├── monitor/           # 文件监控
│   ├── notification/      # 通知系统
│   ├── rule/              # 规则引擎
│   └── utils/             # 工具函数
├── docs/                  # 文档
├── examples/              # 示例
├── prompts/               # 提示词模板
├── scripts/               # 脚本
├── tests/                 # 测试
├── main.go                # 主程序
├── go.mod                 # Go 模块
├── go.sum                 # 依赖校验
├── Makefile              # 构建脚本
└── README.md              # 项目说明
```

### 2. 包设计原则

- **单一职责**: 每个包只负责一个特定功能
- **接口隔离**: 使用小而专一的接口
- **依赖倒置**: 依赖抽象而不是具体实现
- **最小知识**: 减少包之间的相互依赖

### 3. 命名规范

```go
// 包名：小写，简短，有意义
package ai

// 接口名：以 -er 结尾
type Analyzer interface {
    Analyze(log string) (*Result, error)
}

// 结构体名：大驼峰
type AIService struct {
    endpoint string
    apiKey   string
}

// 方法名：大驼峰，动词开头
func (s *AIService) Analyze(log string) (*Result, error) {
    // 实现
}

// 常量：全大写，下划线分隔
const (
    DEFAULT_TIMEOUT = 30 * time.Second
    MAX_RETRIES     = 3
)

// 变量：小驼峰
var (
    defaultConfig = &Config{}
    logger        = log.New()
)
```

## 🧪 测试

### 1. 单元测试

```go
// 测试文件命名：*_test.go
package ai

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestAIService_Analyze(t *testing.T) {
    // 准备测试数据
    service := &AIService{
        endpoint: "https://api.openai.com/v1/chat/completions",
        apiKey:   "test-key",
    }
    
    // 执行测试
    result, err := service.Analyze("ERROR Database connection failed")
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result.Important)
}
```

### 2. 集成测试

```go
func TestIntegration_LogAnalysis(t *testing.T) {
    // 设置测试环境
    config := &Config{
        AIEndpoint: "https://api.openai.com/v1/chat/completions",
        AIAPIKey:   "test-key",
    }
    
    // 创建服务
    service := NewAIService(config)
    
    // 执行集成测试
    result, err := service.Analyze("ERROR Database connection failed")
    
    // 验证结果
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 3. 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/ai

# 运行测试并显示覆盖率
go test -cover ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔧 构建和部署

### 1. 本地构建

```bash
# 构建二进制文件
go build -o aipipe .

# 构建特定平台
GOOS=linux GOARCH=amd64 go build -o aipipe-linux-amd64 .

# 构建并压缩
go build -ldflags="-s -w" -o aipipe .
strip aipipe
```

### 2. 使用 Makefile

```bash
# 查看可用命令
make help

# 构建
make build

# 测试
make test

# 代码检查
make lint

# 格式化代码
make fmt

# 清理
make clean
```

### 3. Docker 构建

```bash
# 构建 Docker 镜像
docker build -t aipipe:latest .

# 运行 Docker 容器
docker run -it --rm aipipe:latest

# 构建多平台镜像
docker buildx build --platform linux/amd64,linux/arm64 -t aipipe:latest .
```

## 📝 代码规范

### 1. Go 代码规范

```go
// 1. 导入顺序
import (
    "context"
    "fmt"
    "time"
    
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

// 2. 函数注释
// Analyze analyzes a log line and returns the analysis result.
// It uses the configured AI service to determine the importance
// of the log line and provides a summary.
func (s *AIService) Analyze(ctx context.Context, logLine string) (*Result, error) {
    // 实现
}

// 3. 错误处理
func (s *AIService) Analyze(ctx context.Context, logLine string) (*Result, error) {
    if logLine == "" {
        return nil, errors.New("log line cannot be empty")
    }
    
    result, err := s.callAI(ctx, logLine)
    if err != nil {
        return nil, fmt.Errorf("failed to call AI service: %w", err)
    }
    
    return result, nil
}
```

### 2. 代码检查

```bash
# 使用 golangci-lint
golangci-lint run

# 使用 go vet
go vet ./...

# 使用 go fmt
go fmt ./...
```

### 3. 提交规范

```bash
# 提交信息格式
<type>(<scope>): <subject>

# 示例
feat(ai): add support for custom prompts
fix(monitor): resolve file permission issue
docs(readme): update installation guide
test(ai): add unit tests for AI service
```

## 🤝 贡献指南

### 1. 贡献流程

```bash
# 1. Fork 仓库
# 在 GitHub 上 Fork 仓库

# 2. 克隆 Fork 的仓库
git clone https://github.com/your-username/aipipe.git
cd aipipe

# 3. 添加上游仓库
git remote add upstream https://github.com/xurenlu/aipipe.git

# 4. 创建功能分支
git checkout -b feature/your-feature

# 5. 提交更改
git add .
git commit -m "feat: add your feature"

# 6. 推送分支
git push origin feature/your-feature

# 7. 创建 Pull Request
# 在 GitHub 上创建 Pull Request
```

### 2. 代码审查

- 确保代码符合项目规范
- 添加必要的测试
- 更新相关文档
- 通过所有 CI 检查

### 3. 问题报告

- 使用 GitHub Issues 报告问题
- 提供详细的复现步骤
- 包含环境信息和日志
- 使用标签分类问题

## 📚 文档

### 1. 代码文档

```go
// Package ai provides AI service management functionality.
// It supports multiple AI providers and load balancing.
package ai

// AIService represents an AI service client.
// It provides methods for analyzing log lines using AI.
type AIService struct {
    endpoint string
    apiKey   string
    client   *http.Client
}

// NewAIService creates a new AI service instance.
// It takes an endpoint URL and API key as parameters.
func NewAIService(endpoint, apiKey string) *AIService {
    return &AIService{
        endpoint: endpoint,
        apiKey:   apiKey,
        client:   &http.Client{Timeout: 30 * time.Second},
    }
}
```

### 2. API 文档

```go
// Analyze analyzes a log line and returns the analysis result.
//
// Parameters:
//   - ctx: context for cancellation and timeout
//   - logLine: the log line to analyze
//
// Returns:
//   - *Result: the analysis result containing importance and summary
//   - error: any error that occurred during analysis
//
// Example:
//   result, err := service.Analyze(ctx, "ERROR Database connection failed")
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Printf("Important: %v\n", result.Important)
func (s *AIService) Analyze(ctx context.Context, logLine string) (*Result, error) {
    // 实现
}
```

### 3. 更新文档

- 代码变更时同步更新文档
- 使用 Markdown 格式
- 提供清晰的示例
- 保持文档的准确性

## 🎉 总结

AIPipe 的开发指南提供了：

- **完整的环境搭建**: 开发工具和配置
- **清晰的代码结构**: 模块化设计和命名规范
- **全面的测试策略**: 单元测试和集成测试
- **规范的开发流程**: 代码规范和贡献指南
- **详细的文档要求**: 代码文档和 API 文档

---

*继续阅读: [16. API 参考](16-api-reference.md)*
