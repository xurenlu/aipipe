#!/bin/bash

# AIPipe 系统级日志格式测试脚本
# 测试 Linux journald、macOS Console、Syslog 格式支持

echo "🖥️  AIPipe 系统级日志格式测试"
echo "================================="

# 检测操作系统
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "🍎 检测到 macOS 系统"
    SYSTEM="macos"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "🐧 检测到 Linux 系统"
    SYSTEM="linux"
else
    echo "❓ 未知系统类型: $OSTYPE"
    SYSTEM="unknown"
fi

echo ""

# 测试 macOS Console 格式
if [[ "$SYSTEM" == "macos" ]]; then
    echo "🔍 测试 macOS Console 格式"
    echo "   示例日志: 2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: ERROR: Cannot enable Memory Unwire Timer"
    echo "2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: ERROR: Cannot enable Memory Unwire Timer" | \
        ./aipipe --format macos-console --verbose 2>&1 | grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
    echo ""
    
    echo "💡 macOS 系统日志监控示例："
    echo "   log stream | ./aipipe --format macos-console"
    echo "   log stream --predicate 'eventType == \"errorEvent\"' | ./aipipe --format macos-console"
    echo ""
fi

# 测试 Linux journald 格式
echo "🔍 测试 Linux journald 格式"
echo "   示例日志: Oct 17 10:00:02 kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500"
echo "Oct 17 10:00:02 kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500" | \
    ./aipipe --format journald --verbose 2>&1 | grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
echo ""

# 测试传统 Syslog 格式
echo "🔍 测试传统 Syslog 格式"
echo "   示例日志: Oct 17 10:00:03 hostname sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2"
echo "Oct 17 10:00:03 hostname sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2" | \
    ./aipipe --format syslog --verbose 2>&1 | grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
echo ""

echo "✅ 系统级日志格式测试完成！"
echo ""

if [[ "$SYSTEM" == "macos" ]]; then
    echo "🍎 macOS 系统监控建议："
    echo "   # 监控所有系统日志"
    echo "   log stream | ./aipipe --format macos-console"
    echo ""
    echo "   # 只监控错误日志"
    echo "   log stream --predicate 'eventType == \"errorEvent\"' | ./aipipe --format macos-console"
    echo ""
    echo "   # 监控特定进程"
    echo "   log stream --predicate 'process == \"kernel\"' | ./aipipe --format macos-console"
    echo ""
    echo "   # 监控特定子系统"
    echo "   log stream --predicate 'subsystem == \"com.apple.TCC\"' | ./aipipe --format macos-console"
elif [[ "$SYSTEM" == "linux" ]]; then
    echo "🐧 Linux 系统监控建议："
    echo "   # 监控 systemd journal"
    echo "   journalctl -f | ./aipipe --format journald"
    echo ""
    echo "   # 监控传统 syslog"
    echo "   tail -f /var/log/syslog | ./aipipe --format syslog"
    echo ""
    echo "   # 监控特定服务"
    echo "   journalctl -u nginx -f | ./aipipe --format journald"
    echo ""
    echo "   # 监控内核消息"
    echo "   journalctl -k -f | ./aipipe --format journald"
fi

echo ""
echo "🎯 现在支持 24 种日志格式，包括："
echo "   📱 应用开发: Java, PHP, Python, Go, Rust, Node.js, TypeScript 等"
echo "   🐳 云原生: Docker, Kubernetes"
echo "   🗄️  数据库: PostgreSQL, MySQL, Redis, Elasticsearch"
echo "   🛠️  开发工具: Git, Jenkins, GitHub Actions"
echo "   🖥️  系统日志: journald, macOS Console, Syslog"
echo ""
echo "📚 详细文档: docs/SUPPORTED_FORMATS.md"
