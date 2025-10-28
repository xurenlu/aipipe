# 12. 监控与维护

> 系统监控、日志管理和性能调优

## 🎯 概述

本指南介绍如何监控和维护 AIPipe 系统，确保稳定运行和最佳性能。

## 📊 系统监控

### 1. 服务状态监控

```bash
# 检查服务状态
systemctl status aipipe

# 查看服务日志
journalctl -u aipipe -f

# 检查服务健康
aipipe dashboard show
```

### 2. 资源监控

```bash
# 监控 CPU 使用
top -p $(pgrep aipipe)

# 监控内存使用
ps aux | grep aipipe

# 监控磁盘使用
df -h /var/lib/aipipe
```

### 3. 网络监控

```bash
# 监控网络连接
netstat -tulpn | grep aipipe

# 监控 API 调用
aipipe ai stats

# 监控通知发送
aipipe notify stats
```

## 📝 日志管理

### 1. 应用日志

```bash
# 查看应用日志
tail -f /var/log/aipipe/aipipe.log

# 查看错误日志
grep ERROR /var/log/aipipe/aipipe.log

# 查看警告日志
grep WARN /var/log/aipipe/aipipe.log
```

### 2. 系统日志

```bash
# 查看系统日志
journalctl -u aipipe --since "1 hour ago"

# 查看启动日志
journalctl -u aipipe -b

# 查看错误日志
journalctl -u aipipe -p err
```

### 3. 日志轮转

```bash
# 配置日志轮转
sudo tee /etc/logrotate.d/aipipe > /dev/null << EOF
/var/log/aipipe/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 aipipe aipipe
    postrotate
        systemctl reload aipipe
    endscript
}
EOF
```

## 🔧 性能调优

### 1. 内存优化

```bash
# 监控内存使用
free -h
ps aux --sort=-%mem | head

# 优化内存配置
aipipe config set --key "memory.max_memory_usage" --value "512MB"
aipipe config set --key "cache.max_size" --value "500"
```

### 2. CPU 优化

```bash
# 监控 CPU 使用
htop
iostat -x 1

# 优化并发配置
aipipe config set --key "concurrency.max_workers" --value "4"
aipipe config set --key "batch_processing.batch_size" --value "10"
```

### 3. 磁盘优化

```bash
# 监控磁盘使用
df -h
du -sh /var/lib/aipipe/*

# 清理旧文件
aipipe cache cleanup
find /var/log/aipipe -name "*.log.*" -mtime +7 -delete
```

## 📈 性能指标

### 1. 关键指标

```bash
# 查看性能指标
aipipe metrics

# 查看缓存命中率
aipipe cache stats

# 查看 API 调用统计
aipipe ai stats
```

### 2. 监控脚本

```bash
# 创建监控脚本
cat > monitor.sh << EOF
#!/bin/bash
# AIPipe 性能监控脚本

echo "=== 系统状态 ==="
systemctl status aipipe --no-pager

echo "=== 内存使用 ==="
free -h

echo "=== 磁盘使用 ==="
df -h /var/lib/aipipe

echo "=== 缓存统计 ==="
aipipe cache stats

echo "=== AI 服务统计 ==="
aipipe ai stats

echo "=== 通知统计 ==="
aipipe notify stats
EOF

chmod +x monitor.sh
```

## 🔄 维护任务

### 1. 定期维护

```bash
# 创建维护脚本
cat > maintenance.sh << EOF
#!/bin/bash
# AIPipe 定期维护脚本

echo "开始维护..."

# 清理缓存
echo "清理缓存..."
aipipe cache cleanup

# 清理日志
echo "清理旧日志..."
find /var/log/aipipe -name "*.log.*" -mtime +7 -delete

# 检查配置
echo "检查配置..."
aipipe config validate

# 测试服务
echo "测试服务..."
aipipe ai test
aipipe notify test

echo "维护完成"
EOF

chmod +x maintenance.sh
```

### 2. 自动维护

```bash
# 添加到 crontab
(crontab -l 2>/dev/null; echo "0 2 * * * /path/to/maintenance.sh") | crontab -

# 或者使用 systemd timer
sudo tee /etc/systemd/system/aipipe-maintenance.timer > /dev/null << EOF
[Unit]
Description=AIPipe Maintenance Timer

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
EOF

sudo systemctl enable aipipe-maintenance.timer
sudo systemctl start aipipe-maintenance.timer
```

## 🚨 告警配置

### 1. 系统告警

```bash
# 创建告警脚本
cat > alert.sh << EOF
#!/bin/bash
# AIPipe 系统告警脚本

# 检查服务状态
if ! systemctl is-active --quiet aipipe; then
    echo "AIPipe 服务停止" | mail -s "AIPipe 告警" admin@example.com
    exit 1
fi

# 检查内存使用
MEMORY_USAGE=$(ps aux | grep aipipe | awk '{sum+=$6} END {print sum/1024}')
if (( $(echo "$MEMORY_USAGE > 1000" | bc -l) )); then
    echo "AIPipe 内存使用过高: ${MEMORY_USAGE}MB" | mail -s "AIPipe 告警" admin@example.com
fi

# 检查磁盘使用
DISK_USAGE=$(df /var/lib/aipipe | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 80 ]; then
    echo "AIPipe 磁盘使用过高: ${DISK_USAGE}%" | mail -s "AIPipe 告警" admin@example.com
fi
EOF

chmod +x alert.sh
```

### 2. 监控集成

```bash
# Prometheus 监控
cat > prometheus.yml << EOF
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'aipipe'
    static_configs:
      - targets: ['localhost:8080']
EOF

# Grafana 仪表板
cat > aipipe-dashboard.json << EOF
{
  "dashboard": {
    "title": "AIPipe 监控面板",
    "panels": [
      {
        "title": "服务状态",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"aipipe\"}"
          }
        ]
      }
    ]
  }
}
EOF
```

## 🔍 故障诊断

### 1. 常见问题

```bash
# 服务无法启动
systemctl status aipipe
journalctl -u aipipe -n 50

# 配置问题
aipipe config validate --verbose

# 网络问题
ping api.openai.com
curl -I https://api.openai.com/v1/models
```

### 2. 调试模式

```bash
# 启用调试模式
export AIPIPE_DEBUG=true
export AIPIPE_LOG_LEVEL=debug

# 重新启动服务
systemctl restart aipipe

# 查看调试日志
journalctl -u aipipe -f
```

### 3. 性能分析

```bash
# 使用 pprof 分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 使用 trace 分析
go tool trace trace.out
```

## 📋 最佳实践

### 1. 监控策略

- 设置关键指标阈值
- 配置自动告警
- 定期检查系统状态
- 建立监控仪表板

### 2. 维护策略

- 定期清理缓存和日志
- 备份重要配置
- 测试服务功能
- 更新系统组件

### 3. 性能优化

- 监控资源使用情况
- 调整配置参数
- 优化缓存策略
- 升级硬件资源

## 🎉 总结

AIPipe 的监控与维护提供了：

- **全面监控**: 系统、应用、性能监控
- **日志管理**: 完整的日志收集和分析
- **性能调优**: 内存、CPU、磁盘优化
- **自动维护**: 定期维护和告警
- **故障诊断**: 完整的诊断工具和流程

---

*继续阅读: [13. 故障排除](13-troubleshooting.md)*
