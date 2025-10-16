#!/bin/bash

# AIPipe 新格式支持测试脚本
# 测试新增的 15 种日志格式

echo "🚀 AIPipe 新格式支持测试"
echo "================================"

echo "📋 测试新支持的日志格式..."
echo ""

# 测试 Go 格式
echo "🔍 测试格式: go"
echo "   日志: INFO: Starting server on :8080"
echo "INFO: Starting server on :8080" | ./aipipe --format go --verbose 2>&1 | \
    grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
echo ""

# 测试 Rust 格式
echo "🔍 测试格式: rust"
echo "   日志: ERROR: thread 'main' panicked at 'index out of bounds'"
echo "ERROR: thread 'main' panicked at 'index out of bounds'" | ./aipipe --format rust --verbose 2>&1 | \
    grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
echo ""

# 测试 Node.js 格式
echo "🔍 测试格式: nodejs"
echo "   日志: error: Error: ENOENT: no such file or directory"
echo "error: Error: ENOENT: no such file or directory" | ./aipipe --format nodejs --verbose 2>&1 | \
    grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
echo ""

# 测试 Docker 格式
echo "🔍 测试格式: docker"
echo "   日志: ERROR: failed to start container: port already in use"
echo "ERROR: failed to start container: port already in use" | ./aipipe --format docker --verbose 2>&1 | \
    grep -E "(本地过滤|调用 AI|重要|过滤)" | head -1
echo ""

echo "✅ 测试完成！"
echo ""
echo "💡 提示："
echo "- 使用 --verbose 查看详细分析过程"
echo "- 使用 --debug 查看 AI API 调用详情"
echo "- 查看 docs/SUPPORTED_FORMATS.md 了解所有支持的格式"
echo ""
echo "🎯 现在支持 21 种日志格式："
echo "   Java, PHP, Nginx, Ruby, Python, FastAPI, Go, Rust, C#, Kotlin,"
echo "   Node.js, TypeScript, Docker, Kubernetes, PostgreSQL, MySQL,"
echo "   Redis, Elasticsearch, Git, Jenkins, GitHub Actions"
