# 04. 文件监控

> 实时监控日志文件，自动分析新增内容

## 🎯 概述

AIPipe 的文件监控功能可以实时监控一个或多个日志文件，自动分析新增的日志内容，并在发现重要日志时及时通知。

## 🏗️ 监控架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   日志文件      │    │   文件监控器    │    │   AI 分析器     │
│                 │    │                 │    │                 │
│ • app.log       │───▶│ • fsnotify      │───▶│ • 实时分析      │
│ • error.log     │    │ • 文件轮转      │    │ • 重要性判断    │
│ • access.log    │    │ • 断点续传      │    │ • 结果输出      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   通知系统      │
                       │                 │
                       │ • 实时告警      │
                       │ • 多渠道通知    │
                       └─────────────────┘
```

## 🚀 快速开始

### 1. 监控单个文件

```bash
# 监控指定文件
aipipe monitor --file /var/log/app.log --format java

# 监控 Nginx 日志
aipipe monitor --file /var/log/nginx/access.log --format nginx

# 监控 Docker 日志
aipipe monitor --file /var/log/docker/container.log --format docker
```

### 2. 监控多个文件

```bash
# 添加监控文件
aipipe dashboard add

# 按提示输入文件信息
文件路径: /var/log/app.log
日志格式: java
优先级: 10

# 启动监控所有配置的文件
aipipe monitor
```

## 📁 文件管理

### 1. 添加监控文件

```bash
# 交互式添加
aipipe dashboard add

# 手动添加（通过配置文件）
aipipe dashboard add --file /var/log/app.log --format java --priority 10
```

### 2. 列出监控文件

```bash
# 显示所有监控文件
aipipe dashboard list

# 显示文件状态
aipipe dashboard show
```

输出示例：
```
📁 监控文件列表 (2 个文件)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. /var/log/app.log
   📝 格式: java
   ⚡ 优先级: 10
   ✅ 状态: 已启用
   📊 大小: 1.2MB
   🕒 修改时间: 2024-01-01 10:00:00

2. /var/log/nginx/access.log
   📝 格式: nginx
   ⚡ 优先级: 20
   ✅ 状态: 已启用
   📊 大小: 5.8MB
   🕒 修改时间: 2024-01-01 10:05:00
```

### 3. 移除监控文件

```bash
# 移除指定文件
aipipe dashboard remove /var/log/app.log

# 交互式移除
aipipe dashboard remove
```

## ⚙️ 监控配置

### 1. 配置文件结构

监控配置保存在 `~/.aipipe-monitor.json`：

```json
{
  "files": [
    {
      "path": "/var/log/app.log",
      "format": "java",
      "enabled": true,
      "priority": 10
    },
    {
      "path": "/var/log/nginx/access.log",
      "format": "nginx",
      "enabled": true,
      "priority": 20
    }
  ]
}
```

### 2. 优先级设置

优先级数字越小，优先级越高：

- **1-10**: 关键系统日志（如系统错误）
- **11-20**: 应用日志（如业务错误）
- **21-30**: 访问日志（如 Web 访问）
- **31-40**: 调试日志（如详细调试信息）

### 3. 文件状态管理

```bash
# 查看文件状态
aipipe dashboard show

# 列出所有监控文件
aipipe dashboard list

# 移除监控文件
aipipe dashboard remove /var/log/app.log
```

## 🔄 监控模式

### 1. 自动模式

监控所有配置的文件：

```bash
# 启动自动监控
aipipe monitor

# 后台运行
nohup aipipe monitor > monitor.log 2>&1 &
```

### 2. 手动模式

监控指定文件：

```bash
# 监控单个文件
aipipe monitor --file /var/log/app.log --format java

# 监控多个文件（需要多次调用）
aipipe monitor --file /var/log/app.log --format java &
aipipe monitor --file /var/log/nginx/access.log --format nginx &
```

## 📊 监控状态

### 1. 实时状态

```bash
# 查看监控状态
aipipe dashboard show
```

输出示例：
```
🔍 AIPipe 系统状态
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📋 配置信息:
  AI端点: https://api.openai.com/v1/chat/completions
  模型: gpt-3.5-turbo
  最大重试: 3
  超时时间: 30秒
  频率限制: 60次/分钟
  本地过滤: true

📁 监听状态:
  🚀 监控模式: 多文件监控
  📝 监控文件: 2 个文件
  📊 总大小: 7.0MB
  🕒 最后更新: 2024-01-01 10:05:00

🔧 模块状态:
  ✅ 缓存系统: 已启用 (15 项目, 85.2% 命中率)
  📢 通知系统: 2 个通知器已启用
  🔍 规则引擎: 3 个规则已启用
  🤖 AI服务: 1 个服务已启用
  📁 文件监控: 2 个文件已监控
```

### 2. 文件状态详情

```bash
# 查看文件状态详情
aipipe dashboard status
```

## 🔧 高级功能

### 1. 文件轮转支持

AIPipe 自动处理日志文件轮转：

```bash
# 监控轮转的日志文件
aipipe monitor --file /var/log/app.log --format java

# 当文件轮转时，自动切换到新文件
# app.log -> app.log.1 -> app.log.2.gz
```

### 2. 断点续传

监控中断后可以从上次位置继续：

```bash
# 启动监控（自动从上次位置继续）
aipipe monitor

# 强制从头开始
aipipe monitor --from-beginning
```

### 3. 文件过滤

```bash
# 只监控特定模式的文件
aipipe monitor --pattern "*.log"

# 排除特定文件
aipipe monitor --exclude "*.tmp"
```

## 📈 性能优化

### 1. 并发监控

```json
{
  "monitor": {
    "max_concurrent_files": 10,
    "buffer_size": 4096,
    "poll_interval": 100
  }
}
```

### 2. 内存管理

```json
{
  "memory": {
    "max_memory_usage": "512MB",
    "gc_interval": 300
  }
}
```

### 3. 批处理优化

```json
{
  "batch_processing": {
    "enabled": true,
    "batch_size": 10,
    "batch_timeout": 5
  }
}
```

## 🔍 故障排除

### 1. 文件不存在

```bash
# 检查文件是否存在
ls -la /var/log/app.log

# 检查文件权限
ls -la /var/log/app.log
```

### 2. 权限问题

```bash
# 检查文件权限
ls -la /var/log/app.log

# 修改权限
sudo chmod 644 /var/log/app.log

# 修改所有者
sudo chown $USER:$USER /var/log/app.log
```

### 3. 文件被占用

```bash
# 检查文件是否被其他进程占用
lsof /var/log/app.log

# 停止占用进程
sudo kill -9 <PID>
```

## 📋 最佳实践

### 1. 文件选择

- 选择重要的日志文件进行监控
- 避免监控过大的文件（>1GB）
- 定期清理旧的日志文件

### 2. 优先级设置

- 系统错误日志：优先级 1-10
- 应用错误日志：优先级 11-20
- 访问日志：优先级 21-30
- 调试日志：优先级 31-40

### 3. 性能考虑

- 合理设置并发监控文件数量
- 启用批处理优化
- 定期清理缓存

### 4. 监控策略

- 监控关键业务日志
- 设置合理的告警阈值
- 定期检查监控状态

## 🎯 使用场景

### 场景一：Web 应用监控

```bash
# 1. 添加应用日志监控
aipipe dashboard add
# 输入: /var/log/myapp/app.log, java, 10

# 2. 添加访问日志监控
aipipe dashboard add
# 输入: /var/log/nginx/access.log, nginx, 20

# 3. 启动监控
aipipe monitor
```

### 场景二：微服务监控

```bash
# 1. 添加多个服务日志
aipipe dashboard add  # user-service.log
aipipe dashboard add  # order-service.log
aipipe dashboard add  # payment-service.log

# 2. 启动监控
aipipe monitor
```

### 场景三：Docker 容器监控

```bash
# 1. 监控容器日志
aipipe monitor --file /var/log/docker/container.log --format docker

# 2. 监控多个容器
aipipe dashboard add  # container1.log
aipipe dashboard add  # container2.log
aipipe monitor
```

## 🎉 总结

AIPipe 的文件监控功能提供了：

- **实时监控**: 自动检测文件变化
- **多文件支持**: 同时监控多个文件
- **智能分析**: AI 驱动的日志分析
- **灵活配置**: 丰富的配置选项
- **高性能**: 优化的监控性能

---

*继续阅读: [05. 通知系统](05-notifications.md)*
