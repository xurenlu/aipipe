# SuperTail 提示词示例说明

## 概述

SuperTail 使用了详细的提示词（Prompt）来指导 AI 判断日志的重要性。提示词中包含了大量实际场景的示例，帮助 AI 更准确地分类日志。

## 提示词结构

### 1. 基本信息
- 日志格式（Java、PHP、Nginx 等）
- 待分析的日志内容
- 期望的 JSON 响应格式

### 2. 应该过滤的日志示例（7 大类，共 20+ 个示例）

#### 类别 1：健康检查和心跳
这类日志表示服务正常运行，无需告警：
```
✅ "Health check endpoint called"
✅ "Heartbeat received from client"
✅ "/health returned 200"
```

**为什么过滤？**
- 频率高（通常每秒或每几秒一次）
- 表示正常状态
- 不需要人工干预

#### 类别 2：应用启动和配置加载
应用启动过程中的正常日志：
```
✅ "Application started successfully"
✅ "Configuration loaded from config.yml"
✅ "Server listening on port 8080"
```

**为什么过滤？**
- 启动时的预期行为
- 只要启动成功就是正常状态
- 如果启动失败会有 ERROR 日志

#### 类别 3：正常的业务操作（INFO/DEBUG）
日常业务流程的正常日志：
```
✅ "User logged in: john@example.com"
✅ "Retrieved 20 records from database"
✅ "Cache hit for key: user_123"
✅ "Request processed in 50ms"
```

**为什么过滤？**
- 表示功能正常工作
- INFO/DEBUG 级别的常规操作
- 没有异常或错误

#### 类别 4：定时任务正常执行
计划任务成功完成：
```
✅ "Scheduled task completed successfully"
✅ "Cleanup job finished, removed 10 items"
```

**为什么过滤？**
- 任务按计划执行
- 结果符合预期
- 如果失败会有 ERROR 日志

#### 类别 5：静态资源请求
Web 服务器提供静态文件：
```
✅ "GET /static/css/style.css 200"
✅ "Serving static file: logo.png"
```

**为什么过滤？**
- 普通的文件服务
- HTTP 200 表示成功
- 不涉及业务逻辑

#### 类别 6：常规数据库操作
正常的数据库交互：
```
✅ "Query executed successfully in 10ms"
✅ "Transaction committed"
```

**为什么过滤？**
- 数据库操作成功
- 响应时间正常（< 1s）
- 没有超时或错误

#### 类别 7：正常的 API 请求响应
HTTP API 正常响应：
```
✅ "GET /api/users 200 OK"
✅ "POST /api/data returned 201"
```

**为什么过滤？**
- HTTP 状态码 2xx 表示成功
- 正常的 RESTful API 调用
- 业务处理正常

---

### 3. 需要关注的日志示例（10 大类，共 40+ 个示例）

#### 类别 1：错误和异常（ERROR 级别）⚠️
任何 ERROR 级别或包含异常的日志：
```
❌ "ERROR: Database connection failed"
❌ "NullPointerException at line 123"
❌ "Failed to connect to Redis"
❌ 任何包含 Exception, Error, Failed 的错误信息
```

**为什么告警？**
- ERROR 级别表示严重问题
- 异常会影响功能正常运行
- 需要立即排查和修复

#### 类别 2：数据库问题 ⚠️
数据库连接、性能、死锁等问题：
```
❌ "Database connection timeout"
❌ "Deadlock detected"
❌ "Slow query: 5000ms"
❌ "Connection pool exhausted"
```

**为什么告警？**
- 数据库是核心依赖
- 问题会影响所有业务
- 可能导致服务不可用

#### 类别 3：认证和授权问题 ⚠️
安全相关的认证失败：
```
❌ "Authentication failed for user admin"
❌ "Invalid token: access denied"
❌ "Permission denied: insufficient privileges"
❌ "Multiple failed login attempts from 192.168.1.100"
```

**为什么告警？**
- 可能是安全攻击
- admin 用户失败特别值得关注
- 多次失败可能是暴力破解

#### 类别 4：性能问题（WARN 级别或慢响应）⚠️
响应慢、超时、资源占用高：
```
❌ "Request timeout after 30s"
❌ "Response time exceeded threshold: 5000ms"
❌ "Memory usage high: 85%"
❌ "Thread pool near capacity: 95/100"
```

**为什么告警？**
- 性能下降影响用户体验
- 可能是资源不足或代码问题
- 需要优化或扩容

#### 类别 5：资源耗尽 ⚠️
内存、磁盘、文件句柄等资源不足：
```
❌ "Out of memory error"
❌ "Disk space low: 95% used"
❌ "Too many open files"
```

**为什么告警？**
- 资源耗尽会导致服务崩溃
- 需要立即处理
- 可能需要重启或清理

#### 类别 6：外部服务调用失败 ⚠️
第三方服务、API 调用问题：
```
❌ "Payment gateway timeout"
❌ "Failed to call external API: 500"
❌ "Third-party service unavailable"
```

**为什么告警？**
- 影响业务功能（如支付）
- 需要联系服务提供商
- 可能需要降级方案

#### 类别 7：业务异常 ⚠️
业务逻辑层面的错误：
```
❌ "Order processing failed: insufficient balance"
❌ "Payment declined: invalid card"
❌ "Data validation failed"
```

**为什么告警？**
- 影响用户操作
- 可能需要人工介入
- 需要分析失败原因

#### 类别 8：安全问题 ⚠️
安全攻击、异常访问：
```
❌ "SQL injection attempt detected"
❌ "Suspicious activity from IP"
❌ "Rate limit exceeded"
❌ "Invalid CSRF token"
```

**为什么告警？**
- 可能是安全攻击
- 需要封禁 IP 或加强防护
- 严重时需要报警

#### 类别 9：数据一致性问题 ⚠️
数据不一致、同步失败：
```
❌ "Data mismatch detected"
❌ "Inconsistent state in transaction"
```

**为什么告警？**
- 数据一致性是关键
- 可能导致业务错误
- 需要立即修复

#### 类别 10：服务降级和熔断 ⚠️
服务保护机制触发：
```
❌ "Circuit breaker opened"
❌ "Service degraded mode activated"
```

**为什么告警？**
- 表示服务异常
- 部分功能不可用
- 需要排查根本原因

---

## 判断规则总结

### 快速判断关键词

**应该过滤（不告警）的关键词：**
- ✅ `INFO`, `DEBUG`
- ✅ `started`, `loaded`, `listening`
- ✅ `successfully`, `completed`, `finished`
- ✅ `health`, `heartbeat`, `ping`
- ✅ `200 OK`, `201 Created`
- ✅ `Cache hit`, `retrieved`, `processed`

**需要告警的关键词：**
- ❌ `ERROR`, `FATAL`
- ❌ `Exception`, `Error`, `Failed`, `Failure`
- ❌ `timeout`, `unable`, `cannot`, `denied`
- ❌ `connection refused`, `unavailable`
- ❌ `exhausted`, `out of`, `low`, `high`
- ❌ `attack`, `injection`, `suspicious`
- ❌ `mismatch`, `inconsistent`

### 特殊注意事项

1. **WARN 级别需要具体分析**
   - 如果是性能、资源、安全相关 → 告警
   - 如果是配置建议、非关键提示 → 可能过滤

2. **HTTP 状态码**
   - 2xx (成功) → 过滤
   - 4xx (客户端错误) → 视情况（429 限流需告警）
   - 5xx (服务器错误) → 告警

3. **响应时间**
   - < 1000ms → 正常
   - 1000ms - 3000ms → 可能需要关注
   - \> 3000ms → 告警

4. **频率考虑**
   - 高频日志（health check）→ 过滤
   - 低频错误 → 更需要关注

## 测试日志文件

项目提供了全面的测试日志：`test-logs-comprehensive.txt`

包含 30 条日志，覆盖所有场景：
- 应该过滤的：12 条
- 需要告警的：18 条

运行测试：
```bash
cat test-logs-comprehensive.txt | ./supertail --format java
```

使用 debug 模式查看 AI 判断依据：
```bash
cat test-logs-comprehensive.txt | ./supertail --format java --debug
```

## 提示词优势

1. **丰富的示例**：40+ 个真实场景示例
2. **清晰的分类**：10 大类需要告警的场景
3. **判断依据**：每类都说明了"为什么"
4. **关键词提示**：帮助 AI 快速识别
5. **特殊规则**：处理边界情况

## 效果验证

使用以下日志测试准确性：

**应该被过滤：**
```
✅ INFO Application started successfully
✅ INFO Health check endpoint called
✅ INFO User logged in: test@example.com
```

**应该被告警：**
```
⚠️  ERROR Database connection failed
⚠️  ERROR NullPointerException at line 123
⚠️  WARN Memory usage high: 85%
```

## 持续优化

如果发现误判，可以：
1. 在提示词中添加更多类似的示例
2. 在"注意事项"部分补充特殊规则
3. 使用 `--debug` 查看 AI 的判断依据
4. 根据实际情况调整示例

---

**最后更新**：2025-10-13  
**版本**：1.0.0

