#!/bin/bash

# AIPipe 通知功能测试脚本

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

# 创建测试日志文件
create_test_log() {
    print_info "创建测试日志文件..."
    
    TEST_LOG_FILE="test-notification.log"
    
    cat > "$TEST_LOG_FILE" << 'EOF'
2025-01-17 10:00:01 INFO Application started successfully
2025-01-17 10:00:02 INFO Database connection established
2025-01-17 10:00:03 INFO User login: john@example.com
2025-01-17 10:00:04 INFO Processing request: GET /api/users
2025-01-17 10:00:05 INFO Health check endpoint called
2025-01-17 10:00:06 ERROR Database connection timeout after 30 seconds
2025-01-17 10:00:07 ERROR Failed to connect to Redis server
2025-01-17 10:00:08 WARN Memory usage high: 85%
2025-01-17 10:00:09 INFO Retry attempt 1
2025-01-17 10:00:10 ERROR Authentication failed for user admin
2025-01-17 10:00:11 INFO Cache hit for key: user_123
2025-01-17 10:00:12 INFO Request processed in 50ms
EOF
    
    print_success "测试日志文件已创建: $TEST_LOG_FILE"
}

# 创建测试配置
create_test_config() {
    print_info "创建测试配置..."
    
    CONFIG_FILE="$HOME/.config/aipipe.json"
    BACKUP_FILE="$HOME/.config/aipipe.json.backup"
    
    # 备份原配置
    if [[ -f "$CONFIG_FILE" ]]; then
        cp "$CONFIG_FILE" "$BACKUP_FILE"
        print_info "原配置文件已备份: $BACKUP_FILE"
    fi
    
    # 创建测试配置（禁用AI调用，只测试通知）
    cat > "$CONFIG_FILE" << 'EOF'
{
  "ai_endpoint": "https://your-ai-server.com/api/v1/chat/completions",
  "token": "test-token",
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
    
    print_success "测试配置已创建"
}

# 恢复原配置
restore_config() {
    CONFIG_FILE="$HOME/.config/aipipe.json"
    BACKUP_FILE="$HOME/.config/aipipe.json.backup"
    
    if [[ -f "$BACKUP_FILE" ]]; then
        cp "$BACKUP_FILE" "$CONFIG_FILE"
        rm "$BACKUP_FILE"
        print_info "原配置文件已恢复"
    fi
}

# 测试系统通知
test_system_notification() {
    print_info "测试系统通知功能..."
    
    # 直接测试通知函数（需要修改代码或创建测试程序）
    print_warning "系统通知测试需要在实际运行中验证"
    print_info "请运行以下命令测试系统通知："
    echo "  ./aipipe -f test-notification.log --format java --verbose"
}

# 显示配置指南
show_config_guide() {
    print_info "通知配置指南："
    echo
    print_info "1. 邮件通知配置："
    echo "   编辑 ~/.config/aipipe.json 中的 email 部分"
    echo "   设置 enabled: true 并配置 SMTP 或 Resend 参数"
    echo
    print_info "2. Webhook 通知配置："
    echo "   编辑 ~/.config/aipipe.json 中的 webhook 部分"
    echo "   设置 enabled: true 并配置相应的 URL"
    echo
    print_info "3. 支持的平台："
    echo "   - 钉钉机器人"
    echo "   - 企业微信机器人"
    echo "   - 飞书机器人"
    echo "   - Slack Webhook"
    echo "   - 自定义 Webhook"
    echo
    print_info "4. 智能识别："
    echo "   AIPipe 会自动识别 webhook URL 类型"
    echo "   无需手动指定平台类型"
}

# 清理测试文件
cleanup() {
    print_info "清理测试文件..."
    
    if [[ -f "test-notification.log" ]]; then
        rm "test-notification.log"
        print_info "测试日志文件已删除"
    fi
    
    restore_config
}

# 主函数
main() {
    echo "🧪 AIPipe 通知功能测试"
    echo "======================="
    echo
    
    # 设置错误处理
    trap cleanup EXIT
    
    # 检查 AIPipe 是否存在
    if [[ ! -f "./aipipe" ]]; then
        print_error "AIPipe 可执行文件不存在，请先编译"
        print_info "运行: go build -o aipipe aipipe.go"
        exit 1
    fi
    
    create_test_log
    create_test_config
    
    echo
    print_info "测试选项："
    echo "1. 查看配置指南"
    echo "2. 运行通知测试"
    echo "3. 退出"
    echo
    
    read -p "请选择 (1-3): " choice
    
    case $choice in
        1)
            show_config_guide
            ;;
        2)
            test_system_notification
            ;;
        3)
            print_info "退出测试"
            ;;
        *)
            print_error "无效选择"
            ;;
    esac
    
    echo
    print_info "测试完成！"
}

# 运行主函数
main "$@"
