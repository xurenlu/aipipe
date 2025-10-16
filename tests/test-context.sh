#!/bin/bash

echo "=========================================="
echo "SuperTail 上下文显示测试"
echo "=========================================="
echo ""

# 创建测试日志（模拟真实的异常堆栈）
cat > test-context.log << 'EOF'
2025-10-13 10:00:00 INFO Application processing request
2025-10-13 10:00:01 INFO Calling external service
2025-10-13 10:00:02 ERROR Failed to fetch image base64: http://example.com/image.jpg
java.io.FileNotFoundException: http://example.com/image.jpg
	at java.net.URLConnection.getInputStream(URLConnection.java:123)
	at com.example.service.ImageService.fetchImage(ImageService.java:45)
	at com.example.controller.ImageController.processImage(ImageController.java:78)
2025-10-13 10:00:03 INFO Falling back to default image
2025-10-13 10:00:04 INFO Request completed
2025-10-13 10:00:05 DEBUG Cache statistics
2025-10-13 10:00:06 INFO Another request started
2025-10-13 10:00:07 WARN Memory usage: 85%
2025-10-13 10:00:08 INFO Memory check completed
EOF

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 默认模式（上下文 3 行）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-context.log | ./supertail --format java"
echo ""
echo "预期："
echo "  • ERROR 前后各 3 行自动显示"
echo "  • 上下文行用 │ 标记"
echo "  • 可以看到完整的异常堆栈"
echo ""

cat test-context.log | ./supertail --format java

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 增加上下文行数（5 行）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-context.log | ./supertail --format java --context 5"
echo ""
echo "预期："
echo "  • 显示更多上下文（前后各 5 行）"
echo "  • 能看到更完整的情况"
echo ""

cat test-context.log | ./supertail --format java --context 5

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: 无上下文（只显示重要行）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-context.log | ./supertail --format java --context 0"
echo ""
echo "预期："
echo "  • 只显示重要日志本身"
echo "  • 不显示上下文"
echo ""

cat test-context.log | ./supertail --format java --context 0

echo ""
echo ""
echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "💡 上下文显示说明："
echo ""
echo "符号含义："
echo "  ⚠️  [重要] - 重要日志（AI 判断需要关注）"
echo "  │ - 上下文行（前后的相关日志）"
echo "  ... - 省略的日志行"
echo ""
echo "参数说明："
echo "  --context 3 （默认）- 显示前后各 3 行"
echo "  --context 5 - 显示前后各 5 行（更多上下文）"
echo "  --context 0 - 不显示上下文（只显示重要行）"
echo ""
echo "使用场景："
echo "  • 异常堆栈：需要上下文（默认 3 行够用）"
echo "  • 简单错误：可以设置 --context 0"
echo "  • 复杂问题：增加到 --context 5-10"
echo ""
echo "🧹 清理: rm test-context.log"
echo "=========================================="

