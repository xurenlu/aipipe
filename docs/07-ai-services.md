# 07. AI 服务管理

> 多 AI 服务支持、负载均衡和故障转移

## 🎯 概述

AIPipe 支持多个 AI 服务，提供负载均衡、故障转移和性能优化功能。

## 🤖 支持的 AI 服务

### 1. OpenAI

```json
{
  "ai_services": [
    {
      "name": "openai-gpt4",
      "endpoint": "https://api.openai.com/v1/chat/completions",
      "api_key": "sk-your-openai-key",
      "model": "gpt-4",
      "enabled": true,
      "priority": 1
    }
  ]
}
```

### 2. Azure OpenAI

```json
{
  "ai_services": [
    {
      "name": "azure-gpt4",
      "endpoint": "https://your-resource.openai.azure.com/openai/deployments/gpt-4/chat/completions",
      "api_key": "your-azure-key",
      "model": "gpt-4",
      "enabled": true,
      "priority": 2
    }
  ]
}
```

### 3. 自定义 API

```json
{
  "ai_services": [
    {
      "name": "custom-api",
      "endpoint": "https://your-api.com/v1/chat/completions",
      "api_key": "your-api-key",
      "model": "custom-model",
      "enabled": true,
      "priority": 3
    }
  ]
}
```

## 🔧 服务管理

### 1. 添加服务

```bash
# 添加 OpenAI 服务
aipipe ai add --name "openai-gpt4" --endpoint "https://api.openai.com/v1/chat/completions" --api-key "sk-your-key" --model "gpt-4"

# 添加 Azure OpenAI 服务
aipipe ai add --name "azure-gpt4" --endpoint "https://your-resource.openai.azure.com/openai/deployments/gpt-4/chat/completions" --api-key "your-azure-key" --model "gpt-4"
```

### 2. 列出服务

```bash
# 查看所有服务
aipipe ai list

# 查看启用的服务
aipipe ai list --enabled
```

### 3. 启用/禁用服务

```bash
# 启用服务
aipipe ai enable --name "openai-gpt4"

# 禁用服务
aipipe ai disable --name "openai-gpt4"
```

### 4. 删除服务

```bash
# 删除服务
aipipe ai remove --name "openai-gpt4"
```

## ⚡ 负载均衡

### 1. 轮询策略

```json
{
  "load_balancing": {
    "strategy": "round_robin",
    "health_check": true,
    "health_check_interval": 30
  }
}
```

### 2. 权重策略

```json
{
  "load_balancing": {
    "strategy": "weighted",
    "weights": {
      "openai-gpt4": 3,
      "azure-gpt4": 2,
      "custom-api": 1
    }
  }
}
```

### 3. 最少连接策略

```json
{
  "load_balancing": {
    "strategy": "least_connections",
    "max_connections_per_service": 10
  }
}
```

## 🔄 故障转移

### 1. 自动故障转移

```json
{
  "failover": {
    "enabled": true,
    "max_retries": 3,
    "retry_delay": 5,
    "circuit_breaker": {
      "enabled": true,
      "failure_threshold": 5,
      "recovery_timeout": 60
    }
  }
}
```

### 2. 健康检查

```bash
# 检查服务健康状态
aipipe ai health

# 检查特定服务
aipipe ai health --name "openai-gpt4"
```

### 3. 服务测试

```bash
# 测试所有服务
aipipe ai test

# 测试特定服务
aipipe ai test --name "openai-gpt4"
```

## 📊 性能监控

### 1. 服务统计

```bash
# 查看服务统计
aipipe ai stats

# 查看特定服务统计
aipipe ai stats --name "openai-gpt4"
```

### 2. 性能指标

```bash
# 查看性能指标
aipipe ai metrics

# 查看实时指标
aipipe ai metrics --realtime
```

### 3. 使用量统计

```bash
# 查看使用量
aipipe ai usage

# 查看成本统计
aipipe ai cost
```

## ⚙️ 配置优化

### 1. 超时设置

```json
{
  "ai_services": [
    {
      "name": "openai-gpt4",
      "timeout": 30,
      "max_retries": 3,
      "retry_delay": 1
    }
  ]
}
```

### 2. 频率限制

```json
{
  "rate_limiting": {
    "enabled": true,
    "requests_per_minute": 60,
    "tokens_per_minute": 90000,
    "burst_limit": 10
  }
}
```

### 3. 缓存配置

```json
{
  "caching": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000,
    "strategy": "lru"
  }
}
```

## 🎯 使用场景

### 场景1: 多服务备份

```bash
# 配置主服务和备份服务
aipipe ai add --name "primary" --endpoint "https://api.openai.com/v1/chat/completions" --priority 1
aipipe ai add --name "backup" --endpoint "https://backup-api.com/v1/chat/completions" --priority 2

# 启用故障转移
# 编辑配置文件 ~/.aipipe/config.json
# 设置故障转移参数
```

### 场景2: 负载均衡

```bash
# 配置多个服务
aipipe ai add --name "service1" --endpoint "https://api1.com/v1/chat/completions" --weight 3
aipipe ai add --name "service2" --endpoint "https://api2.com/v1/chat/completions" --weight 2

# 启用负载均衡
# 编辑配置文件 ~/.aipipe/config.json
# 设置负载均衡策略
```

### 场景3: 成本优化

```bash
# 配置不同成本的模型
aipipe ai add --name "cheap-model" --model "gpt-3.5-turbo" --cost-per-token 0.001
aipipe ai add --name "expensive-model" --model "gpt-4" --cost-per-token 0.03

# 设置成本阈值
# 编辑配置文件 ~/.aipipe/config.json
# 设置成本优化参数
```

## 🔍 故障排除

### 1. 服务连接问题

```bash
# 检查服务连接
aipipe ai test --name "openai-gpt4" --verbose

# 检查网络连接
ping api.openai.com
```

### 2. 认证问题

```bash
# 检查 API 密钥
aipipe ai test --name "openai-gpt4" --check-auth

# 验证 API 密钥
curl -H "Authorization: Bearer sk-your-key" https://api.openai.com/v1/models
```

### 3. 性能问题

```bash
# 检查服务性能
aipipe ai stats --name "openai-gpt4"

# 检查响应时间
aipipe ai metrics --name "openai-gpt4" --metric "response_time"
```

## 📋 最佳实践

### 1. 服务配置

- 配置多个服务作为备份
- 设置合理的超时和重试参数
- 启用健康检查和故障转移

### 2. 性能优化

- 使用缓存减少重复请求
- 设置合理的频率限制
- 监控服务性能和使用量

### 3. 成本控制

- 选择合适成本的模型
- 设置成本阈值和告警
- 定期检查使用量和费用

## 🎉 总结

AIPipe 的 AI 服务管理提供了：

- **多服务支持**: 支持多种 AI 服务提供商
- **负载均衡**: 智能的请求分发策略
- **故障转移**: 自动的服务切换和恢复
- **性能监控**: 详细的性能指标和统计
- **成本控制**: 使用量和成本监控
- **易于管理**: 完整的服务管理命令

---

*继续阅读: [08. 缓存系统](08-caching.md)*
