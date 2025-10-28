# 02. 快速开始

> 5分钟快速上手 AIPipe，体验核心功能

## 🎯 目标

通过本指南，你将学会：
- 基本配置和初始化
- 分析单行日志
- 监控日志文件
- 配置通知系统
- 使用系统面板

## 📋 前置条件

- 已安装 AIPipe（参考 [01. 简介与安装](01-introduction.md)）
- 有效的 AI API 密钥（OpenAI、Azure OpenAI 等）
- 基本的命令行使用经验

## 🚀 步骤一：初始化配置

### 1. 生成配置文件

```bash
# 初始化配置
aipipe config init
```

这会创建 `~/.aipipe/config.json` 配置文件。

### 2. 配置 AI 服务

编辑配置文件，添加你的 AI API 信息：

```bash
# 编辑配置文件
nano ~/.aipipe/config.json
```

示例配置：

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-openai-api-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60,
  "local_filter": true,
  "show_not_important": false
}
```

### 3. 验证配置

```bash
# 查看配置
aipipe config show

# 测试 AI 服务
aipipe ai test
```

## 🔍 步骤二：分析日志

### 1. 分析单行日志

```bash
# 分析错误日志
echo "2024-01-01 10:00:00 ERROR Database connection failed" | aipipe analyze

# 分析警告日志
echo "2024-01-01 10:00:00 WARN High memory usage detected" | aipipe analyze

# 分析信息日志
echo "2024-01-01 10:00:00 INFO User login successful" | aipipe analyze
```

### 2. 分析文件内容

```bash
# 创建测试日志文件
cat > test.log << EOF
2024-01-01 10:00:00 INFO Application started
2024-01-01 10:01:00 WARN High CPU usage: 85%
2024-01-01 10:02:00 ERROR Database connection failed
2024-01-01 10:03:00 INFO Database reconnected
2024-01-01 10:04:00 ERROR Out of memory
EOF

# 分析文件
cat test.log | aipipe analyze --format java
```

### 3. 查看分析结果

AIPipe 会显示：
- ⚠️ **重要日志**: 需要关注的错误和警告
- 🔇 **过滤日志**: 一般信息日志（如果启用 `--show-not-important`）

## 📁 步骤三：监控文件

### 1. 手动监控单个文件

```bash
# 监控指定文件
aipipe monitor --file test.log --format java
```

### 2. 配置多文件监控

```bash
# 添加监控文件
aipipe dashboard add
```

按提示输入：
- 文件路径: `/var/log/app.log`
- 日志格式: `java`
- 优先级: `10`

### 3. 自动监控所有配置的文件

```bash
# 监控所有配置的文件
aipipe monitor
```

## 🔔 步骤四：配置通知

### 1. 配置邮件通知

编辑配置文件，添加邮件设置：

```json
{
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "to": "admin@example.com"
    }
  }
}
```

### 2. 配置系统通知

```json
{
  "notifications": {
    "system": {
      "enabled": true,
      "sound": true
    }
  }
}
```

### 3. 测试通知

```bash
# 测试邮件通知
aipipe notify test --email

# 测试系统通知
aipipe notify test --system
```

## 📊 步骤五：使用系统面板

### 1. 查看系统状态

```bash
# 显示系统状态
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
  📥 标准输入模式 (未监听文件)
  📝 日志格式: java

🔧 模块状态:
  ✅ 缓存系统: 已启用 (0 项目, 0.00% 命中率)
  📢 通知系统: 2 个通知器已启用
  🔍 规则引擎: 0 个规则已启用
  🤖 AI服务: 1 个服务已启用
  📁 文件监控: 0 个文件已监控
```

### 2. 管理监控文件

```bash
# 列出所有监控文件
aipipe dashboard list

# 添加新的监控文件
aipipe dashboard add

# 移除监控文件
aipipe dashboard remove /path/to/file.log
```

## 🎯 实际使用场景

### 场景一：监控应用日志

```bash
# 1. 添加应用日志监控
aipipe dashboard add
# 输入: /var/log/myapp.log, java, 10

# 2. 启动监控
aipipe monitor

# 3. 查看状态
aipipe dashboard show
```

### 场景二：分析 Nginx 访问日志

```bash
# 1. 分析访问日志
tail -f /var/log/nginx/access.log | aipipe analyze --format nginx

# 2. 监控访问日志
aipipe monitor --file /var/log/nginx/access.log --format nginx
```

### 场景三：Docker 容器日志监控

```bash
# 1. 监控容器日志
docker logs -f mycontainer | aipipe analyze --format docker

# 2. 监控多个容器
aipipe dashboard add  # 添加容器日志文件
aipipe monitor
```

## 🔧 常用命令速查

```bash
# 配置管理
aipipe config init          # 初始化配置
aipipe config show          # 显示配置
aipipe config validate      # 验证配置

# 日志分析
aipipe analyze              # 分析标准输入
aipipe analyze --format nginx  # 指定格式

# 文件监控
aipipe monitor              # 监控所有配置的文件
aipipe monitor --file app.log  # 监控指定文件

# 系统面板
aipipe dashboard show       # 显示系统状态
aipipe dashboard add        # 添加监控文件
aipipe dashboard list       # 列出监控文件
aipipe dashboard remove <path>  # 移除监控文件

# 通知测试
aipipe notify test          # 测试所有通知
aipipe notify test --email # 测试邮件通知
```

## 🎉 恭喜！

你已经完成了 AIPipe 的快速入门！现在可以：

1. 继续阅读 [日志分析](03-log-analysis.md) 深入了解分析功能
2. 查看 [文件监控](04-file-monitoring.md) 学习高级监控技巧
3. 配置 [通知系统](05-notifications.md) 设置告警
4. 探索 [规则引擎](06-rule-engine.md) 自定义过滤规则

## ❓ 遇到问题？

如果遇到问题，可以：

1. 查看 [故障排除](13-troubleshooting.md)
2. 检查 [常见问题](20-faq.md)
3. 在 GitHub 上提交 Issue

---

*继续阅读: [03. 日志分析](03-log-analysis.md)*
