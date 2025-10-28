# 18. 示例集合

> 实际使用案例和最佳实践

## 🎯 概述

本章节提供了 AIPipe 的实际使用案例，涵盖不同场景和最佳实践。

## 🚀 基础使用示例

### 示例 1: 分析单行日志

```bash
# 分析错误日志
echo "2024-01-01 10:00:00 ERROR Database connection failed" | aipipe analyze --format java

# 分析警告日志
echo "2024-01-01 10:01:00 WARN High memory usage: 85%" | aipipe analyze --format java

# 分析信息日志
echo "2024-01-01 10:02:00 INFO User login successful" | aipipe analyze --format java
```

### 示例 2: 分析文件内容

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

### 示例 3: 监控文件

```bash
# 监控单个文件
aipipe monitor --file test.log --format java

# 监控多个文件
aipipe dashboard add  # 添加文件
aipipe monitor        # 启动监控
```

## 🌐 Web 应用监控

### 示例 1: Nginx 访问日志监控

```bash
# 1. 创建 Nginx 日志文件
cat > nginx.log << EOF
192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /api/users HTTP/1.1" 200 1234
192.168.1.2 - - [01/Jan/2024:10:01:00 +0000] "POST /api/login HTTP/1.1" 401 567
192.168.1.3 - - [01/Jan/2024:10:02:00 +0000] "GET /api/health HTTP/1.1" 200 89
192.168.1.4 - - [01/Jan/2024:10:03:00 +0000] "GET /api/users HTTP/1.1" 500 0
EOF

# 2. 分析访问日志
cat nginx.log | aipipe analyze --format nginx

# 3. 监控访问日志
aipipe monitor --file nginx.log --format nginx
```

### 示例 2: Apache 访问日志监控

```bash
# 1. 创建 Apache 日志文件
cat > apache.log << EOF
192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /index.html HTTP/1.1" 200 1234
192.168.1.2 - - [01/Jan/2024:10:01:00 +0000] "POST /api/login HTTP/1.1" 401 567
EOF

# 2. 分析 Apache 日志
cat apache.log | aipipe analyze --format apache

# 3. 监控 Apache 日志
aipipe monitor --file apache.log --format apache
```

## 🐳 容器监控

### 示例 1: Docker 容器日志监控

```bash
# 1. 创建 Docker 日志文件
cat > docker.log << EOF
2024-01-01T10:00:00.000Z container_name: ERROR: Service unavailable
2024-01-01T10:01:00.000Z container_name: WARN: High memory usage
2024-01-01T10:02:00.000Z container_name: INFO: Service started
EOF

# 2. 分析 Docker 日志
cat docker.log | aipipe analyze --format docker

# 3. 监控 Docker 日志
aipipe monitor --file docker.log --format docker
```

### 示例 2: Kubernetes Pod 日志监控

```bash
# 1. 创建 K8s 日志文件
cat > k8s.log << EOF
2024-01-01T10:00:00.000Z k8s-pod-123: ERROR: Pod failed to start
2024-01-01T10:01:00.000Z k8s-pod-456: WARN: Resource limit exceeded
EOF

# 2. 分析 K8s 日志
cat k8s.log | aipipe analyze --format kubernetes

# 3. 监控 K8s 日志
aipipe monitor --file k8s.log --format kubernetes
```

## 📱 移动应用监控

### 示例 1: Android 应用日志监控

```bash
# 1. 创建 Android 日志文件
cat > android.log << EOF
01-01 10:00:00.123  1234  5678 E MyApp: Database connection failed
01-01 10:01:00.456  1234  5678 W MyApp: High memory usage
01-01 10:02:00.789  1234  5678 I MyApp: User login successful
EOF

# 2. 分析 Android 日志
cat android.log | aipipe analyze --format android

# 3. 监控 Android 日志
aipipe monitor --file android.log --format android
```

### 示例 2: iOS 应用日志监控

```bash
# 1. 创建 iOS 日志文件
cat > ios.log << EOF
2024-01-01 10:00:00.123 MyApp[1234:5678] ERROR: Database connection failed
2024-01-01 10:01:00.456 MyApp[1234:5678] WARN: High memory usage
EOF

# 2. 分析 iOS 日志
cat ios.log | aipipe analyze --format ios

# 3. 监控 iOS 日志
aipipe monitor --file ios.log --format ios
```

## 📊 结构化日志监控

### 示例 1: JSON 格式日志监控

```bash
# 1. 创建 JSON 日志文件
cat > json.log << EOF
{"timestamp":"2024-01-01T10:00:00Z","level":"ERROR","message":"Database connection failed","service":"api","user_id":12345}
{"timestamp":"2024-01-01T10:01:00Z","level":"WARN","message":"High memory usage","service":"api","memory_usage":85}
{"timestamp":"2024-01-01T10:02:00Z","level":"INFO","message":"User login successful","service":"api","user_id":12345}
EOF

# 2. 分析 JSON 日志
cat json.log | aipipe analyze --format json

# 3. 监控 JSON 日志
aipipe monitor --file json.log --format json
```

### 示例 2: XML 格式日志监控

```bash
# 1. 创建 XML 日志文件
cat > xml.log << EOF
<log timestamp="2024-01-01T10:00:00Z" level="ERROR" service="api">
  <message>Database connection failed</message>
  <user_id>12345</user_id>
</log>
<log timestamp="2024-01-01T10:01:00Z" level="WARN" service="api">
  <message>High memory usage</message>
  <memory_usage>85</memory_usage>
</log>
EOF

# 2. 分析 XML 日志
cat xml.log | aipipe analyze --format xml

# 3. 监控 XML 日志
aipipe monitor --file xml.log --format xml
```

## 🔧 高级配置示例

### 示例 1: 多文件监控配置

```bash
# 1. 添加多个监控文件
aipipe dashboard add
# 输入: /var/log/app.log, java, 10

aipipe dashboard add
# 输入: /var/log/nginx/access.log, nginx, 20

aipipe dashboard add
# 输入: /var/log/docker/container.log, docker, 30

# 2. 启动多文件监控
aipipe monitor

# 3. 查看监控状态
aipipe dashboard show
```

### 示例 2: 通知配置

```bash
# 1. 配置邮件通知
aipipe config set --key "notifications.email.enabled" --value "true"
aipipe config set --key "notifications.email.smtp_host" --value "smtp.gmail.com"
aipipe config set --key "notifications.email.username" --value "your-email@gmail.com"
aipipe config set --key "notifications.email.password" --value "your-app-password"
aipipe config set --key "notifications.email.to" --value "admin@example.com"

# 2. 配置系统通知
aipipe config set --key "notifications.system.enabled" --value "true"
aipipe config set --key "notifications.system.sound" --value "true"

# 3. 测试通知
aipipe notify test
```

### 示例 3: 规则配置

```bash
# 1. 添加过滤规则
aipipe rules add --pattern "DEBUG" --action "ignore"
aipipe rules add --pattern "INFO.*User login" --action "ignore"
aipipe rules add --pattern "ERROR" --action "alert"

# 2. 列出规则
aipipe rules list

# 3. 测试规则
aipipe rules test --pattern "ERROR Database connection failed"
```

## 🎯 实际场景示例

### 场景 1: 电商网站监控

```bash
# 1. 监控应用日志
aipipe dashboard add
# 输入: /var/log/ecommerce/app.log, java, 10

# 2. 监控访问日志
aipipe dashboard add
# 输入: /var/log/nginx/access.log, nginx, 20

# 3. 监控数据库日志
aipipe dashboard add
# 输入: /var/log/mysql/error.log, mysql, 5

# 4. 启动监控
aipipe monitor
```

### 场景 2: 微服务架构监控

```bash
# 1. 监控用户服务
aipipe dashboard add
# 输入: /var/log/user-service.log, java, 10

# 2. 监控订单服务
aipipe dashboard add
# 输入: /var/log/order-service.log, java, 10

# 3. 监控支付服务
aipipe dashboard add
# 输入: /var/log/payment-service.log, java, 10

# 4. 监控网关服务
aipipe dashboard add
# 输入: /var/log/gateway.log, nginx, 15

# 5. 启动监控
aipipe monitor
```

### 场景 3: 云原生应用监控

```bash
# 1. 监控 Kubernetes Pod 日志
aipipe dashboard add
# 输入: /var/log/k8s/pod.log, kubernetes, 10

# 2. 监控 Docker 容器日志
aipipe dashboard add
# 输入: /var/log/docker/container.log, docker, 15

# 3. 监控 AWS CloudWatch 日志
aipipe dashboard add
# 输入: /var/log/cloudwatch/app.log, cloudwatch, 20

# 4. 启动监控
aipipe monitor
```

## 📈 性能优化示例

### 示例 1: 批处理优化

```bash
# 1. 启用批处理
aipipe config set --key "batch_processing.enabled" --value "true"
aipipe config set --key "batch_processing.batch_size" --value "10"
aipipe config set --key "batch_processing.batch_timeout" --value "5"

# 2. 启动监控
aipipe monitor
```

### 示例 2: 缓存优化

```bash
# 1. 启用缓存
aipipe config set --key "cache.enabled" --value "true"
aipipe config set --key "cache.ttl" --value "3600"
aipipe config set --key "cache.max_size" --value "1000"

# 2. 查看缓存统计
aipipe cache stats
```

### 示例 3: 并发优化

```bash
# 1. 设置并发参数
aipipe config set --key "concurrency.max_workers" --value "5"
aipipe config set --key "concurrency.queue_size" --value "100"

# 2. 启动监控
aipipe monitor
```

## 🔍 故障排除示例

### 示例 1: 调试分析问题

```bash
# 1. 启用详细输出
aipipe analyze --verbose

# 2. 启用调试模式
AIPIPE_DEBUG=1 aipipe analyze

# 3. 测试特定日志
echo "ERROR: Database connection failed" | aipipe analyze --format java --verbose
```

### 示例 2: 调试监控问题

```bash
# 1. 检查文件权限
ls -la /var/log/app.log

# 2. 检查文件是否被占用
lsof /var/log/app.log

# 3. 测试文件监控
aipipe monitor --file /var/log/app.log --format java --verbose
```

### 示例 3: 调试通知问题

```bash
# 1. 测试邮件通知
aipipe notify test --email --verbose

# 2. 测试系统通知
aipipe notify test --system --verbose

# 3. 测试 Webhook 通知
aipipe notify test --webhook --verbose
```

## 📋 最佳实践示例

### 示例 1: 生产环境配置

```bash
# 1. 创建生产环境配置
cat > production-config.json << EOF
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-production-key",
  "ai_model": "gpt-4",
  "max_retries": 5,
  "timeout": 60,
  "rate_limit": 100,
  "local_filter": true,
  "show_not_important": false,
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.company.com",
      "smtp_port": 587,
      "username": "alerts@company.com",
      "password": "production-password",
      "to": "admin@company.com"
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
  },
  "batch_processing": {
    "enabled": true,
    "batch_size": 10,
    "batch_timeout": 5
  }
}
EOF

# 2. 应用配置
cp production-config.json ~/.aipipe/config.json

# 3. 验证配置
aipipe config validate
```

### 示例 2: 开发环境配置

```bash
# 1. 创建开发环境配置
cat > development-config.json << EOF
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-dev-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60,
  "local_filter": true,
  "show_not_important": true,
  "notifications": {
    "system": {
      "enabled": true,
      "sound": false
    }
  },
  "cache": {
    "enabled": true,
    "ttl": 1800,
    "max_size": 100
  }
}
EOF

# 2. 应用配置
cp development-config.json ~/.aipipe/config.json

# 3. 验证配置
aipipe config validate
```

## 🎉 总结

本章节提供了丰富的使用示例，包括：

- **基础使用**: 单行日志分析、文件分析、文件监控
- **Web 应用**: Nginx、Apache 日志监控
- **容器平台**: Docker、Kubernetes 日志监控
- **移动应用**: Android、iOS 日志监控
- **结构化日志**: JSON、XML 日志监控
- **高级配置**: 多文件监控、通知配置、规则配置
- **实际场景**: 电商网站、微服务、云原生应用监控
- **性能优化**: 批处理、缓存、并发优化
- **故障排除**: 调试分析、监控、通知问题
- **最佳实践**: 生产环境、开发环境配置

这些示例可以帮助你快速上手 AIPipe，并根据实际需求进行配置和优化。

---

*返回: [文档首页](README.md)*
