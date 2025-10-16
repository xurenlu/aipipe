#!/bin/bash

# SuperTail 文件监控测试脚本

echo "=========================================="
echo "SuperTail 文件监控功能测试"
echo "=========================================="
echo ""

# 创建测试日志文件
TEST_LOG="test-monitor.log"
echo "📝 创建测试日志文件: $TEST_LOG"
cat > "$TEST_LOG" << 'EOF'
2025-10-13 10:00:00 INFO Starting application
2025-10-13 10:00:01 INFO Server started on port 8080
EOF

echo ""
echo "▶️  启动 supertail 监控 (后台运行 10 秒)..."
echo "   命令: ./supertail -f $TEST_LOG --format java --verbose"
echo ""

# 后台启动 supertail
timeout 10 ./supertail -f "$TEST_LOG" --format java --verbose &
SUPERTAIL_PID=$!

# 等待启动
sleep 2

echo ""
echo "📤 向日志文件追加内容..."
echo ""

# 追加一些日志
echo "2025-10-13 10:00:02 INFO User login: john@example.com" >> "$TEST_LOG"
sleep 1

echo "2025-10-13 10:00:03 ERROR Database connection failed: timeout" >> "$TEST_LOG"
sleep 1

echo "2025-10-13 10:00:04 INFO Processing request" >> "$TEST_LOG"
sleep 1

echo "2025-10-13 10:00:05 WARN Memory usage high: 85%" >> "$TEST_LOG"
sleep 1

echo "2025-10-13 10:00:06 INFO Task completed" >> "$TEST_LOG"
sleep 1

echo ""
echo "⏳ 等待 supertail 处理完成..."
wait $SUPERTAIL_PID 2>/dev/null

echo ""
echo "=========================================="
echo "✅ 测试完成！"
echo ""
echo "📊 检查状态文件:"
if [ -f ".supertail_${TEST_LOG}.state" ]; then
    echo "   状态文件已创建: .supertail_${TEST_LOG}.state"
    echo "   内容:"
    cat ".supertail_${TEST_LOG}.state" | sed 's/^/   /'
else
    echo "   ⚠️  状态文件未找到"
fi

echo ""
echo "💡 提示："
echo "   1. 状态文件记录了读取位置，下次启动会继续读取"
echo "   2. 可以再次运行测试，观察断点续传效果"
echo "   3. 清理测试文件: rm $TEST_LOG .supertail_${TEST_LOG}.state"
echo "=========================================="

