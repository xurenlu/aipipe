# 通知声音播放指南

## 声音播放机制

SuperTail 使用了多层次的声音播放策略，确保重要日志告警时一定有声音提醒。

### 播放策略

#### 第一层：通知声音（Notification Sound）

尝试 5 种系统通知声音，按优先级：
1. **Glass** - 清脆的玻璃声（推荐）
2. **Ping** - 短促的提示音
3. **Pop** - 气泡爆裂声
4. **Purr** - 柔和的声音
5. **Bottle** - 瓶子声

通过 `osascript` 在通知中指定声音：
```bash
osascript -e 'display notification "消息" with title "标题" sound name "Glass"'
```

#### 第二层：直接播放音频文件（afplay）

如果通知声音失败，使用 `afplay` 直接播放系统声音文件：
```bash
afplay /System/Library/Sounds/Glass.aiff
```

按优先级尝试：
1. `/System/Library/Sounds/Glass.aiff`
2. `/System/Library/Sounds/Ping.aiff`
3. `/System/Library/Sounds/Pop.aiff`
4. `/System/Library/Sounds/Purr.aiff`
5. `/System/Library/Sounds/Bottle.aiff`
6. `/System/Library/Sounds/Funk.aiff`

#### 第三层：系统蜂鸣声（beep）

如果前两种方式都失败，播放系统蜂鸣声：
```bash
osascript -e "beep 1"
```

### 实现代码

```go
// 发送通知（带多种声音尝试）
func sendNotification(summary, logLine string) {
    // 1. 尝试通知声音
    soundNames := []string{"Glass", "Ping", "Pop", "Purr", "Bottle"}
    for _, soundName := range soundNames {
        script := fmt.Sprintf(`display notification "%s" with title "⚠️ 重要日志告警" subtitle "%s" sound name "%s"`, ...)
        if cmd.Run() == nil {
            go playSystemSound()  // 额外播放确保有声音
            return
        }
    }
    
    // 2. 发送无声通知 + 播放系统音
    go playSystemSound()
}

// 播放系统音（备用方案）
func playSystemSound() {
    // 尝试 afplay
    soundPaths := []string{"/System/Library/Sounds/Glass.aiff", ...}
    for _, path := range soundPaths {
        if exec.Command("afplay", path).Run() == nil {
            return
        }
    }
    
    // 最后尝试 beep
    exec.Command("osascript", "-e", "beep 1").Run()
}
```

## 为什么可能没有声音

### 1. macOS 通知权限未授予

**症状：** 没有通知，也没有声音

**解决方法：**
```bash
# 打开系统设置
open "x-apple.systempreferences:com.apple.preference.notifications"
```

然后：
1. 找到「终端」或「Terminal」
2. 确保启用了「允许通知」
3. 确保启用了「播放通知声音」
4. 选择「横幅」样式（不要选「无」）

### 2. 通知声音被禁用

**症状：** 有通知显示，但没有声音

**检查步骤：**
1. 打开「系统设置 > 通知」
2. 找到「终端」
3. 确保「播放通知声音」已勾选

### 3. 系统音量太低或静音

**症状：** 看起来一切正常但听不到声音

**解决方法：**
```bash
# 检查当前音量
osascript -e "output volume of (get volume settings)"

# 设置音量为 50%
osascript -e "set volume output volume 50"

# 取消静音
osascript -e "set volume with output muted false"
```

### 4. 勿扰模式（专注模式）已开启

**症状：** 有时有声音，有时没有

**解决方法：**
- 关闭勿扰模式/专注模式
- 或在专注模式设置中允许终端的通知

### 5. 声音文件不存在

**症状：** 第一层和第二层都失败，只能听到 beep

**检查方法：**
```bash
# 查看可用的系统声音
ls -l /System/Library/Sounds/

# 手动播放测试
afplay /System/Library/Sounds/Glass.aiff
```

### 6. 终端权限问题

**症状：** 命令执行但无效果

**解决方法：**
- 重启终端
- 给予终端完全磁盘访问权限（在「安全性与隐私」中）

## 测试方法

### 方法 1：运行完整测试脚本

```bash
./test-notification-sound.sh
```

这会测试：
- osascript 通知声音
- afplay 音频播放
- beep 蜂鸣声
- SuperTail 实际告警

### 方法 2：手动测试通知

```bash
# 测试通知 + Glass 声音
osascript -e 'display notification "测试消息" with title "测试" sound name "Glass"'

# 测试通知 + Ping 声音
osascript -e 'display notification "测试消息" with title "测试" sound name "Ping"'

# 测试无声通知
osascript -e 'display notification "测试消息" with title "测试"'
```

### 方法 3：手动测试音频播放

```bash
# 播放 Glass 声音
afplay /System/Library/Sounds/Glass.aiff

# 播放 Ping 声音
afplay /System/Library/Sounds/Ping.aiff

# 列出所有可用声音
ls /System/Library/Sounds/
```

### 方法 4：手动测试 beep

```bash
# 播放一次蜂鸣声
osascript -e "beep 1"

# 播放三次蜂鸣声
osascript -e "beep 3"
```

### 方法 5：使用 SuperTail 测试

```bash
# 触发一个告警（ERROR 日志）
echo "2025-10-13 10:00:00 ERROR Database connection failed" | ./supertail --format java

# 详细模式查看错误
echo "2025-10-13 10:00:00 ERROR Test" | ./supertail --format java --verbose
```

## 可用的系统声音

### 常见声音文件

在 `/System/Library/Sounds/` 目录下：

| 声音名称 | 文件名 | 描述 |
|---------|--------|------|
| Basso | Basso.aiff | 低沉的声音 |
| Blow | Blow.aiff | 吹气声 |
| Bottle | Bottle.aiff | 瓶子声 |
| Frog | Frog.aiff | 青蛙叫声 |
| Funk | Funk.aiff | 节奏声 |
| Glass | Glass.aiff | 玻璃声（推荐）⭐ |
| Hero | Hero.aiff | 英雄音效 |
| Morse | Morse.aiff | 摩斯电码 |
| Ping | Ping.aiff | 短促提示音⭐ |
| Pop | Pop.aiff | 气泡声⭐ |
| Purr | Purr.aiff | 柔和声 |
| Sosumi | Sosumi.aiff | 经典 Mac 声音 |
| Submarine | Submarine.aiff | 潜水艇声 |
| Tink | Tink.aiff | 轻敲声 |

### 查看所有声音

```bash
ls -1 /System/Library/Sounds/
```

### 试听所有声音

```bash
for sound in /System/Library/Sounds/*.aiff; do
    echo "播放: $(basename $sound)"
    afplay "$sound"
    sleep 1
done
```

## 调试技巧

### 1. 检查通知是否发送

```bash
# 查看系统日志中的通知
log show --predicate 'subsystem == "com.apple.notificationcenter"' --last 1m
```

### 2. 查看 osascript 错误

```bash
# 详细模式运行
osascript -l JavaScript -e 'displayNotification("test", {withTitle: "Test", soundName: "Glass"})' 2>&1
```

### 3. 测试音频播放能力

```bash
# 测试 afplay 是否工作
afplay /System/Library/Sounds/Glass.aiff && echo "✅ 成功" || echo "❌ 失败"
```

### 4. 查看 SuperTail 的详细输出

```bash
# 使用 verbose 模式
echo "2025-10-13 ERROR Test" | ./supertail --format java --verbose

# 使用 debug 模式（查看 API 调用）
echo "2025-10-13 ERROR Test" | ./supertail --format java --debug
```

## 常见问题解答

### Q: 为什么有通知但没有声音？

A: 最可能的原因：
1. 系统通知声音被禁用
2. 音量太低或静音
3. 勿扰模式已开启

**解决方法：** 运行 `./test-notification-sound.sh` 进行全面检查

### Q: 可以自定义声音吗？

A: 可以！修改 `supertail.go` 中的声音列表：

```go
soundNames := []string{"Glass", "Ping", "Pop", "Purr", "Bottle"}
// 改为你喜欢的声音
soundNames := []string{"Hero", "Funk", "Sosumi"}
```

然后重新编译：
```bash
go build -o supertail supertail.go
```

### Q: 声音太大/太小怎么办？

A: 调整系统音量：
```bash
# 设置为 30%（较轻）
osascript -e "set volume output volume 30"

# 设置为 70%（较大）
osascript -e "set volume output volume 70"
```

### Q: 能否禁用声音但保留通知？

A: 可以！修改代码移除声音相关调用，或者在系统设置中禁用终端的通知声音。

### Q: 为什么使用三层策略？

A: 为了确保可靠性：
- 第一层（通知声音）：最标准，最美观
- 第二层（afplay）：更可靠，直接播放
- 第三层（beep）：最简单，一定有声音

### Q: 在 SSH 远程会话中会有声音吗？

A: 不会。这些命令需要在本地 macOS 系统上运行。SSH 远程会话无法播放本地声音。

## 推荐设置

### 最佳体验设置

1. **通知权限**：✅ 允许终端通知
2. **通知样式**：横幅（会自动消失）
3. **播放声音**：✅ 已启用
4. **系统音量**：40-60%
5. **推荐声音**：Glass 或 Ping

### 测试命令

```bash
# 完整测试
./test-notification-sound.sh

# 快速测试
echo "2025-10-13 ERROR Database failed" | ./supertail --format java
```

## 相关文件

- `supertail.go` - 实现代码
  - `sendNotification()` - 发送通知函数
  - `playSystemSound()` - 播放声音函数
- `test-notification-sound.sh` - 完整测试脚本
- `README_supertail.md` - 使用文档

---

**最后更新**: 2025-10-13  
**版本**: 1.2.0 - 增强声音播放机制

