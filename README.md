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
- 🔔 **多通道通知** - 支持邮件、钉钉、企业微信、飞书、Slack 等多种通知方式
- 📁 **文件监控** - 类似 `tail -f`，支持断点续传和日志轮转
- 🎯 **上下文显示** - 重要日志自动显示前后上下文，方便排查问题
- 🛡️ **保守策略** - AI 无法确定时默认过滤，避免误报
- 🌍 **多格式支持** - Java、PHP、Nginx、Ruby、Python、FastAPI、journald、syslog
- 🔍 **多行日志合并** - 自动合并异常堆栈等多行日志
- ⚙️ **配置化** - 从 `~/.config/aipipe.json` 读取 AI 服务器配置
- 🎨 **自定义提示词** - 支持用户自定义补充 prompt
- 🌐 **智能识别** - 自动识别 webhook 类型，支持自定义 webhook
- 📰 **系统日志监控** - 直接支持 journalctl，无需手动管道操作
- 🎯 **精确过滤** - 支持服务、级别、时间范围等多维度过滤
- 🔍 **自动检测** - 自动检测多种格式的配置文件，零配置启动
- 🚀 **智能启动** - 自动识别单源/多源监控模式

## 🚀 快速开始

### 零配置启动（推荐）

AIPipe 支持零配置启动，自动检测配置文件：

```bash
# 1. 下载并运行
curl -fsSL https://raw.githubusercontent.com/xurenlu/aipipe/main/install.sh | bash

# 2. 创建配置文件（可选）
mkdir -p ~/.config
cp aipipe.yaml ~/.config/
cp aipipe-sources.yaml ~/.config/

# 3. 直接启动（自动检测配置）
./aipipe
```

### 安装

#### 一键安装（推荐）

```bash
# 使用一键安装脚本
curl -fsSL https://raw.githubusercontent.com/xurenlu/aipipe/main/install.sh | bash
```

#### 手动安装

```bash
# 克隆仓库
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 编译
go build -o aipipe aipipe.go

# 或直接运行
go run aipipe.go -f /var/log/app.log --format java
```

#### Linux 系统服务安装

```bash
# 使用 systemd 安装脚本
sudo ./install-systemd.sh
```

### 配置

首次运行会自动创建配置文件 `~/.config/aipipe.json`：

```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": "请特别注意以下情况：\n1. 数据库连接问题\n2. 内存泄漏警告\n3. 安全相关日志\n4. 性能瓶颈指标\n\n请根据这些特殊要求调整判断标准。",
  "notifiers": {
    "email": {
      "enabled": false,
      "provider": "smtp",
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "from_email": "your-email@gmail.com",
      "to_emails": ["admin@company.com"]
    },
    "dingtalk": {
      "enabled": false,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
    },
    "wechat": {
      "enabled": false,
      "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
    },
    "feishu": {
      "enabled": false,
      "url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN"
    },
    "slack": {
      "enabled": false,
      "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
    }
  }
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

### 多源监控

AIPipe 支持同时监控多个日志源，包括文件、journalctl 和标准输入：

```bash
# 使用多源配置文件
./aipipe --multi-source multi-source-config.json
```

#### 多源配置文件示例

```json
{
  "sources": [
    {
      "name": "Java应用日志",
      "type": "file",
      "path": "/var/log/java-app.log",
      "format": "java",
      "enabled": true,
      "priority": 1,
      "description": "监控Java应用程序日志"
    },
    {
      "name": "PHP应用日志",
      "type": "file",
      "path": "/var/log/php-app.log",
      "format": "php",
      "enabled": true,
      "priority": 2,
      "description": "监控PHP应用程序日志"
    },
    {
      "name": "Nginx错误日志",
      "type": "file",
      "path": "/var/log/nginx/error.log",
      "format": "nginx",
      "enabled": true,
      "priority": 3,
      "description": "监控Nginx错误日志"
    },
    {
      "name": "系统服务监控",
      "type": "journalctl",
      "format": "journald",
      "enabled": true,
      "priority": 4,
      "description": "监控系统服务日志",
      "journal": {
        "services": ["nginx", "docker", "postgresql"],
        "priority": "err",
        "since": "",
        "until": "",
        "user": "",
        "boot": false,
        "kernel": false
      }
    }
  ]
}
```

#### 多源监控特性

- ✅ **并发监控** - 同时监控多个日志源
- ✅ **优先级控制** - 支持源优先级排序
- ✅ **独立格式** - 每个源可以使用不同的日志格式
- ✅ **灵活配置** - 支持启用/禁用特定源
- ✅ **统一处理** - 所有源共享AI分析和通知配置
- ✅ **多格式支持** - 支持JSON、YAML、TOML配置文件格式

#### 配置文件格式示例

**JSON格式 (默认):**
```bash
./aipipe --multi-source config.json
```

**YAML格式:**
```bash
./aipipe --multi-source config.yaml
```

**TOML格式:**
```bash
./aipipe --multi-source config.toml
```

**自动检测格式:**
```bash
# AIPipe会自动检测文件格式
./aipipe --multi-source config  # 无扩展名，自动检测
```

### 零配置启动示例

AIPipe 支持零配置启动，自动检测配置文件：

```bash
# 1. 创建配置文件
mkdir -p ~/.config

# 2. 创建主配置文件
cat > ~/.config/aipipe.yaml << EOF
ai_endpoint: "https://api.openai.com/v1/chat/completions"
token: "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
model: "gpt-4"
custom_prompt: "请特别注意数据库连接、内存泄漏、安全相关日志"

notifiers:
  email:
    enabled: true
    provider: "smtp"
    host: "smtp.gmail.com"
    port: 587
    username: "alerts@company.com"
    password: "your-app-password"
    from_email: "alerts@company.com"
    to_emails: ["admin@company.com"]
EOF

# 3. 创建多源配置文件
cat > ~/.config/aipipe-sources.yaml << EOF
sources:
  - name: "Java应用日志"
    type: "file"
    path: "/var/log/java-app.log"
    format: "java"
    enabled: true
    priority: 1
    description: "监控Java应用程序日志"
  
  - name: "系统服务监控"
    type: "journalctl"
    format: "journald"
    enabled: true
    priority: 2
    description: "监控系统服务日志"
    journal:
      services: ["nginx", "docker", "postgresql"]
      priority: "err"
EOF

# 4. 直接启动（自动检测配置）
./aipipe

# 输出示例：
# 🔍 找到默认配置文件: /home/user/.config/aipipe.yaml
# 🔍 检测到主配置文件格式: yaml
# 🔍 自动检测到多源配置文件: /home/user/.config/aipipe-sources.yaml
# 🔍 检测到配置文件格式: yaml
# 🚀 AIPipe 多源监控启动 - 监控 2 个源
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 📡 源: Java应用日志 (file) - 监控Java应用程序日志
# 📡 源: 系统服务监控 (journalctl) - 监控系统服务日志
# ✅ 启用 2 个监控源
```

### 监控系统日志 (journalctl)

AIPipe 支持直接监控 Linux 系统日志，无需手动使用 `journalctl -f`：

```bash
# 监控所有系统日志
./aipipe --format journald

# 监控特定服务
./aipipe --format journald --journal-services nginx,docker,postgresql

# 只监控错误级别及以上
./aipipe --format journald --journal-priority err

# 监控特定服务 + 错误级别
./aipipe --format journald --journal-services nginx --journal-priority err

# 监控最近1小时的错误日志
./aipipe --format journald --journal-since "1 hour ago" --journal-priority err

# 只监控内核消息
./aipipe --format journald --journal-kernel

# 只监控当前启动的日志
./aipipe --format journald --journal-boot
```

#### journalctl 配置参数

| 参数 | 功能 | 示例 |
|------|------|------|
| `--journal-services` | 监控特定服务 | `nginx,docker,postgresql` |
| `--journal-priority` | 日志级别过滤 | `err`, `warning`, `crit` |
| `--journal-since` | 开始时间 | `"1 hour ago"`, `"2023-10-17 10:00:00"` |
| `--journal-until` | 结束时间 | `"now"`, `"2023-10-17 18:00:00"` |
| `--journal-user` | 用户过滤 | `1000`, `root` |
| `--journal-boot` | 当前启动 | 只监控当前启动的日志 |
| `--journal-kernel` | 内核消息 | 只监控内核相关日志 |

#### 实际使用场景

```bash
# 监控 Web 服务器错误
./aipipe --format journald --journal-services nginx,apache2 --journal-priority err

# 监控数据库服务
./aipipe --format journald --journal-services postgresql,mysql --journal-priority warning

# 监控系统关键问题
./aipipe --format journald --journal-priority crit --journal-kernel

# 监控特定时间范围
./aipipe --format journald --journal-since "1 hour ago" --journal-priority err
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

### journalctl 参数

- `--journal-services` - 监控的systemd服务列表，逗号分隔 (如: nginx,docker,postgresql)
- `--journal-priority` - 监控的日志级别 (emerg,alert,crit,err,warning,notice,info,debug)
- `--journal-since` - 监控开始时间 (如: '1 hour ago', '2023-10-17 10:00:00')
- `--journal-until` - 监控结束时间 (如: 'now', '2023-10-17 18:00:00')
- `--journal-user` - 监控特定用户的日志
- `--journal-boot` - 只监控当前启动的日志
- `--journal-kernel` - 只监控内核消息

### 多源监控参数

- `--multi-source` - 多源监控配置文件路径
- `--config` - 指定主配置文件路径（可选）

### 配置文件格式支持

AIPipe 支持多种配置文件格式，自动检测和解析：

- ✅ **JSON** - 默认格式，支持 `.json` 扩展名
- ✅ **YAML** - 支持 `.yaml` 和 `.yml` 扩展名
- ✅ **TOML** - 支持 `.toml` 扩展名
- ✅ **自动检测** - 根据文件扩展名和内容自动识别格式

### 自动检测默认配置文件

AIPipe 会自动检测 `~/.config/` 目录下的配置文件：

#### 主配置文件检测顺序
1. `~/.config/aipipe.json`
2. `~/.config/aipipe.yaml`
3. `~/.config/aipipe.yml`
4. `~/.config/aipipe.toml`

#### 多源配置文件检测顺序
1. `~/.config/aipipe-sources.json`
2. `~/.config/aipipe-sources.yaml`
3. `~/.config/aipipe-sources.yml`
4. `~/.config/aipipe-sources.toml`
5. `~/.config/aipipe-multi.json`
6. `~/.config/aipipe-multi.yaml`
7. `~/.config/aipipe-multi.yml`
8. `~/.config/aipipe-multi.toml`

#### 自动启动多源监控
如果检测到多源配置文件，AIPipe 会自动启动多源监控模式：

```bash
# 无需指定参数，自动检测并启动
./aipipe

# 输出示例：
# 🔍 找到默认配置文件: /home/user/.config/aipipe.yaml
# 🔍 自动检测到多源配置文件: /home/user/.config/aipipe-sources.yaml
# 🚀 AIPipe 多源监控启动 - 监控 4 个源
```

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

### 不同配置文件的写法

#### 1. 基础配置（仅AI服务）

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "token": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "model": "gpt-4",
  "custom_prompt": ""
}
```

#### 2. 完整配置（包含所有通知方式）

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "token": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "model": "gpt-4",
  "custom_prompt": "请特别注意数据库连接、内存泄漏、安全相关日志",
  "notifiers": {
    "email": {
      "enabled": true,
      "provider": "smtp",
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "alerts@company.com",
      "password": "your-app-password",
      "from_email": "alerts@company.com",
      "to_emails": ["admin@company.com", "devops@company.com"]
    },
    "dingtalk": {
      "enabled": true,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=xxxxxxxx"
    },
    "wechat": {
      "enabled": true,
      "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxxxxxxx"
    },
    "feishu": {
      "enabled": true,
      "url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx"
    },
    "slack": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/xxxxxxxx/xxxxxxxx/xxxxxxxx"
    },
    "custom_webhooks": [
      {
        "enabled": true,
        "url": "https://your-custom-webhook.com/endpoint",
        "secret": "your-webhook-secret"
      }
    ]
  }
}
```

#### 3. 生产环境配置

```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-production-token",
  "model": "gpt-4",
  "custom_prompt": "生产环境监控，请特别关注：\n1. 数据库连接失败\n2. 内存泄漏警告\n3. 安全攻击尝试\n4. 服务启动失败\n5. 性能严重下降",
  "notifiers": {
    "email": {
      "enabled": true,
      "provider": "smtp",
      "host": "smtp.company.com",
      "port": 587,
      "username": "alerts@company.com",
      "password": "secure-password",
      "from_email": "alerts@company.com",
      "to_emails": ["oncall@company.com", "devops@company.com"]
    },
    "feishu": {
      "enabled": true,
      "url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx"
    }
  }
}
```

#### 4. 开发环境配置

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "token": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "model": "gpt-3.5-turbo",
  "custom_prompt": "开发环境调试，请关注：\n1. 编译错误\n2. 依赖问题\n3. 配置错误\n4. 测试失败",
  "notifiers": {
    "slack": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/xxxxxxxx/xxxxxxxx/xxxxxxxx"
    }
  }
}
```

#### 5. 系统监控配置

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "token": "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
  "model": "gpt-4",
  "custom_prompt": "系统级监控，重点关注：\n1. 内核错误和硬件故障\n2. 服务启动失败\n3. 网络连接问题\n4. 磁盘空间不足\n5. 系统资源耗尽",
  "notifiers": {
    "email": {
      "enabled": true,
      "provider": "resend",
      "host": "",
      "port": 0,
      "username": "",
      "password": "re_xxxxxxxxxxxxx",
      "from_email": "system@company.com",
      "to_emails": ["sysadmin@company.com"]
    },
    "dingtalk": {
      "enabled": true,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=xxxxxxxx"
    }
  }
}
```

#### 6. YAML格式配置

```yaml
ai_endpoint: "https://api.openai.com/v1/chat/completions"
token: "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
model: "gpt-4"
custom_prompt: "请特别注意数据库连接、内存泄漏、安全相关日志"

notifiers:
  email:
    enabled: true
    provider: "smtp"
    host: "smtp.gmail.com"
    port: 587
    username: "alerts@company.com"
    password: "your-app-password"
    from_email: "alerts@company.com"
    to_emails: ["admin@company.com", "devops@company.com"]
  
  dingtalk:
    enabled: true
    url: "https://oapi.dingtalk.com/robot/send?access_token=xxxxxxxx"
```

#### 7. TOML格式配置

```toml
ai_endpoint = "https://api.openai.com/v1/chat/completions"
token = "sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
model = "gpt-4"
custom_prompt = "请特别注意数据库连接、内存泄漏、安全相关日志"

[notifiers.email]
enabled = true
provider = "smtp"
host = "smtp.gmail.com"
port = 587
username = "alerts@company.com"
password = "your-app-password"
from_email = "alerts@company.com"
to_emails = ["admin@company.com", "devops@company.com"]

[notifiers.dingtalk]
enabled = true
url = "https://oapi.dingtalk.com/robot/send?access_token=xxxxxxxx"
```

### 通知配置

AIPipe 支持多种通知方式，当检测到重要日志时会自动发送通知：

#### 邮件通知

支持 SMTP 和 Resend 两种方式：

**SMTP 配置：**
```json
"email": {
  "enabled": true,
  "provider": "smtp",
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "from_email": "your-email@gmail.com",
  "to_emails": ["admin@company.com", "devops@company.com"]
}
```

**Resend 配置：**
```json
"email": {
  "enabled": true,
  "provider": "resend",
  "host": "",
  "port": 0,
  "username": "",
  "password": "re_xxxxxxxxxxxxx",
  "from_email": "alerts@yourdomain.com",
  "to_emails": ["admin@company.com"]
}
```

#### Webhook 通知

支持钉钉、企业微信、飞书、Slack 等平台：

**钉钉机器人：**
```json
"dingtalk": {
  "enabled": true,
  "url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
}
```

**企业微信机器人：**
```json
"wechat": {
  "enabled": true,
  "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
}
```

**飞书机器人：**
```json
"feishu": {
  "enabled": true,
  "url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN"
}
```

**Slack Webhook：**
```json
"slack": {
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}
```

**自定义 Webhook：**
```json
"custom_webhooks": [
  {
    "enabled": true,
    "url": "https://your-custom-webhook.com/endpoint",
    "secret": "your-webhook-secret"
  }
]
```

#### 智能识别

AIPipe 会自动识别 webhook URL 类型，无需手动指定。支持的识别规则：

- **钉钉**: 包含 `dingtalk` 关键词
- **企业微信**: 包含 `qyapi.weixin.qq.com` 域名
- **飞书**: 包含 `feishu` 关键词
- **Slack**: 包含 `slack.com` 域名
- **其他**: 自动使用通用格式

#### 通知示例

当检测到重要日志时，各平台会收到如下格式的通知：

**邮件通知：**
```
主题: ⚠️ 重要日志告警: 数据库连接超时

重要日志告警

摘要: 数据库连接超时

日志内容:
2025-10-17 10:00:01 ERROR Database connection timeout after 30 seconds

时间: 2025-10-17 10:00:01
来源: AIPipe 日志监控系统
```

**钉钉/企业微信/飞书通知：**
```
⚠️ 重要日志告警

摘要: 数据库连接超时

日志内容:
2025-10-17 10:00:01 ERROR Database connection timeout after 30 seconds

时间: 2025-10-17 10:00:01
```

**Slack 通知：**
```
⚠️ 重要日志告警

*摘要:* 数据库连接超时

*日志内容:*
```
2025-10-17 10:00:01 ERROR Database connection timeout after 30 seconds
```

*时间:* 2025-10-17 10:00:01
```

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

- [完整安装指南](INSTALL.md)
- [通知功能说明](NOTIFICATION_FEATURES.md)
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

**xurenlu** <m@some.im>

## 🙏 致谢

- AI 服务提供商 - 提供强大的 AI 能力
- fsnotify - 文件监控库
- Go 社区 - 优秀的工具生态

## 🔗 相关链接

- [问题反馈](https://github.com/xurenlu/aipipe/issues)
- [更新日志](CHANGELOG.md)
- [开发文档](docs/)

---

**Star** ⭐ 如果这个项目对你有帮助！