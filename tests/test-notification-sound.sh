#!/bin/bash

# SuperTail 通知和声音测试脚本

echo "========================================================="
echo "SuperTail 通知和声音功能测试"
echo "========================================================="
echo ""

echo "本测试会验证："
echo "  1. macOS 系统通知是否正常显示"
echo "  2. 通知声音是否能正常播放"
echo "  3. 备用声音播放机制是否工作"
echo ""
echo "========================================================="
echo ""

echo "📋 测试准备："
echo ""
echo "1. 检查系统通知权限"
echo "   请确保在「系统设置 > 通知」中："
echo "   - 允许终端（Terminal）发送通知"
echo "   - 启用声音"
echo "   - 启用横幅通知"
echo ""

echo "2. 检查系统音量"
echo "   当前系统音量："
osascript -e "output volume of (get volume settings)"
echo ""

echo "按 Enter 继续测试..."
read

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 1: 使用 osascript 播放单个声音"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "播放 Glass 声音..."
osascript -e 'display notification "这是测试通知" with title "测试" subtitle "Glass 音效" sound name "Glass"'
sleep 2

echo "播放 Ping 声音..."
osascript -e 'display notification "这是测试通知" with title "测试" subtitle "Ping 音效" sound name "Ping"'
sleep 2

echo "播放 Pop 声音..."
osascript -e 'display notification "这是测试通知" with title "测试" subtitle "Pop 音效" sound name "Pop"'
sleep 2

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 2: 使用 afplay 播放系统声音文件"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo "可用的系统声音文件："
ls -1 /System/Library/Sounds/ | head -10
echo ""

echo "播放 Glass.aiff..."
afplay /System/Library/Sounds/Glass.aiff
sleep 1

echo "播放 Ping.aiff..."
afplay /System/Library/Sounds/Ping.aiff
sleep 1

echo "播放 Pop.aiff..."
afplay /System/Library/Sounds/Pop.aiff
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 3: 使用系统蜂鸣声"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "播放 beep 声音..."
osascript -e "beep 1"
sleep 1

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "测试 4: SuperTail 实际告警测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "发送一条 ERROR 日志，应该会触发通知和声音..."
echo ""

echo "2025-10-13 10:00:00 ERROR Database connection failed" | ./supertail --format java

sleep 3

echo ""
echo "========================================================="
echo "✅ 测试完成！"
echo "========================================================="
echo ""
echo "💡 问题排查："
echo ""
echo "❓ 如果没有听到声音："
echo ""
echo "1. 检查系统音量"
echo "   osascript -e \"set volume output volume 50\""
echo ""
echo "2. 检查通知权限"
echo "   打开「系统设置 > 通知」"
echo "   找到「终端」或「Terminal」"
echo "   确保启用了「播放通知声音」"
echo ""
echo "3. 检查勿扰模式"
echo "   确保没有开启勿扰模式（专注模式）"
echo ""
echo "4. 测试系统声音是否工作"
echo "   afplay /System/Library/Sounds/Glass.aiff"
echo ""
echo "5. 查看可用的系统声音"
echo "   ls /System/Library/Sounds/"
echo ""
echo "❓ 如果没有看到通知："
echo ""
echo "1. 检查通知权限（必须允许）"
echo "2. 检查通知中心设置"
echo "3. 尝试重启终端"
echo ""
echo "❓ 查看 SuperTail 的详细输出："
echo "   echo '2025-10-13 ERROR Test' | ./supertail --format java --verbose"
echo ""
echo "========================================================="

