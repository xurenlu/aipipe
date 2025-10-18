# AIPipe 安装指南

本指南提供了多种安装 AIPipe 的方式，包括一键安装脚本和手动安装。

## 🚀 一键安装（推荐）

### 自动安装脚本

我们提供了一键安装脚本，支持 macOS 和 Linux 系统：

```bash
# 下载并运行安装脚本
curl -fsSL https://raw.githubusercontent.com/xurenlu/aipipe/main/install.sh | bash
```

或者手动下载后运行：

```bash
# 克隆仓库
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 运行安装脚本
chmod +x install.sh
./install.sh
```

### 安装脚本功能

- ✅ 自动检测操作系统和架构
- ✅ 自动安装 Go 语言环境（如需要）
- ✅ 从 GitHub 下载最新源码并编译
- ✅ 安装二进制文件到 `/usr/local/bin/`
- ✅ 创建默认配置文件
- ✅ 创建 systemd 服务（Linux）
- ✅ 创建启动脚本

## 🔧 手动安装

### 1. 安装依赖

#### macOS
```bash
# 使用 Homebrew 安装 Go
brew install go
```

#### Linux (Ubuntu/Debian)
```bash
# 安装 Go
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Linux (CentOS/RHEL)
```bash
# 安装 Go
sudo yum install -y golang
```

### 2. 编译安装

```bash
# 克隆仓库
git clone https://github.com/xurenlu/aipipe.git
cd aipipe

# 编译
go mod tidy
go build -o aipipe aipipe.go

# 安装到系统路径
sudo cp aipipe /usr/local/bin/
sudo chmod +x /usr/local/bin/aipipe
```

### 3. 创建配置文件

```bash
# 创建配置目录
mkdir -p ~/.config

# 复制示例配置文件
cp aipipe.json.example ~/.config/aipipe.json

# 编辑配置文件
nano ~/.config/aipipe.json
```

## 🐧 Linux 系统服务安装

### 使用 systemd 安装脚本

```bash
# 下载并运行 systemd 安装脚本
sudo ./install-systemd.sh
```

### 手动配置 systemd 服务

1. **复制服务文件**：
```bash
sudo cp aipipe.service /etc/systemd/system/
```

2. **编辑服务配置**：
```bash
sudo nano /etc/systemd/system/aipipe.service
```

3. **创建 aipipe 用户**：
```bash
sudo useradd -r -s /bin/false -d /home/aipipe -m aipipe
```

4. **创建配置目录**：
```bash
sudo mkdir -p /home/aipipe/.config
sudo chown aipipe:aipipe /home/aipipe/.config
```

5. **创建配置文件**：
```bash
sudo cp aipipe.json.example /home/aipipe/.config/aipipe.json
sudo chown aipipe:aipipe /home/aipipe/.config/aipipe.json
```

6. **启用并启动服务**：
```bash
sudo systemctl daemon-reload
sudo systemctl enable aipipe
sudo systemctl start aipipe
```

## ⚙️ 配置说明

### 基本配置

编辑配置文件 `~/.config/aipipe.json`：

```json
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": ""
}
```

### 通知配置

#### 邮件通知

**SMTP 配置**：
```json
"email": {
  "enabled": true,
  "provider": "smtp",
  "host": "smtp.gmail.com",
  "port": 587,
  "username": "your-email@gmail.com",
  "password": "your-app-password",
  "from_email": "your-email@gmail.com",
  "to_emails": ["admin@company.com"]
}
```

**Resend 配置**：
```json
"email": {
  "enabled": true,
  "provider": "resend",
  "password": "re_xxxxxxxxxxxxx",
  "from_email": "alerts@yourdomain.com",
  "to_emails": ["admin@company.com"]
}
```

#### Webhook 通知

**钉钉机器人**：
```json
"dingtalk": {
  "enabled": true,
  "url": "https://oapi.dingtalk.com/robot/send?access_token=YOUR_TOKEN"
}
```

**企业微信机器人**：
```json
"wechat": {
  "enabled": true,
  "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
}
```

**飞书机器人**：
```json
"feishu": {
  "enabled": true,
  "url": "https://open.feishu.cn/open-apis/bot/v2/hook/YOUR_TOKEN"
}
```

**Slack Webhook**：
```json
"slack": {
  "enabled": true,
  "url": "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
}
```

## 🚀 使用方法

### 基本使用

```bash
# 监控日志文件
aipipe -f /var/log/app.log --format java

# 通过管道输入
tail -f /var/log/app.log | aipipe --format java

# 查看帮助
aipipe --help
```

### 服务管理（Linux）

```bash
# 查看服务状态
sudo systemctl status aipipe

# 启动服务
sudo systemctl start aipipe

# 停止服务
sudo systemctl stop aipipe

# 重启服务
sudo systemctl restart aipipe

# 查看日志
sudo journalctl -u aipipe -f

# 禁用服务
sudo systemctl disable aipipe
```

### 启动脚本

安装脚本会创建启动脚本 `~/aipipe-start.sh`：

```bash
# 使用启动脚本
./aipipe-start.sh /var/log/app.log java
```

## 🔍 故障排除

### 常见问题

1. **编译失败**
   - 确保 Go 版本 >= 1.21
   - 检查网络连接
   - 尝试 `go clean -modcache` 清理模块缓存

2. **权限问题**
   - 确保有读取日志文件的权限
   - Linux 下可能需要将用户添加到 `adm` 组

3. **配置文件错误**
   - 检查 JSON 格式是否正确
   - 验证 AI 服务器端点和 Token

4. **服务启动失败**
   - 检查服务文件路径
   - 查看服务日志：`journalctl -u aipipe -f`
   - 验证日志文件路径

### 调试模式

```bash
# 启用调试模式
aipipe -f /var/log/app.log --format java --debug --verbose

# 查看详细输出
aipipe -f /var/log/app.log --format java --show-not-important
```

## 📚 更多信息

- [完整使用文档](README.md)
- [配置示例](aipipe.json.example)
- [GitHub 仓库](https://github.com/xurenlu/aipipe)
- [问题反馈](https://github.com/xurenlu/aipipe/issues)

## 🆘 获取帮助

如果遇到问题，请：

1. 查看 [故障排除](#故障排除) 部分
2. 查看 [GitHub Issues](https://github.com/xurenlu/aipipe/issues)
3. 提交新的 Issue 描述问题

---

**注意**：首次使用前请务必配置 AI 服务器端点和 Token，否则无法正常工作。
