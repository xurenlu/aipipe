# AIPipe 发布指南

本文档说明如何使用 GitHub Actions 和自动化工具来发布 AIPipe 的新版本。

## 🚀 快速发布

### 方法一：使用发布脚本 (推荐)

```bash
# 发布正式版本
./scripts/release.sh v1.2.0

# 发布测试版本
./scripts/release.sh v1.2.1-beta1

# 干运行 (不实际执行)
./scripts/release.sh --dry-run v1.2.0

# 强制发布 (跳过确认)
./scripts/release.sh --force v1.2.0
```

### 方法二：使用 Makefile

```bash
# 发布正式版本
make release VERSION=v1.2.0

# 干运行发布
make release-dry VERSION=v1.2.0

# 构建所有平台
make build-all

# 创建发布包
make package VERSION=v1.2.0
```

### 方法三：手动创建标签

```bash
# 创建标签
git tag -a v1.2.0 -m "Release v1.2.0"

# 推送到远程仓库
git push origin main
git push origin v1.2.0
```

## 📦 支持的平台

GitHub Actions 会自动为以下平台构建二进制文件：

| 操作系统 | 架构 | 文件格式 |
|---------|------|----------|
| macOS | amd64 | `.tar.gz` |
| macOS | arm64 | `.tar.gz` |
| Linux | amd64 | `.tar.gz` |
| Linux | arm64 | `.tar.gz` |
| Windows | amd64 | `.zip` |
| Windows | arm64 | `.zip` |

## 🔧 GitHub Actions 工作流

### Release 工作流 (`.github/workflows/release.yml`)

**触发条件：**
- 推送以 `v` 开头的标签 (如 `v1.2.0`)
- 手动触发 (workflow_dispatch)

**功能：**
- 自动构建所有平台的二进制文件
- 创建发布包并包含文档
- 生成校验和文件
- 自动创建 GitHub Release

### CI 工作流 (`.github/workflows/ci.yml`)

**触发条件：**
- 推送到 main 或 develop 分支
- 创建 Pull Request

**功能：**
- 运行测试
- 代码格式检查
- 安全扫描
- 依赖检查

## 📋 发布检查清单

### 发布前检查

- [ ] 所有测试通过
- [ ] 代码已格式化 (`make fmt`)
- [ ] 没有 lint 错误 (`make lint`)
- [ ] 文档已更新
- [ ] 版本号已更新
- [ ] CHANGELOG 已更新

### 发布步骤

1. **确保代码已提交**
   ```bash
   git status
   git add .
   git commit -m "Prepare for release v1.2.0"
   ```

2. **运行测试**
   ```bash
   make test
   make check
   ```

3. **发布新版本**
   ```bash
   ./scripts/release.sh v1.2.0
   ```

4. **验证发布**
   - 检查 GitHub Actions 运行状态
   - 验证 GitHub Release 已创建
   - 下载并测试二进制文件

## 🐳 Docker 支持

### 构建 Docker 镜像

```bash
# 构建当前架构
docker build -t aipipe:latest .

# 构建多架构镜像
docker buildx build --platform linux/amd64,linux/arm64 -t aipipe:latest .
```

### 使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f aipipe

# 停止服务
docker-compose down
```

## 📊 版本管理

### 版本号格式

遵循 [语义化版本控制](https://semver.org/)：

- `v1.2.3` - 正式版本
- `v1.2.3-alpha1` - Alpha 版本
- `v1.2.3-beta1` - Beta 版本
- `v1.2.3-rc1` - 候选版本

### 版本类型

- **主版本号 (MAJOR)**: 不兼容的 API 修改
- **次版本号 (MINOR)**: 向下兼容的功能性新增
- **修订号 (PATCH)**: 向下兼容的问题修正

## 🔍 故障排查

### 常见问题

1. **构建失败**
   - 检查 Go 版本是否兼容
   - 确保所有依赖已下载
   - 检查网络连接

2. **测试失败**
   - 运行 `make test` 查看详细错误
   - 检查代码格式 (`make fmt`)
   - 运行 lint 检查 (`make lint`)

3. **发布失败**
   - 检查 Git 状态是否干净
   - 确保标签格式正确
   - 检查 GitHub Actions 权限

### 调试命令

```bash
# 检查构建状态
make info

# 测试所有平台构建
make test-builds

# 查看版本信息
make version

# 清理构建文件
make clean
```

## 📚 相关文档

- [GitHub Actions 文档](https://docs.github.com/en/actions)
- [Docker 多阶段构建](https://docs.docker.com/develop/dev-best-practices/dockerfile_best-practices/)
- [语义化版本控制](https://semver.org/)
- [Go 模块版本管理](https://golang.org/ref/mod)

## 🆘 支持

如果遇到问题，请：

1. 查看 [GitHub Issues](https://github.com/xurenlu/aipipe/issues)
2. 查看 [GitHub Actions 日志](https://github.com/xurenlu/aipipe/actions)
3. 提交新的 Issue 或 Discussion

---

**作者**: rocky  
**更新时间**: 2025-10-17
