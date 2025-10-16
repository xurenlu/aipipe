# SuperTail 快速开始指南

## 🎯 1 分钟快速开始

```bash
# 1. 编译
go build -o supertail supertail.go

# 2. 运行
./supertail -f /var/log/app.log --format java

# 3. 享受智能过滤！
```

## 📋 使用前检查

### ✅ 必须项

- [ ] Go 1.21 或更高版本
- [ ] macOS 系统（通知功能）
- [ ] Azure OpenAI API 密钥

### ✅ 可选项

- [ ] 配置通知权限（系统设置 > 通知 > 终端）
- [ ] 调整系统音量（建议 40-60%）

## 🔧 配置 API

### 方式 1: 直接修改代码

编辑 `supertail.go`：

```go
const (
    AZURE_API_ENDPOINT = "https://your-resource.openai.azure.com/openai/deployments/your-model/chat/completions?api-version=2025-01-01-preview"
    AZURE_API_KEY      = "your-api-key-here"
    AZURE_MODEL        = "gpt-5-mini"
)
```

### 方式 2: 使用环境变量（计划支持）

```bash
export AZURE_OPENAI_ENDPOINT="..."
export AZURE_OPENAI_KEY="..."
```

## 🎬 使用示例

### 示例 1: 基本监控

```bash
./supertail -f /var/log/app.log --format java
```

**效果：**
- 只显示重要日志（ERROR、异常、WARN）
- INFO/DEBUG 自动过滤
- 有错误时通知 + 声音

### 示例 2: 查看上下文

```bash
./supertail -f /var/log/app.log --format java --context 5
```

**效果：**
- 重要日志前后各显示 5 行
- 完整的错误场景
- 方便排查问题

### 示例 3: 调试模式

```bash
./supertail -f /var/log/app.log --format java --show-not-important --verbose
```

**效果：**
- 显示所有日志（包括过滤的）
- 显示过滤原因
- 查看 AI 判断依据

### 示例 4: 高频日志

```bash
./supertail -f /var/log/app.log --format java --batch-size 30 --batch-wait 5s
```

**效果：**
- 大批次处理
- 减少 API 调用
- 节省费用

## 🎓 进阶配置

### 根据日志频率调整

| 日志频率 | 推荐配置 |
|---------|---------|
| 低（< 10/分钟） | `--batch-size 5 --batch-wait 5s` |
| 中（10-50/分钟） | `--batch-size 10 --batch-wait 3s` （默认） |
| 高（50-100/分钟） | `--batch-size 20 --batch-wait 2s` |
| 极高（> 100/分钟） | `--batch-size 30-50 --batch-wait 1s` |

### 根据使用场景调整

**生产监控：**
```bash
./supertail -f /var/log/prod.log --format java \
    --batch-size 20 \
    --batch-wait 5s \
    --context 3
```

**开发调试：**
```bash
./supertail -f /var/log/dev.log --format java \
    --batch-size 5 \
    --batch-wait 1s \
    --context 5 \
    --verbose
```

**历史分析：**
```bash
cat /var/log/old.log | ./supertail --format java \
    --batch-size 50 \
    --batch-wait 10s
```

## 🔍 常见问题

### Q: 如何设置通知权限？

```bash
# 1. 打开系统设置
open "x-apple.systempreferences:com.apple.preference.notifications"

# 2. 找到「终端」
# 3. 开启「允许通知」
# 4. 选择「横幅」样式

# 详细步骤见：
cat docs/NOTIFICATION_SETUP.md
```

### Q: 没有声音怎么办？

```bash
# 测试声音
afplay /System/Library/Sounds/Glass.aiff

# 检查音量
osascript -e "output volume of (get volume settings)"

# 设置音量
osascript -e "set volume output volume 50"
```

### Q: 如何验证批处理是否工作？

```bash
./tests/quick-batch-test.sh

# 应该看到：
# 📦 批次 #1: 处理 N 行日志
# 📋 批次摘要: ...
```

### Q: 如何查看被过滤了什么？

```bash
./supertail -f app.log --format java --show-not-important
```

## 📚 学习资源

### 入门教程

1. **5 分钟入门**（本文件）
2. [完整使用文档](docs/README_supertail.md)
3. [批处理优化说明](docs/批处理优化说明.md)

### 深入了解

1. [本地预过滤优化](docs/本地预过滤优化.md)
2. [保守过滤策略](docs/保守过滤策略.md)
3. [提示词示例说明](docs/PROMPT_EXAMPLES.md)

### 问题排查

1. [通知设置指南](docs/NOTIFICATION_SETUP.md)
2. [声音播放指南](docs/NOTIFICATION_SOUND_GUIDE.md)
3. [中文乱码问题](docs/中文乱码问题解决.md)

## 🎉 下一步

1. **运行测试** - 验证功能
   ```bash
   ./examples/supertail-example.sh
   ```

2. **监控日志** - 开始使用
   ```bash
   ./supertail -f /var/log/your-app.log --format java
   ```

3. **查看文档** - 了解更多
   ```bash
   cat docs/README_supertail.md
   ```

4. **反馈问题** - 帮助改进
   - GitHub Issues
   - Email: m@some.im

---

祝使用愉快！🎊

