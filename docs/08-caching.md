# 08. 缓存系统

> 智能缓存优化，提高性能和减少 API 调用

## 🎯 概述

AIPipe 的缓存系统通过智能缓存分析结果，显著提高性能并减少 API 调用成本。

## 💾 缓存类型

### 1. 内存缓存

```json
{
  "cache": {
    "type": "memory",
    "max_size": 1000,
    "ttl": 3600,
    "strategy": "lru"
  }
}
```

### 2. 文件缓存

```json
{
  "cache": {
    "type": "file",
    "path": "~/.aipipe/cache",
    "max_size": "100MB",
    "ttl": 3600
  }
}
```

### 3. Redis 缓存

```json
{
  "cache": {
    "type": "redis",
    "host": "localhost",
    "port": 6379,
    "password": "your-password",
    "db": 0,
    "ttl": 3600
  }
}
```

## 🔧 缓存管理

### 1. 查看缓存状态

```bash
# 查看缓存统计
aipipe cache stats

# 查看缓存状态
aipipe cache status
```

### 2. 清空缓存

```bash
# 清空所有缓存
aipipe cache clear

# 清空特定缓存
aipipe cache clear --key "error_logs"
```

### 3. 缓存预热

AIPipe 目前不支持缓存预热功能。缓存会在使用过程中自动填充。

## ⚙️ 缓存配置

### 1. 基本配置

```json
{
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000,
    "strategy": "lru",
    "compression": true
  }
}
```

### 2. 高级配置

```json
{
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000,
    "strategy": "lru",
    "compression": true,
    "persistence": {
      "enabled": true,
      "path": "~/.aipipe/cache",
      "sync_interval": 300
    },
    "eviction": {
      "policy": "lru",
      "max_memory": "512MB",
      "cleanup_interval": 600
    }
  }
}
```

## 📊 缓存统计

### 1. 命中率统计

```bash
# 查看命中率
aipipe cache stats --metric "hit_rate"

# 查看详细统计
aipipe cache stats --detailed
```

### 2. 内存使用

```bash
# 查看内存使用
aipipe cache stats --metric "memory_usage"

# 查看缓存大小
aipipe cache stats --metric "cache_size"
```

### 3. 性能指标

```bash
# 查看性能指标
aipipe cache metrics

# 查看实时指标
aipipe cache metrics --realtime
```

## 🎯 使用场景

### 场景1: 重复日志分析

```bash
# 启用缓存
aipipe config set --key "cache.enabled" --value "true"

# 分析重复日志
echo "ERROR Database connection failed" | aipipe analyze --format java
echo "ERROR Database connection failed" | aipipe analyze --format java  # 从缓存获取
```

### 场景2: 批量日志处理

```bash
# 启用批处理缓存
aipipe config set --key "cache.batch_processing" --value "true"

# 批量处理日志
cat logs/*.log | aipipe analyze --format java
```

### 场景3: 规则缓存

```bash
# 启用规则缓存
aipipe config set --key "cache.rules" --value "true"

# 应用规则
aipipe rules apply --file logs/app.log
```

## 🔍 缓存优化

### 1. 缓存策略

```json
{
  "cache_strategies": {
    "frequent": {
      "ttl": 7200,
      "priority": "high"
    },
    "rare": {
      "ttl": 1800,
      "priority": "low"
    },
    "error": {
      "ttl": 3600,
      "priority": "high"
    }
  }
}
```

### 2. 缓存预热

```bash
# 预热常用日志模式
aipipe cache warmup --pattern "ERROR.*"
aipipe cache warmup --pattern "WARN.*"
```

### 3. 缓存清理

```bash
# 定期清理过期缓存
aipipe cache cleanup

# 清理特定模式缓存
aipipe cache cleanup --pattern "DEBUG.*"
```

## 📈 性能监控

### 1. 缓存性能

```bash
# 查看缓存性能
aipipe cache performance

# 查看缓存延迟
aipipe cache latency
```

### 2. 内存监控

```bash
# 查看内存使用
aipipe cache memory

# 查看内存趋势
aipipe cache memory --trend
```

### 3. 成本分析

```bash
# 查看缓存成本
aipipe cache cost

# 查看成本节省
aipipe cache savings
```

## 🔧 故障排除

### 1. 缓存问题

```bash
# 检查缓存状态
aipipe cache status --verbose

# 检查缓存配置
aipipe cache config
```

### 2. 内存问题

```bash
# 检查内存使用
aipipe cache memory --detailed

# 清理内存
aipipe cache cleanup --force
```

### 3. 性能问题

```bash
# 检查缓存性能
aipipe cache performance --detailed

# 优化缓存配置
aipipe cache optimize
```

## 📋 最佳实践

### 1. 缓存配置

- 根据内存大小设置合适的缓存大小
- 设置合理的 TTL 值
- 启用压缩减少内存使用

### 2. 性能优化

- 使用合适的缓存策略
- 定期清理过期缓存
- 监控缓存命中率

### 3. 成本控制

- 使用缓存减少 API 调用
- 监控缓存成本节省
- 优化缓存配置

## 🎉 总结

AIPipe 的缓存系统提供了：

- **多种缓存类型**: 内存、文件、Redis 缓存
- **智能缓存策略**: LRU、LFU 等策略
- **性能优化**: 压缩、持久化、预热
- **监控统计**: 详细的性能指标
- **易于管理**: 完整的缓存管理命令

---

*继续阅读: [09. 配置管理](09-configuration.md)*
