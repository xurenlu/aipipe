# AIPipe - 智能日志监控工具 🚀

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://www.apple.com/macos/)

> 使用 AI 自动分析日志内容，智能过滤噪音，只关注真正重要的问题

AIPipe 是一个智能日志过滤和监控工具，使用可配置的 AI 服务自动分析日志内容，过滤不重要的日志，并对重要事件发送 macOS 通知和声音提醒。

## ✨ 核心特性

- 🤖 **AI 智能分析** - 使用可配置的 AI 服务自动判断日志重要性
- 📦 **批处理模式** - 智能累积多行日志批量分析，节省 70-90% Token
- ⚡ **本地预过滤** - DEBUG/INFO 级别日志本地处理，不调用 API
- 🔔 **系统通知** - 重要日志发送 macOS 原生通知 + 声音提醒
- 📁 **文件监控** - 类似 `tail -f`，支持断点续传和日志轮转
- 🎯 **上下文显示** - 重要日志自动显示前后上下文，方便排查问题
- 🛡️ **保守策略** - AI 无法确定时默认过滤，避免误报
- 🌍 **多格式支持** - Java、PHP、Nginx、Ruby、Python、FastAPI
- 🔍 **多行日志合并** - 自动合并异常堆栈等多行日志
- ⚙️ **配置化** - 从 `~/.config/aipipe.json` 读取 AI 服务器配置
- 🎨 **自定义提示词** - 支持用户自定义补充 prompt

## 🚀 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/your-username/aipipe.git
cd aipipe

# 编译
go build -o aipipe aipipe.go

# 或直接运行
go run aipipe.go -f /var/log/app.log --format java
```

### 配置

首次运行会自动创建配置文件 `~/.config/aipipe.json`：

```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": "请特别注意以下情况：\n1. 数据库连接问题\n2. 内存泄漏警告\n3. 安全相关日志\n4. 性能瓶颈指标\n\n请根据这些特殊要求调整判断标准。"
}
```

### 基本使用

```bash
# 监控日志文件（推荐）
./aipipe -f /var/log/app.log --format java

# 或通过管道
tail -f /var/log/app.log | ./aipipe --format java

# 查看帮助
./aipipe --help
```

## 📖 使用示例

### 监控 Java 应用日志

```bash
./aipipe -f /var/log/tomcat/catalina.out --format java
```

**输出：**
```
🚀 AIPipe 启动 - 监控 java 格式日志
💡 只显示重要日志（过滤的日志不显示）
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 批次摘要: 发现数据库连接错误 (重要日志: 2 条)

   │ 2025-10-13 INFO Processing user request
   │ 2025-10-13 INFO Connecting to database
⚠️  [重要] 2025-10-13 ERROR Database connection timeout
⚠️  [重要] java.sql.SQLException: Connection refused
   │    at com.mysql.jdbc.Connection.connect(...)
   │    at com.example.dao.UserDao.getUser(...)
   │ 2025-10-13 INFO Falling back to cache

⏳ 等待新日志...
```

**同时：**
- 🔔 收到 macOS 通知："发现数据库连接错误"
- 🔊 播放提示音（Glass.aiff）

### 监控 Python/FastAPI 日志

```bash
./aipipe -f /var/log/fastapi.log --format fastapi
```

### 自定义配置

```bash
# 大批次，适合高频日志
./aipipe -f app.log --format java --batch-size 20 --batch-wait 5s

# 更多上下文，适合排查问题
./aipipe -f app.log --format java --context 5

# 显示所有日志（包括过滤的）
./aipipe -f app.log --format java --show-not-important

# 调试模式
./aipipe -f app.log --format java --debug
```

## 🎯 主要功能

### 1. 智能批处理

**问题：** 传统方式逐行分析，API 调用频繁，通知轰炸

**解决：** 批处理模式
- 累积 10 行或等待 3 秒后批量分析
- 一次 API 调用分析多行
- 减少 70-90% Token 消耗
- 一批只发 1 次通知

**性能对比：**
| 指标 | 逐行模式 | 批处理模式 | 提升 |
|------|---------|-----------|------|
| API 调用 | 100 次 | 10 次 | ↓ 90% |
| Token 消耗 | 64,500 | 10,500 | ↓ 83% |
| 通知次数 | 15 次 | 1-2 次 | ↓ 87% |

### 2. 本地预过滤

**问题：** DEBUG/INFO 日志也调用 AI，浪费资源

**解决：** 本地智能识别
- 自动识别 DEBUG、INFO、TRACE 等低级别日志
- 直接本地过滤，不调用 API
- 处理速度提升 10-30 倍（< 0.1秒）
- 但如果包含 ERROR/EXCEPTION 关键词，仍会调用 AI

### 3. 上下文显示

**问题：** 只显示错误行，看不到完整场景

**解决：** 自动显示上下文
- 重要日志前后各显示 3 行（可配置）
- 异常堆栈完整显示
- 用 `│` 标记上下文行
- 方便排查问题

**示例：**
```
   │ INFO Calling service           ← 上下文
⚠️  [重要] ERROR Failed            ← 重要日志
⚠️  [重要] java.sql.SQLException   ← 重要日志（异常）
   │    at com.example...           ← 上下文（堆栈）
   │ INFO Retry attempt              ← 上下文
```

### 4. 多行日志合并

**问题：** Java 堆栈跟踪是多行的，被拆分分析

**解决：** 自动合并
- 识别堆栈跟踪、异常信息等多行日志
- 自动合并为完整日志条目
- 作为一个整体交给 AI 分析
- 支持 Java、Python、Ruby 等格式

### 5. 配置化支持

**问题：** 硬编码的 AI 服务端点，无法灵活配置

**解决：** 配置文件支持
- 从 `~/.config/aipipe.json` 读取配置
- 支持自定义 AI 服务器端点
- 支持自定义 Token 和模型
- 支持用户自定义补充 prompt

## 📋 参数说明

```bash
./aipipe --help
```

### 必选参数

- `--format` - 日志格式：java, php, nginx, ruby, python, fastapi

### 常用参数

- `-f <文件>` - 监控日志文件（类似 tail -f）
- `--context N` - 显示重要日志的上下文行数（默认 3）
- `--show-not-important` - 显示被过滤的日志（默认不显示）

### 批处理参数

- `--batch-size N` - 批处理最大行数（默认 10）
- `--batch-wait 时间` - 批处理等待时间（默认 3s）
- `--no-batch` - 禁用批处理，逐行分析

### 调试参数

- `--verbose` - 显示详细输出
- `--debug` - 调试模式，打印完整 HTTP 请求响应

## 🔧 配置

### 配置文件格式

编辑 `~/.config/aipipe.json`：

```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": "请特别注意以下情况：\n1. 数据库连接问题\n2. 内存泄漏警告\n3. 安全相关日志\n4. 性能瓶颈指标\n\n请根据这些特殊要求调整判断标准。"
}
```

### 配置项说明

- `ai_endpoint`: AI 服务器的 API 端点 URL
- `token`: API 认证 Token
- `model`: 使用的 AI 模型名称
- `custom_prompt`: 用户自定义的补充提示词，会添加到系统提示词中

### 批处理配置

```go
const (
    BATCH_MAX_SIZE  = 10              // 批处理最大行数
    BATCH_WAIT_TIME = 3 * time.Second // 批处理等待时间
)
```

## 📊 判断标准

AIPipe 使用包含 60+ 个真实场景示例的 AI 提示词：

### 会过滤的日志（不显示）
- ✅ DEBUG、INFO、TRACE 级别
- ✅ 健康检查、心跳
- ✅ 应用启动、配置加载
- ✅ 正常的业务操作
- ✅ 静态资源请求

### 需要关注的日志（显示 + 通知）
- ⚠️ ERROR、FATAL 级别
- ⚠️ 异常（Exception、Error）
- ⚠️ WARN 级别（性能、资源）
- ⚠️ 数据库问题
- ⚠️ 认证失败
- ⚠️ 安全问题
- ⚠️ 服务降级、熔断

## 🎬 使用场景

### 生产环境监控

```bash
./aipipe -f /var/log/production.log --format java --batch-size 20
```

**效果：**
- 自动过滤 80% 的噪音日志
- 重要错误立即通知
- 完整的上下文帮助排查
- 节省 API 费用

### 开发调试

```bash
./aipipe -f dev.log --format java --context 5 --verbose
```

**效果：**
- 更多上下文（5 行）
- 详细的分析原因
- 快速定位问题

### 历史日志分析

```bash
cat old.log | ./aipipe --format java --batch-size 50
```

**效果：**
- 快速筛选重要事件
- 大批次高效处理
- 生成问题清单

## 📁 项目结构

```
aipipe-project/
├── aipipe.go                    # 主程序源代码
├── aipipe                      # 编译后的可执行文件
├── README.md                   # 项目说明（本文件）
├── LICENSE                     # MIT 许可证
├── .gitignore                 # Git 忽略文件
├── go.mod                     # Go 模块文件
├── aipipe.json.example        # 配置文件示例
├── docs/                      # 文档目录
│   ├── README_aipipe.md              # 完整使用文档
│   ├── 批处理优化说明.md              # 批处理详解
│   ├── 本地预过滤优化.md              # 本地过滤详解
│   ├── 保守过滤策略.md                # 保守策略说明
│   ├── NOTIFICATION_SETUP.md         # 通知设置指南
│   ├── NOTIFICATION_SOUND_GUIDE.md   # 声音播放指南
│   ├── PROMPT_EXAMPLES.md            # 提示词示例
│   └── ...                           # 其他文档
├── examples/                  # 示例目录
│   ├── test-logs-sample.txt         # 基础示例日志
│   ├── test-logs-comprehensive.txt  # 全面测试日志
│   └── aipipe-example.sh            # 交互式示例
└── tests/                     # 测试目录
    ├── test-batch-processing.sh     # 批处理测试
    ├── test-context.sh              # 上下文显示测试
    ├── test-notification-quick.sh   # 通知设置向导
    └── ...                          # 其他测试
```

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **AI**: 可配置的 AI 服务（支持 OpenAI、Azure OpenAI 等）
- **文件监控**: fsnotify
- **系统通知**: macOS osascript
- **音频播放**: afplay

## 🎯 性能特性

- **内存占用**: < 50MB（流式处理）
- **处理速度**: < 0.1秒（本地过滤）/ 1-3秒（AI 分析）
- **Token 节省**: 70-90%（批处理模式）
- **API 调用减少**: 60-90%（本地预过滤 + 批处理）

## 📝 示例

### 示例 1: 监控生产日志

```bash
# 大批次，节省费用
./aipipe -f /var/log/production.log --format java --batch-size 20 --batch-wait 5s
```

### 示例 2: 排查问题

```bash
# 更多上下文，显示详细原因
./aipipe -f /var/log/app.log --format java --context 10 --verbose
```

### 示例 3: 分析历史日志

```bash
# 快速过滤重要事件
cat /var/log/old/*.log | ./aipipe --format java --batch-size 50
```

## 🧪 运行测试

```bash
# 批处理功能测试
./tests/test-batch-processing.sh

# 上下文显示测试
./tests/test-context.sh

# 通知设置向导
./tests/test-notification-quick.sh

# 完整功能演示
./examples/aipipe-example.sh
```

## 📚 文档

- [完整使用文档](docs/README_aipipe.md)
- [批处理优化说明](docs/批处理优化说明.md)
- [本地预过滤优化](docs/本地预过滤优化.md)
- [通知设置指南](docs/NOTIFICATION_SETUP.md)
- [提示词示例](docs/PROMPT_EXAMPLES.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 👤 作者

**rocky** <m@some.im>

## 🙏 致谢

- AI 服务提供商 - 提供强大的 AI 能力
- fsnotify - 文件监控库
- Go 社区 - 优秀的工具生态

## 🔗 相关链接

- [问题反馈](https://github.com/your-username/aipipe/issues)
- [更新日志](CHANGELOG.md)
- [开发文档](docs/)

---

**Star** ⭐ 如果这个项目对你有帮助！