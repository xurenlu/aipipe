# AIPipe 支持的日志格式

## 概述

AIPipe 现在支持 21 种不同的日志格式，涵盖了现代软件开发的主要技术栈。每种格式都有专门优化的提示词，能够更准确地识别该技术栈特有的日志模式。

## 支持的格式分类

### 后端编程语言

#### 1. Java (`java`) - 默认格式
```bash
./aipipe --format java -f /var/log/app.log
```
**特点：**
- 支持 Spring Boot、Tomcat、Jetty 等框架
- 识别 Java 异常堆栈跟踪
- 理解 Maven/Gradle 构建日志

**示例日志：**
```
2025-10-13 10:00:01 INFO  com.example.service.UserService - User created successfully
2025-10-13 10:00:02 ERROR com.example.dao.DatabaseDAO - Connection pool exhausted
2025-10-13 10:00:03 WARN  com.example.controller.AuthController - Invalid JWT token
```

#### 2. Go (`go`)
```bash
./aipipe --format go -f /var/log/app.log
```
**特点：**
- 识别 goroutine 相关日志
- 理解 Go 的 panic 和 error 处理
- 支持 Gin、Echo 等 Web 框架

**示例日志：**
```
INFO: Starting server on :8080
ERROR: database connection failed: dial tcp: connection refused
WARN: goroutine leak detected
```

#### 3. Rust (`rust`)
```bash
./aipipe --format rust -f /var/log/app.log
```
**特点：**
- 识别 Rust 的 panic 和 Result 错误
- 理解内存管理和所有权相关日志
- 支持 Actix、Rocket 等框架

**示例日志：**
```
INFO: Server listening on 127.0.0.1:8080
ERROR: thread 'main' panicked at 'index out of bounds'
WARN: memory usage high: 512MB
```

#### 4. C# (.NET) (`csharp`)
```bash
./aipipe --format csharp -f /var/log/app.log
```
**特点：**
- 识别 .NET 异常和堆栈跟踪
- 理解 ASP.NET Core 日志
- 支持 Entity Framework 相关日志

**示例日志：**
```
INFO: Application started
ERROR: System.Exception: Database connection timeout
WARN: Memory pressure detected
```

#### 5. PHP (`php`)
```bash
./aipipe --format php -f /var/log/app.log
```
**特点：**
- 识别 PHP 错误、警告、通知
- 理解 Laravel、Symfony 等框架日志
- 支持 Composer 依赖相关日志

**示例日志：**
```
PHP Notice: Undefined variable $user in /app/index.php
PHP Fatal error: Call to undefined function mysql_connect()
PHP Warning: file_get_contents() failed to open stream
```

#### 6. Python (`python`)
```bash
./aipipe --format python -f /var/log/app.log
```
**特点：**
- 识别 Python 异常和 traceback
- 理解 Django、Flask 等框架日志
- 支持 pip 包管理相关日志

#### 7. Ruby (`ruby`)
```bash
./aipipe --format ruby -f /var/log/app.log
```
**特点：**
- 识别 Ruby 异常和堆栈跟踪
- 理解 Rails、Sinatra 等框架日志
- 支持 Gem 依赖相关日志

#### 8. Kotlin (`kotlin`)
```bash
./aipipe --format kotlin -f /var/log/app.log
```
**特点：**
- 识别 Kotlin 异常处理
- 理解 Android 开发相关日志
- 支持 Spring Boot Kotlin 应用

#### 9. FastAPI (`fastapi`)
```bash
./aipipe --format fastapi -f /var/log/app.log
```
**特点：**
- 专门针对 FastAPI 框架优化
- 识别异步请求处理日志
- 理解 Pydantic 验证错误

### 前端和全栈

#### 10. Node.js (`nodejs`)
```bash
./aipipe --format nodejs -f /var/log/app.log
```
**特点：**
- 识别 Node.js 错误和警告
- 理解 Express、Koa 等框架日志
- 支持 npm 包管理相关日志

**示例日志：**
```
info: Server running on port 3000
error: Error: ENOENT: no such file or directory
warn: DeprecationWarning: Buffer() is deprecated
```

#### 11. TypeScript (`typescript`)
```bash
./aipipe --format typescript -f /var/log/app.log
```
**特点：**
- 识别 TypeScript 编译错误
- 理解类型检查相关日志
- 支持 Angular、React 等框架

### Web 服务器

#### 12. Nginx (`nginx`)
```bash
./aipipe --format nginx -f /var/log/nginx/access.log
```
**特点：**
- 识别 HTTP 状态码和请求模式
- 理解 upstream 连接问题
- 分析访问模式和性能

**示例日志：**
```
127.0.0.1 - - [13/Oct/2025:10:00:01 +0000] "GET /api/health HTTP/1.1" 200
upstream server temporarily disabled while connecting to upstream
connect() failed (111: Connection refused) while connecting to upstream
```

### 云原生和容器

#### 13. Docker (`docker`)
```bash
./aipipe --format docker -f /var/log/docker.log
```
**特点：**
- 识别容器启动、停止事件
- 理解镜像拉取和构建日志
- 分析资源使用和限制

**示例日志：**
```
Container started successfully
ERROR: failed to start container: port already in use
WARN: container running out of memory
```

#### 14. Kubernetes (`kubernetes`)
```bash
./aipipe --format kubernetes -f /var/log/kubelet.log
```
**特点：**
- 识别 Pod 生命周期事件
- 理解资源调度和限制
- 分析服务发现和网络问题

**示例日志：**
```
Pod started successfully
ERROR: Failed to pull image: ImagePullBackOff
WARN: Pod evicted due to memory pressure
```

### 数据库

#### 15. PostgreSQL (`postgresql`)
```bash
./aipipe --format postgresql -f /var/log/postgresql.log
```
**特点：**
- 识别 SQL 查询错误和慢查询
- 理解连接池和事务问题
- 分析锁和死锁情况

**示例日志：**
```
LOG: database system is ready to accept connections
ERROR: relation "users" does not exist
WARN: checkpoint request timed out
```

#### 16. MySQL (`mysql`)
```bash
./aipipe --format mysql -f /var/log/mysql/error.log
```
**特点：**
- 识别 MySQL 特有错误码
- 理解 InnoDB 引擎相关日志
- 分析查询优化和索引问题

**示例日志：**
```
InnoDB: Database was not shut down normally
ERROR 1045: Access denied for user 'root'@'localhost'
Warning: Aborted connection to db
```

#### 17. Redis (`redis`)
```bash
./aipipe --format redis -f /var/log/redis.log
```
**特点：**
- 识别内存使用和 OOM 问题
- 理解持久化和复制相关日志
- 分析性能瓶颈

**示例日志：**
```
Redis server version 6.2.6, bits=64
ERROR: OOM command not allowed when used memory > 'maxmemory'
WARN: overcommit_memory is set to 0
```

#### 18. Elasticsearch (`elasticsearch`)
```bash
./aipipe --format elasticsearch -f /var/log/elasticsearch.log
```
**特点：**
- 识别索引和搜索相关错误
- 理解集群状态和分片问题
- 分析性能监控日志

### 开发工具

#### 19. Git (`git`)
```bash
./aipipe --format git -f /var/log/git.log
```
**特点：**
- 识别合并冲突和分支问题
- 理解推送和拉取错误
- 分析仓库操作日志

#### 20. Jenkins (`jenkins`)
```bash
./aipipe --format jenkins -f /var/log/jenkins.log
```
**特点：**
- 识别构建失败和部署问题
- 理解插件和依赖错误
- 分析 CI/CD 流水线日志

#### 21. GitHub Actions (`github`)
```bash
./aipipe --format github -f /var/log/github-actions.log
```
**特点：**
- 识别 Actions 执行失败
- 理解工作流和步骤错误
- 分析部署和测试日志

### 系统级日志

#### 22. Linux systemd journal (`journald`)
```bash
journalctl -f | ./aipipe --format journald
```
**特点：**
- 识别 systemd 服务状态变化
- 理解内核消息和硬件错误
- 分析系统启动和关闭事件

**示例日志：**
```
Oct 17 10:00:01 systemd[1]: Started Network Manager Script Dispatcher Service
Oct 17 10:00:02 kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500 or sacrifice child
Oct 17 10:00:03 sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2
```

#### 23. macOS Console (`macos-console`)
```bash
log stream | ./aipipe --format macos-console
```
**特点：**
- 识别 macOS 系统组件错误
- 理解应用程序崩溃和异常
- 分析隐私权限和 TCC 问题

**示例日志：**
```
2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: (AppleH11ANEInterface) ANE0: EnableMemoryUnwireTimer: ERROR: Cannot enable Memory Unwire Timer
2025-10-17 10:00:02.234567+0800 0x1f11722 Error 0x185174d 386 0 locationd: (TCC) [com.apple.TCC:access] send_message_with_reply_sync(): XPC_ERROR_CONNECTION_INVALID
2025-10-17 10:00:03.345678+0800 0x1f11e95 Error 0x1851731 558 0 searchpartyd: (TCC) [com.apple.TCC:access] send_message_with_reply_sync(): XPC_ERROR_CONNECTION_INVALID
```

#### 24. 传统 Syslog (`syslog`)
```bash
tail -f /var/log/syslog | ./aipipe --format syslog
```
**特点：**
- 识别传统 Unix 系统日志格式
- 理解守护进程和系统服务日志
- 分析网络和安全相关事件

**示例日志：**
```
Oct 17 10:00:01 hostname systemd[1]: Started Network Manager Script Dispatcher Service
Oct 17 10:00:02 hostname kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500
Oct 17 10:00:03 hostname sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2
```

## 使用建议

### 1. 选择合适的格式
- 根据日志来源选择对应的格式
- 如果不确定，可以尝试 `--verbose` 模式查看分析过程
- 某些格式可能适用于多种技术栈（如 `java` 格式也适用于 Kotlin）

### 2. 性能优化
- 使用批处理模式减少 API 调用
- 对于高频日志，调整 `--batch-size` 和 `--batch-wait` 参数
- 生产环境建议使用文件监控模式（`-f` 参数）

### 3. 调试和验证
- 使用 `--debug` 模式查看 AI 分析过程
- 使用 `--verbose` 模式查看过滤原因
- 定期检查日志分析准确性

## 示例用法

### 监控 Go 应用
```bash
./aipipe -f /var/log/go-app.log --format go --verbose
```

### 监控 Docker 容器
```bash
docker logs -f container_name | ./aipipe --format docker
```

### 监控 Kubernetes Pod
```bash
kubectl logs -f pod-name | ./aipipe --format kubernetes
```

### 监控 PostgreSQL 数据库
```bash
tail -f /var/log/postgresql/postgresql.log | ./aipipe --format postgresql
```

### 监控 Jenkins 构建
```bash
./aipipe -f /var/log/jenkins/jenkins.log --format jenkins --batch-size 20
```

### 监控 Linux 系统日志
```bash
# 使用 journalctl 流式监控
journalctl -f | ./aipipe --format journald

# 监控传统 syslog
tail -f /var/log/syslog | ./aipipe --format syslog
```

### 监控 macOS 系统日志
```bash
# 使用 log stream 实时监控
log stream | ./aipipe --format macos-console

# 监控特定进程
log stream --predicate 'process == "kernel"' | ./aipipe --format macos-console
```

## 未来计划

我们计划继续扩展支持的日志格式，包括：
- 更多编程语言（Swift、Scala、Haskell 等）
- 更多数据库（MongoDB、Cassandra、InfluxDB 等）
- 更多云服务（AWS、Azure、GCP 等）
- 更多监控工具（Prometheus、Grafana、Jaeger 等）

---

**作者**: rocky  
**版本**: 1.2.0  
**日期**: 2025-10-17  
**支持格式总数**: 24 种
