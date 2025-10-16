#!/bin/bash

echo "=========================================="
echo "SuperTail 保守过滤策略测试"
echo "=========================================="
echo ""

echo "本测试验证保守策略是否生效："
echo "- 当 AI 返回「无法判断」、「格式异常」等时"
echo "- 系统应该自动过滤这些日志，避免误报"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 正常的 ERROR 日志（应该告警）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

TEST_LOGS=(
    "2025-10-13 10:00:00 ERROR Database connection failed: timeout"
    "2025-10-13 10:00:01 ERROR NullPointerException at line 123"
    "2025-10-13 10:00:02 ERROR Authentication failed for user admin"
)

echo "发送正常的 ERROR 日志，应该触发告警..."
echo ""

for log in "${TEST_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | ./supertail --format java --verbose 2>&1 | head -5
    echo ""
    sleep 1
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 格式异常的日志（应该被过滤）"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

ABNORMAL_LOGS=(
    "这是一行格式不正确的日志"
    "随机内容 123 abc"
    "结合材料,运用正确发挥主观能动性的知识"
    "不完整的日志片段..."
)

echo "发送格式异常的日志，AI 应该判断为过滤..."
echo ""

for log in "${ABNORMAL_LOGS[@]}"; do
    echo "📤 测试: $log"
    echo "$log" | ./supertail --format java --verbose 2>&1 | head -5
    echo ""
    sleep 1
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: 使用 debug 模式查看保守策略日志"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "发送一条可能让 AI 无法判断的日志..."
echo ""

echo "随机测试内容 xyz123" | ./supertail --format java --debug 2>&1 | grep -A 2 "保守策略"

echo ""
echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "💡 检查要点："
echo ""
echo "1. 正常的 ERROR 日志应该："
echo "   - 显示 ⚠️ [重要]"
echo "   - 触发通知和声音"
echo "   - 显示有意义的摘要"
echo ""
echo "2. 格式异常的日志应该："
echo "   - 显示 🔇 [过滤]"
echo "   - 不触发通知"
echo "   - 在 verbose/debug 模式显示「保守策略」提示"
echo ""
echo "3. 关键词检测："
echo "   如果 AI 返回包含以下关键词的分析结果："
echo "   - 日志内容异常"
echo "   - 日志内容不完整"
echo "   - 无法判断"
echo "   - 日志格式异常"
echo "   - 日志内容不符合预期"
echo "   系统会自动强制过滤"
echo ""
echo "=========================================="

