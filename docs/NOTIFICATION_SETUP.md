# macOS 通知权限设置指南

## 问题：为什么看不到通知？

如果运行 `echo '2025-10-13 ERROR Test' | ./aipipe --format java --verbose` 没有看到通知，最常见的原因是**通知权限未授予**。

## 🎯 快速设置（3 步）

### 步骤 1: 打开系统设置

**方法 A - 使用命令（推荐）：**
```bash
open "x-apple.systempreferences:com.apple.preference.notifications"
```

**方法 B - 手动打开：**
1. 点击屏幕左上角的  Apple 菜单
2. 选择「系统设置」(System Settings)
3. 左侧菜单选择「通知」(Notifications)

### 步骤 2: 找到终端应用

在右侧应用列表中，向下滚动找到：
- **「终端」** (如果使用 Terminal.app)
- **「iTerm」** (如果使用 iTerm2)
- 或你正在使用的其他终端应用

**技巧：** 在搜索框输入 "terminal" 或 "终端" 快速定位

### 步骤 3: 启用通知

点击终端应用后，确保以下设置：

✅ **必须开启的选项：**
1. **「允许通知」** - 必须打开（这是最关键的！）
2. **通知样式** - 选择「横幅」或「提醒」（不要选「无」）
3. **「在通知中心显示」** - 建议开启
4. **「播放通知声音」** - 建议开启（虽然我们用 afplay）

❌ **可选的选项：**
- 「在锁定屏幕上显示」- 可选
- 「显示预览」- 可选

## 📸 设置截图说明

### macOS Ventura 及以后版本

```
系统设置 > 通知 > 终端

┌─────────────────────────────────────────┐
│ 终端                                     │
├─────────────────────────────────────────┤
│                                          │
│ ☑ 允许通知                    [最重要！] │
│                                          │
│ 通知样式：                               │
│   ○ 无                                   │
│   ● 横幅              [推荐：横幅或提醒]  │
│   ○ 提醒                                 │
│                                          │
│ ☑ 在通知中心显示                         │
│ ☑ 播放通知声音                           │
│ ☐ 在锁定屏幕上显示                       │
│                                          │
└─────────────────────────────────────────┘
```

### macOS Monterey 及之前版本

```
系统偏好设置 > 通知与专注模式 > 终端

设置与上面类似，界面可能略有不同
```

## 🧪 验证设置

### 测试 1: 手动测试通知

```bash
osascript -e 'display notification "这是一个测试通知" with title "测试"'
```

**预期结果：**
- ✅ 应该在屏幕右上角看到通知横幅
- ✅ 如果没看到，说明权限未正确设置

### 测试 2: 测试 AIPipe（带 verbose）

```bash
echo '2025-10-13 ERROR Database connection failed' | ./aipipe --format java --verbose
```

**预期输出：**
```
🚀 AIPipe 启动 - 监控 java 格式日志
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📥 从标准输入读取日志...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  [重要] 2025-10-13 ERROR Database connection failed
   📝 摘要: 数据库连接失败
✅ 通知已发送                    ← 应该看到这行
🔊 播放声音: /System/Library/Sounds/Glass.aiff  ← 应该听到声音
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 统计: 总计 1 行, 过滤 0 行, 告警 1 次
```

**预期效果：**
- ✅ 应该看到通知横幅
- ✅ 应该听到 Glass.aiff 声音
- ✅ 终端输出「✅ 通知已发送」

### 测试 3: 测试声音播放

```bash
afplay /System/Library/Sounds/Glass.aiff
```

**预期结果：**
- ✅ 应该听到清脆的玻璃声
- ❌ 如果听不到，检查系统音量

## 🔧 常见问题排查

### 问题 1: 设置里找不到「终端」

**可能原因：**
- 终端从未请求过通知权限
- 需要先触发一次通知请求

**解决方法：**
```bash
# 手动触发一次通知
osascript -e 'display notification "测试" with title "测试"'

# 然后重新打开系统设置 > 通知
# 终端应该会出现在列表中
```

### 问题 2: 设置正确但还是看不到通知

**检查步骤：**

1. **检查勿扰模式/专注模式**
   ```bash
   # 查看是否开启了专注模式
   # 在菜单栏右上角查看
   # 如果看到月亮图标，说明勿扰模式已开启
   ```

2. **重启终端**
   ```bash
   # 完全退出终端应用
   # 然后重新打开
   ```

3. **重启通知中心**
   ```bash
   killall NotificationCenter
   # 系统会自动重启通知中心
   ```

4. **检查系统设置继承**
   - 打开「系统设置 > 通知」
   - 查看顶部是否有「专注模式」相关设置
   - 确保专注模式不会阻止终端通知

### 问题 3: 有通知但没有声音

**检查步骤：**

1. **检查系统音量**
   ```bash
   # 查看当前音量
   osascript -e "output volume of (get volume settings)"
   
   # 设置音量为 50%
   osascript -e "set volume output volume 50"
   ```

2. **测试 afplay**
   ```bash
   # 测试声音播放
   afplay /System/Library/Sounds/Glass.aiff
   
   # 如果听不到，可能是：
   # - 音量太低
   # - 静音
   # - 音频输出设备问题
   ```

3. **检查通知声音设置**
   - 在「系统设置 > 通知 > 终端」
   - 确保「播放通知声音」已勾选

### 问题 4: 使用 verbose 没有看到提示

**可能原因：**
- AI 判断该日志应该被过滤（不是错误）
- API 调用失败

**解决方法：**
```bash
# 使用明确的 ERROR 日志测试
echo '2025-10-13 10:00:00 ERROR Database connection failed' | ./aipipe --format java --verbose

# 使用 debug 模式查看完整过程
echo '2025-10-13 10:00:00 ERROR Database connection failed' | ./aipipe --format java --debug
```

## 📋 完整检查清单

在报告问题前，请确认以下所有项目：

- [ ] 系统设置 > 通知 > 终端 > **「允许通知」已开启**
- [ ] 通知样式选择了「横幅」或「提醒」（不是「无」）
- [ ] 没有开启勿扰模式/专注模式
- [ ] 手动测试通知命令可以看到通知：
  ```bash
  osascript -e 'display notification "测试" with title "测试"'
  ```
- [ ] 系统音量不是静音且大于 30%
- [ ] 手动测试声音播放可以听到：
  ```bash
  afplay /System/Library/Sounds/Glass.aiff
  ```
- [ ] 使用 ERROR 日志测试 AIPipe：
  ```bash
  echo '2025-10-13 ERROR Test' | ./aipipe --format java --verbose
  ```

## 🎬 快速设置脚本

运行以下命令自动打开通知设置：

```bash
# 打开通知设置
open "x-apple.systempreferences:com.apple.preference.notifications"

# 等待你设置好后，测试通知
sleep 5
osascript -e 'display notification "如果你看到这个，说明设置成功！" with title "✅ 测试成功"'

# 测试声音
afplay /System/Library/Sounds/Glass.aiff

# 测试 AIPipe
echo '2025-10-13 ERROR Database failed' | ./aipipe --format java --verbose
```

## 📱 不同终端应用的权限

如果你使用的不是默认的 Terminal.app，需要为对应的应用设置权限：

| 应用名称 | 在通知设置中的名称 |
|---------|-------------------|
| Terminal.app | 终端 / Terminal |
| iTerm2 | iTerm |
| Alacritty | Alacritty |
| kitty | kitty |
| Hyper | Hyper |
| WezTerm | WezTerm |

## 🔍 高级调试

### 查看通知中心日志

```bash
# 查看最近 1 分钟的通知中心日志
log show --predicate 'subsystem == "com.apple.notificationcenter"' --last 1m --info
```

### 查看 osascript 详细错误

```bash
# 运行 osascript 并查看错误
osascript -e 'display notification "测试" with title "测试"' 2>&1
```

### 强制重置通知权限

```bash
# 重置特定应用的通知权限（需要关闭应用）
# 注意：这会清除所有设置
tccutil reset Notifications com.apple.Terminal
```

## 💡 提示

1. **第一次使用时**，macOS 可能会弹出权限请求对话框，请选择「允许」
2. **如果修改了设置**，建议重启终端应用使其生效
3. **如果还是不行**，尝试重启 Mac
4. **最重要的是**「允许通知」必须开启！

## 📞 还是不行？

如果按照以上步骤还是无法显示通知，请提供以下信息：

1. macOS 版本：
   ```bash
   sw_vers
   ```

2. 使用的终端：Terminal.app / iTerm2 / 其他

3. 测试命令的输出：
   ```bash
   echo '2025-10-13 ERROR Test' | ./aipipe --format java --verbose
   ```

4. 手动通知测试结果：
   ```bash
   osascript -e 'display notification "测试" with title "测试"'
   ```

5. 系统设置截图：「系统设置 > 通知 > 终端」

---

**更新日期**: 2025-10-13  
**适用版本**: macOS 12.0 (Monterey) 及以上

