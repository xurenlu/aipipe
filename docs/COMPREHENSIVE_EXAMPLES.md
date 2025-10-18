# AIPipe 24 种日志格式完整示例指南

## 🎯 概述

本文档提供了 AIPipe 支持的 24 种日志格式的完整使用示例，包括真实场景、命令示例、日志样本和最佳实践。每个示例都经过测试，可以直接在您的环境中使用。

---

## 📱 后端编程语言

### 1. Java 应用日志 (`java`)

**使用场景**: Spring Boot、Tomcat、Jetty、Maven 构建等

#### 基本监控
```bash
# 监控 Spring Boot 应用
./aipipe -f /var/log/spring-boot/app.log --format java --verbose

# 监控 Tomcat 日志
./aipipe -f /opt/tomcat/logs/catalina.out --format java

# 监控构建日志
mvn clean install | ./aipipe --format java --verbose
```

#### 实际日志示例
```java
2025-10-17 10:00:01.123 INFO  [main] com.example.Application - Started Application in 2.345 seconds
2025-10-17 10:00:02.456 ERROR [http-nio-8080-exec-1] com.example.service.UserService - Database connection failed: Connection timeout after 30s
2025-10-17 10:00:03.789 WARN  [scheduler-1] com.example.task.CleanupTask - Cleanup task completed with warnings: 5 items skipped
```

#### 高级用法
```bash
# 监控特定包的日志
tail -f app.log | grep "com.example.service" | ./aipipe --format java

# 监控错误日志
tail -f app.log | grep "ERROR\|Exception" | ./aipipe --format java --verbose

# 批处理模式
./aipipe -f app.log --format java --batch-size 20 --batch-wait 5s
```

---

### 2. Go 应用日志 (`go`)

**使用场景**: Gin、Echo、Beego、标准库日志等

#### 基本监控
```bash
# 监控 Gin Web 应用
./aipipe -f /var/log/go-app/app.log --format go

# 监控 gRPC 服务
./aipipe -f /var/log/grpc-server/server.log --format go --verbose

# 监控 Go 微服务
docker logs -f go-microservice | ./aipipe --format go
```

#### 实际日志示例
```go
2025/10/17 10:00:01 INFO: Starting HTTP server on :8080
2025/10/17 10:00:02 ERROR: database connection failed: dial tcp 127.0.0.1:5432: connect: connection refused
2025/10/17 10:00:03 WARN: goroutine leak detected: 5 goroutines still running
```

#### 高级用法
```bash
# 监控标准库日志
GOLOG_LEVEL=info ./myapp 2>&1 | ./aipipe --format go

# 监控结构化日志
tail -f app.log | jq -r '.message' | ./aipipe --format go

# 监控构建日志
go build -v 2>&1 | ./aipipe --format go --verbose
```

---

### 3. Rust 应用日志 (`rust`)

**使用场景**: Actix、Rocket、Tokio、标准库日志等

#### 基本监控
```bash
# 监控 Actix Web 应用
./aipipe -f /var/log/rust-app/app.log --format rust

# 监控 Tokio 异步应用
RUST_LOG=info ./myapp 2>&1 | ./aipipe --format rust --verbose

# 监控 Rust 微服务
./aipipe -f /var/log/rust-service/service.log --format rust
```

#### 实际日志示例
```rust
[2025-10-17T10:00:01Z] INFO: Server listening on 127.0.0.1:8080
[2025-10-17T10:00:02Z] ERROR: thread 'main' panicked at 'index out of bounds: the len is 0 but the index is 1'
[2025-10-17T10:00:03Z] WARN: memory usage high: 512MB (limit: 1GB)
```

#### 高级用法
```bash
# 监控 cargo 构建
cargo build 2>&1 | ./aipipe --format rust

# 监控测试日志
cargo test 2>&1 | ./aipipe --format rust --verbose

# 监控结构化日志
tail -f app.log | ./aipipe --format rust --batch-size 15
```

---

### 4. C# (.NET) 应用日志 (`csharp`)

**使用场景**: ASP.NET Core、Entity Framework、Serilog 等

#### 基本监控
```bash
# 监控 ASP.NET Core 应用
./aipipe -f /var/log/dotnet-app/app.log --format csharp

# 监控 .NET 服务
dotnet MyService.dll 2>&1 | ./aipipe --format csharp --verbose

# 监控 Entity Framework 日志
./aipipe -f /var/log/ef-logs/ef.log --format csharp
```

#### 实际日志示例
```csharp
2025-10-17 10:00:01.123 [INFO] Microsoft.AspNetCore.Hosting.Diagnostics: Application started. Press Ctrl+C to shut down.
2025-10-17 10:00:02.456 [ERROR] Microsoft.EntityFrameworkCore.Database.Command: System.Exception: Database connection timeout
2025-10-17 10:00:03.789 [WARN] Microsoft.AspNetCore.Server.Kestrel: Memory pressure detected
```

#### 高级用法
```bash
# 监控 NuGet 包恢复
dotnet restore 2>&1 | ./aipipe --format csharp

# 监控 Entity Framework 迁移
dotnet ef database update 2>&1 | ./aipipe --format csharp --verbose

# 监控结构化日志
tail -f app.log | ./aipipe --format csharp --batch-size 10
```

---

### 5. PHP 应用日志 (`php`)

**使用场景**: Laravel、Symfony、WordPress、Composer 等

#### 基本监控
```bash
# 监控 Laravel 应用
./aipipe -f /var/log/laravel/app.log --format php

# 监控 PHP-FPM 日志
./aipipe -f /var/log/php-fpm/error.log --format php --verbose

# 监控 Composer 操作
composer install 2>&1 | ./aipipe --format php
```

#### 实际日志示例
```php
[2025-10-17 10:00:01] local.INFO: User login successful: john@example.com
[2025-10-17 10:00:02] local.ERROR: PHP Fatal error: Call to undefined function mysql_connect() in /app/index.php:123
[2025-10-17 10:00:03] local.WARNING: PHP Warning: file_get_contents() failed to open stream: No such file or directory
```

#### 高级用法
```bash
# 监控 WordPress 日志
tail -f /var/log/wordpress/debug.log | ./aipipe --format php

# 监控 Symfony 应用
tail -f /var/log/symfony/dev.log | ./aipipe --format php --verbose

# 监控 PHP 错误日志
tail -f /var/log/php/error.log | ./aipipe --format php
```

---

### 6. Python 应用日志 (`python`)

**使用场景**: Django、Flask、Celery、pip 等

#### 基本监控
```bash
# 监控 Django 应用
./aipipe -f /var/log/django/app.log --format python

# 监控 Flask 应用
python app.py 2>&1 | ./aipipe --format python --verbose

# 监控 Celery 任务
./aipipe -f /var/log/celery/celery.log --format python
```

#### 实际日志示例
```python
2025-10-17 10:00:01,123 INFO: Application started successfully
2025-10-17 10:00:02,456 ERROR: Database connection failed: [Errno 111] Connection refused
2025-10-17 10:00:03,789 WARNING: Memory usage high: 85% of available memory
```

#### 高级用法
```bash
# 监控 pip 安装
pip install -r requirements.txt 2>&1 | ./aipipe --format python

# 监控 pytest 测试
pytest -v 2>&1 | ./aipipe --format python --verbose

# 监控 uWSGI 日志
tail -f /var/log/uwsgi/app.log | ./aipipe --format python
```

---

### 7. Ruby 应用日志 (`ruby`)

**使用场景**: Rails、Sinatra、Bundler、RSpec 等

#### 基本监控
```bash
# 监控 Rails 应用
./aipipe -f /var/log/rails/production.log --format ruby

# 监控 Sinatra 应用
ruby app.rb 2>&1 | ./aipipe --format ruby --verbose

# 监控 Bundler 操作
bundle install 2>&1 | ./aipipe --format ruby
```

#### 实际日志示例
```ruby
I, [2025-10-17T10:00:01.123456 #12345]  INFO -- : Started GET "/api/users" for 127.0.0.1 at 2025-10-17 10:00:01
E, [2025-10-17T10:00:02.456789 #12345] ERROR -- : ActiveRecord::ConnectionTimeoutError (could not obtain a database connection within 5.000 seconds)
W, [2025-10-17T10:00:03.789012 #12345]  WARN -- : Memory usage is high: 512MB
```

#### 高级用法
```bash
# 监控 RSpec 测试
bundle exec rspec 2>&1 | ./aipipe --format ruby --verbose

# 监控 Sidekiq 任务
tail -f /var/log/sidekiq/sidekiq.log | ./aipipe --format ruby

# 监控 Capistrano 部署
cap production deploy 2>&1 | ./aipipe --format ruby
```

---

### 8. Kotlin 应用日志 (`kotlin`)

**使用场景**: Spring Boot Kotlin、Android、Ktor 等

#### 基本监控
```bash
# 监控 Spring Boot Kotlin 应用
./aipipe -f /var/log/kotlin-app/app.log --format kotlin

# 监控 Ktor 应用
./aipipe -f /var/log/ktor/app.log --format kotlin --verbose

# 监控 Gradle 构建
./gradlew build 2>&1 | ./aipipe --format kotlin
```

#### 实际日志示例
```kotlin
2025-10-17 10:00:01.123 INFO  [main] com.example.kotlin.Application - Started ApplicationKt in 2.345 seconds
2025-10-17 10:00:02.456 ERROR [http-nio-8080-exec-1] com.example.kotlin.service.UserService - Database connection failed
2025-10-17 10:00:03.789 WARN  [scheduler-1] com.example.kotlin.task.CleanupTask - Cleanup task completed with warnings
```

#### 高级用法
```bash
# 监控 Android 构建
./gradlew assembleDebug 2>&1 | ./aipipe --format kotlin --verbose

# 监控 Ktor 服务器
./aipipe -f /var/log/ktor/ktor.log --format kotlin --batch-size 15

# 监控 Kotlin 编译器
kotlinc app.kt 2>&1 | ./aipipe --format kotlin
```

---

### 9. FastAPI 应用日志 (`fastapi`)

**使用场景**: FastAPI、Uvicorn、Pydantic、异步处理等

#### 基本监控
```bash
# 监控 FastAPI 应用
uvicorn app:app --log-level info 2>&1 | ./aipipe --format fastapi

# 监控异步任务
./aipipe -f /var/log/fastapi/tasks.log --format fastapi --verbose

# 监控 API 请求
./aipipe -f /var/log/fastapi/access.log --format fastapi
```

#### 实际日志示例
```python
INFO:     127.0.0.1:12345 - "GET /api/users HTTP/1.1" 200 OK
ERROR:    Exception in ASGI application: Database connection timeout
WARNING:  Memory usage high: 85% of available memory
```

#### 高级用法
```bash
# 监控 Uvicorn 服务器
uvicorn app:app --reload 2>&1 | ./aipipe --format fastapi --verbose

# 监控 Pydantic 验证
tail -f /var/log/fastapi/validation.log | ./aipipe --format fastapi

# 监控异步任务队列
celery -A app worker 2>&1 | ./aipipe --format fastapi
```

---

## 🌐 前端和全栈

### 10. Node.js 应用日志 (`nodejs`)

**使用场景**: Express、Koa、NestJS、npm 等

#### 基本监控
```bash
# 监控 Express 应用
node app.js 2>&1 | ./aipipe --format nodejs

# 监控 NestJS 应用
npm run start:prod 2>&1 | ./aipipe --format nodejs --verbose

# 监控 npm 操作
npm install 2>&1 | ./aipipe --format nodejs
```

#### 实际日志示例
```javascript
info: Server running on port 3000
error: Error: ENOENT: no such file or directory, open '/app/config.json'
warn: DeprecationWarning: Buffer() is deprecated due to security and usability issues
```

#### 高级用法
```bash
# 监控 PM2 进程
pm2 logs 2>&1 | ./aipipe --format nodejs

# 监控 Jest 测试
npm test 2>&1 | ./aipipe --format nodejs --verbose

# 监控 Webpack 构建
npm run build 2>&1 | ./aipipe --format nodejs
```

---

### 11. TypeScript 应用日志 (`typescript`)

**使用场景**: Angular、React、Vue、tsc 编译等

#### 基本监控
```bash
# 监控 TypeScript 编译
tsc 2>&1 | ./aipipe --format typescript

# 监控 Angular 应用
ng serve 2>&1 | ./aipipe --format typescript --verbose

# 监控 React 构建
npm run build 2>&1 | ./aipipe --format typescript
```

#### 实际日志示例
```typescript
ERROR in src/app/app.component.ts(15,3): error TS2322: Type 'string' is not assignable to type 'number'
WARN in src/app/service.ts(25,10): warning TS6133: Parameter 'unused' is declared but never used
INFO: TypeScript compilation completed successfully
```

#### 高级用法
```bash
# 监控 ESLint 检查
npx eslint . 2>&1 | ./aipipe --format typescript --verbose

# 监控 Prettier 格式化
npx prettier --check . 2>&1 | ./aipipe --format typescript

# 监控 Angular 测试
ng test 2>&1 | ./aipipe --format typescript
```

---

## 🌐 Web 服务器

### 12. Nginx 日志 (`nginx`)

**使用场景**: 反向代理、负载均衡、静态文件服务等

#### 基本监控
```bash
# 监控访问日志
tail -f /var/log/nginx/access.log | ./aipipe --format nginx

# 监控错误日志
./aipipe -f /var/log/nginx/error.log --format nginx --verbose

# 监控特定站点
tail -f /var/log/nginx/site.com.access.log | ./aipipe --format nginx
```

#### 实际日志示例
```nginx
127.0.0.1 - - [17/Oct/2025:10:00:01 +0000] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0"
192.168.1.100 - - [17/Oct/2025:10:00:02 +0000] "POST /api/login HTTP/1.1" 401 567 "-" "curl/7.68.0"
upstream server temporarily disabled while connecting to upstream
```

#### 高级用法
```bash
# 监控上游服务器
tail -f /var/log/nginx/upstream.log | ./aipipe --format nginx

# 监控 SSL 错误
tail -f /var/log/nginx/ssl.log | ./aipipe --format nginx --verbose

# 监控负载均衡
tail -f /var/log/nginx/lb.log | ./aipipe --format nginx
```

---

## 🐳 云原生和容器

### 13. Docker 容器日志 (`docker`)

**使用场景**: 容器运行、镜像构建、容器编排等

#### 基本监控
```bash
# 监控容器日志
docker logs -f container_name | ./aipipe --format docker

# 监控容器构建
docker build . 2>&1 | ./aipipe --format docker --verbose

# 监控 Docker Compose
docker-compose logs -f | ./aipipe --format docker
```

#### 实际日志示例
```docker
Container started successfully
ERROR: failed to start container: port already in use
WARN: container running out of memory (limit: 512MB, usage: 480MB)
```

#### 高级用法
```bash
# 监控特定服务
docker-compose logs -f web | ./aipipe --format docker

# 监控健康检查
docker logs container_name 2>&1 | grep -i health | ./aipipe --format docker --verbose

# 监控镜像拉取
docker pull nginx:latest 2>&1 | ./aipipe --format docker
```

---

### 14. Kubernetes Pod 日志 (`kubernetes`)

**使用场景**: Pod 运行、部署、服务发现等

#### 基本监控
```bash
# 监控 Pod 日志
kubectl logs -f pod-name | ./aipipe --format kubernetes

# 监控部署日志
kubectl logs -f deployment/web-deployment | ./aipipe --format kubernetes --verbose

# 监控特定容器
kubectl logs -f pod-name -c container-name | ./aipipe --format kubernetes
```

#### 实际日志示例
```kubernetes
Pod started successfully
ERROR: Failed to pull image: ImagePullBackOff
WARN: Pod evicted due to memory pressure
```

#### 高级用法
```bash
# 监控所有 Pod
kubectl get pods | grep -v NAME | awk '{print $1}' | xargs -I {} kubectl logs -f {} | ./aipipe --format kubernetes

# 监控事件
kubectl get events --watch | ./aipipe --format kubernetes --verbose

# 监控特定命名空间
kubectl logs -f -n production deployment/web | ./aipipe --format kubernetes
```

---

## 🗄️ 数据库

### 15. PostgreSQL 日志 (`postgresql`)

**使用场景**: 数据库操作、查询优化、连接管理等

#### 基本监控
```bash
# 监控 PostgreSQL 日志
tail -f /var/log/postgresql/postgresql.log | ./aipipe --format postgresql

# 监控慢查询
tail -f /var/log/postgresql/slow.log | ./aipipe --format postgresql --verbose

# 监控连接日志
tail -f /var/log/postgresql/connections.log | ./aipipe --format postgresql
```

#### 实际日志示例
```postgresql
LOG: database system is ready to accept connections
ERROR: relation "users" does not exist
WARN: checkpoint request timed out
```

#### 高级用法
```bash
# 监控特定数据库
tail -f /var/log/postgresql/app_db.log | ./aipipe --format postgresql

# 监控复制日志
tail -f /var/log/postgresql/replication.log | ./aipipe --format postgresql --verbose

# 监控备份日志
pg_dump app_db 2>&1 | ./aipipe --format postgresql
```

---

### 16. MySQL 日志 (`mysql`)

**使用场景**: 数据库操作、InnoDB 引擎、复制等

#### 基本监控
```bash
# 监控 MySQL 错误日志
tail -f /var/log/mysql/error.log | ./aipipe --format mysql

# 监控慢查询日志
tail -f /var/log/mysql/slow.log | ./aipipe --format mysql --verbose

# 监控二进制日志
tail -f /var/log/mysql/mysql-bin.log | ./aipipe --format mysql
```

#### 实际日志示例
```mysql
InnoDB: Database was not shut down normally
ERROR 1045: Access denied for user 'root'@'localhost' (using password: YES)
Warning: Aborted connection to db: 'app_db' user: 'app_user' host: '192.168.1.100'
```

#### 高级用法
```bash
# 监控 InnoDB 状态
tail -f /var/log/mysql/innodb.log | ./aipipe --format mysql

# 监控复制状态
tail -f /var/log/mysql/replication.log | ./aipipe --format mysql --verbose

# 监控备份操作
mysqldump app_db 2>&1 | ./aipipe --format mysql
```

---

### 17. Redis 日志 (`redis`)

**使用场景**: 缓存操作、内存管理、持久化等

#### 基本监控
```bash
# 监控 Redis 日志
tail -f /var/log/redis/redis.log | ./aipipe --format redis

# 监控 Redis 集群
tail -f /var/log/redis/cluster.log | ./aipipe --format redis --verbose

# 监控 Redis 哨兵
tail -f /var/log/redis/sentinel.log | ./aipipe --format redis
```

#### 实际日志示例
```redis
Redis server version 6.2.6, bits=64
ERROR: OOM command not allowed when used memory > 'maxmemory'
WARN: overcommit_memory is set to 0
```

#### 高级用法
```bash
# 监控内存使用
redis-cli monitor 2>&1 | ./aipipe --format redis --verbose

# 监控持久化
tail -f /var/log/redis/persistence.log | ./aipipe --format redis

# 监控复制
tail -f /var/log/redis/replication.log | ./aipipe --format redis
```

---

### 18. Elasticsearch 日志 (`elasticsearch`)

**使用场景**: 索引操作、搜索、集群管理等

#### 基本监控
```bash
# 监控 Elasticsearch 日志
tail -f /var/log/elasticsearch/elasticsearch.log | ./aipipe --format elasticsearch

# 监控慢查询
tail -f /var/log/elasticsearch/slow.log | ./aipipe --format elasticsearch --verbose

# 监控集群状态
tail -f /var/log/elasticsearch/cluster.log | ./aipipe --format elasticsearch
```

#### 实际日志示例
```elasticsearch
[2025-10-17T10:00:01,123][INFO ][o.e.c.r.a.AllocationService] Cluster health status changed from [YELLOW] to [GREEN]
[2025-10-17T10:00:02,456][ERROR][o.e.i.e.Engine] Failed to flush index [users] due to [OutOfMemoryError]
[2025-10-17T10:00:03,789][WARN ][o.e.c.r.a.AllocationService] High disk watermark exceeded
```

#### 高级用法
```bash
# 监控索引操作
tail -f /var/log/elasticsearch/indexing.log | ./aipipe --format elasticsearch

# 监控搜索性能
tail -f /var/log/elasticsearch/search.log | ./aipipe --format elasticsearch --verbose

# 监控备份恢复
tail -f /var/log/elasticsearch/backup.log | ./aipipe --format elasticsearch
```

---

## 🛠️ 开发工具

### 19. Git 操作日志 (`git`)

**使用场景**: 版本控制、合并冲突、推送拉取等

#### 基本监控
```bash
# 监控 Git 操作
git pull 2>&1 | ./aipipe --format git

# 监控合并操作
git merge feature-branch 2>&1 | ./aipipe --format git --verbose

# 监控推送操作
git push origin main 2>&1 | ./aipipe --format git
```

#### 实际日志示例
```git
fatal: repository 'test' does not exist
error: Your local changes to the following files would be overwritten by merge
warning: You have divergent branches and need to specify how to reconcile them
```

#### 高级用法
```bash
# 监控克隆操作
git clone https://github.com/user/repo.git 2>&1 | ./aipipe --format git

# 监控重置操作
git reset --hard HEAD~1 2>&1 | ./aipipe --format git --verbose

# 监控子模块
git submodule update 2>&1 | ./aipipe --format git
```

---

### 20. Jenkins CI/CD 日志 (`jenkins`)

**使用场景**: 构建流水线、部署、测试等

#### 基本监控
```bash
# 监控 Jenkins 构建
./aipipe -f /var/log/jenkins/jenkins.log --format jenkins

# 监控构建日志
curl -s http://jenkins:8080/job/my-job/lastBuild/consoleText | ./aipipe --format jenkins --verbose

# 监控部署日志
tail -f /var/log/jenkins/deployment.log | ./aipipe --format jenkins
```

#### 实际日志示例
```jenkins
Started by user admin
Building on master in workspace /var/jenkins_home/workspace/my-job
ERROR: Failed to checkout repository
WARN: Build failed but continuing with post-build actions
```

#### 高级用法
```bash
# 监控特定作业
./aipipe -f /var/log/jenkins/my-job.log --format jenkins

# 监控插件日志
tail -f /var/log/jenkins/plugins.log | ./aipipe --format jenkins --verbose

# 监控系统日志
tail -f /var/log/jenkins/system.log | ./aipipe --format jenkins
```

---

### 21. GitHub Actions 日志 (`github`)

**使用场景**: CI/CD 流水线、部署、测试等

#### 基本监控
```bash
# 监控 Actions 日志
./aipipe -f /var/log/github-actions/workflow.log --format github

# 监控特定工作流
tail -f /var/log/github-actions/ci.yml.log | ./aipipe --format github --verbose

# 监控部署日志
tail -f /var/log/github-actions/deploy.log | ./aipipe --format github
```

#### 实际日志示例
```github
Run actions/checkout@v3
ERROR: Failed to checkout repository
WARN: Step failed but continuing with next steps
INFO: Deployment completed successfully
```

#### 高级用法
```bash
# 监控测试工作流
tail -f /var/log/github-actions/test.yml.log | ./aipipe --format github

# 监控发布工作流
tail -f /var/log/github-actions/release.yml.log | ./aipipe --format github --verbose

# 监控安全扫描
tail -f /var/log/github-actions/security.yml.log | ./aipipe --format github
```

---

## 🖥️ 系统级日志

### 22. Linux systemd journal (`journald`)

**使用场景**: 系统服务、内核消息、硬件错误等

#### 基本监控
```bash
# 监控所有系统日志
journalctl -f | ./aipipe --format journald

# 监控特定服务
journalctl -u nginx -f | ./aipipe --format journald --verbose

# 监控内核消息
journalctl -k -f | ./aipipe --format journald
```

#### 实际日志示例
```journald
Oct 17 10:00:01 systemd[1]: Started Network Manager Script Dispatcher Service
Oct 17 10:00:02 kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500 or sacrifice child
Oct 17 10:00:03 sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2
```

#### 高级用法
```bash
# 监控特定优先级
journalctl -p err -f | ./aipipe --format journald

# 监控特定时间范围
journalctl --since "1 hour ago" -f | ./aipipe --format journald --verbose

# 监控特定用户
journalctl _UID=1000 -f | ./aipipe --format journald
```

---

### 23. macOS Console (`macos-console`)

**使用场景**: 系统组件、应用程序、权限管理等

#### 基本监控
```bash
# 监控所有系统日志
log stream | ./aipipe --format macos-console

# 监控错误日志
log stream --predicate 'eventType == "errorEvent"' | ./aipipe --format macos-console --verbose

# 监控特定进程
log stream --process kernel | ./aipipe --format macos-console
```

#### 实际日志示例
```macos-console
2025-10-17 10:00:01.123456+0800 0x7b Default 0x0 0 0 kernel: (AppleH11ANEInterface) ANE0: EnableMemoryUnwireTimer: ERROR: Cannot enable Memory Unwire Timer
2025-10-17 10:00:02.234567+0800 0x1f11722 Error 0x185174d 386 0 locationd: (TCC) [com.apple.TCC:access] send_message_with_reply_sync(): XPC_ERROR_CONNECTION_INVALID
2025-10-17 10:00:03.345678+0800 0x1f11e95 Error 0x1851731 558 0 searchpartyd: (TCC) [com.apple.TCC:access] send_message_with_reply_sync(): XPC_ERROR_CONNECTION_INVALID
```

#### 高级用法
```bash
# 监控特定子系统
log stream --predicate 'subsystem == "com.apple.TCC"' | ./aipipe --format macos-console

# 监控特定级别
log stream --level debug | ./aipipe --format macos-console --verbose

# 监控特定用户
log stream --user 501 | ./aipipe --format macos-console
```

---

### 24. 传统 Syslog (`syslog`)

**使用场景**: 传统 Unix 系统、守护进程、系统服务等

#### 基本监控
```bash
# 监控主 syslog 文件
tail -f /var/log/syslog | ./aipipe --format syslog

# 监控认证日志
tail -f /var/log/auth.log | ./aipipe --format syslog --verbose

# 监控内核日志
tail -f /var/log/kern.log | ./aipipe --format syslog
```

#### 实际日志示例
```syslog
Oct 17 10:00:01 hostname systemd[1]: Started Network Manager Script Dispatcher Service
Oct 17 10:00:02 hostname kernel: [ 1234.567890] Out of memory: Kill process 1234 (chrome) score 500
Oct 17 10:00:03 hostname sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2
```

#### 高级用法
```bash
# 监控邮件日志
tail -f /var/log/mail.log | ./aipipe --format syslog

# 监控 cron 日志
tail -f /var/log/cron.log | ./aipipe --format syslog --verbose

# 监控防火墙日志
tail -f /var/log/iptables.log | ./aipipe --format syslog
```

---

## 🚀 综合使用场景

### 1. 全栈应用监控

```bash
# 监控前端构建
npm run build 2>&1 | ./aipipe --format typescript &

# 监控后端 API
./aipipe -f /var/log/api/app.log --format java &

# 监控数据库
tail -f /var/log/postgresql/postgresql.log | ./aipipe --format postgresql &

# 监控 Web 服务器
tail -f /var/log/nginx/error.log | ./aipipe --format nginx &

# 监控系统日志
journalctl -f | ./aipipe --format journald &
```

### 2. 微服务架构监控

```bash
# 监控用户服务
./aipipe -f /var/log/user-service/app.log --format go &

# 监控订单服务
./aipipe -f /var/log/order-service/app.log --format java &

# 监控支付服务
./aipipe -f /var/log/payment-service/app.log --format nodejs &

# 监控消息队列
tail -f /var/log/redis/redis.log | ./aipipe --format redis &

# 监控容器编排
kubectl logs -f deployment/user-service | ./aipipe --format kubernetes &
```

### 3. CI/CD 流水线监控

```bash
# 监控代码检查
npm run lint 2>&1 | ./aipipe --format typescript &

# 监控单元测试
npm test 2>&1 | ./aipipe --format nodejs --verbose &

# 监控构建过程
npm run build 2>&1 | ./aipipe --format nodejs &

# 监控部署过程
kubectl apply -f k8s/ 2>&1 | ./aipipe --format kubernetes &

# 监控集成测试
./integration-tests 2>&1 | ./aipipe --format java &
```

### 4. 系统运维监控

```bash
# 监控系统服务
journalctl -f | ./aipipe --format journald &

# 监控网络服务
journalctl -u NetworkManager -f | ./aipipe --format journald &

# 监控存储服务
journalctl -u systemd-logind -f | ./aipipe --format journald &

# 监控安全事件
tail -f /var/log/auth.log | ./aipipe --format syslog &

# 监控硬件状态
journalctl -k -f | ./aipipe --format journald &
```

---

## 📊 性能优化建议

### 1. 批处理优化
```bash
# 高频日志使用大批次
./aipipe -f app.log --format java --batch-size 30 --batch-wait 5s

# 低频日志使用小批次
./aipipe -f app.log --format java --batch-size 5 --batch-wait 1s

# 实时性要求高的场景
./aipipe -f app.log --format java --no-batch
```

### 2. 过滤优化
```bash
# 使用本地预过滤
journalctl -p err -f | ./aipipe --format journald

# 使用 grep 预过滤
tail -f app.log | grep "ERROR\|Exception" | ./aipipe --format java

# 使用 jq 过滤结构化日志
tail -f app.log | jq -r '.message' | ./aipipe --format java
```

### 3. 资源优化
```bash
# 限制内存使用
./aipipe -f app.log --format java --batch-size 10

# 减少 API 调用
./aipipe -f app.log --format java --batch-wait 10s

# 使用本地过滤
echo "INFO: Application started" | ./aipipe --format java --verbose
```

---

## 🔧 故障排查

### 1. 常见问题

#### API 调用失败
```bash
# 检查网络连接
ping api.example.com

# 使用 debug 模式
./aipipe -f app.log --format java --debug

# 检查配置文件
cat ~/.config/aipipe.json
```

#### 日志解析错误
```bash
# 使用 verbose 模式
./aipipe -f app.log --format java --verbose

# 检查日志格式
head -10 app.log

# 尝试不同格式
./aipipe -f app.log --format syslog --verbose
```

### 2. 调试技巧

#### 验证格式支持
```bash
# 查看支持的格式
./aipipe --help | grep format

# 测试特定格式
echo "ERROR: Test message" | ./aipipe --format java --verbose
```

#### 性能分析
```bash
# 监控 API 调用次数
./aipipe -f app.log --format java --verbose 2>&1 | grep "调用 AI"

# 监控处理速度
time ./aipipe -f app.log --format java --no-batch
```

---

## 📚 相关文档

- [SUPPORTED_FORMATS.md](SUPPORTED_FORMATS.md) - 格式支持详细说明
- [SYSTEM_LOG_EXAMPLES.md](SYSTEM_LOG_EXAMPLES.md) - 系统级日志监控示例
- [README_aipipe.md](README_aipipe.md) - 主要使用文档
- [aipipe-quickstart.md](aipipe-quickstart.md) - 快速入门指南

---

**作者**: xurenlu  
**版本**: v1.2.0  
**日期**: 2025-10-17  
**支持格式**: 24 种日志格式  
**适用场景**: 全栈开发、DevOps、系统运维、CI/CD
