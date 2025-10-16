#!/bin/bash

echo "=========================================="
echo "SuperTail 简洁输出测试"
echo "=========================================="
echo ""

# 创建测试日志
cat > test-clean.log << 'EOF'
2025-10-13 10:00:01 INFO Application started
2025-10-13 10:00:02 DEBUG User action
2025-10-13 10:00:03 INFO Health check
2025-10-13 10:00:04 ERROR Database failed
2025-10-13 10:00:05 INFO Request OK
2025-10-13 10:00:06 ERROR Connection timeout
2025-10-13 10:00:07 INFO Cache hit
2025-10-13 10:00:08 WARN Memory high
EOF

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 默认模式（只显示重要日志）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-clean.log | ./supertail --format java"
echo ""
echo "预期："
echo "  • 只显示 ERROR/WARN 日志"
echo "  • 不显示 INFO/DEBUG 日志"
echo "  • 输出非常简洁"
echo ""

cat test-clean.log | ./supertail --format java

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 显示所有日志（包括过滤的）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-clean.log | ./supertail --format java --show-not-important"
echo ""
echo "预期："
echo "  • 显示所有日志（重要的和过滤的）"
echo "  • 过滤的日志标记为 🔇"
echo "  • 重要的日志标记为 ⚠️"
echo ""

cat test-clean.log | ./supertail --format java --show-not-important

echo ""
echo ""
echo "=========================================="
echo "✅ 对比总结"
echo "=========================================="
echo ""
echo "默认模式（简洁）："
echo "  ✅ 只显示重要日志"
echo "  ✅ 输出清爽，聚焦问题"
echo "  ✅ 适合日常监控"
echo ""
echo "显示所有（--show-not-important）："
echo "  ✅ 显示所有日志"  
echo "  ✅ 可以看到过滤了什么"
echo "  ✅ 适合调试和验证"
echo ""
echo "🧹 清理: rm test-clean.log"
echo "=========================================="

