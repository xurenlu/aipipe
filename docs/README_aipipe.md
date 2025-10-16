# AIPipe - 智能日志监控工具

AIPipe 是一个智能日志过滤和监控工具，使用 AI 自动分析日志内容，过滤不重要的日志，并对重要事件发送 macOS 通知。

## 功能特性

- 🤖 **AI 智能分析**：使用 Poe API 分析日志内容，自动判断重要性
- 🔍 **多格式支持**：支持 Java、PHP、Nginx、Ruby、FastAPI、Python 等多种日志格式
- 🔔 **实时通知**：对重要日志发送 macOS 系统通知（带声音提醒）
- 🎯 **智能过滤**：自动过滤 INFO、DEBUG 等常规日志，只关注错误和异常
- 📊 **统计信息**：实时显示过滤和告警统计
- 📁 **直接文件监控**：支持直接监控日志文件（-f 参数），类似 `tail -f`
- 💾 **断点续传**：记住上次读取位置，重启后继续监控
- 🔄 **日志轮转支持**：自动检测并处理日志轮转（rotate）情况
- 📦 **批处理模式**：智能累积多行日志批量分析，大幅节省 Token 和减少通知 ⭐
- ⚡ **本地预过滤**：DEBUG/INFO 级别日志本地处理，不调用 AI
- 🛡️ **保守策略**：AI 无法确定时默认过滤，避免误报

## 安装

```bash
cd /Users/rocky/bin
go build -o aipipe aipipe.go
```

## 使用方法

### 基本用法

#### 方式一：直接监控文件（推荐）

```bash
# 监控 Java 日志文件（会记住读取位置）
./aipipe -f /var/log/application.log --format java

# 监控 PHP 日志
./aipipe -f /var/log/php-fpm.log --format php

# 监控 Nginx 日志
./aipipe -f /var/log/nginx/error.log --format nginx

# 详细模式监控
./aipipe -f /var/log/app.log --format java --verbose
```

#### 方式二：从标准输入读取

```bash
# 通过管道监控
tail -f application.log | ./aipipe --format java

# 分析历史日志
cat old.log | ./aipipe --format php

# 监控多个日志文件
tail -f /var/log/*.log | ./aipipe --format java
```

### 参数说明

- `-f <文件路径>`：直接监控指定的日志文件（类似 `tail -f`）
  - 自动记住读取位置，重启后继续监控
  - 自动处理日志轮转（rotate）
  - 推荐使用这种方式

- `--format`：指定日志格式，支持：
  - `java` - Java 应用日志（默认）
  - `php` - PHP 日志
  - `nginx` - Nginx 日志
  - `ruby` - Ruby 日志
  - `python` - Python 日志
  - `fastapi` - FastAPI 日志

- `--verbose`：显示详细输出，包括过滤原因

- `--debug`：调试模式，打印完整的 HTTP 请求和响应详情
  - 显示请求 URL、Headers、Body
  - 显示响应状态码、耗时、Headers、Body
  - 用于调试 API 调用问题
  - 默认关闭

- `--batch-size N`：批处理最大行数（默认 10）
  - 累积 N 行日志后立即批量分析
  - 建议值：5-50

- `--batch-wait 时间`：批处理等待时间（默认 3s）
  - 累积日志后等待指定时间自动处理
  - 支持格式：1s, 500ms, 1m 等
  - 建议值：1s-10s

- `--no-batch`：禁用批处理，逐行分析（默认启用批处理）
  - 每行立即分析，无延迟
  - 增加 API 调用和通知次数
  - 适合调试和低频日志

### 示例

```bash
# 监控生产环境 Java 应用（断点续传）
./aipipe -f /var/log/tomcat/catalina.out --format java

# 监控带日志轮转的文件
./aipipe -f /var/log/app.log --format java

# 详细模式，显示过滤原因
./aipipe -f /var/log/app.log --format java --verbose

# 调试模式，显示 HTTP 请求响应详情
./aipipe -f /var/log/app.log --format java --debug

# 组合使用 verbose 和 debug
./aipipe -f /var/log/app.log --format java --verbose --debug

# 批处理模式（默认，推荐）
./aipipe -f /var/log/app.log --format java

# 自定义批处理参数（大批次，适合高频日志）
./aipipe -f /var/log/app.log --format java --batch-size 20 --batch-wait 2s

# 禁用批处理（逐行分析，实时性更好）
./aipipe -f /var/log/app.log --format java --no-batch

# 从标准输入读取
tail -f /var/log/app.log | ./aipipe --format java
```

## 输出说明

- 🔇 **[过滤]**：不重要的日志，会被过滤（灰色显示）
- ⚠️ **[重要]**：需要关注的日志（红色显示 + 系统通知）

重要日志会触发：
1. 终端输出摘要
2. macOS 系统通知
3. 声音提醒（多层次播放策略）
   - 第一层：通知声音（Glass/Ping/Pop 等）
   - 第二层：直接播放系统音频文件（afplay）
   - 第三层：系统蜂鸣声（beep）

详细说明：`NOTIFICATION_SOUND_GUIDE.md`

## 判断标准

AIPipe 使用了包含 60+ 个真实场景示例的增强提示词，帮助 AI 更准确地判断日志重要性。

### 保守过滤策略 ⭐

**核心原则：只提示确认重要的信息，不确定的一律过滤**

当 AI 无法确定日志重要性时（如格式异常、内容不完整等），系统会自动过滤，避免误报。

**自动过滤的情况：**
- AI 返回包含"日志内容异常"、"日志格式异常"
- AI 返回包含"无法判断"、"不确定"
- AI 返回包含"日志内容不完整"、"日志内容不符合预期"
- 其他表示不确定的关键词

详细说明：`保守过滤策略.md`

### 本地预过滤优化 ⚡

**对于明确的低级别日志，直接在本地过滤，不调用 AI**

- **支持的级别**：DEBUG、INFO、TRACE、VERBOSE
- **性能提升**：处理速度提高 10-30倍（< 0.1秒 vs 1-3秒）
- **节省费用**：减少 60-80% 的 API 调用
- **智能检测**：包含错误关键词的日志仍会调用 AI

**示例：**
```bash
# DEBUG 日志 → 本地过滤（不调用 AI）
echo '2025-10-13 DEBUG User action' | ./aipipe --format java

# ERROR 日志 → 调用 AI 分析
echo '2025-10-13 ERROR Database failed' | ./aipipe --format java
```

详细说明：`本地预过滤优化.md`

### 会被过滤的日志（7 大类，20+ 示例）

1. **健康检查和心跳**
   - Health check, Heartbeat, /health 等

2. **应用启动和配置**
   - Application started, Configuration loaded 等

3. **正常的业务操作**（INFO/DEBUG）
   - User logged in, Retrieved records, Cache hit 等

4. **定时任务正常执行**
   - Scheduled task completed 等

5. **静态资源请求**
   - GET /static/css/style.css 200 等

6. **常规数据库操作**
   - Query executed successfully, Transaction committed 等

7. **正常的 API 请求响应**
   - GET /api/users 200 OK 等

### 需要关注的日志（10 大类，40+ 示例）

1. **错误和异常**（ERROR 级别）
   - ERROR, Exception, Failed 等

2. **数据库问题**
   - Connection timeout, Deadlock, Slow query 等

3. **认证和授权问题**
   - Authentication failed, Permission denied 等

4. **性能问题**（WARN 级别或慢响应）
   - Request timeout, Memory high, Thread pool full 等

5. **资源耗尽**
   - Out of memory, Disk space low 等

6. **外部服务调用失败**
   - Payment gateway timeout, API 500 等

7. **业务异常**
   - Order failed, Payment declined 等

8. **安全问题**
   - SQL injection, Suspicious activity, Rate limit 等

9. **数据一致性问题**
   - Data mismatch, Inconsistent state 等

10. **服务降级和熔断**
    - Circuit breaker opened, Service degraded 等

### 判断规则

- ✅ ERROR 级别或包含 Exception/Error → **告警**
- ✅ 包含 "failed", "timeout", "unable" 等负面词汇 → **仔细判断**
- ✅ WARN 级别 → **根据具体内容判断严重程度**
- ✅ 健康检查、心跳、正常 INFO 日志 → **过滤**

详细示例请查看：`PROMPT_EXAMPLES.md`

## API 配置

工具使用 Poe API 进行日志分析：
- API 地址：https://cdnproxy.shifen.de/api.poe.com/v1
- 模型：Assistant

### System Prompt 设计

AIPipe 使用了符合 OpenAI 最佳实践的 **system + user** 消息结构：

**请求格式：**
```json
{
  "model": "Assistant",
  "messages": [
    {
      "role": "system",
      "content": "你是一个专业的日志分析助手，专门分析 java 格式的日志。\n\n【应该过滤的日志】...\n【需要关注的日志】..."
    },
    {
      "role": "user",
      "content": "请分析以下日志：\n\n2025-10-13 ERROR Database failed"
    }
  ]
}
```

**优势：**
- ✅ **结构清晰**：角色定义在 system，任务在 user
- ✅ **效果更好**：AI 更重视 system 消息中的指令
- ✅ **符合标准**：遵循 OpenAI API 最佳实践
- ✅ **易于维护**：修改判断标准只需改 system 消息

详细说明：`SYSTEM_PROMPT_IMPROVEMENT.md`

如需修改 API 配置，请编辑 `aipipe.go` 中的常量。

## 技术实现

- 语言：Go
- API：Poe (OpenAI 兼容接口)
- 通知：macOS osascript
- 架构：流式处理，低内存占用

## 断点续传与日志轮转

### 状态文件

当使用 `-f` 参数监控文件时，aipipe 会在日志文件所在目录创建状态文件：

```bash
# 监控 /var/log/app.log 时会创建
/var/log/.aipipe_app.log.state
```

状态文件包含：
- 上次读取位置（offset）
- 文件 inode（用于检测轮转）
- 最后更新时间

### 日志轮转处理

工具会自动检测并处理以下日志轮转场景：

1. **文件被重命名或删除**：通过 fsnotify 监控文件事件，自动切换到新文件
2. **文件被截断**：定期检查文件大小，如果小于当前位置则重新打开
3. **inode 变化**：检测到 inode 改变时，认为是新文件，从头开始读取

### 工作流程

```
启动 → 检查状态文件 → 从上次位置继续 → 监控新内容
                    ↓
              （如果没有状态）
                    ↓
            从文件末尾开始监控
```

### 使用建议

- **长期监控**：使用 `-f` 参数，支持断点续传
- **临时查看**：使用管道方式，不保存状态
- **日志分析**：先用 `cat` 分析历史，再用 `-f` 实时监控

## 示例场景

### 监控生产环境（推荐）
```bash
# 直接监控，支持断点续传
./aipipe -f /var/log/app.log --format java

# 或通过 SSH（不支持断点续传）
ssh production "tail -f /var/log/app.log" | ./aipipe --format java
```

### 调试开发环境
```bash
npm run dev 2>&1 | ./aipipe --format fastapi --verbose
```

### 分析错误日志
```bash
grep ERROR app.log | ./aipipe --format java
```

### 监控带轮转的日志
```bash
# logrotate 配置的日志文件
./aipipe -f /var/log/app.log --format java
# 自动处理 app.log.1、app.log.2 等轮转文件
```

## Debug 调试模式

### 使用方法

添加 `--debug` 参数开启调试模式：

```bash
./aipipe -f /var/log/app.log --format java --debug
```

### 输出内容

调试模式会显示每次 API 调用的完整信息：

**HTTP 请求详情**
- 请求 URL
- 请求方法（POST）
- 请求 Headers（Content-Type, Authorization）
- 请求 Body（格式化的 JSON）

**HTTP 响应详情**
- 响应状态码
- 响应耗时
- Content-Length
- 响应 Headers
- 响应 Body（格式化的 JSON）

### 示例输出

```
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
      "content": "你是一个专业的日志分析助手..."
    }
  ]
}
================================================================================
⏳ 发送请求中...
================================================================================
🔍 DEBUG: HTTP 响应详情
================================================================================
Status Code: 200 OK
Response Time: 1.234s
Content-Length: 567 bytes

Response Headers:
  Content-Type: application/json
  Date: Mon, 13 Oct 2025 10:23:45 GMT

Response Body:
{
  "choices": [
    {
      "message": {
        "content": "{\"should_filter\":false,\"summary\":\"数据库连接超时\",\"reason\":\"ERROR级别错误\"}"
      }
    }
  ]
}
================================================================================
```

### 使用场景

1. **调试 API 调用问题**
   - 查看请求是否正确发送
   - 检查 API 返回的错误信息

2. **验证提示词**
   - 查看发送给 AI 的完整提示词
   - 优化提示词内容

3. **分析 API 响应**
   - 查看 AI 返回的原始内容
   - 调试 JSON 解析问题

4. **性能分析**
   - 查看每次 API 调用的耗时
   - 识别慢响应问题

### 测试脚本

```bash
# 运行 debug 模式演示
./test-debug-mode.sh
```

## 注意事项

1. 需要网络连接以调用 Poe API
2. 首次使用可能需要等待 API 响应
3. 日志行过长时会自动截断显示
4. 通知频率较高时可能会被系统限制
5. Debug 模式会产生大量输出，仅用于调试

## 统计输出

程序结束时会显示统计信息：
- 总计处理行数
- 过滤的日志数量
- 发送的告警次数

## JSON 解析兼容性

AIPipe 能自动处理多种 AI 响应格式，无需担心格式问题。

### 支持的格式

✅ **纯 JSON**
```json
{"should_filter": true, "summary": "摘要", "reason": "原因"}
```

✅ **Markdown 代码块（常见）**
```
```json
{"should_filter": true, "summary": "摘要", "reason": "原因"}
```
```

✅ **带前后文本**
```
分析结果：{"should_filter": true, ...}
```

✅ **带空白字符**
```
  
  {"should_filter": true, ...}
  
```

### 解析流程

1. 自动检测 Markdown 代码块（` ```json ... ``` `）
2. 提取代码块内容
3. 智能定位 JSON 起始和结束位置
4. 解析并返回结构化数据

### 错误提示

如果解析失败，会显示详细错误：
```
解析 JSON 失败: invalid character...
原始响应: ```json\n{...
提取的JSON: {...
```

详细说明请查看：`JSON_PARSE_FIX.md`

## 故障排除

### API 调用失败
- 检查网络连接
- 确认 API token 有效
- 查看 `--verbose` 输出的详细错误
- 使用 `--debug` 查看完整请求响应

### 通知不显示
- 确认系统通知权限已启用
- 检查"系统设置 > 通知 > 终端"
- 确保选择了「横幅」样式（不要选「无」）

### 通知显示但没有声音
**常见原因：**
1. 系统通知声音被禁用
2. 音量太低或静音
3. 勿扰模式（专注模式）已开启
4. 终端的通知声音未启用

**解决方法：**
```bash
# 运行声音测试脚本
./test-notification-sound.sh

# 检查音量
osascript -e "output volume of (get volume settings)"

# 设置音量
osascript -e "set volume output volume 50"

# 手动测试通知声音
osascript -e 'display notification "测试" with title "测试" sound name "Glass"'

# 手动测试音频播放
afplay /System/Library/Sounds/Glass.aiff
```

**检查步骤：**
1. 打开「系统设置 > 通知」
2. 找到「终端」或「Terminal」
3. 确保「播放通知声音」已勾选
4. 关闭勿扰模式

### 日志解析错误
- 尝试不同的 `--format` 参数
- 使用 `--verbose` 查看详细信息
- 使用 `--debug` 查看 AI 返回的原始内容

### JSON 解析失败
- 使用 `--debug` 查看 API 返回的原始响应
- 检查错误信息中的"提取的JSON"部分
- 查看 `JSON_PARSE_FIX.md` 了解支持的格式

## 版本信息

- 版本：1.0.0
- 作者：rocky
- 日期：2025-10-13

