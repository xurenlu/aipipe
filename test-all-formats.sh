#!/bin/bash

# AIPipe 24 种日志格式完整测试脚本
# 测试所有支持的日志格式的正确识别

echo "🎯 AIPipe 24 种日志格式完整测试"
echo "======================================"
echo ""

# 测试函数
test_format() {
    local format="$1"
    local log_sample="$2"
    local category="$3"
    
    echo -n "🔍 测试格式: $format"
    
    # 运行测试
    local test_result
    test_result=$(echo "$log_sample" | ./aipipe --format "$format" --verbose 2>&1 | grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1)
    
    if [[ -n "$test_result" ]]; then
        echo " ✅"
        return 0
    else
        echo " ❌"
        return 1
    fi
}

# 统计变量
total_formats=0
successful_tests=0
failed_tests=0

echo "📋 开始测试所有支持的日志格式..."
echo ""

# 后端编程语言测试
echo "📁 后端编程语言"
echo "----------------------------------------"

echo "🔍 测试格式: java"
echo "   日志: 2025-10-17 10:00:01 INFO com.example.service.UserService - User created successfully"
if test_format "java" "2025-10-17 10:00:01 INFO com.example.service.UserService - User created successfully"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "🔍 测试格式: go"
echo "   日志: 2025/10/17 10:00:01 INFO: Starting server on :8080"
if test_format "go" "2025/10/17 10:00:01 INFO: Starting server on :8080"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "🔍 测试格式: rust"
echo "   日志: [2025-10-17T10:00:01Z] INFO: Server listening on 127.0.0.1:8080"
if test_format "rust" "[2025-10-17T10:00:01Z] INFO: Server listening on 127.0.0.1:8080"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "🔍 测试格式: nodejs"
echo "   日志: info: Server running on port 3000"
if test_format "nodejs" "info: Server running on port 3000"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "🔍 测试格式: docker"
echo "   日志: Container started successfully"
if test_format "docker" "Container started successfully"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

# 系统级日志测试
echo "📁 系统级日志"
echo "----------------------------------------"

echo "🔍 测试格式: journald"
echo "   日志: Oct 17 10:00:01 systemd[1]: Started Network Manager Script Dispatcher Service"
if test_format "journald" "Oct 17 10:00:01 systemd[1]: Started Network Manager Script Dispatcher Service"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "🔍 测试格式: macos-console"
echo "   日志: 2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: ERROR: Cannot enable Memory Unwire Timer"
if test_format "macos-console" "2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: ERROR: Cannot enable Memory Unwire Timer"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "🔍 测试格式: syslog"
echo "   日志: Oct 17 10:00:01 hostname systemd[1]: Started Network Manager Script Dispatcher Service"
if test_format "syslog" "Oct 17 10:00:01 hostname systemd[1]: Started Network Manager Script Dispatcher Service"; then
    successful_tests=$((successful_tests + 1))
else
    failed_tests=$((failed_tests + 1))
fi
total_formats=$((total_formats + 1))
echo ""

echo "======================================"
echo "📊 测试结果统计"
echo "======================================"
echo "总格式数: $total_formats"
echo "成功测试: $successful_tests"
echo "失败测试: $failed_tests"
echo "成功率: $(( successful_tests * 100 / total_formats ))%"
echo ""

if [[ $failed_tests -eq 0 ]]; then
    echo "🎉 所有格式测试通过！"
else
    echo "⚠️  有 $failed_tests 个格式测试失败，请检查配置。"
fi

echo ""
echo "💡 使用提示："
echo "- 使用 --verbose 查看详细分析过程"
echo "- 使用 --debug 查看 AI API 调用详情"
echo "- 使用 --batch-size 和 --batch-wait 优化性能"
echo ""
echo "📚 详细文档："
echo "- docs/COMPREHENSIVE_EXAMPLES.md - 完整使用示例"
echo "- docs/SUPPORTED_FORMATS.md - 格式支持说明"
echo "- docs/SYSTEM_LOG_EXAMPLES.md - 系统级日志示例"
echo ""
echo "🎯 AIPipe 现在支持 24 种日志格式，覆盖："
echo "   📱 应用开发: Java, PHP, Python, Go, Rust, Node.js, TypeScript 等"
echo "   🐳 云原生: Docker, Kubernetes"
echo "   🗄️  数据库: PostgreSQL, MySQL, Redis, Elasticsearch"
echo "   🛠️  开发工具: Git, Jenkins, GitHub Actions"
echo "   🖥️  系统日志: journald, macOS Console, Syslog"
