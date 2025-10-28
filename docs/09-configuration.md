# 09. 配置管理

> 灵活的配置系统，支持多种配置方式

## 🎯 概述

AIPipe 提供了灵活的配置管理系统，支持多种配置方式和动态配置更新。

## 📁 配置文件

### 1. 主配置文件

**位置**: `~/.aipipe/config.json`

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-api-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60,
  "local_filter": true,
  "show_not_important": false,
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.gmail.com",
      "smtp_port": 587,
      "username": "your-email@gmail.com",
      "password": "your-app-password",
      "to": "admin@example.com"
    }
  }
}
```

### 2. 监控配置文件

**位置**: `~/.aipipe-monitor.json`

```json
{
  "files": [
    {
      "path": "/var/log/app.log",
      "format": "java",
      "enabled": true,
      "priority": 10
    }
  ]
}
```

## 🔧 配置管理命令

### 1. 初始化配置

```bash
# 创建默认配置文件
aipipe config init

# 从模板创建配置
aipipe config init --template production
```

### 2. 查看配置

```bash
# 查看所有配置
aipipe config show

# 查看特定配置
aipipe config show --key "ai_endpoint"

# 查看配置摘要
aipipe config summary
```

### 3. 设置配置

AIPipe 目前不支持通过命令行直接设置配置值。需要手动编辑配置文件：

```bash
# 编辑配置文件
nano ~/.aipipe/config.json

# 或者使用其他编辑器
vim ~/.aipipe/config.json
```

### 4. 验证配置

```bash
# 验证配置文件
aipipe config validate

# 验证特定配置
aipipe config validate --key "ai_endpoint"
```

## 🌍 环境变量

### 1. 基本环境变量

```bash
# AI 配置
export OPENAI_API_KEY="sk-your-api-key"
export AIPIPE_AI_ENDPOINT="https://api.openai.com/v1/chat/completions"
export AIPIPE_AI_MODEL="gpt-3.5-turbo"

# 应用配置
export AIPIPE_CONFIG_FILE="~/.aipipe/config.json"
export AIPIPE_LOG_LEVEL="info"
export AIPIPE_DEBUG="false"
```

### 2. 通知配置

```bash
# 邮件配置
export AIPIPE_EMAIL_SMTP_HOST="smtp.gmail.com"
export AIPIPE_EMAIL_SMTP_PORT="587"
export AIPIPE_EMAIL_USERNAME="your-email@gmail.com"
export AIPIPE_EMAIL_PASSWORD="your-app-password"

# 系统通知
export AIPIPE_SYSTEM_NOTIFICATION="true"
export AIPIPE_SYSTEM_SOUND="true"
```

### 3. 缓存配置

```bash
# 缓存配置
export AIPIPE_CACHE_ENABLED="true"
export AIPIPE_CACHE_TTL="3600"
export AIPIPE_CACHE_MAX_SIZE="1000"
```

## 📋 配置模板

### 1. 开发环境模板

```bash
# 创建开发环境配置
aipipe config template --env development
```

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-dev-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 3,
  "timeout": 30,
  "rate_limit": 60,
  "local_filter": true,
  "show_not_important": true,
  "debug": true,
  "log_level": "debug"
}
```

### 2. 生产环境模板

```bash
# 创建生产环境配置
aipipe config template --env production
```

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-production-key",
  "ai_model": "gpt-4",
  "max_retries": 5,
  "timeout": 60,
  "rate_limit": 100,
  "local_filter": true,
  "show_not_important": false,
  "debug": false,
  "log_level": "info",
  "notifications": {
    "email": {
      "enabled": true,
      "smtp_host": "smtp.company.com",
      "smtp_port": 587,
      "username": "alerts@company.com",
      "password": "production-password",
      "to": "admin@company.com"
    }
  }
}
```

### 3. 测试环境模板

```bash
# 创建测试环境配置
aipipe config template --env testing
```

```json
{
  "ai_endpoint": "https://api.openai.com/v1/chat/completions",
  "ai_api_key": "sk-your-test-key",
  "ai_model": "gpt-3.5-turbo",
  "max_retries": 1,
  "timeout": 10,
  "rate_limit": 10,
  "local_filter": true,
  "show_not_important": true,
  "debug": true,
  "log_level": "debug",
  "notifications": {
    "system": {
      "enabled": true,
      "sound": false
    }
  }
}
```

## 🔄 配置更新

### 1. 热重载

```bash
# 启用热重载
aipipe config set --key "hot_reload" --value "true"

# 重新加载配置
aipipe config reload
```

### 2. 配置备份

```bash
# 备份配置
aipipe config backup

# 恢复配置
aipipe config restore --backup "2024-01-01-10-00-00"
```

### 3. 配置同步

```bash
# 同步配置到远程
aipipe config sync --remote "https://config.company.com"

# 从远程拉取配置
aipipe config pull --remote "https://config.company.com"
```

## 📊 配置监控

### 1. 配置状态

```bash
# 查看配置状态
aipipe config status

# 查看配置变更历史
aipipe config history
```

### 2. 配置验证

```bash
# 验证配置完整性
aipipe config validate --full

# 检查配置冲突
aipipe config check --conflicts
```

### 3. 配置统计

```bash
# 查看配置统计
aipipe config stats

# 查看配置使用情况
aipipe config usage
```

## 🎯 使用场景

### 场景1: 多环境配置

```bash
# 开发环境
export AIPIPE_ENV="development"
aipipe config init --env development

# 生产环境
export AIPIPE_ENV="production"
aipipe config init --env production
```

### 场景2: 配置管理

```bash
# 备份当前配置
aipipe config backup

# 修改配置
aipipe config set --key "ai_model" --value "gpt-4"

# 验证配置
aipipe config validate

# 应用配置
aipipe config reload
```

### 场景3: 配置同步

```bash
# 从远程同步配置
aipipe config pull --remote "https://config.company.com"

# 推送到远程
aipipe config push --remote "https://config.company.com"
```

## 🔍 故障排除

### 1. 配置文件问题

```bash
# 检查配置文件
aipipe config validate --verbose

# 检查配置文件权限
ls -la ~/.aipipe/config.json
```

### 2. 环境变量问题

```bash
# 检查环境变量
aipipe config env

# 检查环境变量覆盖
aipipe config env --show-overrides
```

### 3. 配置冲突

```bash
# 检查配置冲突
aipipe config check --conflicts

# 解决配置冲突
aipipe config resolve --conflicts
```

## 📋 最佳实践

### 1. 配置管理

- 使用版本控制管理配置文件
- 定期备份配置文件
- 使用环境变量覆盖敏感配置

### 2. 安全配置

- 保护配置文件权限
- 使用环境变量存储敏感信息
- 定期轮换 API 密钥

### 3. 性能优化

- 启用配置缓存
- 使用配置模板
- 监控配置变更

## 🎉 总结

AIPipe 的配置管理提供了：

- **多种配置方式**: 文件、环境变量、命令行
- **配置模板**: 开发、测试、生产环境模板
- **动态更新**: 热重载和配置同步
- **配置验证**: 完整的配置验证和检查
- **易于管理**: 完整的配置管理命令

---

*继续阅读: [10. 提示词管理](10-prompt-management.md)*
