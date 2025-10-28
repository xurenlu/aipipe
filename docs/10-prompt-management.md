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

AIPipe 支持通过配置文件指定自定义提示词文件：

```bash
# 编辑配置文件，添加提示词文件路径
nano ~/.aipipe/config.json

# 在配置文件中添加：
# {
#   "prompt_file": "prompts/custom.txt"
# }

# 使用配置的提示词文件
aipipe analyze --format java
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
# 测试提示词效果
echo "ERROR Database connection failed" | aipipe analyze --format java

# 比较不同提示词的效果
aipipe config set --key "prompt_file" --value "prompts/default.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java

aipipe config set --key "prompt_file" --value "prompts/custom.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java
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

### 1. 手动优化

通过编辑提示词文件来优化效果：

```bash
# 编辑提示词文件
nano prompts/custom.txt

# 测试优化效果
echo "ERROR Database connection failed" | aipipe analyze --format java
```

### 2. 效果测试

```bash
# 使用不同提示词测试
aipipe config set --key "prompt_file" --value "prompts/default.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java

aipipe config set --key "prompt_file" --value "prompts/custom.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java
```

### 3. 变量优化

通过调整提示词变量来优化效果：

```json
{
  "prompt_variables": {
    "format": "java",
    "environment": "production",
    "critical_keywords": "ERROR,FATAL,CRITICAL,Exception"
  }
}
```

## 📊 提示词效果

### 1. 测试不同提示词

```bash
# 测试默认提示词
aipipe config set --key "prompt_file" --value "prompts/default.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java

# 测试高级提示词
aipipe config set --key "prompt_file" --value "prompts/advanced.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java

# 测试简单提示词
aipipe config set --key "prompt_file" --value "prompts/simple.txt"
echo "ERROR Database connection failed" | aipipe analyze --format java
```

### 2. 效果对比

通过比较不同提示词的分析结果来选择最佳提示词：

```bash
# 创建测试日志文件
cat > test-logs.txt << EOF
ERROR Database connection failed
WARN High memory usage detected
INFO User login successful
DEBUG Processing request
EOF

# 使用不同提示词测试
for prompt in prompts/*.txt; do
    echo "测试提示词: $prompt"
    aipipe config set --key "prompt_file" --value "$prompt"
    cat test-logs.txt | aipipe analyze --format java
    echo "---"
done
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
# 检查提示词文件是否存在
ls -la prompts/custom.txt

# 检查提示词文件内容
cat prompts/custom.txt

# 检查配置文件中的提示词文件路径
aipipe config show --key "prompt_file"
```

### 2. 效果问题

```bash
# 测试提示词效果
echo "ERROR Database connection failed" | aipipe analyze --format java --verbose

# 检查提示词变量替换
grep -n "{format}" prompts/custom.txt
grep -n "{log_line}" prompts/custom.txt
```

### 3. 配置问题

```bash
# 检查配置文件
aipipe config validate

# 重新设置提示词文件
aipipe config set --key "prompt_file" --value "prompts/custom.txt"

# 测试配置
aipipe analyze --format java
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
