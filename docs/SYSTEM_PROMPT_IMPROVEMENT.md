# 系统提示词改进

## 改进概述

将原来单一的 `user` 消息拆分为 `system` 和 `user` 两条消息，更符合 OpenAI API 的最佳实践。

## 改进前后对比

### 改进前（单一 user 消息）

```json
{
  "model": "Assistant",
  "messages": [
    {
      "role": "user",
      "content": "你是一个专业的日志分析助手。请分析以下 java 格式的日志...\n\n日志内容：\n2025-10-13 ERROR Database failed\n\n请以 JSON 格式回复...\n\n【应该过滤的日志】...\n【需要关注的日志】..."
    }
  ]
}
```

**问题：**
- 角色定义、判断标准、示例都混在用户消息中
- 结构不清晰
- 每次请求都要重复发送所有示例（浪费 tokens）

### 改进后（system + user 消息）

```json
{
  "model": "Assistant",
  "messages": [
    {
      "role": "system",
      "content": "你是一个专业的日志分析助手，专门分析 java 格式的日志。\n\n你的任务是判断日志是否需要关注...\n\n【应该过滤的日志】...\n【需要关注的日志】..."
    },
    {
      "role": "user",
      "content": "请分析以下日志：\n\n2025-10-13 ERROR Database failed"
    }
  ]
}
```

**优势：**
- ✅ 结构清晰：角色和标准在 system，实际任务在 user
- ✅ 符合最佳实践：system 消息会被 AI 更重视
- ✅ 语义明确：角色定义 vs 具体任务分离
- ✅ 更好的效果：AI 能更好地理解和遵循 system 指令

## 技术实现

### 1. 拆分提示词构建函数

**`buildSystemPrompt(format string)`**
- 定义 AI 的角色（日志分析助手）
- 说明任务目标（判断日志是否需要关注）
- 提供返回格式要求
- 列举所有判断标准和示例（7 大类过滤 + 10 大类告警）
- 说明注意事项

**`buildUserPrompt(logLine string)`**
- 简洁地提出任务
- 只包含待分析的日志内容

### 2. 修改 API 调用

```go
// 调用 Poe API
func callPoeAPI(systemPrompt, userPrompt string) (string, error) {
    reqBody := ChatRequest{
        Model: "Assistant",
        Messages: []ChatMessage{
            {
                Role:    "system",
                Content: systemPrompt,
            },
            {
                Role:    "user",
                Content: userPrompt,
            },
        },
    }
    // ...
}
```

### 3. 更新分析流程

```go
func analyzeLog(logLine string, format string) (*LogAnalysis, error) {
    // 分别构建 system 和 user 消息
    systemPrompt := buildSystemPrompt(format)
    userPrompt := buildUserPrompt(logLine)
    
    // 调用 API
    response, err := callPoeAPI(systemPrompt, userPrompt)
    // ...
}
```

## 消息内容

### System 消息（系统提示词）

```
你是一个专业的日志分析助手，专门分析 {format} 格式的日志。

你的任务是判断日志是否需要关注，并以 JSON 格式返回分析结果。

返回格式：
{
  "should_filter": true/false,
  "summary": "简短摘要（20字内）",
  "reason": "判断原因"
}

判断标准和示例：

【应该过滤的日志】(should_filter=true) - 正常运行状态，无需告警：
1. 健康检查和心跳
   - "Health check endpoint called"
   ...

【需要关注的日志】(should_filter=false) - 异常情况，需要告警：
1. 错误和异常（ERROR级别）
   - "ERROR: Database connection failed"
   ...

注意：
- 如果日志级别是 ERROR 或包含 Exception/Error，通常需要关注
- ...

只返回 JSON，不要其他内容。
```

**特点：**
- 包含所有 60+ 个示例
- 定义 AI 的专业角色
- 明确输出格式要求

### User 消息（用户提示词）

```
请分析以下日志：

2025-10-13 10:00:09 ERROR [http-nio-8080-exec-4] Database connection timeout after 30s
```

**特点：**
- 简洁明了
- 只包含要分析的日志
- 节省 tokens

## 优势分析

### 1. 更好的 AI 理解

**System 消息的特殊性：**
- OpenAI 模型对 system 消息有特殊处理
- System 消息通常被视为"永久指令"
- AI 会更严格地遵循 system 中的要求

### 2. 结构化设计

```
System：我是谁？我要做什么？有什么规则？
User：具体任务是什么？
```

这种结构更符合人类的思维模式，也更符合 AI 训练时的模式。

### 3. Token 优化（未来可能）

虽然目前每次请求都发送完整的 system 消息，但：
- 某些 API 支持缓存 system 消息
- 未来可能支持会话级别的 system 消息复用
- 结构清晰为优化打下基础

### 4. 可维护性

**分离的好处：**
- 修改判断标准只需改 `buildSystemPrompt()`
- 修改任务描述只需改 `buildUserPrompt()`
- 职责清晰，易于测试

### 5. 扩展性

未来可以轻松添加功能：
- 多轮对话（保持 system 不变）
- 不同日志格式使用不同 system 消息
- A/B 测试不同的 system prompt

## Debug 模式查看

使用 `--debug` 可以看到完整的请求结构：

```bash
echo "2025-10-13 ERROR Test" | ./aipipe --format java --debug
```

输出会显示：
```json
{
  "model": "Assistant",
  "messages": [
    {
      "role": "system",
      "content": "你是一个专业的日志分析助手..."
    },
    {
      "role": "user",
      "content": "请分析以下日志：\n\n2025-10-13 ERROR Test"
    }
  ]
}
```

## 最佳实践对比

### OpenAI 官方建议

```
✅ 推荐：
- system: 设置 AI 的行为和角色
- user: 用户的具体请求
- assistant: AI 的回复（多轮对话）

❌ 不推荐：
- 把所有内容都放在 user 消息中
- system 消息过于简短
- 每次在 user 中重复说明角色
```

### 我们的实现

```
✅ System 消息：
- 定义专业角色（日志分析助手）
- 明确任务目标（判断重要性）
- 提供详细标准（60+ 示例）
- 说明输出格式（JSON）

✅ User 消息：
- 简洁的任务描述
- 待分析的日志内容
- 无需重复角色和规则
```

## 对比其他实现

### 方案 A：单一 user 消息（改进前）
```
优点：简单直接
缺点：结构混乱，AI 可能不够重视规则
```

### 方案 B：system + user（改进后）⭐
```
优点：结构清晰，AI 效果更好
缺点：需要两个函数（可接受）
```

### 方案 C：极简 system
```
{
  "role": "system",
  "content": "你是日志分析助手"
}

缺点：缺少详细指导，效果可能不好
```

## 实际效果预期

使用 system prompt 后，预期改进：

1. **更准确的判断**
   - AI 更重视 system 中的标准
   - 示例作为"永久参考"效果更好

2. **更稳定的输出**
   - JSON 格式要求在 system 中更容易遵守
   - 减少格式错误

3. **更好的一致性**
   - 不同日志使用同一套标准
   - AI 行为更可预测

## 相关文件

- `aipipe.go` - 实现代码
  - `buildSystemPrompt()` - 构建系统提示词
  - `buildUserPrompt()` - 构建用户提示词
  - `callPoeAPI()` - 支持两条消息的 API 调用

- `PROMPT_EXAMPLES.md` - 提示词示例说明
- `DEBUG_MODE_EXAMPLE.md` - 如何使用 debug 查看

## 测试验证

```bash
# 基本测试
echo "2025-10-13 INFO Health check" | ./aipipe --format java

# Debug 模式查看消息结构
echo "2025-10-13 ERROR Database failed" | ./aipipe --format java --debug

# 全面测试
cat test-logs-comprehensive.txt | ./aipipe --format java
```

## 版本历史

- **v1.0.0** - 初始版本，单一 user 消息
- **v1.1.0** - 改进版本，使用 system + user 消息 ⭐

---

**最后更新**: 2025-10-13  
**状态**: ✅ 已实现并测试
**推荐**: ⭐⭐⭐⭐⭐ 强烈推荐这种方式

