#!/bin/bash

# AIPipe systemd 服务安装脚本
# 用于在 Linux 系统上安装和配置 AIPipe 作为系统服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "此脚本需要 root 权限运行"
        print_info "请使用: sudo $0"
        exit 1
    fi
}

# 检查系统是否为 Linux
check_linux() {
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        print_error "此脚本仅支持 Linux 系统"
        exit 1
    fi
}

# 检查 systemd 是否可用
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        print_error "systemd 不可用，请确保系统使用 systemd"
        exit 1
    fi
    print_success "systemd 可用"
}

# 创建 aipipe 用户
create_user() {
    if ! id "aipipe" &>/dev/null; then
        print_info "创建 aipipe 用户..."
        useradd -r -s /bin/false -d /home/aipipe -m aipipe
        print_success "aipipe 用户已创建"
    else
        print_info "aipipe 用户已存在"
    fi
}

# 创建必要的目录
create_directories() {
    print_info "创建必要的目录..."
    
    # 创建配置目录
    mkdir -p /home/aipipe/.config
    chown aipipe:aipipe /home/aipipe/.config
    
    # 创建日志目录（如果需要）
    mkdir -p /var/log/aipipe
    chown aipipe:aipipe /var/log/aipipe
    
    print_success "目录创建完成"
}

# 安装服务文件
install_service() {
    print_info "安装 systemd 服务文件..."
    
    SERVICE_FILE="/etc/systemd/system/aipipe.service"
    
    # 复制服务文件
    cp aipipe.service "$SERVICE_FILE"
    
    print_success "服务文件已安装: $SERVICE_FILE"
}

# 配置服务
configure_service() {
    print_info "配置服务参数..."
    
    SERVICE_FILE="/etc/systemd/system/aipipe.service"
    
    # 询问用户配置参数
    echo
    print_info "请配置 AIPipe 服务参数:"
    echo
    
    # 日志文件路径
    read -p "请输入要监控的日志文件路径 [/var/log/app.log]: " LOG_FILE
    LOG_FILE=${LOG_FILE:-/var/log/app.log}
    
    # 日志格式
    echo "可用的日志格式: java, php, nginx, ruby, python, fastapi, go, rust, csharp, kotlin, nodejs, typescript, docker, kubernetes, postgresql, mysql, redis, elasticsearch, git, jenkins, github, journald, macos-console, syslog"
    read -p "请输入日志格式 [java]: " LOG_FORMAT
    LOG_FORMAT=${LOG_FORMAT:-java}
    
    # 批处理大小
    read -p "请输入批处理大小 [10]: " BATCH_SIZE
    BATCH_SIZE=${BATCH_SIZE:-10}
    
    # 批处理等待时间
    read -p "请输入批处理等待时间 [3s]: " BATCH_WAIT
    BATCH_WAIT=${BATCH_WAIT:-3s}
    
    # 上下文行数
    read -p "请输入上下文行数 [3]: " CONTEXT_LINES
    CONTEXT_LINES=${CONTEXT_LINES:-3}
    
    # 构建启动命令
    START_COMMAND="/usr/local/bin/aipipe -f $LOG_FILE --format $LOG_FORMAT --batch-size $BATCH_SIZE --batch-wait $BATCH_WAIT --context $CONTEXT_LINES --verbose"
    
    # 更新服务文件
    sed -i "s|ExecStart=.*|ExecStart=$START_COMMAND|" "$SERVICE_FILE"
    
    print_success "服务配置完成"
    print_info "启动命令: $START_COMMAND"
}

# 创建配置文件
create_config() {
    print_info "创建配置文件..."
    
    CONFIG_FILE="/home/aipipe/.config/aipipe.json"
    
    if [[ ! -f "$CONFIG_FILE" ]]; then
        cat > "$CONFIG_FILE" << 'EOF'
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "your-api-token-here",
  "model": "gpt-4",
  "custom_prompt": "",
  "notifiers": {
    "email": {
      "enabled": false,
      "provider": "smtp",
      "host": "smtp.gmail.com",
      "port": 587,
      "username": "",
      "password": "",
      "from_email": "",
      "to_emails": []
    },
    "dingtalk": {
      "enabled": false,
      "url": ""
    },
    "wechat": {
      "enabled": false,
      "url": ""
    },
    "feishu": {
      "enabled": false,
      "url": ""
    },
    "slack": {
      "enabled": false,
      "url": ""
    },
    "custom_webhooks": []
  }
}
EOF
        chown aipipe:aipipe "$CONFIG_FILE"
        print_success "配置文件已创建: $CONFIG_FILE"
    else
        print_info "配置文件已存在: $CONFIG_FILE"
    fi
}

# 重新加载 systemd
reload_systemd() {
    print_info "重新加载 systemd..."
    systemctl daemon-reload
    print_success "systemd 已重新加载"
}

# 启用服务
enable_service() {
    print_info "启用 AIPipe 服务..."
    systemctl enable aipipe
    print_success "服务已启用（开机自启）"
}

# 启动服务
start_service() {
    print_info "启动 AIPipe 服务..."
    systemctl start aipipe
    
    # 等待服务启动
    sleep 2
    
    # 检查服务状态
    if systemctl is-active --quiet aipipe; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败"
        print_info "查看服务状态: systemctl status aipipe"
        print_info "查看服务日志: journalctl -u aipipe -f"
        exit 1
    fi
}

# 显示服务信息
show_service_info() {
    echo
    print_success "🎉 AIPipe 服务安装完成!"
    echo
    print_info "服务管理命令:"
    print_info "  查看状态: systemctl status aipipe"
    print_info "  启动服务: systemctl start aipipe"
    print_info "  停止服务: systemctl stop aipipe"
    print_info "  重启服务: systemctl restart aipipe"
    print_info "  查看日志: journalctl -u aipipe -f"
    print_info "  禁用服务: systemctl disable aipipe"
    echo
    print_info "配置文件位置: /home/aipipe/.config/aipipe.json"
    print_info "服务文件位置: /etc/systemd/system/aipipe.service"
    echo
    print_info "重要提醒:"
    print_info "  1. 请编辑配置文件设置 AI 服务器端点和 Token"
    print_info "  2. 配置通知方式（可选）"
    print_info "  3. 确保日志文件路径正确且有读取权限"
    print_info "  4. 服务会在开机时自动启动"
    echo
    print_info "配置完成后，请重启服务: systemctl restart aipipe"
}

# 主函数
main() {
    echo "🔧 AIPipe systemd 服务安装脚本"
    echo "=================================="
    echo
    
    check_root
    check_linux
    check_systemd
    create_user
    create_directories
    install_service
    configure_service
    create_config
    reload_systemd
    enable_service
    start_service
    show_service_info
}

# 运行主函数
main "$@"
