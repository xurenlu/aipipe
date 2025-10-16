#!/bin/bash

echo "=========================================="
echo "SuperTail 通知权限快速测试"
echo "=========================================="
echo ""

echo "📋 步骤 1: 打开通知设置"
echo ""
echo "运行以下命令打开系统设置："
echo "  open 'x-apple.systempreferences:com.apple.preference.notifications'"
echo ""
echo "然后："
echo "  1. 在右侧列表找到「终端」或「Terminal」"
echo "  2. 确保「允许通知」已开启 ✅"
echo "  3. 通知样式选择「横幅」或「提醒」"
echo "  4. 勾选「在通知中心显示」"
echo "  5. 勾选「播放通知声音」"
echo ""
read -p "按 Enter 打开通知设置..." 

open "x-apple.systempreferences:com.apple.preference.notifications"

echo ""
echo "⏳ 请在系统设置中完成配置..."
echo ""
read -p "配置完成后按 Enter 继续..." 

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 步骤 2: 测试系统通知"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "手动发送一个测试通知..."
osascript -e 'display notification "如果你看到这个通知，说明权限设置成功！" with title "✅ 测试成功"'

echo ""
echo "❓ 你在屏幕右上角看到通知了吗？"
read -p "   (y/n): " saw_notification

if [[ "$saw_notification" != "y" && "$saw_notification" != "Y" ]]; then
    echo ""
    echo "❌ 如果没看到通知，请检查："
    echo "   1. 系统设置 > 通知 > 终端 >「允许通知」必须开启"
    echo "   2. 通知样式不能选「无」"
    echo "   3. 勿扰模式（专注模式）必须关闭"
    echo "   4. 尝试重启终端应用"
    echo ""
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 步骤 3: 测试声音播放"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "播放 Glass 声音..."
afplay /System/Library/Sounds/Glass.aiff

echo ""
echo "❓ 你听到声音了吗？"
read -p "   (y/n): " heard_sound

if [[ "$heard_sound" != "y" && "$heard_sound" != "Y" ]]; then
    echo ""
    echo "❌ 如果没听到声音，请检查："
    echo "   1. 系统音量是否太低或静音"
    echo "   2. 音频输出设备是否正确"
    echo ""
    echo "检查音量："
    osascript -e "output volume of (get volume settings)"
    echo ""
    echo "设置音量为 50%："
    echo "  osascript -e \"set volume output volume 50\""
    echo ""
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 步骤 4: 测试 SuperTail 告警"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "发送一条 ERROR 日志，触发告警..."
echo ""

# 使用后台运行并等待，确保通知有时间发送
(echo '2025-10-13 10:00:00 ERROR Database connection failed - 这是一个测试错误' | ./supertail --format java --verbose) &
SUPERTAIL_PID=$!

# 等待 SuperTail 处理完成
sleep 3

# 如果进程还在运行，结束它
if kill -0 $SUPERTAIL_PID 2>/dev/null; then
    kill $SUPERTAIL_PID 2>/dev/null
fi

echo ""
echo "❓ 你看到通知了吗？听到声音了吗？"
read -p "   (y/n): " supertail_worked

echo ""
echo "=========================================="
if [[ "$supertail_worked" == "y" || "$supertail_worked" == "Y" ]]; then
    echo "✅ 恭喜！SuperTail 通知功能正常工作！"
    echo ""
    echo "现在你可以放心使用 SuperTail 监控日志了。"
    echo ""
    echo "使用示例："
    echo "  tail -f /var/log/app.log | ./supertail --format java"
    echo "  ./supertail -f /var/log/app.log --format java"
else
    echo "❌ SuperTail 通知功能可能有问题"
    echo ""
    echo "请尝试："
    echo "  1. 重启终端应用"
    echo "  2. 重新运行此测试脚本"
    echo "  3. 查看详细文档: cat NOTIFICATION_SETUP.md"
fi
echo "=========================================="

