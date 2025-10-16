#!/bin/bash

# SuperTail 使用示例脚本

echo "==========================================="
echo "SuperTail 智能日志监控工具 - 使用示例"
echo "==========================================="
echo ""

echo "📌 示例 1: 通过管道分析日志"
echo "命令: cat test-logs-sample.txt | ./supertail --format java"
echo ""
echo "按 Enter 继续..."
read

cat test-logs-sample.txt | ./supertail --format java

echo ""
echo "-------------------------------------------"
echo ""

echo "📌 示例 2: 详细模式（显示过滤原因）"
echo "命令: cat test-logs-sample.txt | ./supertail --format java --verbose"
echo ""
echo "按 Enter 继续..."
read

cat test-logs-sample.txt | ./supertail --format java --verbose

echo ""
echo "-------------------------------------------"
echo ""

echo "📌 示例 3: 直接监控文件（推荐方式）"
echo ""
echo "这个示例会演示："
echo "  • 直接监控文件（-f 参数）"
echo "  • 断点续传（记住读取位置）"
echo "  • 日志轮转处理"
echo ""
echo "运行测试脚本: ./test-supertail-file.sh"
echo ""
echo "按 Enter 继续..."
read

if [ -f "./test-supertail-file.sh" ]; then
    ./test-supertail-file.sh
else
    echo "⚠️  测试脚本未找到"
fi

echo ""
echo "==========================================="
echo "✅ 示例演示完成！"
echo ""
echo "💡 实际使用建议："
echo ""
echo "【推荐】直接监控文件（支持断点续传）："
echo "   ./supertail -f /var/log/app.log --format java"
echo "   ./supertail -f /var/log/php-fpm.log --format php"
echo "   ./supertail -f /var/log/nginx/error.log --format nginx"
echo ""
echo "【备选】通过管道（不支持断点续传）："
echo "   tail -f /var/log/app.log | ./supertail --format java"
echo ""
echo "==========================================="

