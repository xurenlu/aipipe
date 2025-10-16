# SuperTail - 智能日志监控工具 🚀

> 使用 AI 自动分析日志内容，智能过滤噪音，只关注真正重要的问题

[English](README.md) | 简体中文

## ⭐ 项目亮点

SuperTail 是一个专为开发者和运维人员设计的智能日志监控工具，它能：

- 🎯 **自动过滤 80% 的噪音日志**（DEBUG、INFO、健康检查等）
- 🤖 **AI 智能判断**重要性（使用 Azure OpenAI gpt-5-mini）
- 📦 **批量处理**日志，节省 70-90% Token 和费用
- 🔔 **重要日志立即通知**，配合声音提醒
- 📋 **自动显示上下文**，完整的错误场景一目了然
- ⚡ **极速处理**，本地过滤 < 0.1秒

## 🚀 5 分钟快速开始

### 1. 安装

```bash
# 克隆仓库
git clone https://github.com/your-username/supertail.git
cd supertail

# 编译
go build -o supertail supertail.go
```

### 2. 配置 API

编辑 `supertail.go` 设置你的 Azure OpenAI API：

```go
const (
    AZURE_API_ENDPOINT = "your-endpoint"
    AZURE_API_KEY      = "your-api-key"
)
```

### 3. 运行

```bash
# 监控 Java 日志
./supertail -f /var/log/app.log --format java

# 或通过管道
tail -f /var/log/app.log | ./supertail --format java
```

### 4. 查看效果

```
📋 批次摘要: 发现数据库连接错误 (重要日志: 2 条)

   │ 2025-10-13 INFO Connecting to database
⚠️  [重要] 2025-10-13 ERROR Connection timeout
⚠️  [重要] java.sql.SQLException: ...
   │    at com.example.dao...
   │ 2025-10-13 INFO Retry attempt
```

同时：
- 🔔 macOS 通知："发现数据库连接错误"
- 🔊 播放提示音

## 💡 为什么选择 SuperTail

### 传统方式的问题

```bash
tail -f app.log | grep ERROR
```

**问题：**
- ❌ 只能过滤固定关键词，不够智能
- ❌ 看不到完整的错误上下文
- ❌ 无法区分重要性（所有 ERROR 一视同仁）
- ❌ 没有通知提醒
- ❌ 多行异常堆栈被拆分

### SuperTail 的优势

```bash
./supertail -f app.log --format java
```

**优势：**
- ✅ AI 智能判断重要性（60+ 场景示例）
- ✅ 自动显示完整上下文（前后 3 行）
- ✅ 异常堆栈自动合并
- ✅ 系统通知 + 声音提醒
- ✅ 批量处理，节省费用
- ✅ 输出简洁，聚焦问题

## 🎯 核心功能详解

### 1. 批处理模式（默认）

**工作原理：**
- 累积 10 行或等待 3 秒后批量分析
- 一次 API 调用处理多行
- 减少通知频率

**效果对比：**

| 指标 | 逐行模式 | 批处理模式 | 提升 |
|------|---------|-----------|------|
| API 调用 | 100 次 | 10 次 | ↓ 90% |
| Token 消耗 | 64,500 | 10,500 | ↓ 83% |
| 通知次数 | 15 次 | 1-2 次 | ↓ 87% |
| 处理速度 | 60-90秒 | 6-9秒 | ↑ 10倍 |

### 2. 本地预过滤

**智能识别：**
- DEBUG、INFO、TRACE、VERBOSE 级别
- 健康检查、心跳、启动日志
- 正常的业务操作

**直接过滤：**
- 不调用 AI API
- 处理速度 < 0.1秒
- 节省 60-80% API 调用

**安全机制：**
- INFO 中包含 "error" 等关键词 → 仍调用 AI

### 3. 上下文显示

**问题场景：**
```
⚠️ [重要] ERROR Database failed
```
看不出是什么操作导致的错误。

**SuperTail 方案：**
```
   │ INFO User login request
   │ INFO Connecting to database
⚠️  [重要] ERROR Connection timeout
⚠️  [重要] java.sql.SQLException: ...
   │    at com.example.dao.UserDao...
   │ INFO Retry attempt 1
```

完整显示了：
- 用户登录请求
- 数据库连接
- 连接超时错误
- 异常堆栈
- 重试信息

### 4. 多行日志合并

**Java 异常堆栈：**
```
2025-10-13 ERROR Failed to process
java.lang.NullPointerException: Cannot invoke...
    at com.example.Service.process(Service.java:123)
    at com.example.Controller.handle(Controller.java:456)
```

**传统方式：** 4 行被拆分，逐行分析，失去上下文

**SuperTail：** 自动合并为 1 条完整日志，整体分析

## 📊 判断标准

### 会被过滤（不显示）

✅ **7 大类，20+ 示例：**
1. 健康检查和心跳
2. 应用启动和配置
3. 正常的业务操作（INFO/DEBUG）
4. 定时任务正常执行
5. 静态资源请求
6. 常规数据库操作
7. 正常的 API 响应（200 OK）

### 需要关注（显示 + 通知）

⚠️ **10 大类，40+ 示例：**
1. 错误和异常（ERROR、Exception）
2. 数据库问题（timeout、deadlock）
3. 认证和授权问题
4. 性能问题（慢查询、高内存）
5. 资源耗尽（OOM、磁盘满）
6. 外部服务失败
7. 业务异常
8. 安全问题（SQL注入、可疑活动）
9. 数据一致性问题
10. 服务降级和熔断

详见：[提示词示例](docs/PROMPT_EXAMPLES.md)

## 🔧 命令行参数

### 基础参数

```bash
--format string       # 日志格式（java, php, nginx, ruby, python, fastapi）
-f string             # 监控的日志文件路径
--context N           # 上下文行数（默认 3）
--show-not-important  # 显示被过滤的日志
--verbose             # 详细输出
--debug               # 调试模式
```

### 批处理参数

```bash
--batch-size N        # 批处理最大行数（默认 10）
--batch-wait 时间     # 批处理等待时间（默认 3s）
--no-batch            # 禁用批处理
```

## 📚 使用场景

### 场景 1: 生产环境监控

```bash
# 24/7 监控，断点续传，自动处理日志轮转
./supertail -f /var/log/production.log --format java --batch-size 20
```

- 重要错误立即通知
- 大批次节省费用
- 完整上下文方便排查

### 场景 2: 开发调试

```bash
# 实时监控，更多上下文，显示详细原因
./supertail -f dev.log --format java --context 5 --verbose
```

- 快速定位问题
- 查看过滤原因
- 完整的错误场景

### 场景 3: 历史日志分析

```bash
# 快速筛选重要事件
cat /var/log/old/*.log | ./supertail --format java --batch-size 50
```

- 从海量日志中提取关键信息
- 大批次高效处理
- 生成问题清单

### 场景 4: 多个服务监控

```bash
# 启动多个实例分别监控
./supertail -f /var/log/service1.log --format java &
./supertail -f /var/log/service2.log --format php &
./supertail -f /var/log/nginx/error.log --format nginx &
```

## 🧪 测试

```bash
# 运行所有测试
./tests/test-batch-processing.sh     # 批处理测试
./tests/test-context.sh              # 上下文测试
./tests/test-local-filter.sh         # 本地过滤测试
./tests/test-notification-quick.sh   # 通知设置向导

# 快速测试
./tests/test-batch-clean.sh          # 简洁输出测试
./tests/quick-batch-test.sh          # 快速批处理测试
```

## 🎓 进阶使用

### 自定义过滤规则

修改 `buildSystemPrompt` 函数中的示例，调整 AI 判断标准。

### 集成到监控系统

```bash
# systemd 服务（Linux）
[Unit]
Description=SuperTail Log Monitor
After=network.target

[Service]
Type=simple
User=youruser
ExecStart=/path/to/supertail -f /var/log/app.log --format java
Restart=always

[Install]
WantedBy=multi-user.target
```

### 性能调优

```bash
# 高频日志（100+ 条/分钟）
./supertail -f app.log --format java \
    --batch-size 30 \
    --batch-wait 2s \
    --context 2

# 低频日志（< 10 条/分钟）
./supertail -f app.log --format java \
    --batch-size 5 \
    --batch-wait 10s \
    --context 5
```

## 📖 完整文档

- [完整使用文档](docs/README_supertail.md)
- [批处理优化说明](docs/批处理优化说明.md)
- [本地预过滤优化](docs/本地预过滤优化.md)
- [保守过滤策略](docs/保守过滤策略.md)
- [通知设置指南](docs/NOTIFICATION_SETUP.md)
- [上下文显示说明](docs/)
- [多行日志合并](docs/)

## 🤝 贡献指南

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 💬 反馈

- 问题反馈：[GitHub Issues](https://github.com/your-username/supertail/issues)
- 功能建议：欢迎提 Issue
- 技术交流：m@some.im

## 🎖️ 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 👤 作者

**rocky** 
- Email: m@some.im
- GitHub: [@rocky](https://github.com/rocky)

## 🌟 Star History

如果这个项目对你有帮助，请给一个 ⭐ Star！

---

Made with ❤️ by rocky

