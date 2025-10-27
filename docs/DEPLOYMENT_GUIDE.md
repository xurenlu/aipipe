# AIPipe 部署和运维指南 🚀

**版本**: v1.1.0  
**更新时间**: 2024年10月28日  
**状态**: 生产就绪

## 📋 目录

1. [系统要求](#系统要求)
2. [安装部署](#安装部署)
3. [配置管理](#配置管理)
4. [服务管理](#服务管理)
5. [监控运维](#监控运维)
6. [故障排除](#故障排除)
7. [性能调优](#性能调优)
8. [安全配置](#安全配置)
9. [备份恢复](#备份恢复)
10. [升级维护](#升级维护)

## 💻 系统要求

### 最低要求

| 组件 | 最低要求 | 推荐配置 |
|------|----------|----------|
| **操作系统** | Linux 3.10+, macOS 10.14+, Windows 10+ | Linux 5.4+, macOS 12+, Windows 11+ |
| **CPU** | 2核心 | 4核心+ |
| **内存** | 512MB | 2GB+ |
| **磁盘** | 1GB | 10GB+ |
| **网络** | 100Mbps | 1Gbps+ |

### 依赖软件

- **Go**: 1.21+ (编译时)
- **Docker**: 20.10+ (容器部署)
- **Systemd**: 240+ (服务管理)
- **Git**: 2.0+ (源码管理)

### 网络要求

- **出站连接**: HTTPS (443端口) - AI服务API
- **入站连接**: 无特殊要求
- **代理支持**: HTTP/HTTPS/SOCKS5代理

## 🚀 安装部署

### 1. 二进制安装 (推荐)

#### Linux/macOS

```bash
# 下载最新版本
curl -L https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-linux-amd64 -o aipipe
chmod +x aipipe
sudo mv aipipe /usr/local/bin/

# 验证安装
aipipe --version
```

#### Windows

```powershell
# 下载最新版本
Invoke-WebRequest -Uri "https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-windows-amd64.exe" -OutFile "aipipe.exe"

# 添加到PATH
$env:PATH += ";C:\path\to\aipipe"
```

### 2. 源码编译

```bash
# 克隆仓库
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 编译
go build -o aipipe .

# 安装
sudo cp aipipe /usr/local/bin/
```

### 3. Docker部署

```bash
# 拉取镜像
docker pull xurenlu/aipipe:latest

# 运行容器
docker run -d \
  --name aipipe \
  -v /var/log:/var/log:ro \
  -v ~/.config/aipipe.json:/app/config.json \
  xurenlu/aipipe:latest
```

### 4. 包管理器安装

#### Ubuntu/Debian

```bash
# 添加仓库
curl -fsSL https://packages.aipipe.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/aipipe.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/aipipe.gpg] https://packages.aipipe.io/ubuntu focal main" | sudo tee /etc/apt/sources.list.d/aipipe.list

# 安装
sudo apt update
sudo apt install aipipe
```

#### CentOS/RHEL

```bash
# 添加仓库
sudo yum-config-manager --add-repo https://packages.aipipe.io/rpm/aipipe.repo

# 安装
sudo yum install aipipe
```

## ⚙️ 配置管理

### 1. 初始配置

```bash
# 启动配置向导
aipipe --config-init

# 或手动创建配置
mkdir -p ~/.config
cp aipipe.json.example ~/.config/aipipe.json
```

### 2. 配置文件结构

```json
{
  "ai": {
    "endpoint": "https://api.openai.com/v1/chat/completions",
    "token": "your-api-token",
    "model": "gpt-3.5-turbo",
    "max_tokens": 1000,
    "temperature": 0.7
  },
  "cache": {
    "enabled": true,
    "size": 1000,
    "ttl": "1h"
  },
  "worker": {
    "pool_size": 4,
    "queue_size": 100
  },
  "memory": {
    "limit": "512MB",
    "gc_threshold": 0.8
  },
  "output_format": {
    "format": "table",
    "color": true
  },
  "log_level": {
    "enabled": true,
    "min_level": "info",
    "max_level": "error"
  }
}
```

### 3. 环境变量配置

```bash
# AI配置
export AIPIPE_AI_ENDPOINT="https://api.openai.com/v1/chat/completions"
export AIPIPE_AI_TOKEN="your-api-token"

# 缓存配置
export AIPIPE_CACHE_SIZE="1000"
export AIPIPE_CACHE_TTL="1h"

# 工作池配置
export AIPIPE_WORKER_POOL_SIZE="4"
export AIPIPE_WORKER_QUEUE_SIZE="100"
```

### 4. 配置验证

```bash
# 验证配置
aipipe --config-test

# 显示当前配置
aipipe --config-show

# 生成配置模板
aipipe --config-template yaml > config.yaml
```

## 🔧 服务管理

### 1. Systemd服务

#### 创建服务文件

```bash
sudo tee /etc/systemd/system/aipipe.service > /dev/null <<EOF
[Unit]
Description=AIPipe Log Analysis Service
After=network.target

[Service]
Type=simple
User=aipipe
Group=aipipe
ExecStart=/usr/local/bin/aipipe --config /etc/aipipe/config.json
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
```

#### 管理服务

```bash
# 重载配置
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start aipipe

# 停止服务
sudo systemctl stop aipipe

# 重启服务
sudo systemctl restart aipipe

# 查看状态
sudo systemctl status aipipe

# 开机自启
sudo systemctl enable aipipe
```

### 2. Docker Compose

```yaml
version: '3.8'
services:
  aipipe:
    image: xurenlu/aipipe:latest
    container_name: aipipe
    restart: unless-stopped
    volumes:
      - /var/log:/var/log:ro
      - ./config:/app/config
      - ./logs:/app/logs
    environment:
      - AIPIPE_AI_ENDPOINT=https://api.openai.com/v1/chat/completions
      - AIPIPE_AI_TOKEN=${AI_TOKEN}
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "aipipe", "--config-test"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### 3. Kubernetes部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aipipe
spec:
  replicas: 2
  selector:
    matchLabels:
      app: aipipe
  template:
    metadata:
      labels:
        app: aipipe
    spec:
      containers:
      - name: aipipe
        image: xurenlu/aipipe:latest
        ports:
        - containerPort: 8080
        env:
        - name: AIPIPE_AI_TOKEN
          valueFrom:
            secretKeyRef:
              name: aipipe-secrets
              key: ai-token
        volumeMounts:
        - name: config
          mountPath: /app/config
        - name: logs
          mountPath: /var/log
      volumes:
      - name: config
        configMap:
          name: aipipe-config
      - name: logs
        hostPath:
          path: /var/log
```

## 📊 监控运维

### 1. 日志监控

```bash
# 查看服务日志
sudo journalctl -u aipipe -f

# 查看应用日志
tail -f /var/log/aipipe/app.log

# 查看错误日志
tail -f /var/log/aipipe/error.log
```

### 2. 性能监控

```bash
# 查看性能统计
aipipe --performance-stats

# 查看内存使用
aipipe --memory-stats

# 查看缓存统计
aipipe --cache-stats

# 查看工作池统计
aipipe --worker-stats
```

### 3. 健康检查

```bash
# 检查服务健康
curl http://localhost:8080/health

# 检查配置
aipipe --config-test

# 检查AI连接
aipipe --ai-test
```

### 4. 监控脚本

```bash
#!/bin/bash
# 监控脚本示例

# 检查服务状态
if ! systemctl is-active --quiet aipipe; then
    echo "AIPipe service is not running"
    systemctl restart aipipe
fi

# 检查内存使用
MEMORY_USAGE=$(aipipe --memory-stats | grep "Memory Usage" | awk '{print $3}')
if [ "$MEMORY_USAGE" -gt 80 ]; then
    echo "High memory usage: $MEMORY_USAGE%"
    aipipe --memory-gc
fi

# 检查错误率
ERROR_RATE=$(aipipe --performance-stats | grep "Error Rate" | awk '{print $3}')
if [ "$ERROR_RATE" -gt 5 ]; then
    echo "High error rate: $ERROR_RATE%"
    # 发送告警
fi
```

## 🔍 故障排除

### 1. 常见问题

#### 服务无法启动

```bash
# 检查配置文件
aipipe --config-test

# 检查权限
ls -la /usr/local/bin/aipipe
ls -la ~/.config/aipipe.json

# 检查依赖
ldd /usr/local/bin/aipipe
```

#### 内存使用过高

```bash
# 查看内存统计
aipipe --memory-stats

# 强制垃圾回收
aipipe --memory-gc

# 调整内存限制
export AIPIPE_MEMORY_LIMIT="256MB"
```

#### AI服务连接失败

```bash
# 测试AI连接
aipipe --ai-test

# 检查网络连接
curl -I https://api.openai.com/v1/chat/completions

# 检查代理设置
echo $https_proxy
echo $http_proxy
```

### 2. 调试模式

```bash
# 启用调试模式
aipipe --debug --verbose /var/log/app.log

# 查看详细日志
aipipe --log-level debug /var/log/app.log

# 性能分析
aipipe --profile /var/log/app.log
```

### 3. 日志分析

```bash
# 分析错误日志
grep "ERROR" /var/log/aipipe/app.log | tail -20

# 分析性能日志
grep "Performance" /var/log/aipipe/app.log | tail -20

# 分析内存日志
grep "Memory" /var/log/aipipe/app.log | tail -20
```

## ⚡ 性能调优

### 1. 系统级优化

#### 文件描述符限制

```bash
# 增加文件描述符限制
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf
```

#### 内核参数优化

```bash
# 优化网络参数
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

### 2. 应用级优化

#### 工作池配置

```json
{
  "worker": {
    "pool_size": 8,
    "queue_size": 200,
    "timeout": "30s"
  }
}
```

#### 缓存配置

```json
{
  "cache": {
    "enabled": true,
    "size": 2000,
    "ttl": "2h",
    "cleanup_interval": "10m"
  }
}
```

#### 内存配置

```json
{
  "memory": {
    "limit": "1GB",
    "gc_threshold": 0.7,
    "pool_size": 100
  }
}
```

### 3. 监控指标

```bash
# 创建监控脚本
cat > /usr/local/bin/aipipe-monitor.sh << 'EOF'
#!/bin/bash
while true; do
    echo "$(date): $(aipipe --performance-stats)"
    sleep 60
done
EOF

chmod +x /usr/local/bin/aipipe-monitor.sh
```

## 🔒 安全配置

### 1. 用户权限

```bash
# 创建专用用户
sudo useradd -r -s /bin/false aipipe

# 设置文件权限
sudo chown -R aipipe:aipipe /etc/aipipe
sudo chmod 600 /etc/aipipe/config.json
```

### 2. 网络安全

```bash
# 配置防火墙
sudo ufw allow 8080/tcp
sudo ufw deny 22/tcp

# 配置SSL证书
sudo cp ssl.crt /etc/ssl/certs/aipipe.crt
sudo cp ssl.key /etc/ssl/private/aipipe.key
```

### 3. 配置加密

```bash
# 加密配置文件
gpg --symmetric --cipher-algo AES256 ~/.config/aipipe.json

# 使用加密配置
aipipe --config ~/.config/aipipe.json.gpg --decrypt
```

## 💾 备份恢复

### 1. 配置备份

```bash
# 备份配置
tar -czf aipipe-config-$(date +%Y%m%d).tar.gz ~/.config/aipipe.json

# 恢复配置
tar -xzf aipipe-config-20241028.tar.gz -C ~/.config/
```

### 2. 数据备份

```bash
# 备份日志数据
tar -czf aipipe-logs-$(date +%Y%m%d).tar.gz /var/log/aipipe/

# 备份缓存数据
tar -czf aipipe-cache-$(date +%Y%m%d).tar.gz ~/.cache/aipipe/
```

### 3. 自动备份

```bash
# 创建备份脚本
cat > /usr/local/bin/aipipe-backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/var/backups/aipipe"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR

# 备份配置
cp ~/.config/aipipe.json $BACKUP_DIR/config_$DATE.json

# 备份日志
tar -czf $BACKUP_DIR/logs_$DATE.tar.gz /var/log/aipipe/

# 清理旧备份
find $BACKUP_DIR -name "*.tar.gz" -mtime +7 -delete
EOF

chmod +x /usr/local/bin/aipipe-backup.sh

# 添加到crontab
echo "0 2 * * * /usr/local/bin/aipipe-backup.sh" | crontab -
```

## 🔄 升级维护

### 1. 版本升级

```bash
# 备份当前版本
sudo cp /usr/local/bin/aipipe /usr/local/bin/aipipe.backup

# 下载新版本
curl -L https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-linux-amd64 -o aipipe

# 停止服务
sudo systemctl stop aipipe

# 替换二进制
sudo cp aipipe /usr/local/bin/
sudo chmod +x /usr/local/bin/aipipe

# 启动服务
sudo systemctl start aipipe

# 验证升级
aipipe --version
```

### 2. 配置迁移

```bash
# 检查配置兼容性
aipipe --config-validate

# 迁移配置
aipipe --config-migrate

# 测试新配置
aipipe --config-test
```

### 3. 回滚操作

```bash
# 停止服务
sudo systemctl stop aipipe

# 恢复旧版本
sudo cp /usr/local/bin/aipipe.backup /usr/local/bin/aipipe

# 启动服务
sudo systemctl start aipipe
```

## 📞 支持与维护

### 1. 日志收集

```bash
# 收集诊断信息
aipipe --diagnose > aipipe-diagnose-$(date +%Y%m%d).log

# 收集性能数据
aipipe --performance-stats > aipipe-performance-$(date +%Y%m%d).log
```

### 2. 问题报告

```bash
# 生成问题报告
aipipe --bug-report > aipipe-bug-report-$(date +%Y%m%d).log
```

### 3. 社区支持

- **GitHub Issues**: https://github.com/xurenlu/aipipe/issues
- **文档**: https://docs.aipipe.io
- **社区**: https://community.aipipe.io

---

**部署状态**: ✅ 生产就绪  
**维护状态**: ✅ 活跃维护  
**文档状态**: ✅ 完整详细  
**支持状态**: ✅ 社区支持  

*最后更新: 2024年10月28日*
