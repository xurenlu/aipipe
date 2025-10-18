# AIPipe 快速入门

## 5 分钟快速开始

### 1. 基本使用

```bash
# 直接监控日志文件（推荐）
./aipipe -f /var/log/app.log --format java

# 通过管道监控
tail -f /var/log/app.log | ./aipipe --format java
```

### 2. 支持的日志格式

**后端语言：**
- `java` - Java 应用日志（默认）
- `php` - PHP 日志
- `ruby` - Ruby 日志
- `python` - Python 日志
- `fastapi` - FastAPI 日志
- `go` - Go 语言应用日志
- `rust` - Rust 应用日志
- `csharp` - C#/.NET 应用日志
- `kotlin` - Kotlin 应用日志

**前端和全栈：**
- `nodejs` - Node.js 应用日志
- `typescript` - TypeScript 应用日志

**Web 服务器：**
- `nginx` - Nginx 日志

**云原生和容器：**
- `docker` - Docker 容器日志
- `kubernetes` - Kubernetes Pod 日志

**数据库：**
- `postgresql` - PostgreSQL 数据库日志
- `mysql` - MySQL 数据库日志
- `redis` - Redis 日志
- `elasticsearch` - Elasticsearch 日志

**开发工具：**
- `git` - Git 操作日志
- `jenkins` - Jenkins CI/CD 日志
- `github` - GitHub Actions 日志

**系统级日志：**
- `journald` - Linux systemd journal 日志
- `macos-console` - macOS Console 统一日志系统
- `syslog` - 传统 Syslog 日志格式

### 3. 运行示例

```bash
# 运行交互式示例
./aipipe-example.sh

# 测试文件监控功能
./test-aipipe-file.sh

# 测试增强提示词效果（30 条全面测试日志）
./test-prompt-examples.sh
```

## 核心功能

### ✨ 智能过滤

自动识别日志重要性：
- ✅ **保留**：ERROR、Exception、WARN、性能问题、安全问题
- 🔇 **过滤**：INFO、DEBUG、常规请求、健康检查

### 💾 断点续传

使用 `-f` 参数时会自动：
- 记住上次读取位置
- 重启后继续监控
- 状态文件：`.aipipe_文件名.state`

### 🔄 日志轮转

自动检测并处理：
- 文件重命名/删除
- 文件截断
- inode 变化

### 🔔 系统通知

重要日志会触发：
- macOS 系统通知
- 声音提醒（Glass 音效）
- 摘要显示

## 实用场景

### 监控生产环境

```bash
# 本地监控
./aipipe -f /var/log/tomcat/catalina.out --format java

# 远程监控（SSH）
ssh server "tail -f /var/log/app.log" | ./aipipe --format java
```

### 开发调试

```bash
# 详细模式（显示过滤原因）
./aipipe -f logs/development.log --format ruby --verbose

# 监控应用输出
npm run dev 2>&1 | ./aipipe --format fastapi
```

### 分析历史日志

```bash
# 分析整个日志文件
cat app.log | ./aipipe --format java

# 过滤特定内容后分析
grep "2025-10-13" app.log | ./aipipe --format java
```

### 带日志轮转的文件

```bash
# 自动处理 logrotate 配置的文件
./aipipe -f /var/log/app.log --format java
```

## 输出说明

```
🚀 AIPipe 启动 - 监控 java 格式日志
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📁 监控文件: /var/log/app.log
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📌 从文件末尾开始监控新内容
⏳ 等待新日志...

🔇 [过滤] 2025-10-13 10:00:00 INFO Application started
⚠️  [重要] 2025-10-13 10:00:01 ERROR Database connection failed
   📝 摘要: 数据库连接失败，需要检查配置
```

## 状态文件示例

```json
{
  "path": "/var/log/app.log",
  "offset": 12345,
  "inode": 98765,
  "time": "2025-10-13T10:23:45+08:00"
}
```

## 常见问题

### Q: 如何停止监控？
A: 按 `Ctrl+C` 停止

### Q: 状态文件存放在哪里？
A: 与被监控的日志文件在同一目录，文件名为 `.aipipe_原文件名.state`

### Q: 如何重新从头开始读取？
A: 删除状态文件后重启 aipipe

```bash
rm .aipipe_app.log.state
./aipipe -f /var/log/app.log --format java
```

### Q: 能同时监控多个文件吗？
A: 每个 aipipe 实例监控一个文件，需要多个文件时启动多个实例

```bash
./aipipe -f /var/log/app1.log --format java &
./aipipe -f /var/log/app2.log --format php &
```

### Q: API 调用失败怎么办？
A: 检查网络连接，使用 `--verbose` 查看详细错误信息

### Q: 如何自定义过滤规则？
A: 当前版本由 AI 自动判断，后续版本会支持自定义规则

### Q: 如何调试 API 调用问题？
A: 使用 `--debug` 参数查看完整的 HTTP 请求和响应

```bash
./aipipe -f /var/log/app.log --format java --debug
```

这会显示：
- 完整的请求 URL、Headers、Body
- 响应状态码、耗时、Headers、Body
- 方便调试 API 问题和验证提示词

## 性能说明

- **内存占用**：< 50MB（流式处理）
- **API 延迟**：1-3 秒/条（取决于网络）
- **适用场景**：中低频率日志监控（< 100 条/分钟）

## 更多信息

详细文档：`README_aipipe.md`

---

**作者**: xurenlu  
**版本**: 1.0.0  
**日期**: 2025-10-13

