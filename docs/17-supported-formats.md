# 17. 支持格式

> AIPipe 支持的日志格式完整列表

## 📋 格式概览

AIPipe 支持 20+ 种常见的日志格式，每种格式都有专门的分析规则和优化。

## 🔧 应用日志格式

### Java 应用日志

**格式标识**: `java`

**示例**:
```
2024-01-01 10:00:00 ERROR com.example.Service: Database connection failed
2024-01-01 10:01:00 WARN  com.example.Service: High memory usage: 85%
2024-01-01 10:02:00 INFO  com.example.Service: User login successful
```

**特点**:
- 时间戳: `yyyy-MM-dd HH:mm:ss`
- 日志级别: `ERROR`, `WARN`, `INFO`, `DEBUG`
- 类名: 完整的包路径
- 消息: 具体的日志内容

### Python 应用日志

**格式标识**: `python`

**示例**:
```
2024-01-01 10:00:00,123 ERROR: Database connection failed
2024-01-01 10:01:00,456 WARNING: High memory usage: 85%
2024-01-01 10:02:00,789 INFO: User login successful
```

**特点**:
- 时间戳: `yyyy-MM-dd HH:mm:ss,SSS`
- 日志级别: `ERROR`, `WARNING`, `INFO`, `DEBUG`
- 消息: 具体的日志内容

### Node.js 应用日志

**格式标识**: `nodejs`

**示例**:
```
2024-01-01T10:00:00.123Z ERROR: Database connection failed
2024-01-01T10:01:00.456Z WARN: High memory usage: 85%
2024-01-01T10:02:00.789Z INFO: User login successful
```

**特点**:
- 时间戳: ISO 8601 格式
- 日志级别: `ERROR`, `WARN`, `INFO`, `DEBUG`
- 消息: 具体的日志内容

## 🌐 Web 服务器日志

### Nginx 访问日志

**格式标识**: `nginx`

**示例**:
```
192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /api/users HTTP/1.1" 200 1234
192.168.1.2 - - [01/Jan/2024:10:01:00 +0000] "POST /api/login HTTP/1.1" 401 567
192.168.1.3 - - [01/Jan/2024:10:02:00 +0000] "GET /api/health HTTP/1.1" 200 89
```

**特点**:
- IP 地址: 客户端 IP
- 时间戳: `[dd/MMM/yyyy:HH:mm:ss +0000]`
- HTTP 方法: `GET`, `POST`, `PUT`, `DELETE`
- 状态码: `200`, `404`, `500` 等
- 响应大小: 字节数

### Apache 访问日志

**格式标识**: `apache`

**示例**:
```
192.168.1.1 - - [01/Jan/2024:10:00:00 +0000] "GET /index.html HTTP/1.1" 200 1234
192.168.1.2 - - [01/Jan/2024:10:01:00 +0000] "POST /api/login HTTP/1.1" 401 567
```

**特点**:
- 格式与 Nginx 类似
- 时间戳格式相同
- 支持自定义日志格式

### IIS 访问日志

**格式标识**: `iis`

**示例**:
```
2024-01-01 10:00:00 192.168.1.1 GET /api/users 200 1234
2024-01-01 10:01:00 192.168.1.2 POST /api/login 401 567
```

**特点**:
- 时间戳: `yyyy-MM-dd HH:mm:ss`
- IP 地址: 客户端 IP
- HTTP 方法: `GET`, `POST`, `PUT`, `DELETE`
- 状态码: HTTP 状态码

## 🐳 容器和云平台日志

### Docker 容器日志

**格式标识**: `docker`

**示例**:
```
2024-01-01T10:00:00.000Z container_name: ERROR: Service unavailable
2024-01-01T10:01:00.000Z container_name: WARN: High memory usage
2024-01-01T10:02:00.000Z container_name: INFO: Service started
```

**特点**:
- 时间戳: ISO 8601 格式
- 容器名: 容器标识符
- 日志级别: `ERROR`, `WARN`, `INFO`, `DEBUG`
- 消息: 具体的日志内容

### Kubernetes 日志

**格式标识**: `kubernetes`

**示例**:
```
2024-01-01T10:00:00.000Z k8s-pod-123: ERROR: Pod failed to start
2024-01-01T10:01:00.000Z k8s-pod-456: WARN: Resource limit exceeded
```

**特点**:
- 时间戳: ISO 8601 格式
- Pod 标识: Kubernetes Pod 名称
- 日志级别: `ERROR`, `WARN`, `INFO`, `DEBUG`

### AWS CloudWatch 日志

**格式标识**: `cloudwatch`

**示例**:
```
2024-01-01T10:00:00.000Z [ERROR] Lambda function failed
2024-01-01T10:01:00.000Z [WARN] High memory usage detected
```

**特点**:
- 时间戳: ISO 8601 格式
- 日志级别: `[ERROR]`, `[WARN]`, `[INFO]`, `[DEBUG]`
- 服务标识: AWS 服务名称

## 📊 系统日志格式

### Syslog 格式

**格式标识**: `syslog`

**示例**:
```
Jan 1 10:00:00 hostname systemd[1]: Started Network Manager
Jan 1 10:01:00 hostname kernel: [12345.678901] ERROR: Out of memory
Jan 1 10:02:00 hostname sshd[1234]: Failed password for user
```

**特点**:
- 时间戳: `MMM dd HH:mm:ss`
- 主机名: 系统主机名
- 进程名: 进程名称和 PID
- 消息: 具体的日志内容

### Windows 事件日志

**格式标识**: `windows`

**示例**:
```
2024-01-01 10:00:00 ERROR Application: Database connection failed
2024-01-01 10:01:00 WARN  System: High CPU usage detected
2024-01-01 10:02:00 INFO  Security: User login successful
```

**特点**:
- 时间戳: `yyyy-MM-dd HH:mm:ss`
- 日志级别: `ERROR`, `WARN`, `INFO`
- 来源: 日志来源（Application, System, Security）
- 消息: 具体的日志内容

## 📱 移动应用日志

### Android 日志

**格式标识**: `android`

**示例**:
```
01-01 10:00:00.123  1234  5678 E MyApp: Database connection failed
01-01 10:01:00.456  1234  5678 W MyApp: High memory usage
01-01 10:02:00.789  1234  5678 I MyApp: User login successful
```

**特点**:
- 时间戳: `MM-dd HH:mm:ss.SSS`
- 进程 ID: 进程标识符
- 线程 ID: 线程标识符
- 日志级别: `E`, `W`, `I`, `D`
- 标签: 应用或组件名称

### iOS 日志

**格式标识**: `ios`

**示例**:
```
2024-01-01 10:00:00.123 MyApp[1234:5678] ERROR: Database connection failed
2024-01-01 10:01:00.456 MyApp[1234:5678] WARN: High memory usage
```

**特点**:
- 时间戳: `yyyy-MM-dd HH:mm:ss.SSS`
- 应用名: 应用程序名称
- 进程 ID: 进程标识符
- 线程 ID: 线程标识符
- 日志级别: `ERROR`, `WARN`, `INFO`, `DEBUG`

## 📄 结构化日志格式

### JSON 格式

**格式标识**: `json`

**示例**:
```json
{"timestamp":"2024-01-01T10:00:00Z","level":"ERROR","message":"Database connection failed","service":"api","user_id":12345}
{"timestamp":"2024-01-01T10:01:00Z","level":"WARN","message":"High memory usage","service":"api","memory_usage":85}
```

**特点**:
- 结构化数据
- 标准字段: `timestamp`, `level`, `message`
- 自定义字段: 业务相关字段
- 易于解析和查询

### XML 格式

**格式标识**: `xml`

**示例**:
```xml
<log timestamp="2024-01-01T10:00:00Z" level="ERROR" service="api">
  <message>Database connection failed</message>
  <user_id>12345</user_id>
</log>
```

**特点**:
- 结构化数据
- 标签化格式
- 支持嵌套结构
- 易于验证和转换

### CSV 格式

**格式标识**: `csv`

**示例**:
```
2024-01-01T10:00:00Z,ERROR,api,Database connection failed,12345
2024-01-01T10:01:00Z,WARN,api,High memory usage,12345
```

**特点**:
- 逗号分隔值
- 固定字段顺序
- 易于导入数据库
- 轻量级格式

## 🔧 自定义格式

### 正则表达式格式

**格式标识**: `regex`

**示例**:
```bash
# 自定义格式: 时间戳 级别 消息
aipipe analyze --format regex --pattern "(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) (\w+) (.+)"
```

**特点**:
- 支持自定义正则表达式
- 灵活的模式匹配
- 可提取特定字段
- 适用于特殊格式

### 分隔符格式

**格式标识**: `delimiter`

**示例**:
```bash
# 自定义分隔符格式
aipipe analyze --format delimiter --delimiter "|" --fields "timestamp,level,message"
```

**特点**:
- 支持自定义分隔符
- 指定字段顺序
- 适用于固定格式
- 易于配置

## 🎯 格式选择指南

### 根据日志来源选择

- **应用日志**: `java`, `python`, `nodejs`
- **Web 服务器**: `nginx`, `apache`, `iis`
- **容器平台**: `docker`, `kubernetes`
- **云平台**: `cloudwatch`, `azure`
- **系统日志**: `syslog`, `windows`
- **移动应用**: `android`, `ios`

### 根据日志结构选择

- **结构化日志**: `json`, `xml`, `csv`
- **非结构化日志**: `java`, `python`, `syslog`
- **自定义格式**: `regex`, `delimiter`

### 根据分析需求选择

- **错误分析**: 选择支持日志级别的格式
- **性能分析**: 选择包含时间戳的格式
- **业务分析**: 选择包含业务字段的格式

## 🔍 格式检测

### 自动检测

```bash
# 自动检测日志格式
aipipe analyze --format auto
```

### 手动检测

```bash
# 检测特定格式
aipipe analyze --format java --test
```

## 📋 格式配置

### 配置文件

```json
{
  "formats": {
    "java": {
      "pattern": "^(\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}) (\\w+) (.+)$",
      "fields": ["timestamp", "level", "message"],
      "time_format": "2006-01-02 15:04:05"
    },
    "nginx": {
      "pattern": "^(\\S+) - - \\[(.+)\\] \"(\\S+) (\\S+) HTTP/\\d\\.\\d\" (\\d+) (\\d+)$",
      "fields": ["ip", "timestamp", "method", "path", "status", "size"],
      "time_format": "02/Jan/2006:15:04:05 -0700"
    }
  }
}
```

### 自定义格式

```bash
# 添加自定义格式
aipipe config add-format --name "custom" --pattern "^(\\d{4}-\\d{2}-\\d{2}) (\\w+) (.+)$" --fields "date,level,message"
```

## 🎉 总结

AIPipe 支持 20+ 种日志格式，包括：

- **应用日志**: Java, Python, Node.js 等
- **Web 服务器**: Nginx, Apache, IIS 等
- **容器平台**: Docker, Kubernetes 等
- **云平台**: AWS CloudWatch, Azure 等
- **系统日志**: Syslog, Windows 事件日志等
- **移动应用**: Android, iOS 等
- **结构化日志**: JSON, XML, CSV 等
- **自定义格式**: 正则表达式, 分隔符等

每种格式都有专门的分析规则和优化，确保最佳的分析效果。

---

*返回: [文档首页](README.md)*
