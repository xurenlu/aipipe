# AIPipe 项目重构说明

## 项目结构

项目已经重构为标准的 Go 项目结构，主要变化如下：

### 新的目录结构

```
aipipe/
├── main.go                    # 主入口文件
├── go.mod                     # Go 模块定义
├── go.sum                     # 依赖锁定文件
├── internal/                  # 内部包目录
│   ├── ai/                    # AI 服务相关
│   │   └── ai.go
│   ├── cache/                 # 缓存管理
│   │   └── cache.go
│   ├── config/                # 配置管理
│   │   └── config.go
│   ├── concurrency/           # 并发控制
│   │   ├── backpressure.go
│   │   ├── concurrency.go
│   │   ├── loadbalancer.go
│   │   └── priority_queue.go
│   ├── io/                    # I/O 优化
│   │   └── io.go
│   ├── memory/                # 内存管理
│   │   └── mem.go
│   ├── notification/          # 通知服务
│   │   ├── mail.go
│   │   └── webhook.go
│   ├── rule/                  # 规则引擎
│   │   └── rule.go
│   ├── utils/                 # 工具函数
│   │   ├── cmd.go
│   │   ├── fm.go
│   │   └── metric.go
│   └── worker/                # 工作池
│       └── job.go
├── aipipe.go                  # 原始主文件（保留作为参考）
├── aipipe_test.go             # 测试文件
└── README_REFACTOR.md         # 本文件
```

## 使用方法

### 基本运行

```bash
# 显示帮助信息
go run ./main.go --help

# 显示当前配置
go run ./main.go --config-show

# 显示配置模板
go run ./main.go --config-template

# 测试配置文件
go run ./main.go --config-test

# 验证配置文件
go run ./main.go --config-validate
```

### 配置管理

配置文件位置：`~/.config/aipipe.json`

```bash
# 启动配置向导（开发中）
go run ./main.go --config-init

# 显示配置模板并保存到文件
go run ./main.go --config-template > ~/.config/aipipe.json
```

## 重构状态

### ✅ 已完成

1. **项目结构重组**：将根目录下的 Go 文件移动到 `internal/` 目录
2. **包名修复**：所有 internal 包都使用正确的包名
3. **主入口文件**：创建了新的 `main.go` 作为程序入口
4. **配置管理**：实现了基本的配置加载和管理功能
5. **基本功能**：程序可以正常启动和显示帮助信息

### 🚧 开发中

1. **核心功能迁移**：AI 分析、日志处理等核心功能正在迁移中
2. **类型定义**：需要将原始文件中的类型定义移动到相应的包中
3. **函数导出**：需要将处理函数导出为包级别的公共函数
4. **依赖关系**：需要修复包之间的依赖关系

### 📋 待完成

1. **完整功能实现**：恢复所有原始功能
2. **测试覆盖**：添加单元测试
3. **文档完善**：更新使用文档
4. **性能优化**：优化包结构和依赖关系

## 开发说明

### 当前状态

- ✅ 项目可以正常编译和运行
- ✅ 配置管理功能完整
- ✅ 命令行参数解析正常
- ⚠️ 核心日志处理功能暂时不可用（显示开发中提示）

### 下一步计划

1. 逐步迁移核心功能到相应的包中
2. 修复类型定义和函数导出
3. 实现完整的日志处理流程
4. 添加错误处理和日志记录

## 贡献指南

1. 每个功能模块应该放在对应的 `internal/` 子目录中
2. 所有导出的函数和类型都应该有适当的文档注释
3. 保持包的职责单一，避免循环依赖
4. 添加适当的单元测试

## 提示词管理

AIPipe 支持自定义 AI 提示词，可以通过以下方式配置：

### 1. 使用提示词文件

```bash
# 在配置文件中指定提示词文件
{
  "prompt_file": "prompts/advanced.txt"
}

# 或者使用环境变量
AIPIPE_CONFIG=my-config.json go run ./main.go
```

### 2. 提示词文件格式

提示词文件支持以下占位符：
- `{format}`: 日志格式（如 java、php、nginx 等）
- `{log_line}`: 要分析的日志行

示例提示词文件 (`prompts/advanced.txt`)：
```
你是一个专业的日志分析专家，具有丰富的系统运维和故障排查经验。请分析以下 {format} 格式的日志行，判断其重要性。

## 分析规则

### 重要日志（should_filter: false）：
- 错误级别：ERROR、FATAL、CRITICAL、PANIC
- 异常信息：Exception、Error、Failed、Timeout
- 安全事件：Security、Auth、Permission、Access denied
- 性能问题：Slow、Timeout、Memory、CPU、Disk
- 系统故障：Crash、Restart、Shutdown、Connection refused
- 业务异常：Transaction failed、Payment error、User blocked

### 不重要日志（should_filter: true）：
- 调试信息：DEBUG、TRACE、VERBOSE
- 正常操作：INFO、Started、Stopped、Connected
- 健康检查：Health check、Ping、Pong、Heartbeat
- 常规统计：Statistics、Metrics、Count、Rate
- 配置加载：Config loaded、Settings applied

## 输出格式

请返回JSON格式：
{
  "should_filter": true/false,
  "summary": "简要摘要（不超过50字）",
  "reason": "详细判断原因（不超过200字）",
  "confidence": 0.0-1.0
}

请分析这条日志行：
{log_line}
```

### 3. 内置提示词

如果不指定 `prompt_file`，AIPipe 会使用内置的专业提示词。

### 4. 自定义提示词

也可以通过 `custom_prompt` 配置项直接指定提示词：

```json
{
  "custom_prompt": "你是一个日志分析助手，请判断以下日志是否重要..."
}
```

## 注意事项

- 原始的 `aipipe.go` 文件保留作为参考，但不再使用
- 所有新的开发都应该基于新的包结构
- 配置文件的格式保持不变，向后兼容
- 提示词文件支持 UTF-8 编码，建议使用 `.txt` 扩展名
