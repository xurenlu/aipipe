# 06. 规则引擎

> 灵活的过滤规则和自定义分析逻辑

## 🎯 概述

AIPipe 的规则引擎提供了强大的日志过滤和自定义分析功能，允许用户定义复杂的过滤规则和自定义分析逻辑。

## 🔧 规则类型

### 1. 过滤规则

```bash
# 添加过滤规则
aipipe rules add --pattern "DEBUG" --action "ignore"
aipipe rules add --pattern "ERROR" --action "alert"
aipipe rules add --pattern "INFO.*User login" --action "ignore"
```

### 2. 正则表达式规则

```bash
# 使用正则表达式
aipipe rules add --pattern "ERROR.*Database" --action "alert"
aipipe rules add --pattern "WARN.*Memory" --action "notify"
```

### 3. 自定义规则

```bash
# 添加自定义规则
aipipe rules add --pattern ".*" --action "custom" --script "custom_analysis.js"
```

## 📋 规则管理

### 1. 列出规则

```bash
# 查看所有规则
aipipe rules list

# 查看特定规则
aipipe rules list --pattern "ERROR"
```

### 2. 启用/禁用规则

```bash
# 启用规则
aipipe rules enable --id 1

# 禁用规则
aipipe rules disable --id 1
```

### 3. 删除规则

```bash
# 删除规则
aipipe rules remove --id 1

# 删除所有规则
aipipe rules clear
```

## 🧪 规则测试

### 1. 测试规则

```bash
# 测试规则
aipipe rules test --pattern "ERROR Database connection failed"

# 测试特定规则
aipipe rules test --id 1 --input "ERROR Database connection failed"
```

### 2. 规则统计

```bash
# 查看规则统计
aipipe rules stats
```

## 📊 规则配置

### 1. 配置文件

```json
{
  "rules": [
    {
      "id": 1,
      "pattern": "DEBUG",
      "action": "ignore",
      "enabled": true,
      "priority": 10
    },
    {
      "id": 2,
      "pattern": "ERROR",
      "action": "alert",
      "enabled": true,
      "priority": 1
    }
  ]
}
```

### 2. 规则优先级

```bash
# 设置规则优先级
aipipe rules set-priority --id 1 --priority 5
```

## 🎯 使用场景

### 场景1: 过滤调试日志

```bash
# 忽略所有DEBUG日志
aipipe rules add --pattern "DEBUG" --action "ignore"

# 忽略特定服务的DEBUG日志
aipipe rules add --pattern "DEBUG.*MyService" --action "ignore"
```

### 场景2: 告警重要错误

```bash
# 告警所有ERROR日志
aipipe rules add --pattern "ERROR" --action "alert"

# 告警特定错误
aipipe rules add --pattern "ERROR.*Database" --action "alert"
```

### 场景3: 自定义分析

```bash
# 自定义分析规则
aipipe rules add --pattern ".*" --action "custom" --script "analysis.js"
```

## 🔍 高级功能

### 1. 条件规则

```bash
# 条件规则
aipipe rules add --pattern "ERROR" --condition "memory_usage > 80" --action "alert"
```

### 2. 组合规则

```bash
# 组合规则
aipipe rules add --pattern "ERROR|FATAL|CRITICAL" --action "alert"
```

### 3. 时间窗口规则

```bash
# 时间窗口规则
aipipe rules add --pattern "ERROR" --window "5m" --threshold "10" --action "alert"
```

## 📈 性能优化

### 1. 规则缓存

```json
{
  "rule_cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  }
}
```

### 2. 规则编译

```bash
# 编译规则
aipipe rules compile

# 预编译规则
aipipe rules precompile
```

## 🎉 总结

AIPipe 的规则引擎提供了：

- **灵活过滤**: 支持正则表达式和自定义规则
- **多种动作**: 忽略、告警、通知、自定义分析
- **优先级管理**: 支持规则优先级和条件规则
- **性能优化**: 规则缓存和预编译
- **易于管理**: 完整的规则管理命令

---

*继续阅读: [07. AI服务管理](07-ai-services.md)*
