# 11. 部署指南

> 生产环境部署、Docker 和系统服务配置

## 🎯 概述

本指南介绍如何在不同环境中部署 AIPipe，包括生产环境、Docker 容器和系统服务。

## 🚀 生产环境部署

### 1. 系统要求

- **操作系统**: Linux (Ubuntu 20.04+, CentOS 8+)
- **内存**: 最少 1GB，推荐 2GB+
- **CPU**: 最少 2 核，推荐 4 核+
- **磁盘**: 最少 10GB 可用空间
- **网络**: 需要访问 AI API 服务

### 2. 安装步骤

```bash
# 1. 下载最新版本
wget https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-linux-amd64
chmod +x aipipe-linux-amd64
sudo mv aipipe-linux-amd64 /usr/local/bin/aipipe

# 2. 创建用户
sudo useradd -r -s /bin/false aipipe
sudo mkdir -p /var/lib/aipipe
sudo chown aipipe:aipipe /var/lib/aipipe

# 3. 创建配置目录
sudo mkdir -p /etc/aipipe
sudo chown aipipe:aipipe /etc/aipipe

# 4. 初始化配置
sudo -u aipipe aipipe config init
```

### 3. 配置系统服务

```bash
# 创建 systemd 服务文件
sudo tee /etc/systemd/system/aipipe.service > /dev/null << EOF
[Unit]
Description=AIPipe Log Analysis Service
After=network.target

[Service]
Type=simple
User=aipipe
Group=aipipe
WorkingDirectory=/var/lib/aipipe
ExecStart=/usr/local/bin/aipipe monitor
Restart=always
RestartSec=5
Environment=AIPIPE_CONFIG_FILE=/etc/aipipe/config.json

[Install]
WantedBy=multi-user.target
EOF

# 启用服务
sudo systemctl daemon-reload
sudo systemctl enable aipipe
sudo systemctl start aipipe
```

## 🐳 Docker 部署

### 1. 使用官方镜像

```bash
# 拉取镜像
docker pull xurenlu/aipipe:latest

# 运行容器
docker run -d \
  --name aipipe \
  -v /var/log:/var/log:ro \
  -v ~/.aipipe:/root/.aipipe \
  -e OPENAI_API_KEY="sk-your-api-key" \
  xurenlu/aipipe:latest
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
      - ./config:/root/.aipipe
      - ./monitor:/root/.aipipe-monitor.json
    environment:
      - OPENAI_API_KEY=sk-your-api-key
      - AIPIPE_AI_MODEL=gpt-3.5-turbo
      - AIPIPE_LOG_LEVEL=info
    command: monitor
    depends_on:
      - redis
    networks:
      - aipipe-network

  redis:
    image: redis:7-alpine
    container_name: aipipe-redis
    restart: unless-stopped
    volumes:
      - redis-data:/data
    networks:
      - aipipe-network

volumes:
  redis-data:

networks:
  aipipe-network:
    driver: bridge
```

### 3. 运行 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f aipipe

# 停止服务
docker-compose down
```

## 🔧 配置管理

### 1. 环境变量配置

```bash
# 创建环境变量文件
cat > .env << EOF
OPENAI_API_KEY=sk-your-api-key
AIPIPE_AI_MODEL=gpt-3.5-turbo
AIPIPE_LOG_LEVEL=info
AIPIPE_CACHE_ENABLED=true
AIPIPE_CACHE_TTL=3600
AIPIPE_EMAIL_SMTP_HOST=smtp.gmail.com
AIPIPE_EMAIL_SMTP_PORT=587
AIPIPE_EMAIL_USERNAME=your-email@gmail.com
AIPIPE_EMAIL_PASSWORD=your-app-password
EOF
```

### 2. 配置文件管理

```bash
# 创建生产环境配置
cat > config/production.json << EOF
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-production-key",
  "ai_model": "gpt-4",
  "max_retries": 5,
  "timeout": 60,
  "rate_limit": 100,
  "local_filter": true,
  "show_not_important": false,
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.company.com",
      "smtp_port": 587,
      "username": "alerts@company.com",
      "password": "production-password",
      "to": "admin@company.com"
    }
  },
  "cache": {
    "enabled": true,
    "ttl": 3600,
    "max_size": 1000
  }
}
EOF
```

## 📊 监控和日志

### 1. 日志配置

```bash
# 创建日志目录
sudo mkdir -p /var/log/aipipe
sudo chown aipipe:aipipe /var/log/aipipe

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
}
EOF
```

### 2. 监控配置

```bash
# 安装监控工具
sudo apt-get install htop iotop nethogs

# 创建监控脚本
cat > monitor.sh << EOF
#!/bin/bash
# AIPipe 监控脚本

echo "=== AIPipe 状态 ==="
systemctl status aipipe

echo "=== 内存使用 ==="
ps aux | grep aipipe

echo "=== 日志文件 ==="
ls -la /var/log/aipipe/

echo "=== 配置状态 ==="
aipipe config status
EOF

chmod +x monitor.sh
```

## 🔒 安全配置

### 1. 用户权限

```bash
# 创建专用用户
sudo useradd -r -s /bin/false aipipe
sudo usermod -aG docker aipipe

# 设置文件权限
sudo chown -R aipipe:aipipe /var/lib/aipipe
sudo chmod 700 /var/lib/aipipe
sudo chmod 600 /etc/aipipe/config.json
```

### 2. 网络安全

```bash
# 配置防火墙
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# 限制网络访问
sudo iptables -A OUTPUT -p tcp --dport 443 -j ACCEPT
sudo iptables -A OUTPUT -p tcp --dport 587 -j ACCEPT
sudo iptables -A OUTPUT -j DROP
```

### 3. 密钥管理

```bash
# 使用密钥管理服务
export OPENAI_API_KEY=$(vault kv get -field=api_key secret/aipipe)
export AIPIPE_EMAIL_PASSWORD=$(vault kv get -field=password secret/email)
```

## 📈 性能优化

### 1. 系统优化

```bash
# 优化系统参数
echo 'vm.max_map_count=262144' | sudo tee -a /etc/sysctl.conf
echo 'fs.file-max=65536' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# 优化文件描述符
echo '* soft nofile 65536' | sudo tee -a /etc/security/limits.conf
echo '* hard nofile 65536' | sudo tee -a /etc/security/limits.conf
```

### 2. 应用优化

```json
{
  "performance": {
    "max_workers": 4,
    "queue_size": 1000,
    "batch_size": 10,
    "cache_size": 1000,
    "gc_interval": 300
  }
}
```

## 🔄 更新和维护

### 1. 自动更新

```bash
# 创建更新脚本
cat > update.sh << EOF
#!/bin/bash
# AIPipe 自动更新脚本

echo "停止服务..."
sudo systemctl stop aipipe

echo "备份配置..."
sudo cp /etc/aipipe/config.json /etc/aipipe/config.json.backup

echo "下载新版本..."
wget -O aipipe-new https://github.com/xurenlu/aipipe/releases/latest/download/aipipe-linux-amd64
chmod +x aipipe-new

echo "替换二进制文件..."
sudo mv aipipe-new /usr/local/bin/aipipe

echo "启动服务..."
sudo systemctl start aipipe

echo "检查状态..."
sudo systemctl status aipipe
EOF

chmod +x update.sh
```

### 2. 健康检查

```bash
# 创建健康检查脚本
cat > health-check.sh << EOF
#!/bin/bash
# AIPipe 健康检查脚本

# 检查服务状态
if ! systemctl is-active --quiet aipipe; then
    echo "AIPipe 服务未运行"
    exit 1
fi

# 检查配置
if ! aipipe config validate; then
    echo "配置文件验证失败"
    exit 1
fi

# 检查 AI 服务
if ! aipipe ai test; then
    echo "AI 服务测试失败"
    exit 1
fi

echo "AIPipe 健康检查通过"
exit 0
EOF

chmod +x health-check.sh
```

## 🎯 使用场景

### 场景1: 单机部署

```bash
# 简单单机部署
sudo ./install.sh
sudo systemctl start aipipe
sudo systemctl enable aipipe
```

### 场景2: 集群部署

```bash
# 使用 Docker Swarm
docker swarm init
docker stack deploy -c docker-compose.yml aipipe
```

### 场景3: Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aipipe
spec:
  replicas: 3
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
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: aipipe-secrets
              key: openai-api-key
        volumeMounts:
        - name: config
          mountPath: /root/.aipipe
      volumes:
      - name: config
        configMap:
          name: aipipe-config
```

## 🎉 总结

AIPipe 的部署指南提供了：

- **多种部署方式**: 系统服务、Docker、Kubernetes
- **完整配置**: 环境变量、配置文件、安全设置
- **监控维护**: 日志管理、健康检查、自动更新
- **性能优化**: 系统优化、应用优化
- **安全配置**: 用户权限、网络安全、密钥管理

---

*继续阅读: [12. 监控与维护](12-monitoring.md)*
