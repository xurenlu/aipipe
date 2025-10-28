# 01. 简介与安装

## 📖 项目简介

AIPipe 是一个基于 AI 的智能日志分析工具，能够实时监控和分析各种格式的日志文件，自动识别重要日志并提供智能通知。

### 🎯 核心特性

- **🤖 AI 智能分析**: 使用大语言模型自动判断日志重要性
- **📁 多格式支持**: 支持 20+ 种日志格式（Java、Nginx、Apache、Docker 等）
- **⚡ 实时监控**: 支持标准输入和文件监控两种模式
- **🔔 智能通知**: 多渠道通知系统（邮件、Webhook、系统通知等）
- **🎛️ 规则引擎**: 灵活的正则表达式过滤和自定义规则
- **💾 缓存优化**: 智能缓存减少 API 调用，提高性能
- **🔄 负载均衡**: 支持多个 AI 服务，自动故障转移
- **📊 监控面板**: 实时系统状态和统计信息

### 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   日志源        │    │   AIPipe 核心   │    │   通知系统      │
│                 │    │                 │    │                 │
│ • 标准输入      │───▶│ • AI 分析引擎   │───▶│ • 邮件通知      │
│ • 文件监控      │    │ • 规则引擎      │    │ • Webhook       │
│ • 多源支持      │    │ • 缓存系统      │    │ • 系统通知      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   AI 服务       │
                       │                 │
                       │ • OpenAI        │
                       │ • Azure OpenAI  │
                       │ • 自定义 API    │
                       └─────────────────┘
```

## 🚀 快速安装

### 系统要求

- **操作系统**: Linux、macOS、Windows
- **Go 版本**: 1.19 或更高版本
- **内存**: 最少 512MB，推荐 1GB+
- **网络**: 需要访问 AI API 服务

### 安装方式

#### 方式一：从源码编译（推荐）

```bash
# 克隆仓库
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 编译
go build -o aipipe .

# 安装到系统路径（可选）
sudo cp aipipe /usr/local/bin/
```

#### 方式二：下载预编译二进制

```bash
# 下载最新版本
wget https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-linux-amd64
chmod +x aipipe-linux-amd64
sudo mv aipipe-linux-amd64 /usr/local/bin/aipipe
```

#### 方式三：使用 Docker

```bash
# 拉取镜像
docker pull xurenlu/aipipe:latest

# 运行容器
docker run -it --rm xurenlu/aipipe:latest
```

### 验证安装

```bash
# 检查版本
aipipe --version

# 查看帮助
aipipe --help

# 查看子命令
aipipe --help
```

## ⚙️ 基本配置

### 1. 初始化配置

```bash
# 生成默认配置文件
aipipe config init

# 查看当前配置
aipipe config show
```

### 2. 配置 AI 服务

编辑配置文件 `~/.aipipe/config.json`：

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "your-api-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60
}
```

### 3. 测试配置

```bash
# 测试 AI 服务连接
aipipe ai test

# 测试通知系统
aipipe notify test
```

## 🎯 快速体验

### 1. 分析标准输入

```bash
# 分析单行日志
echo "2024-01-01 10:00:00 ERROR Database connection failed" | aipipe analyze

# 分析文件内容
cat app.log | aipipe analyze --format java
```

### 2. 监控文件

```bash
# 监控单个文件
aipipe monitor --file /var/log/app.log --format java

# 监控多个文件（需要先配置）
aipipe dashboard add
aipipe monitor
```

### 3. 查看系统状态

```bash
# 查看系统状态
aipipe dashboard show

# 查看监控文件列表
aipipe dashboard list
```

## 📋 支持的日志格式

AIPipe 支持以下日志格式：

| 格式 | 描述 | 示例 |
|------|------|------|
| `java` | Java 应用日志 | `2024-01-01 10:00:00 ERROR com.example.Service: Database error` |
| `nginx` | Nginx 访问日志 | `192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /api/users HTTP/1.1" 200 1234` |
| `apache` | Apache 访问日志 | `192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /index.html HTTP/1.1" 200 1234` |
| `docker` | Docker 容器日志 | `2024-01-01T10:00:00.000Z container_name: ERROR: Service unavailable` |
| `syslog` | 系统日志 | `Jan 1 10:00:00 hostname systemd[1]: Started Network Manager` |
| `json` | JSON 格式日志 | `{"timestamp":"2024-01-01T10:00:00Z","level":"ERROR","message":"Database error"}` |

更多格式支持请参考 [支持格式](17-supported-formats.md)。

## 🔧 命令行工具

AIPipe 提供了丰富的命令行工具：

```bash
# 核心功能
aipipe analyze          # 分析日志
aipipe monitor          # 监控文件
aipipe dashboard        # 系统面板

# 配置管理
aipipe config init      # 初始化配置
aipipe config show      # 显示配置
aipipe config validate  # 验证配置

# 规则管理
aipipe rules add        # 添加规则
aipipe rules list       # 列出规则
aipipe rules test       # 测试规则

# 通知管理
aipipe notify test      # 测试通知
aipipe notify send      # 发送通知

# AI 服务管理
aipipe ai list          # 列出 AI 服务
aipipe ai add           # 添加 AI 服务
aipipe ai test          # 测试 AI 服务

# 缓存管理
aipipe cache stats      # 缓存统计
aipipe cache clear      # 清空缓存
```

## 🎉 下一步

现在你已经完成了 AIPipe 的安装和基本配置，可以：

1. 阅读 [快速开始](02-quick-start.md) 进行首次使用
2. 查看 [日志分析](03-log-analysis.md) 了解分析功能
3. 参考 [文件监控](04-file-monitoring.md) 设置监控
4. 配置 [通知系统](05-notifications.md) 接收告警

---

*继续阅读: [02. 快速开始](02-quick-start.md)*
