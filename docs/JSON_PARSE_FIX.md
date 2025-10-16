# JSON 解析兼容性改进

## 问题描述

Poe API 返回的内容可能包含 Markdown 代码块格式，导致 JSON 解析失败。

### 问题示例

**API 返回的内容：**
```json
{
  "choices": [
    {
      "message": {
        "content": "```json\n{\n  \"should_filter\": true,\n  \"summary\": \"任务状态更新为处理中\",\n  \"reason\": \"日志为INFO级别，表示正常业务操作，任务状态变更无需关注\"\n}\n```"
      }
    }
  ]
}
```

注意 `content` 字段的值是：
```
```json
{
  "should_filter": true,
  "summary": "任务状态更新为处理中",
  "reason": "日志为INFO级别，表示正常业务操作，任务状态变更无需关注"
}
```
```

而不是纯 JSON。

## 解决方案

增强 `parseAnalysisResponse()` 函数，支持多种格式：

### 1. 处理 Markdown 代码块

**格式 1：带 json 标记**
```
```json
{ ... }
```
```

**格式 2：不带标记**
```
```
{ ... }
```
```

### 2. 智能提取 JSON

即使没有代码块，也能找到实际的 JSON 内容：
- 自动查找第一个 `{` 或 `[`
- 自动查找最后一个 `}` 或 `]`
- 清理前后的空白字符

### 3. 改进的解析逻辑

```go
// 1. 提取代码块内容
if strings.Contains(response, "```json") {
    // 找到 ```json 后面的内容
    start := strings.Index(response, "```json") + 7
    remaining := response[start:]
    end := strings.Index(remaining, "```")
    jsonStr = remaining[:end]
}

// 2. 清理空白字符
jsonStr = strings.TrimSpace(jsonStr)

// 3. 确保以 { 或 [ 开头
if jsonStr[0] != '{' && jsonStr[0] != '[' {
    start := strings.Index(jsonStr, "{")
    if start != -1 {
        jsonStr = jsonStr[start:]
    }
}

// 4. 确保以 } 或 ] 结尾
if jsonStr[len(jsonStr)-1] != '}' && jsonStr[len(jsonStr)-1] != ']' {
    end := strings.LastIndex(jsonStr, "}")
    if end != -1 {
        jsonStr = jsonStr[:end+1]
    }
}
```

## 支持的格式

### ✅ 格式 1：纯 JSON
```json
{
  "should_filter": true,
  "summary": "摘要",
  "reason": "原因"
}
```

### ✅ 格式 2：Markdown 代码块（带 json 标记）
```
```json
{
  "should_filter": true,
  "summary": "摘要",
  "reason": "原因"
}
```
```

### ✅ 格式 3：Markdown 代码块（不带标记）
```
```
{
  "should_filter": true,
  "summary": "摘要",
  "reason": "原因"
}
```
```

### ✅ 格式 4：带前后文本
```
这是一个分析结果：
{
  "should_filter": true,
  "summary": "摘要",
  "reason": "原因"
}
以上是分析。
```

### ✅ 格式 5：带换行符
```json
{
  "should_filter": true,
  "summary": "摘要",
  "reason": "原因"
}

```

## 错误处理

如果解析失败，会显示详细的错误信息：

```
解析 JSON 失败: invalid character 'x' looking for beginning of value
原始响应: ```json\n{...
提取的JSON: {...
```

这样可以方便调试，了解：
1. 原始 API 响应是什么
2. 提取出来的 JSON 是什么
3. 具体的解析错误是什么

## 测试方法

### 使用 --debug 模式查看解析过程

```bash
echo "2025-10-13 INFO Test log" | ./aipipe --format java --debug
```

在 debug 输出中可以看到：
1. API 返回的原始响应
2. 提取的 JSON 内容
3. 解析结果

### 手动测试各种格式

创建测试文件 `test-json-parse.go`：

```go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
)

type LogAnalysis struct {
    ShouldFilter bool   `json:"should_filter"`
    Summary      string `json:"summary"`
    Reason       string `json:"reason"`
}

func parseAnalysisResponse(response string) (*LogAnalysis, error) {
    // [实际的解析代码]
    // ...
}

func main() {
    // 测试用例
    testCases := []string{
        // 纯 JSON
        `{"should_filter":true,"summary":"test","reason":"test"}`,
        
        // Markdown 代码块
        "```json\n{\"should_filter\":true,\"summary\":\"test\",\"reason\":\"test\"}\n```",
        
        // 带前后空白
        "\n\n  {\"should_filter\":true,\"summary\":\"test\",\"reason\":\"test\"}  \n\n",
        
        // 带文本
        "分析结果：\n{\"should_filter\":true,\"summary\":\"test\",\"reason\":\"test\"}\n完成",
    }
    
    for i, tc := range testCases {
        result, err := parseAnalysisResponse(tc)
        if err != nil {
            fmt.Printf("测试 %d 失败: %v\n", i+1, err)
        } else {
            fmt.Printf("测试 %d 成功: should_filter=%v\n", i+1, result.ShouldFilter)
        }
    }
}
```

## 优势

1. **更强的兼容性**：支持多种 AI 返回格式
2. **更好的容错性**：即使格式不标准也能尽力提取
3. **更清晰的错误**：解析失败时显示详细信息
4. **零配置**：用户无需关心 API 返回格式

## 相关文件

- `aipipe.go` - 包含 `parseAnalysisResponse()` 函数
- `DEBUG_MODE_EXAMPLE.md` - 使用 debug 模式查看解析过程

## 版本历史

- **v1.0.0** - 初始版本，基本的代码块处理
- **v1.1.0** - 增强版本，支持更多格式和智能提取

---

**最后更新**: 2025-10-13  
**状态**: ✅ 已修复并测试

