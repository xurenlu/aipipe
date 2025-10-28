# AIPipe CLI 使用指南

AIPipe 现在使用 Cobra 框架提供了完整的子命令管理系统，所有功能都可以通过命令行直接操作，无需修改配置文件。

## 🚀 快速开始

### 基本用法
```bash
# 分析标准输入的日志
echo "ERROR: Database connection failed" | aipipe analyze

# 监控日志文件
aipipe monitor --file /var/log/app.log

# 显示帮助
aipipe --help
```

## 📋 子命令详解

### 1. analyze - 日志分析
分析从标准输入读取的日志内容。

```bash
# 基本分析
tail -f app.log | aipipe analyze

# 指定日志格式
echo "ERROR: Database failed" | aipipe analyze --format nginx

# 显示被过滤的日志
echo "INFO: User login" | aipipe analyze --show-not-important
```

**全局标志:**
- `--format, -f`: 日志格式 (默认: java)
- `--show-not-important`: 显示被过滤的日志
- `--verbose, -v`: 显示详细输出

### 2. monitor - 文件监控
实时监控指定的日志文件。

```bash
# 监控单个文件
aipipe monitor --file /var/log/app.log

# 指定日志格式
aipipe monitor --file /var/log/nginx/access.log --format nginx

# 监控系统日志
aipipe monitor --file /var/log/syslog --format syslog
```

### 3. config - 配置管理
管理 AIPipe 的配置文件。

```bash
# 显示当前配置
aipipe config show

# 启动配置向导
aipipe config init

# 验证配置文件
aipipe config validate

# 生成配置模板
aipipe config template

# 测试配置
aipipe config test
```

### 4. rules - 规则管理
管理过滤规则，包括添加、删除、启用、禁用和测试规则。

```bash
# 列出所有规则
aipipe rules list

# 添加新规则
aipipe rules add --pattern "ERROR" --action alert --description "匹配错误日志" --priority 1

# 添加过滤规则
aipipe rules add --pattern "INFO" --action filter --description "过滤信息日志" --priority 100

# 启用规则
aipipe rules enable rule_1

# 禁用规则
aipipe rules disable rule_1

# 删除规则
aipipe rules remove rule_1

# 测试规则
aipipe rules test rule_1 "ERROR: Database connection failed"

# 显示规则统计
aipipe rules stats
```

**规则参数:**
- `--pattern`: 规则模式 (正则表达式)
- `--action`: 规则动作 (filter, alert, ignore, highlight)
- `--priority`: 规则优先级 (数字越小优先级越高)
- `--description`: 规则描述
- `--category`: 规则分类
- `--color`: 高亮颜色
- `--enabled`: 是否启用规则
- `--id`: 规则ID (可选)

### 5. notify - 通知管理
管理通知系统，包括测试通知、配置通知器和发送测试消息。

```bash
# 显示通知状态
aipipe notify status

# 发送测试通知
aipipe notify test

# 发送自定义通知
aipipe notify send "重要告警" "系统出现严重错误，请立即处理"
```

### 6. cache - 缓存管理
管理缓存系统，包括查看统计信息、清理缓存和配置缓存。

```bash
# 显示缓存状态
aipipe cache status

# 显示缓存统计
aipipe cache stats

# 清理缓存
aipipe cache clear
```

### 7. ai - AI服务管理
管理AI服务，包括添加、删除、启用、禁用AI服务。

```bash
# 列出所有AI服务
aipipe ai list

# 添加AI服务
aipipe ai add --name "openai" --endpoint "https://api.openai.com/v1/chat/completions" --token "sk-xxx" --model "gpt-4" --priority 1

# 启用AI服务
aipipe ai enable openai

# 禁用AI服务
aipipe ai disable openai

# 删除AI服务
aipipe ai remove openai

# 测试AI服务
aipipe ai test openai

# 显示AI服务统计
aipipe ai stats
```

**AI服务参数:**
- `--name`: 服务名称
- `--endpoint`: API端点
- `--token`: API Token
- `--model`: 模型名称
- `--priority`: 优先级 (数字越小优先级越高)
- `--enabled`: 是否启用服务

## 🎯 使用场景

### 场景1: 实时日志监控
```bash
# 监控应用日志
aipipe monitor --file /var/log/app.log --format java

# 监控Nginx访问日志
aipipe monitor --file /var/log/nginx/access.log --format nginx
```

### 场景2: 日志分析管道
```bash
# 分析系统日志
journalctl -f | aipipe analyze --format journald

# 分析Docker日志
docker logs -f container_name | aipipe analyze --format docker
```

### 场景3: 规则配置
```bash
# 添加错误告警规则
aipipe rules add --pattern "ERROR|FATAL|CRITICAL" --action alert --priority 1

# 添加过滤规则
aipipe rules add --pattern "DEBUG|INFO" --action filter --priority 100

# 查看规则效果
aipipe rules list
```

### 场景4: 通知配置
```bash
# 测试通知系统
aipipe notify test

# 发送自定义告警
aipipe notify send "系统告警" "检测到异常流量"
```

## 🔧 配置示例

### 基本配置
```bash
# 显示当前配置
aipipe config show

# 启动配置向导
aipipe config init
```

### 规则配置
```bash
# 添加多个规则
aipipe rules add --pattern "ERROR" --action alert --priority 1 --description "错误日志告警"
aipipe rules add --pattern "WARN" --action alert --priority 2 --description "警告日志告警"
aipipe rules add --pattern "INFO" --action filter --priority 100 --description "信息日志过滤"
```

### AI服务配置
```bash
# 添加多个AI服务
aipipe ai add --name "openai" --endpoint "https://api.openai.com/v1/chat/completions" --token "sk-xxx" --model "gpt-4" --priority 1
aipipe ai add --name "claude" --endpoint "https://api.anthropic.com/v1/messages" --token "sk-xxx" --model "claude-3" --priority 2
```

## 📊 监控和统计

### 查看系统状态
```bash
# 缓存统计
aipipe cache stats

# 规则统计
aipipe rules stats

# AI服务统计
aipipe ai stats

# 通知状态
aipipe notify status
```

## 🚨 故障排除

### 常见问题
1. **配置加载失败**: 使用 `aipipe config init` 重新初始化配置
2. **规则不生效**: 检查规则优先级和模式是否正确
3. **通知发送失败**: 使用 `aipipe notify test` 测试通知系统
4. **AI服务不可用**: 使用 `aipipe ai test` 测试AI服务连接

### 调试模式
```bash
# 启用详细输出
aipipe analyze --verbose

# 显示被过滤的日志
aipipe analyze --show-not-important
```

## 🎉 总结

AIPipe 的子命令系统让所有功能都可以通过命令行直接操作，无需手动编辑配置文件。这大大提高了使用效率和灵活性，特别适合自动化脚本和CI/CD流水线集成。
