#!/bin/bash

# AIPipe 一键安装脚本
# 支持 macOS 和 Linux 系统
# 从 GitHub Release 下载预编译二进制文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
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

# 检查系统类型
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS="darwin"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
    else
        print_error "不支持的操作系统: $OSTYPE"
        exit 1
    fi
    print_info "检测到操作系统: $OS"
}

# 检查架构
detect_arch() {
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    print_info "检测到架构: $ARCH"
}

# 获取最新版本
get_latest_version() {
    print_info "获取最新版本信息..."
    
    # 使用 GitHub API 获取最新 release
    LATEST_VERSION=$(curl -s https://api.github.com/repos/xurenlu/aipipe/releases/latest | grep '"tag_name"' | cut -d'"' -f4)
    
    if [[ -z "$LATEST_VERSION" ]]; then
        print_error "无法获取最新版本信息"
        exit 1
    fi
    
    print_success "最新版本: $LATEST_VERSION"
    VERSION="$LATEST_VERSION"
}

# 下载并安装二进制文件
download_and_install() {
    print_info "下载 AIPipe 二进制文件..."
    
    # 构建下载URL
    if [[ "$OS" == "darwin" ]]; then
        if [[ "$ARCH" == "arm64" ]]; then
            FILENAME="aipipe-${VERSION}-darwin-arm64.tar.gz"
        else
            FILENAME="aipipe-${VERSION}-darwin-amd64.tar.gz"
        fi
    elif [[ "$OS" == "linux" ]]; then
        FILENAME="aipipe-${VERSION}-linux-amd64.tar.gz"
    fi
    
    DOWNLOAD_URL="https://github.com/xurenlu/aipipe/releases/download/${VERSION}/${FILENAME}"
    
    print_info "下载地址: $DOWNLOAD_URL"
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # 下载文件
    print_info "正在下载 $FILENAME..."
    if ! curl -L -o "$FILENAME" "$DOWNLOAD_URL"; then
        print_error "下载失败，请检查网络连接或版本信息"
        exit 1
    fi
    
    # 解压文件
    print_info "解压文件..."
    tar -xzf "$FILENAME"
    
    # 检查二进制文件
    if [[ ! -f "aipipe" ]]; then
        print_error "解压后未找到 aipipe 二进制文件"
        exit 1
    fi
    
    # 安装二进制文件
    print_info "安装 AIPipe 到 /usr/local/bin..."
    sudo cp aipipe /usr/local/bin/
    sudo chmod +x /usr/local/bin/aipipe
    
    print_success "AIPipe 已安装到 /usr/local/bin/aipipe"
    
    # 验证安装
    if command -v aipipe &> /dev/null; then
        INSTALLED_VERSION=$(aipipe --version 2>/dev/null || echo "unknown")
        print_success "安装成功! 版本: $INSTALLED_VERSION"
    else
        print_warning "安装完成，但无法验证版本"
    fi
}

# 创建配置目录和文件
setup_config() {
    print_info "设置配置文件..."
    
    CONFIG_DIR="$HOME/.config"
    CONFIG_FILE="$CONFIG_DIR/aipipe.json"
    
    # 创建配置目录
    mkdir -p "$CONFIG_DIR"
    
    # 如果配置文件不存在，创建默认配置
    if [[ ! -f "$CONFIG_FILE" ]]; then
        print_info "创建默认配置文件..."
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
        print_success "配置文件已创建: $CONFIG_FILE"
    else
        print_info "配置文件已存在: $CONFIG_FILE"
    fi
}

# 创建 systemd 服务 (仅限 Linux)
setup_systemd_service() {
    if [[ "$OS" == "linux" ]]; then
        print_info "创建 systemd 服务..."
        
        SERVICE_FILE="/etc/systemd/system/aipipe.service"
        SERVICE_USER="${SUDO_USER:-$USER}"
        
        sudo tee "$SERVICE_FILE" > /dev/null << EOF
[Unit]
Description=AIPipe Log Monitor
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=/tmp
ExecStart=/usr/local/bin/aipipe -f /var/log/your-app.log --format java
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
        
        # 重新加载 systemd
        sudo systemctl daemon-reload
        
        print_success "systemd 服务已创建: $SERVICE_FILE"
        print_info "使用以下命令管理服务:"
        print_info "  启动服务: sudo systemctl start aipipe"
        print_info "  停止服务: sudo systemctl stop aipipe"
        print_info "  开机自启: sudo systemctl enable aipipe"
        print_info "  查看状态: sudo systemctl status aipipe"
        print_info "  查看日志: sudo journalctl -u aipipe -f"
    fi
}

# 创建启动脚本
create_startup_script() {
    print_info "创建启动脚本..."
    
    STARTUP_SCRIPT="$HOME/aipipe-start.sh"
    cat > "$STARTUP_SCRIPT" << 'EOF'
#!/bin/bash

# AIPipe 启动脚本
# 使用方法: ./aipipe-start.sh [日志文件路径] [日志格式]

LOG_FILE="${1:-/var/log/app.log}"
LOG_FORMAT="${2:-java}"

echo "启动 AIPipe 监控 $LOG_FILE (格式: $LOG_FORMAT)"

# 检查日志文件是否存在
if [[ ! -f "$LOG_FILE" ]]; then
    echo "警告: 日志文件 $LOG_FILE 不存在"
    echo "请确保日志文件路径正确，或创建测试日志文件"
fi

# 启动 AIPipe
aipipe -f "$LOG_FILE" --format "$LOG_FORMAT" --verbose
EOF
    
    chmod +x "$STARTUP_SCRIPT"
    print_success "启动脚本已创建: $STARTUP_SCRIPT"
}

# 显示安装完成信息
show_completion_info() {
    print_success "🎉 AIPipe 安装完成!"
    echo
    print_info "安装位置: /usr/local/bin/aipipe"
    print_info "配置文件: $HOME/.config/aipipe.json"
    print_info "启动脚本: $HOME/aipipe-start.sh"
    echo
    print_info "使用方法:"
    print_info "  1. 编辑配置文件: nano $HOME/.config/aipipe.json"
    print_info "  2. 设置 AI 服务器端点和 Token"
    print_info "  3. 配置通知方式（可选）"
    print_info "  4. 启动监控: ./aipipe-start.sh /path/to/your/logfile java"
    echo
    print_info "快速测试:"
    print_info "  aipipe --help"
    echo
    if [[ "$OS" == "linux" ]]; then
        print_info "systemd 服务管理:"
        print_info "  sudo systemctl start aipipe    # 启动服务"
        print_info "  sudo systemctl enable aipipe   # 开机自启"
        print_info "  sudo journalctl -u aipipe -f   # 查看日志"
        echo
    fi
    print_info "更多信息请查看: https://github.com/xurenlu/aipipe"
}

# 清理临时文件
cleanup() {
    if [[ -n "$TEMP_DIR" && -d "$TEMP_DIR" ]]; then
        rm -rf "$TEMP_DIR"
    fi
}

# 主函数
main() {
    echo "🚀 AIPipe 一键安装脚本"
    echo "=========================="
    echo
    
    # 设置错误处理
    trap cleanup EXIT
    
    # 检查是否为 root 用户（Linux 需要）
    if [[ "$EUID" -eq 0 && "$OS" == "linux" ]]; then
        print_warning "检测到 root 用户，建议使用普通用户运行此脚本"
        print_info "请使用: sudo -u $SUDO_USER $0"
        exit 1
    fi
    
    # 执行安装步骤
    detect_os
    detect_arch
    get_latest_version
    download_and_install
    setup_config
    setup_systemd_service
    create_startup_script
    show_completion_info
    
    print_success "安装完成! 🎉"
}

# 运行主函数
main "$@"