# 20. 常见问题

> 常见问题解答和故障排除指南

## 🎯 概述

本章节收集了 AIPipe 使用过程中的常见问题和解决方案。

## ❓ 安装和配置问题

### Q1: 如何安装 AIPipe？

**A**: 有多种安装方式：

```bash
# 方式一：从源码编译
git clone https://github.com/xurenlu/aipipe.git
cd aipipe
go build -o aipipe .

# 方式二：下载预编译二进制
wget https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-linux-amd64
chmod +x aipipe-linux-amd64
sudo mv aipipe-linux-amd64 /usr/local/bin/aipipe

# 方式三：使用 Docker
docker pull xurenlu/aipipe:latest
```

### Q2: 如何配置 AI API 密钥？

**A**: 通过配置文件设置：

```bash
# 初始化配置
aipipe config init

# 编辑配置文件
nano ~/.aipipe/config.json

# 添加 API 密钥
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-api-key",
  "ai_model": "gpt-3.5-turbo"
}
```

### Q3: 支持哪些 AI 服务？

**A**: 支持多种 AI 服务：

- **OpenAI**: GPT-3.5, GPT-4
- **Azure OpenAI**: GPT-3.5, GPT-4
- **自定义 API**: 兼容 OpenAI 格式的 API

### Q4: 如何验证配置是否正确？

**A**: 使用配置验证命令：

```bash
# 验证配置文件
aipipe config validate

# 测试 AI 服务
aipipe ai test

# 测试通知
aipipe notify test
```

## 🔍 日志分析问题

### Q5: 如何选择合适的日志格式？

**A**: 根据日志来源选择：

- **Java 应用**: `--format java`
- **Python 应用**: `--format python`
- **Nginx 日志**: `--format nginx`
- **Docker 日志**: `--format docker`
- **JSON 日志**: `--format json`
- **不确定格式**: `--format auto`

### Q6: 为什么分析结果不准确？

**A**: 可能的原因和解决方案：

1. **格式选择错误**: 选择正确的日志格式
2. **提示词不合适**: 使用自定义提示词
3. **API 限制**: 检查 API 使用量和限制
4. **网络问题**: 检查网络连接

```bash
# 使用自定义提示词
aipipe analyze --format java --prompt-file prompts/custom.txt

# 启用详细输出
aipipe analyze --verbose
```

### Q7: 如何提高分析性能？

**A**: 优化配置：

```json
{
  "local_filter": true,
  "cache": {
    "enabled": true,
    "ttl": 3600
  },
  "batch_processing": {
    "enabled": true,
    "batch_size": 10
  }
}
```

### Q8: 如何自定义分析规则？

**A**: 使用规则引擎：

```bash
# 添加过滤规则
aipipe rules add --pattern "DEBUG" --action "ignore"
aipipe rules add --pattern "ERROR" --action "alert"

# 测试规则
aipipe rules test --pattern "ERROR Database connection failed"
```

## 📁 文件监控问题

### Q9: 如何监控多个文件？

**A**: 使用配置文件管理：

```bash
# 添加监控文件
aipipe dashboard add

# 启动监控
aipipe monitor

# 查看监控状态
aipipe dashboard show
```

### Q10: 文件监控失败怎么办？

**A**: 检查以下问题：

1. **文件权限**: 确保有读取权限
2. **文件存在**: 检查文件是否存在
3. **文件被占用**: 检查是否有其他进程占用
4. **磁盘空间**: 确保有足够的磁盘空间

```bash
# 检查文件权限
ls -la /var/log/app.log

# 检查文件是否被占用
lsof /var/log/app.log

# 测试文件监控
aipipe monitor --file /var/log/app.log --format java --verbose
```

### Q11: 如何处理日志文件轮转？

**A**: AIPipe 自动处理文件轮转：

```bash
# 监控轮转的日志文件
aipipe monitor --file /var/log/app.log --format java

# 当文件轮转时，自动切换到新文件
# app.log -> app.log.1 -> app.log.2.gz
```

### Q12: 如何设置监控优先级？

**A**: 在添加监控文件时设置：

```bash
# 添加高优先级文件
aipipe dashboard add
# 输入: /var/log/system.log, syslog, 1

# 添加低优先级文件
aipipe dashboard add
# 输入: /var/log/debug.log, java, 40
```

## 🔔 通知问题

### Q13: 如何配置邮件通知？

**A**: 配置 SMTP 设置：

```json
{
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "to": "admin@example.com"
    }
  }
}
```

### Q14: 邮件发送失败怎么办？

**A**: 检查以下问题：

1. **SMTP 配置**: 检查 SMTP 服务器和端口
2. **认证信息**: 检查用户名和密码
3. **网络连接**: 检查网络连接
4. **防火墙**: 检查防火墙设置

```bash
# 测试邮件配置
aipipe notify test --email --verbose

# 检查网络连接
telnet smtp.gmail.com 587
```

### Q15: 如何配置系统通知？

**A**: 配置系统通知：

```json
{
  "notifications": {
    "system": {
      "enabled": true,
      "sound": true,
      "title": "AIPipe 告警"
    }
  }
}
```

### Q16: 系统通知不显示怎么办？

**A**: 检查系统设置：

```bash
# 测试系统通知
aipipe notify test --system --verbose

# 检查通知权限
# macOS: 系统偏好设置 > 通知
# Linux: 检查 notify-send 命令
```

## 🔧 性能问题

### Q17: 如何优化性能？

**A**: 多种优化方式：

1. **启用缓存**: 减少重复分析
2. **批处理**: 批量处理日志
3. **本地过滤**: 减少 API 调用
4. **并发控制**: 合理设置并发数

```json
{
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  },
  "batch_processing": {
    "enabled": true,
    "batch_size": 10,
    "batch_timeout": 5
  },
  "local_filter": true,
  "concurrency": {
    "max_workers": 5,
    "queue_size": 100
  }
}
```

### Q18: 内存使用过高怎么办？

**A**: 优化内存使用：

```json
{
  "memory": {
    "max_memory_usage": "512MB",
    "gc_interval": 300
  },
  "cache": {
    "max_size": 100
  }
}
```

### Q19: API 调用过多怎么办？

**A**: 减少 API 调用：

1. **启用本地过滤**: 过滤掉不重要的日志
2. **使用缓存**: 缓存分析结果
3. **批处理**: 批量处理日志
4. **调整频率限制**: 设置合理的频率限制

```json
{
  "local_filter": true,
  "cache": {
    "enabled": true,
    "ttl": 3600
  },
  "rate_limit": 60
}
```

## 🐛 故障排除

### Q20: 如何启用调试模式？

**A**: 使用调试选项：

```bash
# 启用详细输出
aipipe analyze --verbose

# 启用调试模式
AIPIPE_DEBUG=1 aipipe analyze

# 查看日志
tail -f ~/.aipipe/aipipe.log
```

### Q21: 如何查看错误日志？

**A**: 查看日志文件：

```bash
# 查看应用日志
tail -f ~/.aipipe/aipipe.log

# 查看错误日志
grep ERROR ~/.aipipe/aipipe.log

# 查看警告日志
grep WARN ~/.aipipe/aipipe.log
```

### Q22: 如何重置配置？

**A**: 重置配置文件：

```bash
# 备份当前配置
cp ~/.aipipe/config.json ~/.aipipe/config.json.backup

# 删除配置文件
rm ~/.aipipe/config.json

# 重新初始化
aipipe config init
```

### Q23: 如何清理缓存？

**A**: 清理缓存数据：

```bash
# 查看缓存统计
aipipe cache stats

# 清空缓存
aipipe cache clear

# 查看缓存状态
aipipe cache status
```

## 🔒 安全问题

### Q24: 如何保护 API 密钥？

**A**: 安全存储 API 密钥：

1. **环境变量**: 使用环境变量存储
2. **配置文件权限**: 设置合适的文件权限
3. **加密存储**: 使用加密存储

```bash
# 使用环境变量
export OPENAI_API_KEY="sk-your-api-key"

# 设置配置文件权限
chmod 600 ~/.aipipe/config.json
```

### Q25: 如何限制访问权限？

**A**: 设置访问控制：

```bash
# 设置文件权限
chmod 600 ~/.aipipe/config.json

# 设置目录权限
chmod 700 ~/.aipipe/

# 使用用户权限
sudo -u aipipe aipipe monitor
```

## 📊 监控和维护

### Q26: 如何监控 AIPipe 状态？

**A**: 使用状态命令：

```bash
# 查看系统状态
aipipe dashboard show

# 查看监控状态
aipipe dashboard status

# 查看统计信息
aipipe cache stats
aipipe ai stats
```

### Q27: 如何设置日志轮转？

**A**: 配置日志轮转：

```bash
# 创建 logrotate 配置
sudo nano /etc/logrotate.d/aipipe

# 配置内容
/var/log/aipipe/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 aipipe aipipe
}
```

### Q28: 如何备份配置？

**A**: 备份重要配置：

```bash
# 备份配置文件
cp ~/.aipipe/config.json ~/.aipipe/config.json.backup

# 备份监控配置
cp ~/.aipipe-monitor.json ~/.aipipe-monitor.json.backup

# 创建备份脚本
cat > backup-aipipe.sh << EOF
#!/bin/bash
BACKUP_DIR="/backup/aipipe/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"
cp ~/.aipipe/config.json "$BACKUP_DIR/"
cp ~/.aipipe-monitor.json "$BACKUP_DIR/"
echo "Backup completed: $BACKUP_DIR"
EOF
chmod +x backup-aipipe.sh
```

## 🎯 最佳实践

### Q29: 生产环境部署建议？

**A**: 生产环境最佳实践：

1. **使用专用用户**: 创建专用用户运行 AIPipe
2. **设置日志轮转**: 避免日志文件过大
3. **监控资源使用**: 监控 CPU、内存、磁盘使用
4. **配置告警**: 设置系统告警
5. **定期备份**: 定期备份配置文件

### Q30: 如何优化成本？

**A**: 成本优化建议：

1. **使用本地过滤**: 减少 API 调用
2. **启用缓存**: 避免重复分析
3. **选择合适模型**: 根据需求选择模型
4. **设置频率限制**: 控制 API 调用频率
5. **监控使用量**: 定期检查 API 使用量

## 🎉 总结

本章节涵盖了 AIPipe 使用过程中的常见问题和解决方案，包括：

- **安装配置**: 安装、配置、验证
- **日志分析**: 格式选择、性能优化、规则配置
- **文件监控**: 多文件监控、故障排除、优先级设置
- **通知系统**: 邮件、系统通知配置和故障排除
- **性能优化**: 缓存、批处理、并发控制
- **故障排除**: 调试模式、日志查看、配置重置
- **安全维护**: 权限控制、备份恢复、监控维护
- **最佳实践**: 生产环境部署、成本优化

如果遇到其他问题，可以：

1. 查看 [故障排除](13-troubleshooting.md) 章节
2. 在 GitHub 上提交 Issue
3. 查看项目文档和示例

---

*返回: [文档首页](README.md)*
