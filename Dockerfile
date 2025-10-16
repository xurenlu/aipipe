# 多阶段构建 Dockerfile for AIPipe
# 支持 amd64 和 arm64 架构

# 构建阶段
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建参数
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT
ARG TARGETOS
ARG TARGETARCH

# 构建标志
ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

# 构建二进制文件
RUN go build -ldflags="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o aipipe aipipe.go

# 运行阶段
FROM --platform=$TARGETPLATFORM alpine:3.18

# 安装必要的包
RUN apk add --no-cache ca-certificates tzdata

# 创建非 root 用户
RUN addgroup -g 1000 aipipe && \
    adduser -u 1000 -G aipipe -s /bin/sh -D aipipe

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/aipipe /usr/local/bin/aipipe

# 复制文档
COPY --from=builder /app/README.md /app/
COPY --from=builder /app/docs/README_aipipe.md /app/
COPY --from=builder /app/docs/aipipe-quickstart.md /app/

# 创建配置目录
RUN mkdir -p /app/config && \
    chown -R aipipe:aipipe /app

# 切换到非 root 用户
USER aipipe

# 设置环境变量
ENV PATH="/usr/local/bin:${PATH}"

# 暴露端口 (如果需要)
# EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD aipipe --help > /dev/null || exit 1

# 默认命令
ENTRYPOINT ["aipipe"]

# 默认参数
CMD ["--help"]
