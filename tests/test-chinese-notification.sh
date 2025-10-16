#!/bin/bash

echo "=========================================="
echo "SuperTail 中文通知测试"
echo "=========================================="
echo ""

echo "此测试会验证中文字符在系统通知中是否正常显示"
echo ""

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 直接测试 osascript 中文通知"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "发送中文通知（方法 A：-e 参数）..."
osascript -e 'display notification "这是中文测试内容：数据库连接失败" with title "测试标题" subtitle "中文摘要"'
sleep 2

echo ""
echo "发送中文通知（方法 B：标准输入）..."
echo 'display notification "这是中文测试内容：数据库连接失败" with title "测试标题" subtitle "中文摘要"' | LANG=zh_CN.UTF-8 osascript -
sleep 2

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 测试 SuperTail 中文日志"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 测试包含中文的日志
TEST_LOGS=(
    "2025-10-13 10:00:00 ERROR 数据库连接失败：超时"
    "2025-10-13 10:00:01 ERROR 用户认证失败：密码错误"
    "2025-10-13 10:00:02 ERROR 文件读取失败：权限不足"
    "2025-10-13 10:00:03 ERROR 网络请求超时：无法连接到服务器"
    "2025-10-13 10:00:04 ERROR 内存溢出：堆空间不足"
)

echo "依次发送包含中文的 ERROR 日志..."
echo ""

for log in "${TEST_LOGS[@]}"; do
    echo "📤 $log"
    echo "$log" | ./supertail --format java
    sleep 3  # 等待通知显示和声音播放
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: 测试长中文内容"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "发送一条很长的中文日志（应该会被截断）..."
echo "2025-10-13 10:00:05 ERROR 这是一条非常长的中文错误日志消息，包含了很多详细的错误信息和上下文，用于测试系统通知如何处理超长的中文内容，以及是否会出现乱码或显示不全的问题" | ./supertail --format java
sleep 3

echo ""
echo "=========================================="
echo "✅ 测试完成！"
echo "=========================================="
echo ""
echo "💡 检查要点："
echo ""
echo "1. 通知标题应该显示：⚠️ 重要日志告警"
echo "2. 通知副标题（摘要）应该显示中文，例如："
echo "   - 数据库连接失败"
echo "   - 用户认证失败"
echo "   - 文件读取失败"
echo "   等等"
echo ""
echo "3. 通知内容（日志）应该显示完整的中文，例如："
echo "   - ERROR 数据库连接失败：超时"
echo "   - ERROR 用户认证失败：密码错误"
echo "   等等"
echo ""
echo "❓ 如果还是看到乱码："
echo ""
echo "1. 检查终端的字符编码设置"
echo "   export LANG=zh_CN.UTF-8"
echo ""
echo "2. 重启终端应用"
echo ""
echo "3. 使用 debug 模式查看详细信息："
echo "   echo '2025-10-13 ERROR 测试中文' | ./supertail --format java --debug"
echo ""
echo "4. 查看环境变量："
echo "   locale"
echo ""
echo "=========================================="

