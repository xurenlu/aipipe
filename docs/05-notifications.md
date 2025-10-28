# 05. 通知系统

> 多渠道智能通知，确保重要日志及时告警

## 🎯 概述

AIPipe 的通知系统支持多种通知渠道，当检测到重要日志时，会自动发送通知给相关人员，确保问题能够及时处理。

## 🔔 支持的通知渠道

### 1. 邮件通知

```json
{
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "to": "admin@example.com",
      "subject": "AIPipe 日志告警"
    }
  }
}
```

### 2. 系统通知

```json
{
  "notifications": {
    "system": {
      "enabled": true,
      "sound": true,
      "title": "AIPipe 告警"
    }
  }
}
```

### 3. Webhook 通知

```json
{
  "notifications": {
    "webhook": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/xxx",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      }
    }
  }
}
```

## ⚙️ 配置通知

### 1. 初始化通知配置

```bash
# 初始化配置
aipipe config init

# 编辑配置文件
nano ~/.aipipe/config.json
```

### 2. 测试通知

```bash
# 测试所有通知
aipipe notify test

# 测试特定通知
aipipe notify test --email
aipipe notify test --system
aipipe notify test --webhook
```

## 📧 邮件通知配置

### 1. Gmail 配置

```json
{
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "to": "admin@example.com",
      "subject": "AIPipe 日志告警 - {timestamp}",
      "template": "email-template.html"
    }
  }
}
```

### 2. 企业邮箱配置

```json
{
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "mail.company.com",
      "smtp_port": 25,
      "username": "alerts@company.com",
      "password": "password",
      "to": "admin@company.com",
      "subject": "系统告警 - {level} - {service}"
    }
  }
}
```

## 🔔 系统通知配置

### 1. macOS 通知

```json
{
  "notifications": {
    "system": {
      "enabled": true,
      "sound": true,
      "title": "AIPipe 告警",
      "subtitle": "发现重要日志",
      "message": "{summary}"
    }
  }
}
```

### 2. Linux 通知

```json
{
  "notifications": {
    "system": {
      "enabled": true,
      "sound": true,
      "title": "AIPipe 告警",
      "message": "{summary}",
      "urgency": "critical"
    }
  }
}
```

## 🌐 Webhook 通知配置

### 1. Slack 集成

```json
{
  "notifications": {
    "webhook": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK/URL",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "text": "AIPipe 告警: {summary}",
        "channel": "#alerts",
        "username": "AIPipe",
        "icon_emoji": ":warning:"
      }
    }
  }
}
```

### 2. 钉钉集成

```json
{
  "notifications": {
    "webhook": {
      "enabled": true,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
      "method": "POST",
      "headers": {
        "Content-Type": "application/json"
      },
      "body": {
        "msgtype": "text",
        "text": {
          "content": "AIPipe 告警: {summary}"
        }
      }
    }
  }
}
```

## 📱 移动端通知

### 1. 企业微信

```json
{
  "notifications": {
    "webhook": {
      "enabled": true,
      "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx",
      "method": "POST",
      "body": {
        "msgtype": "text",
        "text": {
          "content": "AIPipe 告警: {summary}",
          "mentioned_list": ["@all"]
        }
      }
    }
  }
}
```

### 2. 飞书

```json
{
  "notifications": {
    "webhook": {
      "enabled": true,
      "url": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
      "method": "POST",
      "body": {
        "msg_type": "text",
        "content": {
          "text": "AIPipe 告警: {summary}"
        }
      }
    }
  }
}
```

## 🎨 通知模板

### 1. 邮件模板

```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>AIPipe 日志告警</title>
</head>
<body>
    <h2>🚨 AIPipe 日志告警</h2>
    <p><strong>时间:</strong> {timestamp}</p>
    <p><strong>级别:</strong> {level}</p>
    <p><strong>服务:</strong> {service}</p>
    <p><strong>摘要:</strong> {summary}</p>
    <p><strong>原始日志:</strong></p>
    <pre>{log_line}</pre>
    <p><strong>建议:</strong></p>
    <ul>
        {suggestions}
    </ul>
</body>
</html>
```

### 2. 文本模板

```text
🚨 AIPipe 日志告警
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
时间: {timestamp}
级别: {level}
服务: {service}
摘要: {summary}

原始日志:
{log_line}

建议:
{suggestions}
```

## 🔧 通知管理

### 1. 查看通知状态

```bash
# 查看所有通知器状态
aipipe notify status

# 查看特定通知器
aipipe notify status --email
```

### 2. 发送测试通知

```bash
# 发送测试通知
aipipe notify send --message "这是一条测试通知"

# 发送到特定渠道
aipipe notify send --email --message "邮件测试"
aipipe notify send --system --message "系统通知测试"
```

### 3. 启用/禁用通知

```bash
# 启用通知
aipipe notify enable --email
aipipe notify enable --system

# 禁用通知
aipipe notify disable --email
aipipe notify disable --system
```

## 📊 通知统计

### 1. 查看通知统计

```bash
# 查看通知统计
aipipe notify stats
```

输出示例：
```
📊 通知统计
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📧 邮件通知: 15 条已发送
🔔 系统通知: 8 条已发送
🌐 Webhook: 12 条已发送
📈 成功率: 98.5%
⏱️ 平均延迟: 150ms
```

### 2. 通知历史

```bash
# 查看通知历史
aipipe notify history

# 查看最近的通知
aipipe notify history --limit 10
```

## 🚨 告警策略

### 1. 告警级别

```json
{
  "alert_levels": {
    "critical": {
      "channels": ["email", "system", "webhook"],
      "rate_limit": 60
    },
    "warning": {
      "channels": ["email", "system"],
      "rate_limit": 300
    },
    "info": {
      "channels": ["system"],
      "rate_limit": 600
    }
  }
}
```

### 2. 频率限制

```json
{
  "rate_limiting": {
    "enabled": true,
    "max_notifications_per_minute": 10,
    "cooldown_period": 300
  }
}
```

### 3. 告警聚合

```json
{
  "alert_aggregation": {
    "enabled": true,
    "aggregation_window": 300,
    "max_alerts_per_window": 5
  }
}
```

## 🔍 故障排除

### 1. 邮件发送失败

```bash
# 检查 SMTP 配置
aipipe notify test --email --verbose

# 检查网络连接
telnet smtp.gmail.com 587

# 检查认证信息
openssl s_client -connect smtp.gmail.com:587 -starttls smtp
```

### 2. 系统通知不显示

```bash
# 检查系统通知权限
aipipe notify test --system --verbose

# 检查通知中心设置
# macOS: 系统偏好设置 > 通知
# Linux: 检查 notify-send 命令
```

### 3. Webhook 失败

```bash
# 测试 Webhook URL
curl -X POST "https://hooks.slack.com/services/xxx" \
  -H "Content-Type: application/json" \
  -d '{"text":"测试消息"}'

# 检查网络连接
ping hooks.slack.com
```

## 📋 最佳实践

### 1. 通知配置

- 为不同级别的日志配置不同的通知渠道
- 设置合理的频率限制避免通知轰炸
- 使用模板确保通知格式一致

### 2. 告警策略

- 只对真正重要的日志发送通知
- 设置告警聚合避免重复通知
- 定期检查和调整告警阈值

### 3. 监控和维护

- 定期测试通知功能
- 监控通知发送成功率
- 及时更新通知配置

## 🎉 总结

AIPipe 的通知系统提供了：

- **多渠道支持**: 邮件、系统通知、Webhook
- **灵活配置**: 丰富的配置选项
- **智能告警**: 基于日志重要性的告警策略
- **高性能**: 优化的通知发送机制
- **易维护**: 完善的监控和故障排除工具

---

*继续阅读: [06. 规则引擎](06-rule-engine.md)*
