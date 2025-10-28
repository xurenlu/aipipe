# 13. 故障排除

> 常见问题诊断和解决方案

## 🎯 概述

本指南提供 AIPipe 常见问题的诊断和解决方案，帮助快速解决使用过程中遇到的问题。

## 🔧 安装问题

### 问题1: 编译失败

**症状**: `go build` 失败

**解决方案**:
```bash
# 检查 Go 版本
go version

# 更新 Go 版本
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 设置环境变量
export PATH=$PATH:/usr/local/go/bin

# 重新编译
go build -o aipipe .
```

### 问题2: 依赖问题

**症状**: 模块依赖错误

**解决方案**:
```bash
# 清理模块缓存
go clean -modcache

# 重新下载依赖
go mod download

# 更新依赖
go get -u all
```

## ⚙️ 配置问题

### 问题1: 配置文件错误

**症状**: `aipipe config validate` 失败

**解决方案**:
```bash
# 检查配置文件语法
aipipe config validate --verbose

# 重置配置文件
rm ~/.aipipe/config.json
aipipe config init

# 检查文件权限
ls -la ~/.aipipe/config.json
chmod 600 ~/.aipipe/config.json
```

### 问题2: API 密钥问题

**症状**: AI 服务连接失败

**解决方案**:
```bash
# 检查 API 密钥
echo $OPENAI_API_KEY

# 测试 API 连接
curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models

# 更新 API 密钥
# 编辑配置文件 ~/.aipipe/config.json
# 修改 "ai_api_key" 字段的值
```

## 🔍 分析问题

### 问题1: 分析结果不准确

**症状**: AI 分析结果不符合预期

**解决方案**:
```bash
# 检查日志格式
aipipe analyze --format java --verbose

# 使用自定义提示词
aipipe analyze --format java --prompt-file prompts/custom.txt

# 调整分析参数
# 编辑配置文件 ~/.aipipe/config.json
# 修改 "ai_model" 字段的值为 "gpt-4"
```

### 问题2: 分析速度慢

**症状**: 分析响应时间过长

**解决方案**:
```bash
# 启用本地过滤
# 编辑配置文件 ~/.aipipe/config.json
# 设置 "local_filter": true

# 启用缓存
# 编辑配置文件 ~/.aipipe/config.json
# 设置 "cache": {"enabled": true}

# 调整超时设置
# 编辑配置文件 ~/.aipipe/config.json
# 设置 "timeout": 30
```

## 📁 监控问题

### 问题1: 文件监控失败

**症状**: 监控文件无法启动

**解决方案**:
```bash
# 检查文件权限
ls -la /var/log/app.log
sudo chmod 644 /var/log/app.log

# 检查文件是否被占用
lsof /var/log/app.log

# 测试文件监控
aipipe monitor --file /var/log/app.log --format java --verbose
```

### 问题2: 监控配置丢失

**症状**: 添加的监控文件消失

**解决方案**:
```bash
# 检查监控配置文件
ls -la ~/.aipipe-monitor.json

# 重新添加监控文件
aipipe dashboard add

# 检查配置权限
chmod 600 ~/.aipipe-monitor.json
```

## 🔔 通知问题

### 问题1: 邮件发送失败

**症状**: 邮件通知无法发送

**解决方案**:
```bash
# 测试邮件配置
aipipe notify test --email --verbose

# 检查 SMTP 设置
aipipe config show --key "notifications.email"

# 测试网络连接
telnet smtp.gmail.com 587
```

### 问题2: 系统通知不显示

**症状**: 系统通知不出现

**解决方案**:
```bash
# 测试系统通知
aipipe notify test --system --verbose

# 检查通知权限
# macOS: 系统偏好设置 > 通知
# Linux: 检查 notify-send 命令

# 安装通知工具
sudo apt-get install libnotify-bin
```

## 🚀 性能问题

### 问题1: 内存使用过高

**症状**: 系统内存不足

**解决方案**:
```bash
# 检查内存使用
ps aux | grep aipipe
free -h

# 调整内存配置
# 编辑配置文件 ~/.aipipe/config.json
# 设置内存和缓存限制

# 重启服务
systemctl restart aipipe
```

### 问题2: CPU 使用过高

**症状**: CPU 使用率过高

**解决方案**:
```bash
# 检查 CPU 使用
top -p $(pgrep aipipe)

# 调整并发配置
# 编辑配置文件 ~/.aipipe/config.json
# 设置并发和批处理参数
```

## 🔄 服务问题

### 问题1: 服务无法启动

**症状**: systemd 服务启动失败

**解决方案**:
```bash
# 检查服务状态
systemctl status aipipe

# 查看详细日志
journalctl -u aipipe -n 50

# 检查配置文件
aipipe config validate

# 手动启动测试
sudo -u aipipe aipipe monitor
```

### 问题2: 服务频繁重启

**症状**: 服务不断重启

**解决方案**:
```bash
# 检查重启原因
journalctl -u aipipe --since "1 hour ago"

# 调整重启策略
sudo systemctl edit aipipe
# 添加:
# [Service]
# RestartSec=10
# StartLimitInterval=60
# StartLimitBurst=3

# 重新加载配置
sudo systemctl daemon-reload
```

## 🐳 Docker 问题

### 问题1: 容器无法启动

**症状**: Docker 容器启动失败

**解决方案**:
```bash
# 检查容器日志
docker logs aipipe

# 检查镜像
docker images | grep aipipe

# 重新拉取镜像
docker pull xurenlu/aipipe:latest

# 重新运行容器
docker run -d --name aipipe xurenlu/aipipe:latest
```

### 问题2: 容器内文件权限问题

**症状**: 容器内无法访问文件

**解决方案**:
```bash
# 检查文件权限
docker exec aipipe ls -la /var/log/

# 修改文件权限
docker exec aipipe chmod 644 /var/log/app.log

# 使用正确的用户运行
docker run -u root xurenlu/aipipe:latest
```

## 🔍 调试技巧

### 1. 启用调试模式

```bash
# 设置调试环境变量
export AIPIPE_DEBUG=true
export AIPIPE_LOG_LEVEL=debug

# 重新启动服务
systemctl restart aipipe

# 查看调试日志
journalctl -u aipipe -f
```

### 2. 详细输出

```bash
# 使用详细模式
aipipe analyze --verbose
aipipe monitor --verbose
aipipe config show --verbose
```

### 3. 日志分析

```bash
# 查看错误日志
grep ERROR /var/log/aipipe/aipipe.log

# 查看警告日志
grep WARN /var/log/aipipe/aipipe.log

# 查看特定时间日志
journalctl -u aipipe --since "2024-01-01 10:00:00"
```

## 📋 常见错误码

### 错误码列表

| 错误码 | 描述 | 解决方案 |
|--------|------|----------|
| 1001 | 配置文件错误 | 检查配置文件语法和权限 |
| 1002 | API 密钥无效 | 验证 API 密钥和权限 |
| 1003 | 网络连接失败 | 检查网络连接和防火墙 |
| 1004 | 文件权限错误 | 检查文件权限和所有者 |
| 1005 | 内存不足 | 调整内存配置或增加内存 |
| 1006 | 服务启动失败 | 检查配置和依赖 |

### 错误处理

```bash
# 查看错误详情
aipipe analyze --format java 2>&1 | grep "ERROR"

# 检查错误日志
tail -f /var/log/aipipe/aipipe.log | grep "ERROR"

# 重置错误状态
aipipe config reset
```

## 🎯 预防措施

### 1. 定期检查

```bash
# 创建健康检查脚本
cat > health-check.sh << EOF
#!/bin/bash
# AIPipe 健康检查

# 检查服务状态
if ! systemctl is-active --quiet aipipe; then
    echo "服务未运行"
    exit 1
fi

# 检查配置
if ! aipipe config validate; then
    echo "配置错误"
    exit 1
fi

# 检查 AI 服务
if ! aipipe ai test; then
    echo "AI 服务异常"
    exit 1
fi

echo "健康检查通过"
EOF

chmod +x health-check.sh
```

### 2. 监控告警

```bash
# 设置监控告警
cat > monitor.sh << EOF
#!/bin/bash
# AIPipe 监控脚本

# 检查内存使用
MEMORY_USAGE=$(ps aux | grep aipipe | awk '{sum+=$6} END {print sum/1024}')
if (( $(echo "$MEMORY_USAGE > 1000" | bc -l) )); then
    echo "内存使用过高: ${MEMORY_USAGE}MB"
    # 发送告警
fi

# 检查磁盘使用
DISK_USAGE=$(df /var/lib/aipipe | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "磁盘使用过高: ${DISK_USAGE}%"
    # 发送告警
fi
EOF

chmod +x monitor.sh
```

## 🎉 总结

AIPipe 的故障排除提供了：

- **问题诊断**: 系统性的问题诊断流程
- **解决方案**: 针对性的解决方案
- **调试技巧**: 实用的调试方法和工具
- **预防措施**: 主动的监控和预防策略
- **错误处理**: 完整的错误码和处理方法

---

*继续阅读: [14. 架构设计](14-architecture.md)*
