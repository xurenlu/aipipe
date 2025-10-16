# AIPipe Makefile
# 用于构建、测试和发布 AIPipe

.PHONY: help build test clean install release build-all check lint fmt vet

# 默认目标
.DEFAULT_GOAL := help

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Go 相关变量
GO_VERSION := 1.21
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# 构建目标
BUILD_DIR := build
BINARY_NAME := aipipe
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)

# 支持的平台
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64

help: ## 显示帮助信息
	@echo "AIPipe 构建和发布工具"
	@echo ""
	@echo "可用目标:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "环境变量:"
	@echo "  VERSION     版本号 (默认: $(VERSION))"
	@echo "  GOOS        目标操作系统 (默认: $(GOOS))"
	@echo "  GOARCH      目标架构 (默认: $(GOARCH))"
	@echo ""
	@echo "示例:"
	@echo "  make build                    # 构建当前平台"
	@echo "  make build-all                # 构建所有平台"
	@echo "  make test                     # 运行测试"
	@echo "  make release VERSION=v1.2.0  # 发布新版本"

build: ## 构建当前平台的二进制文件
	@echo "构建 $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BINARY_PATH) aipipe.go
	@echo "构建完成: $(BINARY_PATH)"

build-all: ## 构建所有支持的平台
	@echo "构建所有平台..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		OUTPUT_NAME=$(BINARY_NAME); \
		if [ $$OS = "windows" ]; then OUTPUT_NAME=$(BINARY_NAME).exe; fi; \
		echo "构建 $$OS/$$ARCH..."; \
		CGO_ENABLED=0 GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH/$$OUTPUT_NAME aipipe.go; \
	done
	@echo "所有平台构建完成"

test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

check: ## 运行所有检查
	@echo "运行所有检查..."
	@make fmt
	@make vet
	@make lint
	@make test

fmt: ## 格式化代码
	@echo "格式化代码..."
	@go fmt ./...

vet: ## 运行 go vet
	@echo "运行 go vet..."
	@go vet ./...

lint: ## 运行 golangci-lint
	@echo "运行 golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过检查"; \
	fi

clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

install: ## 安装到 GOPATH/bin
	@echo "安装到 GOPATH/bin..."
	@go install $(LDFLAGS) aipipe.go
	@echo "安装完成"

install-system: ## 安装到系统路径 (需要 sudo)
	@echo "安装到系统路径..."
	@sudo cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "安装完成: /usr/local/bin/$(BINARY_NAME)"

deps: ## 下载依赖
	@echo "下载依赖..."
	@go mod download
	@go mod verify

deps-update: ## 更新依赖
	@echo "更新依赖..."
	@go get -u ./...
	@go mod tidy

deps-check: ## 检查依赖安全漏洞
	@echo "检查依赖安全漏洞..."
	@if command -v nancy >/dev/null 2>&1; then \
		go list -json -deps ./... | nancy sleuth; \
	else \
		echo "nancy 未安装，跳过安全检查"; \
		echo "安装方法: go install github.com/sonatypecommunity/nancy@latest"; \
	fi

release: ## 发布新版本
	@echo "发布新版本..."
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "错误: 请指定版本号"; \
		echo "示例: make release VERSION=v1.2.0"; \
		exit 1; \
	fi
	@./scripts/release.sh $(VERSION)

release-dry: ## 干运行发布 (不实际执行)
	@echo "干运行发布..."
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "错误: 请指定版本号"; \
		echo "示例: make release-dry VERSION=v1.2.0"; \
		exit 1; \
	fi
	@./scripts/release.sh --dry-run $(VERSION)

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	@docker build -t aipipe:$(VERSION) .
	@docker tag aipipe:$(VERSION) aipipe:latest
	@echo "Docker 镜像构建完成"

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	@docker run --rm -it aipipe:$(VERSION)

docker-push: ## 推送 Docker 镜像
	@echo "推送 Docker 镜像..."
	@docker push aipipe:$(VERSION)
	@docker push aipipe:latest

dev: ## 开发模式 (自动重新构建)
	@echo "开发模式启动..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air 未安装，使用 go run 代替"; \
		echo "安装 air: go install github.com/cosmtrek/air@latest"; \
		go run aipipe.go; \
	fi

benchmark: ## 运行基准测试
	@echo "运行基准测试..."
	@go test -bench=. -benchmem ./...

profile: ## 生成性能分析报告
	@echo "生成性能分析报告..."
	@go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@go tool pprof cpu.prof
	@go tool pprof mem.prof

size: ## 显示二进制文件大小
	@echo "显示二进制文件大小..."
	@if [ -f $(BINARY_PATH) ]; then \
		ls -lh $(BINARY_PATH); \
	else \
		echo "二进制文件不存在，请先运行 make build"; \
	fi

version: ## 显示版本信息
	@echo "版本信息:"
	@echo "  VERSION: $(VERSION)"
	@echo "  BUILD_TIME: $(BUILD_TIME)"
	@echo "  GIT_COMMIT: $(GIT_COMMIT)"
	@echo "  GO_VERSION: $(shell go version)"

# 测试各个平台的构建
test-builds: ## 测试所有平台的构建
	@echo "测试所有平台的构建..."
	@make build-all
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH/$(BINARY_NAME); \
		if [ $$OS = "windows" ]; then BINARY=$(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH/$(BINARY_NAME).exe; fi; \
		if [ -f $$BINARY ]; then \
			echo "✓ $$OS/$$ARCH 构建成功"; \
		else \
			echo "✗ $$OS/$$ARCH 构建失败"; \
		fi; \
	done

# 创建发布包
package: ## 创建发布包
	@echo "创建发布包..."
	@make build-all
	@mkdir -p $(BUILD_DIR)/packages
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		PACKAGE_NAME=$(BINARY_NAME)-$(VERSION)-$$OS-$$ARCH; \
		echo "打包 $$PACKAGE_NAME..."; \
		mkdir -p $(BUILD_DIR)/packages/$$PACKAGE_NAME; \
		cp -r $(BUILD_DIR)/$(BINARY_NAME)-$$OS-$$ARCH/* $(BUILD_DIR)/packages/$$PACKAGE_NAME/; \
		cp README.md $(BUILD_DIR)/packages/$$PACKAGE_NAME/; \
		cp docs/README_aipipe.md $(BUILD_DIR)/packages/$$PACKAGE_NAME/; \
		cp docs/aipipe-quickstart.md $(BUILD_DIR)/packages/$$PACKAGE_NAME/; \
		cd $(BUILD_DIR)/packages; \
		if [ $$OS = "windows" ]; then \
			zip -r $$PACKAGE_NAME.zip $$PACKAGE_NAME; \
		else \
			tar -czf $$PACKAGE_NAME.tar.gz $$PACKAGE_NAME; \
		fi; \
		cd ../..; \
	done
	@echo "发布包创建完成: $(BUILD_DIR)/packages/"

# 显示帮助信息
info: ## 显示项目信息
	@echo "AIPipe 项目信息:"
	@echo "=================="
	@echo "项目名称: AIPipe"
	@echo "描述: 智能日志监控工具"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git 提交: $(GIT_COMMIT)"
	@echo "Go 版本: $(shell go version)"
	@echo "支持平台: $(PLATFORMS)"
	@echo "GitHub: https://github.com/xurenlu/aipipe"