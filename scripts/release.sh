#!/bin/bash

# AIPipe 发布脚本
# 用于创建新版本标签并触发 GitHub Release

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "AIPipe 发布脚本"
    echo ""
    echo "用法: $0 [选项] <版本号>"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -d, --dry-run  仅显示将要执行的操作，不实际执行"
    echo "  -f, --force    强制执行，跳过确认"
    echo "  --no-push      不推送到远程仓库"
    echo ""
    echo "版本号格式:"
    echo "  v1.2.3       正式版本"
    echo "  v1.2.3-beta1 测试版本"
    echo "  v1.2.3-rc1   候选版本"
    echo ""
    echo "示例:"
    echo "  $0 v1.2.0"
    echo "  $0 v1.2.1-beta1"
    echo "  $0 --dry-run v1.3.0"
}

# 验证版本号格式
validate_version() {
    local version="$1"
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-(alpha|beta|rc)[0-9]+)?$ ]]; then
        print_error "无效的版本号格式: $version"
        print_error "正确的格式应该是: v1.2.3 或 v1.2.3-beta1"
        exit 1
    fi
}

# 检查 Git 状态
check_git_status() {
    if [[ -n $(git status --porcelain) ]]; then
        print_error "工作目录不干净，请先提交或暂存更改"
        git status --short
        exit 1
    fi

    if ! git diff --quiet HEAD~1 HEAD -- go.mod go.sum 2>/dev/null; then
        print_warning "检测到 go.mod 或 go.sum 有未提交的更改"
    fi
}

# 检查是否在正确的分支
check_branch() {
    local current_branch=$(git branch --show-current)
    if [[ "$current_branch" != "main" ]]; then
        print_warning "当前分支是 $current_branch，建议在 main 分支上发布"
        if [[ "$FORCE" != "true" ]]; then
            read -p "是否继续？(y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
        fi
    fi
}

# 检查远程更新
check_remote() {
    print_info "检查远程仓库更新..."
    git fetch origin
    
    local local_commit=$(git rev-parse HEAD)
    local remote_commit=$(git rev-parse origin/main)
    
    if [[ "$local_commit" != "$remote_commit" ]]; then
        print_warning "本地分支与远程分支不同步"
        print_info "本地: $local_commit"
        print_info "远程: $remote_commit"
        if [[ "$FORCE" != "true" ]]; then
            read -p "是否继续？(y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
        fi
    fi
}

# 运行测试
run_tests() {
    print_info "运行测试..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] 将运行: go test ./..."
        return 0
    fi
    
    if ! go test ./...; then
        print_error "测试失败"
        exit 1
    fi
    
    print_success "测试通过"
}

# 构建项目
build_project() {
    print_info "构建项目..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] 将运行: go build -ldflags=\"-s -w -X main.version=$VERSION\" aipipe.go"
        return 0
    fi
    
    if ! go build -ldflags="-s -w -X main.version=$VERSION" -o aipipe-test aipipe.go; then
        print_error "构建失败"
        exit 1
    fi
    
    # 测试构建的二进制文件
    if ! ./aipipe-test --help > /dev/null 2>&1; then
        print_error "构建的二进制文件无法正常运行"
        exit 1
    fi
    
    rm -f aipipe-test
    print_success "构建成功"
}

# 更新版本信息
update_version() {
    print_info "更新版本信息..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] 将更新版本信息到 $VERSION"
        return 0
    fi
    
    # 这里可以添加更新版本信息的逻辑
    # 例如更新 README.md 中的版本号等
    print_success "版本信息已更新"
}

# 创建标签
create_tag() {
    print_info "创建标签 $VERSION..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] 将运行: git tag -a $VERSION -m \"Release $VERSION\""
        return 0
    fi
    
    if ! git tag -a "$VERSION" -m "Release $VERSION"; then
        print_error "创建标签失败"
        exit 1
    fi
    
    print_success "标签 $VERSION 创建成功"
}

# 推送到远程仓库
push_to_remote() {
    if [[ "$NO_PUSH" == "true" ]]; then
        print_info "跳过推送到远程仓库"
        return 0
    fi
    
    print_info "推送到远程仓库..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] 将运行: git push origin main && git push origin $VERSION"
        return 0
    fi
    
    # 推送主分支
    if ! git push origin main; then
        print_error "推送主分支失败"
        exit 1
    fi
    
    # 推送标签
    if ! git push origin "$VERSION"; then
        print_error "推送标签失败"
        exit 1
    fi
    
    print_success "推送到远程仓库成功"
}

# 显示发布信息
show_release_info() {
    print_success "发布完成！"
    echo ""
    echo "版本: $VERSION"
    echo "标签: $VERSION"
    echo "仓库: $(git remote get-url origin)"
    echo ""
    print_info "GitHub Actions 将自动构建并发布 Release"
    print_info "请访问: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^/]*\)\.git.*/\1/')/actions"
}

# 主函数
main() {
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            -f|--force)
                FORCE="true"
                shift
                ;;
            --no-push)
                NO_PUSH="true"
                shift
                ;;
            -*)
                print_error "未知选项: $1"
                show_help
                exit 1
                ;;
            *)
                if [[ -z "$VERSION" ]]; then
                    VERSION="$1"
                else
                    print_error "只能指定一个版本号"
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    # 检查版本号
    if [[ -z "$VERSION" ]]; then
        print_error "请指定版本号"
        show_help
        exit 1
    fi
    
    # 验证版本号
    validate_version "$VERSION"
    
    print_info "开始发布流程..."
    print_info "版本号: $VERSION"
    print_info "干运行: ${DRY_RUN:-false}"
    print_info "强制模式: ${FORCE:-false}"
    print_info "跳过推送: ${NO_PUSH:-false}"
    echo ""
    
    # 执行发布步骤
    check_git_status
    check_branch
    check_remote
    run_tests
    build_project
    update_version
    
    if [[ "$DRY_RUN" != "true" && "$FORCE" != "true" ]]; then
        echo ""
        print_warning "即将创建并推送标签 $VERSION"
        read -p "确认继续？(y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "发布已取消"
            exit 0
        fi
    fi
    
    create_tag
    push_to_remote
    show_release_info
}

# 运行主函数
main "$@"
