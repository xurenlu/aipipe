# AIPipe 实施指南 🚀

## 📋 快速开始

### 第一步：环境准备
```bash
# 1. 确保 Go 环境
go version  # 需要 1.21+

# 2. 克隆项目
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 3. 安装依赖
go mod tidy

# 4. 编译项目
go build -o aipipe aipipe.go
```

### 第二步：基础配置
```bash
# 1. 首次运行创建配置文件
./aipipe --format java --verbose

# 2. 编辑配置文件
nano ~/.config/aipipe.json

# 3. 测试配置
./aipipe config test
```

### 第三步：开始使用
```bash
# 监控日志文件
./aipipe -f /var/log/app.log --format java

# 或通过管道
tail -f /var/log/app.log | ./aipipe --format java
```

## 🔧 配置详解

### 基础配置文件
```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": "请特别注意以下情况：\n1. 数据库连接问题\n2. 内存泄漏警告\n3. 安全相关日志\n4. 性能瓶颈指标"
}
```

### 高级配置选项
```json
{
  "ai": {
    "services": [
      {
        "name": "primary",
        "endpoint": "https://api.openai.com/v1/chat/completions",
        "token": "sk-xxx",
        "model": "gpt-4",
        "priority": 1,
        "enabled": true
      },
      {
        "name": "backup",
        "endpoint": "https://api.anthropic.com/v1/messages",
        "token": "sk-ant-xxx",
        "model": "claude-3-sonnet",
        "priority": 2,
        "enabled": true
      }
    ],
    "timeout": 30,
    "retries": 3,
    "rate_limit": 100
  },
  "processing": {
    "batch_size": 10,
    "batch_timeout": "3s",
    "workers": 4,
    "local_filter": true,
    "context_lines": 3
  },
  "output": {
    "format": "text",
    "color": true,
    "show_filtered": false,
    "notifications": true
  },
  "cache": {
    "enabled": true,
    "ttl": "1h",
    "max_size": 10000
  }
}
```

## 🎯 使用场景

### 场景1：生产环境监控
```bash
# 大批次处理，节省成本
./aipipe -f /var/log/production.log \
  --format java \
  --batch-size 20 \
  --batch-wait 5s \
  --context 5
```

**配置要点**:
- 使用大批次减少 API 调用
- 增加上下文行数便于排查
- 启用本地预过滤

### 场景2：开发调试
```bash
# 详细模式，更多信息
./aipipe -f dev.log \
  --format java \
  --context 10 \
  --verbose \
  --show-not-important
```

**配置要点**:
- 显示所有日志包括过滤的
- 增加上下文行数
- 启用详细输出

### 场景3：历史日志分析
```bash
# 快速分析历史日志
cat /var/log/old/*.log | ./aipipe \
  --format java \
  --batch-size 50 \
  --no-batch
```

**配置要点**:
- 大批次处理历史数据
- 禁用批处理等待时间
- 快速获得结果

## 🔧 高级功能

### 1. 自定义提示词
```json
{
  "custom_prompt": "请特别注意以下情况：\n1. 数据库连接问题\n2. 内存泄漏警告\n3. 安全相关日志\n4. 性能瓶颈指标\n\n请根据这些特殊要求调整判断标准。"
}
```

### 2. 多 AI 服务配置
```json
{
  "ai": {
    "services": [
      {
        "name": "openai",
        "endpoint": "https://api.openai.com/v1/chat/completions",
        "token": "sk-xxx",
        "model": "gpt-4",
        "priority": 1
      },
      {
        "name": "azure",
        "endpoint": "https://your-resource.openai.azure.com/openai/deployments/gpt-4/chat/completions",
        "token": "your-azure-key",
        "model": "gpt-4",
        "priority": 2
      }
    ]
  }
}
```

### 3. 规则引擎配置
```json
{
  "rules": [
    {
      "name": "database_errors",
      "pattern": ".*(database|db|mysql|postgres).*error.*",
      "action": "alert",
      "priority": 1,
      "enabled": true
    },
    {
      "name": "debug_logs",
      "pattern": ".*\\[DEBUG\\].*",
      "action": "filter",
      "priority": 10,
      "enabled": true
    }
  ]
}
```

## 📊 性能调优

### 1. 批处理优化
```bash
# 高频日志 - 大批次
--batch-size 50 --batch-wait 10s

# 低频日志 - 小批次
--batch-size 5 --batch-wait 1s

# 实时处理 - 禁用批处理
--no-batch
```

### 2. 内存优化
```bash
# 限制内存使用
--max-memory 512MB

# 启用流式处理
--stream-mode

# 清理缓存
--cache-clean
```

### 3. 并发优化
```bash
# 调整工作线程数
--workers 8

# 调整队列大小
--queue-size 1000

# 调整超时时间
--timeout 30s
```

## 🚨 故障排除

### 常见问题

#### 1. 配置文件错误
```bash
# 检查配置文件格式
./aipipe config validate

# 测试配置
./aipipe config test

# 重置配置
./aipipe config reset
```

#### 2. AI 服务连接问题
```bash
# 检查网络连接
curl -H "Authorization: Bearer $TOKEN" $ENDPOINT

# 测试 API 调用
./aipipe --debug --verbose

# 查看详细日志
./aipipe --log-level debug
```

#### 3. 性能问题
```bash
# 监控资源使用
./aipipe --metrics

# 调整批处理参数
--batch-size 10 --batch-wait 3s

# 启用本地过滤
--local-filter
```

### 调试模式
```bash
# 完整调试信息
./aipipe -f app.log --format java --debug --verbose

# 只显示错误
./aipipe -f app.log --format java --log-level error

# 性能分析
./aipipe -f app.log --format java --profile
```

## 📈 监控与告警

### 1. 性能指标
```bash
# 查看实时指标
./aipipe --metrics

# 导出指标
./aipipe --metrics --format json > metrics.json

# 监控特定指标
./aipipe --metrics --filter "api_calls,processing_time"
```

### 2. 健康检查
```bash
# 检查服务状态
./aipipe health

# 检查配置
./aipipe config validate

# 检查 AI 服务
./aipipe ai test
```

### 3. 告警配置
```json
{
  "alerts": [
    {
      "name": "high_error_rate",
      "condition": "error_rate > 0.1",
      "action": "notify",
      "enabled": true
    },
    {
      "name": "api_failure",
      "condition": "api_failures > 5",
      "action": "fallback",
      "enabled": true
    }
  ]
}
```

## 🔄 部署方案

### 1. 单机部署
```bash
# 直接运行
./aipipe -f /var/log/app.log --format java

# 后台运行
nohup ./aipipe -f /var/log/app.log --format java > aipipe.log 2>&1 &

# 系统服务
sudo systemctl enable aipipe
sudo systemctl start aipipe
```

### 2. Docker 部署
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o aipipe aipipe.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/aipipe .
CMD ["./aipipe"]
```

```bash
# 构建镜像
docker build -t aipipe .

# 运行容器
docker run -d \
  -v /var/log:/var/log \
  -v ~/.config:/root/.config \
  --name aipipe \
  aipipe -f /var/log/app.log --format java
```

### 3. Kubernetes 部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aipipe
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aipipe
  template:
    metadata:
      labels:
        app: aipipe
    spec:
      containers:
      - name: aipipe
        image: aipipe:latest
        args: ["-f", "/var/log/app.log", "--format", "java"]
        volumeMounts:
        - name: logs
          mountPath: /var/log
        - name: config
          mountPath: /root/.config
      volumes:
      - name: logs
        hostPath:
          path: /var/log
      - name: config
        configMap:
          name: aipipe-config
```

## 📚 最佳实践

### 1. 配置管理
- 使用版本控制管理配置文件
- 定期备份配置
- 使用环境变量管理敏感信息
- 定期验证配置有效性

### 2. 监控策略
- 设置关键指标告警
- 定期检查服务健康状态
- 监控资源使用情况
- 记录和分析性能数据

### 3. 安全考虑
- 保护 API 密钥安全
- 使用 HTTPS 连接
- 定期轮换密钥
- 限制网络访问

### 4. 性能优化
- 根据日志量调整批处理参数
- 启用本地过滤减少 API 调用
- 使用缓存提高响应速度
- 监控内存和 CPU 使用

## 🔗 相关资源

- [项目主页](https://github.com/xurenlu/aipipe)
- [问题反馈](https://github.com/xurenlu/aipipe/issues)
- [功能请求](https://github.com/xurenlu/aipipe/discussions)
- [技术文档](docs/)
- [API 文档](docs/API.md)
- [配置参考](docs/CONFIGURATION.md)

---

**💡 提示**: 如果您在使用过程中遇到问题，请先查看故障排除部分，或提交 Issue 获取帮助。
