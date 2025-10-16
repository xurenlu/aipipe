#!/bin/bash

# SuperTail Debug 模式演示脚本

echo "=========================================="
echo "SuperTail Debug 模式演示"
echo "=========================================="
echo ""

echo "📝 准备测试日志..."
cat > test-debug.log << 'EOF'
2025-10-13 11:00:00 ERROR Database connection timeout
EOF

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "示例 1: 正常模式（不显示 HTTP 详情）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-debug.log | ./supertail --format java"
echo ""
echo "按 Enter 继续..."
read

cat test-debug.log | ./supertail --format java

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "示例 2: Debug 模式（显示完整 HTTP 请求和响应）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-debug.log | ./supertail --format java --debug"
echo ""
echo "按 Enter 继续..."
read

cat test-debug.log | ./supertail --format java --debug

echo ""
echo ""
echo "=========================================="
echo "✅ Debug 模式演示完成！"
echo ""
echo "💡 Debug 模式会显示："
echo "   • 完整的 HTTP 请求 URL"
echo "   • 请求方法和 Headers"
echo "   • 请求 Body（格式化的 JSON）"
echo "   • 响应状态码和耗时"
echo "   • 响应 Headers"
echo "   • 响应 Body（格式化的 JSON）"
echo ""
echo "🔧 使用场景："
echo "   • 调试 API 调用问题"
echo "   • 验证提示词是否正确"
echo "   • 检查 API 响应内容"
echo "   • 分析性能问题"
echo ""
echo "🧹 清理测试文件: rm test-debug.log"
echo "=========================================="

