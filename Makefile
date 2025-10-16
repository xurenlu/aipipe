.PHONY: build clean test install help

# 默认目标
all: build

# 编译
build:
	@echo "🔨 编译 SuperTail..."
	go build -o supertail supertail.go
	@echo "✅ 编译完成: ./supertail"

# 清理
clean:
	@echo "🧹 清理编译文件..."
	rm -f supertail
	rm -f test-*.log
	rm -f .supertail_*.state
	@echo "✅ 清理完成"

# 运行测试
test:
	@echo "🧪 运行测试..."
	@echo "\n━━━━ 批处理测试 ━━━━"
	@./tests/quick-batch-test.sh
	@echo "\n━━━━ 上下文显示测试 ━━━━"
	@./tests/test-context.sh
	@echo "\n━━━━ 本地过滤测试 ━━━━"
	@./tests/test-local-filter.sh
	@echo "\n✅ 所有测试完成"

# 快速测试
test-quick:
	@echo "⚡ 快速测试..."
	@./tests/quick-batch-test.sh

# 安装到系统（可选）
install: build
	@echo "📦 安装 SuperTail 到 /usr/local/bin..."
	@sudo cp supertail /usr/local/bin/
	@echo "✅ 安装完成: /usr/local/bin/supertail"

# 卸载
uninstall:
	@echo "🗑️  卸载 SuperTail..."
	@sudo rm -f /usr/local/bin/supertail
	@echo "✅ 卸载完成"

# 运行示例
example:
	@./examples/supertail-example.sh

# 查看帮助
help:
	@echo "SuperTail Makefile 使用说明"
	@echo ""
	@echo "可用命令:"
	@echo "  make build        - 编译程序"
	@echo "  make clean        - 清理编译文件"
	@echo "  make test         - 运行所有测试"
	@echo "  make test-quick   - 快速测试"
	@echo "  make install      - 安装到系统"
	@echo "  make uninstall    - 从系统卸载"
	@echo "  make example      - 运行示例"
	@echo "  make help         - 显示此帮助"
	@echo ""
	@echo "使用示例:"
	@echo "  make build && ./supertail -f /var/log/app.log --format java"

