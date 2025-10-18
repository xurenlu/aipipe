#!/bin/bash

# AIPipe 一键安装脚本
# 支持 macOS 和 Linux 系统

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
        OS="macos"
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
        armv7l)
            ARCH="arm"
            ;;
        *)
            print_error "不支持的架构: $ARCH"
            exit 1
            ;;
    esac
    print_info "检测到架构: $ARCH"
}

# 检查依赖
check_dependencies() {
    print_info "检查依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        print_warning "Go 未安装，正在安装..."
        install_go
    else
        GO_VERSION=$(go version | cut -d' ' -f3)
        print_success "Go 已安装: $GO_VERSION"
    fi
    
    # 检查 Git
    if ! command -v git &> /dev/null; then
        print_error "Git 未安装，请先安装 Git"
        exit 1
    fi
    print_success "Git 已安装"
}

# 安装 Go (仅限 macOS 和 Linux)
install_go() {
    if [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            print_info "使用 Homebrew 安装 Go..."
            brew install go
        else
            print_error "请先安装 Homebrew 或手动安装 Go"
            exit 1
        fi
    elif [[ "$OS" == "linux" ]]; then
        print_info "下载并安装 Go..."
        GO_VERSION="1.21.0"
        wget -q "https://golang.org/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "go${GO_VERSION}.linux-${ARCH}.tar.gz"
        rm "go${GO_VERSION}.linux-${ARCH}.tar.gz"
        
        # 添加到 PATH
        if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        if ! grep -q "/usr/local/go/bin" ~/.profile; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
        fi
        export PATH=$PATH:/usr/local/go/bin
    fi
}

# 下载并编译 AIPipe
build_aipipe() {
    print_info "下载并编译 AIPipe..."
    
    # 创建临时目录
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # 克隆仓库
    print_info "从 GitHub 下载源码..."
    git clone https://github.com/xurenlu/aipipe.git .
    
    # 编译
    print_info "编译 AIPipe..."
    go mod tidy
    go build -o aipipe aipipe.go
    
    # 检查编译结果
    if [[ -f "aipipe" ]]; then
        print_success "编译成功"
    else
        print_error "编译失败"
        exit 1
    fi
}

# 安装二进制文件
install_binary() {
    print_info "安装 AIPipe 二进制文件..."
    
    # 创建安装目录
    INSTALL_DIR="/usr/local/bin"
    if [[ "$OS" == "macos" ]]; then
        INSTALL_DIR="/usr/local/bin"
    elif [[ "$OS" == "linux" ]]; then
        INSTALL_DIR="/usr/local/bin"
    fi
    
    # 复制二进制文件
    sudo cp aipipe "$INSTALL_DIR/"
    sudo chmod +x "$INSTALL_DIR/aipipe"
    
    print_success "AIPipe 已安装到 $INSTALL_DIR/aipipe"
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
WorkingDirectory=$HOME
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
    check_dependencies
    build_aipipe
    install_binary
    setup_config
    setup_systemd_service
    create_startup_script
    show_completion_info
    
    print_success "安装完成! 🎉"
}

# 运行主函数
main "$@"