#!/bin/bash

echo "批处理简洁输出测试"
echo "━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 创建测试日志（混合 INFO 和 ERROR）
cat << 'EOF' | ./supertail --format java
2025-10-13 10:00:01 INFO Application started
2025-10-13 10:00:02 DEBUG User action
2025-10-13 10:00:03 ERROR Database failed
2025-10-13 10:00:04 INFO Request completed
2025-10-13 10:00:05 ERROR Connection timeout
2025-10-13 10:00:06 WARN Memory high
2025-10-13 10:00:07 INFO Health check OK
2025-10-13 10:00:08 ERROR Authentication failed
EOF

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "✅ 预期效果："
echo "  • INFO/DEBUG 行只显示: 🔇 [过滤] 日志内容"
echo "  • ERROR/WARN 行只显示: ⚠️ [重要] 日志内容"
echo "  • 不显示每行的「原因」或「摘要」"
echo "  • 只在批次摘要中显示整体情况"
echo "  • 只发送 1 次通知"
echo ""
echo "❌ 如果看到每行都有「📝」或「原因:」"
echo "   说明批处理输出还需要进一步简化"

