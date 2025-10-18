# AIPipe 通知功能说明

## 🎯 功能概述

AIPipe 现在支持多种通知方式，当检测到重要日志时会自动发送通知到指定渠道。

## 🔔 支持的通知方式

### 1. 邮件通知

支持两种邮件发送方式：

#### SMTP 邮件
- 支持 Gmail、Outlook、企业邮箱等
- 支持 SSL/TLS 加密
- 支持 STARTTLS

#### Resend API
- 现代化的邮件发送服务
- 高送达率
- 简单易用

### 2. Webhook 通知

#### 钉钉机器人
- 支持钉钉群机器人
- 自动格式化消息
- 支持 @ 提醒

#### 企业微信机器人
- 支持企业微信群机器人
- 自动格式化消息
- 支持 @ 提醒

#### 飞书机器人
- 支持飞书群机器人
- 自动格式化消息
- 支持 @ 提醒

#### Slack Webhook
- 支持 Slack 频道
- Markdown 格式支持
- 自定义用户名和图标

#### 自定义 Webhook
- 支持任意 HTTP webhook
- 通用 JSON 格式
- 可配置签名验证

## 🧠 智能识别功能

AIPipe 会自动识别 webhook URL 类型：

- **钉钉**: 包含 `dingtalk` 关键词
- **企业微信**: 包含 `qyapi.weixin.qq.com` 域名
- **飞书**: 包含 `feishu` 关键词
- **Slack**: 包含 `slack.com` 域名
- **其他**: 自动使用通用格式

## 📝 通知格式

### 邮件通知格式
```
主题: ⚠️ 重要日志告警: [摘要]

重要日志告警

摘要: [AI分析的摘要]

日志内容:
[原始日志内容]

时间: [时间戳]
来源: AIPipe 日志监控系统
```

### Webhook 通知格式

#### 钉钉/企业微信/飞书
```
⚠️ 重要日志告警

摘要: [AI分析的摘要]

日志内容:
[原始日志内容]

时间: [时间戳]
```

#### Slack
```
⚠️ 重要日志告警

*摘要:* [AI分析的摘要]

*日志内容:*
```
[原始日志内容]
```

*时间:* [时间戳]
```

#### 自定义 Webhook
```json
{
  "summary": "[AI分析的摘要]",
  "log_line": "[原始日志内容]",
  "timestamp": "[时间戳]",
  "source": "AIPipe",
  "level": "warning"
}
```

## ⚙️ 配置示例

### 完整配置示例

```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": "",
  "notifiers": {
    "email": {
      "enabled": true,
      "provider": "smtp",
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "from_email": "your-email@gmail.com",
      "to_emails": ["admin@company.com", "devops@company.com"]
    },
    "dingtalk": {
      "enabled": true,
      "url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
    },
    "wechat": {
      "enabled": true,
      "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
    },
    "feishu": {
      "enabled": true,
      "url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN"
    },
    "slack": {
      "enabled": true,
      "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
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

## 🚀 使用场景

### 生产环境监控
- 重要错误自动通知运维团队
- 支持多种通知渠道，确保及时响应
- 批量通知，避免通知轰炸

### 开发环境调试
- 开发团队及时了解应用状态
- 集成到 CI/CD 流程
- 支持自定义 webhook 集成

### 企业级应用
- 支持企业微信、钉钉等企业通讯工具
- 邮件通知支持企业邮箱
- 可配置签名验证，确保安全性

## 🔧 高级功能

### 批量通知
- 智能累积多条重要日志
- 一次发送批量摘要
- 减少通知频率，提高效率

### 智能过滤
- AI 分析日志重要性
- 只对真正重要的事件发送通知
- 避免误报和通知疲劳

### 上下文显示
- 重要日志自动显示前后上下文
- 便于快速定位问题
- 提高排查效率

## 🛠️ 配置指南

### 1. 邮件配置

#### Gmail SMTP
```json
"email": {
  "enabled": true,
  "provider": "smtp",
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "from_email": "your-email@gmail.com",
  "to_emails": ["admin@company.com"]
}
```

#### Resend API
```json
"email": {
  "enabled": true,
  "provider": "resend",
  "password": "re_xxxxxxxxxxxxx",
  "from_email": "alerts@yourdomain.com",
  "to_emails": ["admin@company.com"]
}
```

### 2. Webhook 配置

#### 获取 Webhook URL

**钉钉机器人**：
1. 在钉钉群中添加自定义机器人
2. 选择"自定义关键词"
3. 复制 Webhook URL

**企业微信机器人**：
1. 在企业微信群中添加机器人
2. 复制 Webhook URL

**飞书机器人**：
1. 在飞书群中添加自定义机器人
2. 复制 Webhook URL

**Slack Webhook**：
1. 在 Slack 中创建 Incoming Webhook
2. 复制 Webhook URL

### 3. 测试配置

使用测试脚本验证配置：

```bash
./test-notifications.sh
```

## 🔍 故障排除

### 常见问题

1. **邮件发送失败**
   - 检查 SMTP 服务器配置
   - 验证用户名和密码
   - 检查网络连接

2. **Webhook 发送失败**
   - 验证 Webhook URL 是否正确
   - 检查网络连接
   - 查看详细日志

3. **通知格式错误**
   - 检查 JSON 配置格式
   - 验证字段名称和类型
   - 查看错误日志

### 调试模式

启用调试模式查看详细信息：

```bash
./aipipe -f /var/log/app.log --format java --debug --verbose
```

## 📚 更多信息

- [完整安装指南](INSTALL.md)
- [配置示例](aipipe.json.example)
- [测试脚本](test-notifications.sh)
- [GitHub 仓库](https://github.com/xurenlu/aipipe)

---

**注意**: 通知功能是可选的，如果不配置通知器，AIPipe 仍然可以正常工作，只是不会发送通知。
