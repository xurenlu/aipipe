# AIPipe 系统级日志监控示例

## 概述

AIPipe 现在支持监控 Linux 和 macOS 系统的核心日志，包括 systemd journal、macOS Console 统一日志系统和传统 syslog。这些系统级日志包含了操作系统、内核、服务等关键组件的运行状态，是系统监控和故障排查的重要数据源。

---

## 🐧 Linux 系统日志监控

### 1. systemd journal (journald)

systemd journal 是现代 Linux 发行版的标准日志系统，提供了结构化、索引化的日志存储。

#### 基本监控
```bash
# 监控所有系统日志
journalctl -f | ./aipipe --format journald

# 监控特定服务
journalctl -u nginx -f | ./aipipe --format journald
journalctl -u docker -f | ./aipipe --format journald

# 监控内核消息
journalctl -k -f | ./aipipe --format journald

# 监控特定优先级
journalctl -p err -f | ./aipipe --format journald
```

#### 高级过滤
```bash
# 监控特定时间范围的日志
journalctl --since "1 hour ago" -f | ./aipipe --format journald

# 监控特定用户/进程
journalctl _UID=1000 -f | ./aipipe --format journald

# 监控特定子系统
journalctl -u systemd-* -f | ./aipipe --format journald

# 监控网络相关日志
journalctl -u NetworkManager -f | ./aipipe --format journald
```

#### 实际示例
```bash
# 监控 Web 服务器
journalctl -u apache2 -f | ./aipipe --format journald --verbose

# 监控数据库服务
journalctl -u postgresql -f | ./aipipe --format journald

# 监控容器运行时
journalctl -u containerd -f | ./aipipe --format journald
```

### 2. 传统 Syslog

对于使用传统 syslog 的系统或特定日志文件：

```bash
# 监控主 syslog 文件
tail -f /var/log/syslog | ./aipipe --format syslog

# 监控 auth 日志（安全相关）
tail -f /var/log/auth.log | ./aipipe --format syslog

# 监控内核日志
tail -f /var/log/kern.log | ./aipipe --format syslog

# 监控邮件日志
tail -f /var/log/mail.log | ./aipipe --format syslog
```

---

## 🍎 macOS 系统日志监控

### macOS Console 统一日志系统

macOS 使用统一日志系统（Unified Logging System）来管理所有系统和应用日志。

#### 基本监控
```bash
# 监控所有系统日志
log stream | ./aipipe --format macos-console

# 监控错误日志
log stream --predicate 'eventType == "errorEvent"' | ./aipipe --format macos-console

# 监控特定进程
log stream --process kernel | ./aipipe --format macos-console
log stream --process systemd | ./aipipe --format macos-console
```

#### 高级过滤
```bash
# 监控特定子系统
log stream --predicate 'subsystem == "com.apple.TCC"' | ./aipipe --format macos-console
log stream --predicate 'subsystem == "com.apple.security"' | ./aipipe --format macos-console

# 监控特定级别
log stream --level debug | ./aipipe --format macos-console

# 监控特定用户
log stream --user 501 | ./aipipe --format macos-console

# 监控特定活动
log stream --predicate 'activity == "0x12345"' | ./aipipe --format macos-console
```

#### 实际示例
```bash
# 监控系统启动问题
log stream --predicate 'process == "kernel"' | ./aipipe --format macos-console

# 监控权限问题
log stream --predicate 'subsystem CONTAINS "TCC"' | ./aipipe --format macos-console

# 监控网络问题
log stream --predicate 'process == "networkd"' | ./aipipe --format macos-console

# 监控存储问题
log stream --predicate 'process == "diskmanagementd"' | ./aipipe --format macos-console
```

---

## 🔍 监控策略建议

### 1. 分层监控

```bash
# 系统级监控（后台运行）
journalctl -f | ./aipipe --format journald --batch-size 20 > /var/log/system-monitor.log 2>&1 &

# 应用级监控
tail -f /var/log/nginx/error.log | ./aipipe --format nginx --verbose &

# 数据库监控
tail -f /var/log/postgresql/postgresql.log | ./aipipe --format postgresql &
```

### 2. 关键服务监控

```bash
# 监控关键系统服务
for service in nginx postgresql redis docker; do
    journalctl -u $service -f | ./aipipe --format journald &
done

# 监控安全相关日志
tail -f /var/log/auth.log | ./aipipe --format syslog &
journalctl -u sshd -f | ./aipipe --format journald &
```

### 3. 性能监控

```bash
# 监控系统资源使用
journalctl -u systemd-oomd -f | ./aipipe --format journald

# 监控磁盘空间
journalctl -u systemd-logind -f | ./aipipe --format journald

# 监控内存使用
log stream --predicate 'process == "kernel" AND message CONTAINS "memory"' | ./aipipe --format macos-console
```

---

## 🚨 重要日志类型识别

### Linux 系统重要日志

#### 系统错误
- **OOM (Out of Memory)**: `Out of memory: Kill process`
- **硬件错误**: `Hardware Error`, `Machine Check Exception`
- **文件系统错误**: `EXT4-fs error`, `XFS error`
- **网络错误**: `NetworkManager`, `sshd`

#### 安全事件
- **登录失败**: `Failed password for`
- **权限提升**: `sudo`, `su`
- **防火墙**: `iptables`, `ufw`

### macOS 系统重要日志

#### 系统错误
- **内核错误**: `kernel: ERROR`, `kernel: PANIC`
- **内存错误**: `memory pressure`, `out of memory`
- **磁盘错误**: `diskmanagementd`, `fsck`

#### 安全事件
- **权限问题**: `TCC`, `access denied`
- **应用沙盒**: `sandbox`, `security`
- **网络问题**: `networkd`, `firewall`

---

## 📊 监控最佳实践

### 1. 日志轮转处理
```bash
# 使用 logrotate 管理日志文件
# /etc/logrotate.d/aipipe-monitor
/var/log/system-monitor.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

### 2. 性能优化
```bash
# 使用批处理减少 API 调用
journalctl -f | ./aipipe --format journald --batch-size 30 --batch-wait 5s

# 过滤低级别日志
log stream --level info | ./aipipe --format macos-console

# 使用本地预过滤
journalctl -p err -f | ./aipipe --format journald
```

### 3. 告警集成
```bash
# 集成到监控系统
journalctl -f | ./aipipe --format journald | tee -a /var/log/alerts.log

# 发送到外部系统
log stream | ./aipipe --format macos-console | curl -X POST -d @- http://monitoring.example.com/alerts
```

---

## 🛠️ 故障排查示例

### 系统启动问题
```bash
# 监控启动日志
journalctl -b -f | ./aipipe --format journald
log show --last boot | ./aipipe --format macos-console
```

### 网络问题
```bash
# 监控网络服务
journalctl -u NetworkManager -f | ./aipipe --format journald
log stream --predicate 'process == "networkd"' | ./aipipe --format macos-console
```

### 存储问题
```bash
# 监控存储服务
journalctl -u systemd-logind -f | ./aipipe --format journald
log stream --predicate 'process == "diskmanagementd"' | ./aipipe --format macos-console
```

---

## 📚 相关文档

- [SUPPORTED_FORMATS.md](SUPPORTED_FORMATS.md) - 完整的格式支持说明
- [README_aipipe.md](README_aipipe.md) - 主要使用文档
- [aipipe-quickstart.md](aipipe-quickstart.md) - 快速入门指南

---

**作者**: rocky  
**版本**: v1.2.0  
**日期**: 2025-10-17
