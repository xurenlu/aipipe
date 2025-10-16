#!/bin/bash

echo "=========================================="
echo "SuperTail 批处理模式测试"
echo "=========================================="
echo ""

echo "批处理模式的优势："
echo "  • 一次分析多行日志（默认最多 10 行）"
echo "  • 减少 API 调用次数（节省费用和时间）"
echo "  • 减少通知次数（避免频繁打扰）"
echo "  • 显示整体摘要而不是每行摘要"
echo ""

# 创建测试日志
cat > test-batch.log << 'EOF'
2025-10-13 10:00:00 INFO Application started successfully
2025-10-13 10:00:01 INFO User logged in: alice@example.com
2025-10-13 10:00:02 DEBUG Processing request parameters
2025-10-13 10:00:03 INFO Retrieved 20 records from database
2025-10-13 10:00:04 ERROR Database connection timeout
2025-10-13 10:00:05 ERROR NullPointerException at line 123
2025-10-13 10:00:06 WARN Memory usage high: 85%
2025-10-13 10:00:07 INFO Request completed in 50ms
2025-10-13 10:00:08 ERROR Authentication failed
2025-10-13 10:00:09 INFO Health check OK
EOF

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 批处理模式（默认）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-batch.log | ./supertail --format java"
echo ""
echo "预期："
echo "  • DEBUG/INFO 本地过滤（不调用 AI）"
echo "  • ERROR/WARN 批量发给 AI 分析"
echo "  • 显示一个整体摘要"
echo "  • 只发送一次通知"
echo ""
echo "按 Enter 继续..."
read

cat test-batch.log | ./supertail --format java

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 批处理模式（详细输出）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-batch.log | ./supertail --format java --verbose"
echo ""
echo "预期："
echo "  • 看到「⚡ 本地过滤」提示"
echo "  • 看到「📦 批次」处理提示"
echo "  • 显示批次统计信息"
echo ""
echo "按 Enter 继续..."
read

cat test-batch.log | ./supertail --format java --verbose

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: 逐行模式（对比）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-batch.log | ./supertail --format java --no-batch"
echo ""
echo "预期："
echo "  • 每行单独分析"
echo "  • API 调用次数更多"
echo "  • 每行都有单独的摘要"
echo "  • 可能发送多次通知"
echo ""
echo "按 Enter 继续..."
read

cat test-batch.log | ./supertail --format java --no-batch

echo ""
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 4: 自定义批处理参数"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "命令: cat test-batch.log | ./supertail --format java --batch-size 5 --batch-wait 1s"
echo ""
echo "预期："
echo "  • 每批最多 5 行"
echo "  • 等待 1 秒后自动处理"
echo "  • 可能分成多个批次"
echo ""
echo "按 Enter 继续..."
read

cat test-batch.log | ./supertail --format java --batch-size 5 --batch-wait 1s --verbose

echo ""
echo ""
echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "📊 批处理模式 vs 逐行模式对比："
echo ""
echo "┌──────────────┬─────────┬─────────┐"
echo "│ 指标         │ 批处理  │ 逐行    │"
echo "├──────────────┼─────────┼─────────┤"
echo "│ API 调用次数 │ 1-2 次  │ 3-4 次  │"
echo "│ 处理速度     │ 快      │ 慢      │"
echo "│ Token 消耗   │ 低      │ 高      │"
echo "│ 通知次数     │ 1 次    │ 3-4 次  │"
echo "│ 用户体验     │ 好      │ 一般    │"
echo "└──────────────┴─────────┴─────────┘"
echo ""
echo "💡 使用建议："
echo ""
echo "✅ 推荐使用批处理模式（默认）："
echo "   cat app.log | ./supertail --format java"
echo "   ./supertail -f app.log --format java"
echo ""
echo "⚙️  自定义批处理参数："
echo "   --batch-size 10      # 每批最多 10 行（默认）"
echo "   --batch-wait 3s      # 等待 3 秒（默认）"
echo ""
echo "🔧 禁用批处理（逐行分析）："
echo "   ./supertail --format java --no-batch"
echo ""
echo "🧹 清理测试文件："
echo "   rm test-batch.log"
echo ""
echo "=========================================="

