# 10. 提示词管理

> 自定义 AI 提示词，优化分析效果

## 🎯 概述

AIPipe 的提示词管理系统允许用户自定义 AI 分析提示词，优化日志分析效果。

## 📝 提示词文件

### 1. 默认提示词

**位置**: `prompts/default.txt`

```
你是一个专业的日志分析专家，专门分析 {format} 格式的日志。

请分析以下日志行，判断其重要性：
- 如果是错误、异常、警告等需要关注的问题，标记为重要
- 如果是正常的信息日志，标记为不重要
- 提供简洁的摘要和关键词

日志行: {log_line}
```

### 2. 高级提示词

**位置**: `prompts/advanced.txt`

```
你是一个资深的系统运维专家，具有丰富的日志分析经验。

请分析以下 {format} 格式的日志行：

分析要求：
1. 判断日志重要性（重要/不重要）
2. 识别日志类型（错误/警告/信息/调试）
3. 提取关键信息（服务名、错误码、用户ID等）
4. 评估业务影响（高/中/低）
5. 提供处理建议

日志行: {log_line}

请以JSON格式返回分析结果：
{
  "important": true/false,
  "level": "ERROR/WARN/INFO/DEBUG",
  "summary": "简洁摘要",
  "keywords": ["关键词1", "关键词2"],
  "business_impact": "高/中/低",
  "suggestions": ["建议1", "建议2"]
}
```

### 3. 简单提示词

**位置**: `prompts/simple.txt`

```
分析日志重要性：{log_line}

重要：ERROR, FATAL, CRITICAL, Exception
不重要：INFO, DEBUG, TRACE

返回：重要/不重要
```

## 🔧 提示词管理

### 1. 使用提示词文件

```bash
# 使用默认提示词
aipipe analyze --format java

# 使用自定义提示词
aipipe analyze --format java --prompt-file prompts/custom.txt

# 使用提示词文件配置
aipipe config set --key "prompt_file" --value "prompts/advanced.txt"
```

### 2. 创建提示词

```bash
# 创建自定义提示词
cat > prompts/custom.txt << EOF
你是一个专业的 {format} 日志分析专家。

请分析以下日志行：
{log_line}

分析要求：
1. 判断重要性
2. 提取关键信息
3. 提供建议

返回JSON格式结果。
EOF
```

### 3. 测试提示词

```bash
# 测试提示词
aipipe analyze --format java --prompt-file prompts/custom.txt --test

# 测试特定日志
echo "ERROR Database connection failed" | aipipe analyze --format java --prompt-file prompts/custom.txt
```

## 📋 提示词变量

### 1. 内置变量

- `{format}`: 日志格式
- `{log_line}`: 日志行内容
- `{timestamp}`: 当前时间戳
- `{service}`: 服务名称
- `{environment}`: 环境名称

### 2. 自定义变量

```json
{
  "prompt_variables": {
    "format": "java",
    "environment": "production",
    "service": "api-gateway",
    "critical_keywords": "ERROR,FATAL,CRITICAL,Exception"
  }
}
```

### 3. 变量使用

```
你是一个专业的 {format} 日志分析专家，专门分析 {environment} 环境的 {service} 服务日志。

请分析以下日志行，特别关注包含 {critical_keywords} 的日志：

日志行: {log_line}
```

## 🎨 提示词模板

### 1. 错误分析模板

```
错误日志分析模板：

日志格式: {format}
日志内容: {log_line}

分析步骤：
1. 识别错误类型
2. 分析错误原因
3. 评估影响范围
4. 提供解决方案

请返回详细的分析结果。
```

### 2. 性能分析模板

```
性能日志分析模板：

日志格式: {format}
日志内容: {log_line}

分析重点：
1. 性能指标
2. 响应时间
3. 资源使用
4. 瓶颈识别

请提供性能分析报告。
```

### 3. 安全分析模板

```
安全日志分析模板：

日志格式: {format}
日志内容: {log_line}

安全分析：
1. 威胁等级
2. 攻击类型
3. 影响评估
4. 应对措施

请提供安全分析结果。
```

## 🔄 提示词优化

### 1. 性能优化

```bash
# 优化提示词长度
aipipe prompt optimize --file prompts/custom.txt --max-tokens 1000

# 压缩提示词
aipipe prompt compress --file prompts/custom.txt
```

### 2. 效果优化

```bash
# 测试提示词效果
aipipe prompt test --file prompts/custom.txt --test-data test-logs.txt

# 比较提示词效果
aipipe prompt compare prompts/default.txt prompts/advanced.txt
```

### 3. 自动优化

```bash
# 自动优化提示词
aipipe prompt auto-optimize --file prompts/custom.txt

# 基于历史数据优化
aipipe prompt optimize --file prompts/custom.txt --history
```

## 📊 提示词统计

### 1. 使用统计

```bash
# 查看提示词使用统计
aipipe prompt stats

# 查看特定提示词统计
aipipe prompt stats --file prompts/custom.txt
```

### 2. 效果统计

```bash
# 查看分析效果
aipipe prompt effectiveness

# 查看准确率
aipipe prompt accuracy
```

### 3. 成本统计

```bash
# 查看提示词成本
aipipe prompt cost

# 查看令牌使用
aipipe prompt tokens
```

## 🎯 使用场景

### 场景1: 自定义分析逻辑

```bash
# 创建业务特定提示词
cat > prompts/business.txt << EOF
你是一个电商系统日志分析专家。

请分析以下 {format} 格式的电商日志：
{log_line}

重点关注：
1. 订单处理错误
2. 支付异常
3. 库存问题
4. 用户行为异常

请提供业务分析结果。
EOF

# 使用业务提示词
aipipe analyze --format java --prompt-file prompts/business.txt
```

### 场景2: 多环境配置

```bash
# 开发环境提示词
aipipe config set --key "prompt_file" --value "prompts/development.txt"

# 生产环境提示词
aipipe config set --key "prompt_file" --value "prompts/production.txt"
```

### 场景3: 动态提示词

```bash
# 根据日志格式选择提示词
aipipe analyze --format java --prompt-file prompts/java.txt
aipipe analyze --format nginx --prompt-file prompts/nginx.txt
```

## 🔍 故障排除

### 1. 提示词问题

```bash
# 验证提示词格式
aipipe prompt validate --file prompts/custom.txt

# 检查提示词变量
aipipe prompt check-variables --file prompts/custom.txt
```

### 2. 效果问题

```bash
# 检查分析效果
aipipe prompt test --file prompts/custom.txt --verbose

# 调试提示词
aipipe prompt debug --file prompts/custom.txt
```

### 3. 性能问题

```bash
# 检查提示词性能
aipipe prompt performance --file prompts/custom.txt

# 优化提示词
aipipe prompt optimize --file prompts/custom.txt
```

## 📋 最佳实践

### 1. 提示词设计

- 明确分析目标和要求
- 使用清晰的指令和格式
- 包含具体的示例和模板
- 考虑不同日志格式的特点

### 2. 变量使用

- 合理使用内置变量
- 定义有意义的自定义变量
- 避免变量冲突和重复

### 3. 性能优化

- 控制提示词长度
- 使用高效的指令格式
- 定期测试和优化效果

## 🎉 总结

AIPipe 的提示词管理提供了：

- **自定义提示词**: 支持自定义分析逻辑
- **模板系统**: 多种预设模板
- **变量支持**: 灵活的参数化
- **效果优化**: 自动优化和测试
- **统计分析**: 详细的使用统计
- **易于管理**: 完整的提示词管理命令

---

*继续阅读: [11. 部署指南](11-deployment.md)*
