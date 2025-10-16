# Debug 模式使用示例

## 概述

`--debug` 参数会显示完整的 HTTP 请求和响应信息，用于调试 API 调用问题。

## 对比示例

### 正常模式（无 --debug）

```bash
./supertail -f /var/log/app.log --format java
```

**输出：**
```
🚀 SuperTail 启动 - 监控 java 格式日志
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 监控文件: /var/log/app.log
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📌 从文件末尾开始监控新内容
⏳ 等待新日志...

⚠️  [重要] 2025-10-13 10:00:01 ERROR Database connection failed
   📝 摘要: 数据库连接失败
```

### Debug 模式（带 --debug）

```bash
./supertail -f /var/log/app.log --format java --debug
```

**输出：**
```
🚀 SuperTail 启动 - 监控 java 格式日志
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 监控文件: /var/log/app.log
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📌 从文件末尾开始监控新内容
⏳ 等待新日志...

================================================================================
🔍 DEBUG: HTTP 请求详情
================================================================================
URL: https://cdnproxy.shifen.de/api.poe.com/v1/chat/completions
Method: POST
Headers:
  Content-Type: application/json
  Authorization: Bearer _dJurJWPM...ZmQjq1Y

Request Body:
{
  "model": "Assistant",
  "messages": [
    {
      "role": "user",
      "content": "你是一个专业的日志分析助手。请分析以下 java 格式的日志，判断是否应该过滤掉。\n\n日志内容：\n2025-10-13 10:00:01 ERROR Database connection failed\n\n请以 JSON 格式回复..."
    }
  ]
}
================================================================================
⏳ 发送请求中...
================================================================================
🔍 DEBUG: HTTP 响应详情
================================================================================
Status Code: 200 OK
Response Time: 1.456s
Content-Length: 234 bytes

Response Headers:
  Content-Type: application/json
  Date: Mon, 13 Oct 2025 10:23:45 GMT
  X-Request-Id: req_abc123

Response Body:
{
  "id": "chatcmpl-123",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\n  \"should_filter\": false,\n  \"summary\": \"数据库连接失败\",\n  \"reason\": \"ERROR级别，需要关注\"\n}"
      },
      "finish_reason": "stop"
    }
  ]
}
================================================================================

⚠️  [重要] 2025-10-13 10:00:01 ERROR Database connection failed
   📝 摘要: 数据库连接失败
```

## Debug 模式显示的信息

### 1. HTTP 请求详情
- ✅ **URL**：完整的 API 端点地址
- ✅ **Method**：HTTP 方法（POST）
- ✅ **Headers**：
  - Content-Type
  - Authorization（Token 会部分隐藏，只显示前后各 10 个字符）
- ✅ **Request Body**：格式化的 JSON，包含完整的提示词

### 2. HTTP 响应详情
- ✅ **Status Code**：HTTP 状态码和状态文本
- ✅ **Response Time**：API 响应耗时（精确到毫秒）
- ✅ **Content-Length**：响应体大小（字节）
- ✅ **Response Headers**：所有响应头信息
- ✅ **Response Body**：格式化的 JSON 响应内容

## 使用场景

### 1. 调试 API 连接问题

当 API 调用失败时，debug 模式可以帮助你：

```bash
./supertail -f /var/log/app.log --format java --debug
```

检查：
- URL 是否正确
- Authorization Token 是否有效
- 请求是否成功发送
- 错误信息是什么

### 2. 验证提示词

查看发送给 AI 的完整提示词内容：

```json
{
  "model": "Assistant",
  "messages": [
    {
      "role": "user",
      "content": "你是一个专业的日志分析助手。请分析以下 java 格式的日志..."
    }
  ]
}
```

这样可以：
- 确认提示词格式正确
- 验证日志内容是否完整传递
- 优化提示词以提高准确性

### 3. 分析 AI 响应

查看 AI 返回的原始 JSON：

```json
{
  "should_filter": false,
  "summary": "数据库连接失败",
  "reason": "ERROR级别，需要关注"
}
```

这样可以：
- 验证 AI 判断是否合理
- 调试 JSON 解析问题
- 了解 AI 的决策依据

### 4. 性能分析

查看每次 API 调用的耗时：

```
Response Time: 1.456s
```

这样可以：
- 识别慢响应
- 评估 API 性能
- 优化调用频率

## 快速测试

运行测试脚本体验 debug 模式：

```bash
./test-debug-mode.sh
```

这个脚本会：
1. 先展示正常模式的输出
2. 再展示 debug 模式的输出
3. 让你直观对比两种模式的差异

## 组合使用

Debug 模式可以与其他参数组合：

```bash
# Debug + Verbose（最详细）
./supertail -f /var/log/app.log --format java --debug --verbose

# Debug + 从管道读取
tail -f /var/log/app.log | ./supertail --format java --debug

# Debug + 不同日志格式
./supertail -f /var/log/nginx/error.log --format nginx --debug
```

## 注意事项

⚠️ **Debug 模式会产生大量输出**
- 每条日志都会打印完整的 HTTP 请求和响应
- 建议仅在调试时使用
- 生产环境监控不建议开启 debug

⚠️ **敏感信息**
- API Token 会部分隐藏（只显示首尾）
- 日志内容会完整显示
- 请注意保护敏感信息

⚠️ **性能影响**
- Debug 输出会增加终端 I/O
- 对 API 调用本身无影响
- 主要影响是输出到终端的时间

## 实用技巧

### 1. 重定向 Debug 输出到文件

```bash
./supertail -f /var/log/app.log --format java --debug > debug.log 2>&1
```

### 2. 只查看 HTTP 调用部分

```bash
./supertail -f /var/log/app.log --format java --debug | grep -A 20 "DEBUG:"
```

### 3. 测试单条日志

```bash
echo "2025-10-13 10:00:00 ERROR Test" | ./supertail --format java --debug
```

---

**提示**：正常使用时不需要 `--debug`，只在遇到问题时开启用于诊断。

