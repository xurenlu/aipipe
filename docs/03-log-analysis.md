# 03. 日志分析

> AIPipe 的核心功能：AI 驱动的智能日志分析

## 🎯 概述

AIPipe 的日志分析功能使用大语言模型（LLM）自动判断日志的重要性，帮助运维人员快速识别需要关注的问题。

## 🧠 AI 分析原理

### 分析流程

```
日志输入 → 格式识别 → AI 分析 → 重要性判断 → 结果输出
    ↓           ↓         ↓         ↓         ↓
  原始日志   格式解析   智能分析   重要/过滤   通知/显示
```

### 分析维度

AIPipe 从以下维度分析日志：

1. **严重程度**: ERROR > WARN > INFO > DEBUG
2. **关键词匹配**: 错误、异常、失败、超时等
3. **上下文分析**: 结合前后文判断重要性
4. **模式识别**: 识别常见的错误模式
5. **业务影响**: 评估对业务的影响程度

## 📝 支持的日志格式

### Java 应用日志

```bash
# 分析 Java 日志
echo "2024-01-01 10:00:00 ERROR com.example.Service: Database connection failed" | aipipe analyze --format java
```

**格式特点**:
- 时间戳: `2024-01-01 10:00:00`
- 日志级别: `ERROR`, `WARN`, `INFO`, `DEBUG`
- 类名: `com.example.Service`
- 消息: `Database connection failed`

### Nginx 访问日志

```bash
# 分析 Nginx 日志
echo '192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /api/users HTTP/1.1" 200 1234' | aipipe analyze --format nginx
```

**格式特点**:
- IP 地址: `192.168.1.1`
- 时间戳: `[01/Jan/2024:10:00:00 +0000]`
- HTTP 方法: `GET`, `POST`, `PUT`, `DELETE`
- 状态码: `200`, `404`, `500` 等
- 响应大小: `1234`

### Docker 容器日志

```bash
# 分析 Docker 日志
echo "2024-01-01T10:00:00.000Z container_name: ERROR: Service unavailable" | aipipe analyze --format docker
```

**格式特点**:
- ISO 时间戳: `2024-01-01T10:00:00.000Z`
- 容器名: `container_name`
- 日志级别: `ERROR`, `WARN`, `INFO`
- 消息: `Service unavailable`

### JSON 格式日志

```bash
# 分析 JSON 日志
echo '{"timestamp":"2024-01-01T10:00:00Z","level":"ERROR","message":"Database error","service":"api"}' | aipipe analyze --format json
```

**格式特点**:
- 结构化数据
- 标准字段: `timestamp`, `level`, `message`
- 自定义字段: `service`, `user_id`, `request_id` 等

## 🔧 分析配置

### 基本配置

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "your-api-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60,
  "local_filter": true,
  "show_not_important": false
}
```

### 高级配置

```json
{
  "ai_analyzer": {
    "confidence_threshold": 0.7,
    "max_tokens": 1000,
    "temperature": 0.1,
    "custom_prompt": "你是一个专业的日志分析专家...",
    "prompt_file": "prompts/custom.txt"
  }
}
```

## 🎛️ 分析选项

### 命令行选项

```bash
# 基本分析
aipipe analyze

# 指定格式
aipipe analyze --format java

# 显示所有日志（包括不重要的）
aipipe analyze --show-not-important

# 详细输出
aipipe analyze --verbose

# 从文件读取
cat app.log | aipipe analyze --format java
```

### 环境变量

```bash
# 设置默认格式
export AIPIPE_DEFAULT_FORMAT=java

# 设置 API 密钥
export OPENAI_API_KEY=your-api-key

# 设置端点
export AIPIPE_AI_ENDPOINT=https://api.openai.com/v1/chat/completions
```

## 📊 分析结果

### 结果格式

```json
{
  "important": true,
  "summary": "数据库连接失败，需要立即处理",
  "confidence": 0.95,
  "severity": "ERROR",
  "keywords": ["database", "connection", "failed"],
  "suggestions": [
    "检查数据库服务状态",
    "验证连接配置",
    "查看数据库日志"
  ]
}
```

### 显示格式

```
⚠️  [重要] 2024-01-01 10:00:00 ERROR Database connection failed
   📝 摘要: 数据库连接失败，需要立即处理
   🔍 关键词: database, connection, failed
   💡 建议: 检查数据库服务状态
```

## 🎯 分析策略

### 1. 本地预过滤

启用本地过滤可以减少 API 调用：

```json
{
  "local_filter": true,
  "filter_rules": [
    {
      "pattern": "DEBUG",
      "action": "ignore"
    },
    {
      "pattern": "INFO.*User login",
      "action": "ignore"
    }
  ]
}
```

### 2. 关键词过滤

```json
{
  "keyword_filter": {
    "important_keywords": ["ERROR", "FATAL", "CRITICAL", "Exception"],
    "ignore_keywords": ["DEBUG", "TRACE", "INFO.*login"]
  }
}
```

### 3. 正则表达式过滤

```json
{
  "regex_filter": [
    {
      "pattern": "ERROR|FATAL|CRITICAL",
      "action": "analyze"
    },
    {
      "pattern": "DEBUG|TRACE",
      "action": "ignore"
    }
  ]
}
```

## 🔄 批处理分析

### 批量分析文件

```bash
# 分析多个文件
for file in *.log; do
  echo "分析文件: $file"
  cat "$file" | aipipe analyze --format java
done
```

### 实时流式分析

```bash
# 实时分析日志流
tail -f app.log | aipipe analyze --format java
```

### 并行分析

```bash
# 使用 GNU parallel 并行分析
find . -name "*.log" | parallel "cat {} | aipipe analyze --format java"
```

## 📈 性能优化

### 1. 缓存策略

```json
{
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  }
}
```

### 2. 批处理优化

```json
{
  "batch_processing": {
    "enabled": true,
    "batch_size": 10,
    "batch_timeout": 5
  }
}
```

### 3. 并发控制

```json
{
  "concurrency": {
    "max_workers": 5,
    "queue_size": 100
  }
}
```

## 🎨 自定义提示词

### 1. 使用提示词文件

```bash
# 创建自定义提示词
cat > prompts/custom.txt << EOF
你是一个专业的日志分析专家，专门分析 {format} 格式的日志。

请分析以下日志行，判断其重要性：
- 如果是错误、异常、警告等需要关注的问题，标记为重要
- 如果是正常的信息日志，标记为不重要
- 提供简洁的摘要和关键词

日志行: {log_line}
EOF

# 使用自定义提示词
aipipe analyze --format java --prompt-file prompts/custom.txt
```

### 2. 配置提示词文件

```json
{
  "prompt_file": "prompts/custom.txt",
  "prompt_variables": {
    "format": "java",
    "environment": "production"
  }
}
```

## 🔍 调试分析

### 1. 启用调试模式

```bash
# 详细输出
aipipe analyze --verbose

# 调试模式
AIPIPE_DEBUG=1 aipipe analyze
```

### 2. 分析统计

```bash
# 查看分析统计
aipipe cache stats
```

### 3. 测试分析

```bash
# 测试特定日志
echo "ERROR: Database connection failed" | aipipe analyze --format java --verbose
```

## 📋 最佳实践

### 1. 格式选择

- 根据实际日志格式选择正确的 `--format` 参数
- 不确定格式时，可以尝试 `auto` 自动检测

### 2. 性能考虑

- 启用本地过滤减少 API 调用
- 使用缓存提高响应速度
- 合理设置批处理大小

### 3. 错误处理

- 设置合理的重试次数和超时时间
- 监控 API 使用量和费用
- 配置备用 AI 服务

## 🎉 总结

AIPipe 的日志分析功能提供了：

- **智能分析**: 基于 AI 的重要性判断
- **多格式支持**: 支持 20+ 种日志格式
- **灵活配置**: 丰富的配置选项
- **高性能**: 缓存和批处理优化
- **可扩展**: 自定义提示词和规则

---

*继续阅读: [04. 文件监控](04-file-monitoring.md)*
