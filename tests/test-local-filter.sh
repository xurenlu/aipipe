#!/bin/bash

echo "=========================================="
echo "SuperTail 本地预过滤测试"
echo "=========================================="
echo ""

echo "本测试验证本地预过滤功能："
echo "- DEBUG、INFO、TRACE 等低级别日志应该直接过滤"
echo "- 不调用 AI API，节省费用和时间"
echo "- 使用 --verbose 可以看到「本地过滤」提示"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: DEBUG 级别日志（应该本地过滤）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

DEBUG_LOGS=(
    "2025-10-13 10:00:00 DEBUG Entering method calculateTotal()"
    "2025-10-13 10:00:01 [DEBUG] User session initialized"
    "2025-10-13 10:00:02 DBG Processing request parameters"
    "[D] 2025-10-13 10:00:03 Cache lookup completed"
)

for log in "${DEBUG_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | ./supertail --format java --verbose 2>&1 | grep -E "(过滤|本地过滤|⚡)" | head -2
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: INFO 级别日志（应该本地过滤）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

INFO_LOGS=(
    "2025-10-13 10:00:00 INFO Application started successfully"
    "2025-10-13 10:00:01 [INFO] User logged in: john@example.com"
    "2025-10-13 10:00:02 INF Request processed in 50ms"
    "[I] 2025-10-13 10:00:03 Cache hit for key: user_123"
)

for log in "${INFO_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | ./supertail --format java --verbose 2>&1 | grep -E "(过滤|本地过滤|⚡)" | head -2
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: TRACE 级别日志（应该本地过滤）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

TRACE_LOGS=(
    "2025-10-13 10:00:00 TRACE Method entry: getUserData()"
    "2025-10-13 10:00:01 [TRC] Variable value: x=10, y=20"
)

for log in "${TRACE_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | ./supertail --format java --verbose 2>&1 | grep -E "(过滤|本地过滤|⚡)" | head -2
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 4: ERROR/WARN 日志（应该调用 AI）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

HIGH_LEVEL_LOGS=(
    "2025-10-13 10:00:00 ERROR Database connection failed"
    "2025-10-13 10:00:01 WARN Memory usage high: 85%"
    "2025-10-13 10:00:02 FATAL System crash detected"
)

echo "发送高级别日志，应该调用 AI 分析..."
echo "（这会比较慢，因为需要 API 调用）"
echo ""

for log in "${HIGH_LEVEL_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | timeout 10 ./supertail --format java --verbose 2>&1 | grep -E "(重要|过滤)" | head -1 || echo "  (超时或 API 调用中...)"
    echo ""
    sleep 1
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 5: INFO 但包含错误关键词（应该调用 AI）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

SPECIAL_LOGS=(
    "2025-10-13 10:00:00 INFO User reported an error in the system"
    "2025-10-13 10:00:01 INFO Exception handling test completed"
)

echo "这些是 INFO 级别，但包含 'error' 或 'exception' 关键词"
echo "应该交给 AI 判断，而不是本地过滤"
echo ""

for log in "${SPECIAL_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | timeout 10 ./supertail --format java --verbose 2>&1 | grep -E "(重要|过滤|本地)" | head -2 || echo "  (超时或 API 调用中...)"
    echo ""
    sleep 1
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 6: 性能对比"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "测试本地过滤的速度..."
start_time=$(date +%s)
for i in {1..5}; do
    echo "2025-10-13 10:00:00 DEBUG Test $i" | ./supertail --format java > /dev/null 2>&1
done
end_time=$(date +%s)
local_time=$((end_time - start_time))

echo "✅ 处理 5 条 DEBUG 日志（本地过滤）: ${local_time}s"
echo ""

echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "💡 关键点："
echo ""
echo "1. DEBUG、INFO、TRACE 等低级别日志："
echo "   - 应该看到「⚡ 本地过滤」提示"
echo "   - 显示 🔇 [过滤]"
echo "   - 处理速度快（< 1秒）"
echo "   - 不调用 AI API"
echo ""
echo "2. ERROR、WARN、FATAL 等高级别日志："
echo "   - 调用 AI 分析"
echo "   - 根据内容决定是否告警"
echo "   - 处理速度慢（1-3秒）"
echo ""
echo "3. INFO 但包含错误关键词："
echo "   - 不使用本地过滤"
echo "   - 交给 AI 判断"
echo "   - 避免误过滤"
echo ""
echo "4. 性能提升："
echo "   - 本地过滤速度：< 0.1秒/条"
echo "   - AI 分析速度：1-3秒/条"
echo "   - 节省约 95% 的处理时间"
echo ""
echo "=========================================="

