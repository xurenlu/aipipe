package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// AI 服务配置
type AIService struct {
	Name     string `json:"name"`     // 服务名称
	Endpoint string `json:"endpoint"` // API 端点
	Token    string `json:"token"`    // API Token
	Model    string `json:"model"`    // 模型名称
	Priority int    `json:"priority"` // 优先级（数字越小优先级越高）
	Enabled  bool   `json:"enabled"`  // 是否启用
}

// AI 服务管理器
type AIServiceManager struct {
	services    []AIService
	current     int
	fallback    bool
	rateLimiter map[string]time.Time
	mutex       sync.RWMutex
}

// 启用/禁用服务
func (asm *AIServiceManager) SetServiceEnabled(serviceName string, enabled bool) error {
	asm.mutex.Lock()
	defer asm.mutex.Unlock()

	for i := range asm.services {
		if asm.services[i].Name == serviceName {
			asm.services[i].Enabled = enabled
			return nil
		}
	}

	return fmt.Errorf("服务 %s 不存在", serviceName)
}

// 获取服务列表
func (asm *AIServiceManager) GetServices() []AIService {
	asm.mutex.RLock()
	defer asm.mutex.RUnlock()

	services := make([]AIService, len(asm.services))
	copy(services, asm.services)
	return services
}

// AI服务管理命令处理函数

// 列出所有AI服务
func handleAIList() {
	fmt.Println("🤖 AI 服务列表:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	services := aiServiceManager.GetServices()
	if len(services) == 0 {
		fmt.Println("没有配置AI服务")
		return
	}

	for i, service := range services {
		status := "❌ 禁用"
		if service.Enabled {
			status = "✅ 启用"
		}

		fmt.Printf("%d. %s %s\n", i+1, status, service.Name)
		fmt.Printf("   端点: %s\n", service.Endpoint)
		fmt.Printf("   模型: %s\n", service.Model)
		fmt.Printf("   Token: %s...%s\n", service.Token[:min(8, len(service.Token))], service.Token[max(0, len(service.Token)-8):])
		fmt.Printf("   优先级: %d\n", service.Priority)
		fmt.Println()
	}
}

// 测试所有AI服务
func handleAITest() {
	fmt.Println("🧪 测试所有AI服务...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	services := aiServiceManager.GetServices()
	if len(services) == 0 {
		fmt.Println("没有配置AI服务")
		return
	}

	successCount := 0
	for _, service := range services {
		if !service.Enabled {
			fmt.Printf("⏭️  跳过禁用的服务: %s\n", service.Name)
			continue
		}

		fmt.Printf("🔗 测试服务: %s...", service.Name)

		// 创建测试请求
		testPrompt := "请回复 'OK' 表示连接正常"
		reqBody := ChatRequest{
			Model: service.Model,
			Messages: []ChatMessage{
				{
					Role:    "user",
					Content: testPrompt,
				},
			},
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			fmt.Printf(" ❌ 构建请求失败\n")
			continue
		}

		// 创建HTTP请求
		req, err := http.NewRequest("POST", service.Endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf(" ❌ 创建请求失败\n")
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("api-key", service.Token)

		// 发送请求
		client := &http.Client{
			Timeout: time.Duration(globalConfig.Timeout) * time.Second,
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf(" ❌ 请求失败: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf(" ❌ API错误 %d: %s\n", resp.StatusCode, string(body))
			continue
		}

		fmt.Printf(" ✅ 成功\n")
		successCount++
	}

	fmt.Printf("\n📊 测试结果: %d/%d 服务可用\n", successCount, len(services))
	if successCount == 0 {
		os.Exit(1)
	}
}

// 显示AI服务统计信息
func handleAIStats() {
	fmt.Println("📊 AI 服务统计信息:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载配置
	if err := loadConfig(); err != nil {
		fmt.Printf("❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	stats := aiServiceManager.GetStats()
	fmt.Printf("总服务数: %d\n", stats["total_services"])
	fmt.Printf("启用服务数: %d\n", stats["enabled_services"])
	fmt.Printf("当前索引: %d\n", stats["current_index"])
	fmt.Printf("故障转移模式: %t\n", stats["fallback_mode"])

	// 显示服务详情
	services := aiServiceManager.GetServices()
	if len(services) > 0 {
		fmt.Println("\n服务详情:")
		for _, service := range services {
			status := "❌ 禁用"
			if service.Enabled {
				status = "✅ 启用"
			}
			fmt.Printf("  %s %s (优先级: %d)\n", status, service.Name, service.Priority)
		}
	}
}

// 测试 AI 服务连接
func testAIConnection() error {
	// 创建一个简单的测试请求
	testPrompt := "请回复 'OK' 表示连接正常"

	// 构建请求
	reqBody := ChatRequest{
		Model: globalConfig.Model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: testPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", globalConfig.AIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", globalConfig.Token)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(globalConfig.Timeout) * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 构建系统提示词（定义角色和判断标准）
func buildSystemPrompt(format string) string {
	formatExamples := getFormatSpecificExamples(format)

	basePrompt := fmt.Sprintf(`你是一个专业的日志分析助手，专门分析 %s 格式的日志。

你的任务是判断日志是否需要关注，并以 JSON 格式返回分析结果。

返回格式：
{
  "should_filter": true/false,  // true 表示应该过滤（不重要），false 表示需要关注
  "summary": "简短摘要（20字内）",
  "reason": "判断原因"
}

判断标准和示例：

【应该过滤的日志】(should_filter=true) - 正常运行状态，无需告警：
1. 健康检查和心跳
   - "Health check endpoint called"
   - "Heartbeat received from client"
   - "/health returned 200"
   
2. 应用启动和配置加载
   - "Application started successfully"
   - "Configuration loaded from config.yml"
   - "Server listening on port 8080"
   
3. 正常的业务操作（INFO/DEBUG）
   - "User logged in: john@example.com"
   - "Retrieved 20 records from database"
   - "Cache hit for key: user_123"
   - "Request processed in 50ms"
   
4. 定时任务正常执行
   - "Scheduled task completed successfully"
   - "Cleanup job finished, removed 10 items"
   
5. 静态资源请求
   - "GET /static/css/style.css 200"
   - "Serving static file: logo.png"

6. 常规数据库操作
   - "Query executed successfully in 10ms"
   - "Transaction committed"
   
7. 正常的API请求响应
   - "GET /api/users 200 OK"
   - "POST /api/data returned 201"

【需要关注的日志】(should_filter=false) - 异常情况，需要告警：
1. 错误和异常（ERROR级别）
   - "ERROR: Database connection failed"
   - "NullPointerException at line 123"
   - "Failed to connect to Redis"
   - 任何包含 Exception, Error, Failed 的错误信息
   
2. 数据库问题
   - "Database connection timeout"
   - "Deadlock detected"
   - "Slow query: 5000ms"
   - "Connection pool exhausted"
   
3. 认证和授权问题
   - "Authentication failed for user admin"
   - "Invalid token: access denied"
   - "Permission denied: insufficient privileges"
   - "Multiple failed login attempts from 192.168.1.100"
   
4. 性能问题（WARN级别或慢响应）
   - "Request timeout after 30s"
   - "Response time exceeded threshold: 5000ms"
   - "Memory usage high: 85%%"
   - "Thread pool near capacity: 95/100"
   
5. 资源耗尽
   - "Out of memory error"
   - "Disk space low: 95%% used"
   - "Too many open files"
   
6. 外部服务调用失败
   - "Payment gateway timeout"
   - "Failed to call external API: 500"
   - "Third-party service unavailable"
   
7. 业务异常
   - "Order processing failed: insufficient balance"
   - "Payment declined: invalid card"
   - "Data validation failed"
   
8. 安全问题
   - "SQL injection attempt detected"
   - "Suspicious activity from IP"
   - "Rate limit exceeded"
   - "Invalid CSRF token"
   
9. 数据一致性问题
   - "Data mismatch detected"
   - "Inconsistent state in transaction"
   
10. 服务降级和熔断
    - "Circuit breaker opened"
    - "Service degraded mode activated"`, format)

	// 添加格式特定的示例
	if formatExamples != "" {
		basePrompt += "\n\n" + formatExamples
	}

	basePrompt += `

注意：
- 如果日志级别是 ERROR 或包含 Exception/Error，通常需要关注
- 如果包含 "failed", "timeout", "unable", "cannot" 等负面词汇，需要仔细判断
- 如果是 WARN 级别，需要根据具体内容判断严重程度
- 健康检查、心跳、正常的 INFO 日志通常可以过滤

重要原则（保守策略）：
- 如果日志内容不完整、格式异常或无法确定重要性，请设置 should_filter=true
- 在 summary 或 reason 中明确说明"日志内容异常"、"无法判断"等原因
- 我们采取保守策略：只提示确认重要的信息，不确定的一律过滤

只返回 JSON，不要其他内容。`

	// 如果有自定义提示词，添加到系统提示词中
	if globalConfig.CustomPrompt != "" {
		basePrompt += "\n\n" + globalConfig.CustomPrompt
	}

	return basePrompt
}

// 构建用户提示词（实际要分析的日志）
func buildUserPrompt(logLine string) string {
	return fmt.Sprintf("请分析以下日志：\n\n%s", logLine)
}

// 构建批量用户提示词
func buildBatchUserPrompt(logLines []string) string {
	var sb strings.Builder
	sb.WriteString("请批量分析以下日志，对每一行给出判断：\n\n")

	for i, line := range logLines {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, line))
	}

	sb.WriteString("\n请返回 JSON 格式：\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"results\": [\n")
	sb.WriteString("    {\"should_filter\": true/false, \"summary\": \"摘要\", \"reason\": \"原因\"},\n")
	sb.WriteString("    ...\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"overall_summary\": \"这批日志的整体摘要（20字内）\",\n")
	sb.WriteString(fmt.Sprintf("  \"important_count\": 0  // 重要日志数量（%d 条中有几条）\n", len(logLines)))
	sb.WriteString("}\n")
	sb.WriteString("\n注意：results 数组必须包含 " + fmt.Sprintf("%d", len(logLines)) + " 个元素，按顺序对应每一行日志。")

	return sb.String()
}

// 调用 AI API
func callAIAPI(systemPrompt, userPrompt string) (string, error) {
	// 获取AI服务
	service, err := aiServiceManager.GetNextService()
	if err != nil {
		return "", fmt.Errorf("获取AI服务失败: %w", err)
	}

	// 记录服务调用
	aiServiceManager.RecordCall(service.Name)

	// 构建请求，使用 system 和 user 两条消息
	reqBody := ChatRequest{
		Model: service.Model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Debug: 打印请求信息
	if *debug {
		fmt.Println("\n" + strings.Repeat("=", 80))
		fmt.Println("🔍 DEBUG: HTTP 请求详情")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("服务: %s\n", service.Name)
		fmt.Printf("URL: %s\n", service.Endpoint)
		fmt.Printf("Method: POST\n")
		fmt.Printf("Headers:\n")
		fmt.Printf("  Content-Type: application/json\n")
		fmt.Printf("  api-key: %s...%s\n", service.Token[:min(10, len(service.Token))], service.Token[max(0, len(service.Token)-10):])
		fmt.Printf("\nRequest Body:\n")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonData, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(jsonData))
		}
		fmt.Println(strings.Repeat("=", 80))
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", service.Endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", service.Token)

	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(globalConfig.Timeout) * time.Second,
	}

	if *debug {
		fmt.Printf("⏳ 发送请求中...\n")
	}

	startTime := time.Now()
	var resp *http.Response
	var httpErr error

	// 重试机制
	for i := 0; i < globalConfig.MaxRetries; i++ {
		resp, httpErr = client.Do(req)
		if httpErr == nil {
			break
		}

		// 使用错误处理器处理网络错误
		if handledErr := errorHandler.Handle(httpErr, map[string]interface{}{
			"operation":   "ai_api_call",
			"service":     service.Name,
			"endpoint":    service.Endpoint,
			"retry":       i + 1,
			"max_retries": globalConfig.MaxRetries,
		}); handledErr != nil {
			if i == globalConfig.MaxRetries-1 {
				if *debug {
					fmt.Printf("❌ 请求失败 (重试 %d/%d): %v\n", i+1, globalConfig.MaxRetries, handledErr)
					fmt.Println(strings.Repeat("=", 80) + "\n")
				}
				return "", handledErr
			}
			time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
		} else {
			// 错误已恢复，重试
			continue
		}
	}

	if httpErr != nil {
		if *debug {
			fmt.Printf("❌ 请求失败: %v\n", httpErr)
			fmt.Println(strings.Repeat("=", 80) + "\n")
		}
		return "", httpErr
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Debug: 打印响应信息
	if *debug {
		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("🔍 DEBUG: HTTP 响应详情")
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("Status Code: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("Response Time: %v\n", elapsed)
		fmt.Printf("Content-Length: %d bytes\n", len(body))
		fmt.Printf("\nResponse Headers:\n")
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
		fmt.Printf("\nResponse Body:\n")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Println(string(body))
		}
		fmt.Println(strings.Repeat("=", 80) + "\n")
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))

		// 使用错误处理器处理 API 错误
		if handledErr := errorHandler.Handle(apiErr, map[string]interface{}{
			"operation":     "ai_api_response",
			"service":       service.Name,
			"status_code":   resp.StatusCode,
			"endpoint":      service.Endpoint,
			"response_body": string(body),
		}); handledErr != nil {
			return "", handledErr
		}
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API 响应中没有内容")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// 解析批量 AI 响应
func parseBatchAnalysisResponse(response string, expectedCount int) (*BatchLogAnalysis, error) {
	// 提取 JSON（处理 markdown 代码块）
	jsonStr := extractJSON(response)

	var batchAnalysis BatchLogAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &batchAnalysis); err != nil {
		return nil, fmt.Errorf("解析批量 JSON 失败: %w\n原始响应: %s\n提取的JSON: %s", err, response, jsonStr)
	}

	// 验证结果数量
	if len(batchAnalysis.Results) != expectedCount {
		if *verbose || *debug {
			log.Printf("⚠️  批量分析结果数量不匹配：期望 %d 条，实际 %d 条", expectedCount, len(batchAnalysis.Results))
		}

		// 如果结果少于预期，补充默认结果（过滤）
		for len(batchAnalysis.Results) < expectedCount {
			batchAnalysis.Results = append(batchAnalysis.Results, LogAnalysis{
				ShouldFilter: true,
				Summary:      "结果缺失",
				Reason:       "批量分析返回结果数量不足",
			})
		}
	}

	return &batchAnalysis, nil
}

// 提取 JSON（从可能包含 markdown 代码块的响应中）
func extractJSON(response string) string {
	jsonStr := response

	// 处理 ```json ... ``` 格式
	if strings.Contains(response, "```json") {
		start := strings.Index(response, "```json")
		if start != -1 {
			start += 7
			remaining := response[start:]
			end := strings.Index(remaining, "```")
			if end != -1 {
				jsonStr = remaining[:end]
			}
		}
	} else if strings.Contains(response, "```") {
		start := strings.Index(response, "```")
		if start != -1 {
			start += 3
			remaining := response[start:]
			end := strings.Index(remaining, "```")
			if end != -1 {
				jsonStr = remaining[:end]
			}
		}
	}

	// 清理字符串
	jsonStr = strings.TrimSpace(jsonStr)

	// 智能定位 JSON 起始和结束
	if len(jsonStr) > 0 && jsonStr[0] != '{' && jsonStr[0] != '[' {
		startBrace := strings.Index(jsonStr, "{")
		startBracket := strings.Index(jsonStr, "[")

		start := -1
		if startBrace != -1 && (startBracket == -1 || startBrace < startBracket) {
			start = startBrace
		} else if startBracket != -1 {
			start = startBracket
		}

		if start != -1 {
			jsonStr = jsonStr[start:]
		}
	}

	if len(jsonStr) > 0 && jsonStr[len(jsonStr)-1] != '}' && jsonStr[len(jsonStr)-1] != ']' {
		endBrace := strings.LastIndex(jsonStr, "}")
		endBracket := strings.LastIndex(jsonStr, "]")

		end := -1
		if endBrace != -1 && endBrace > endBracket {
			end = endBrace
		} else if endBracket != -1 {
			end = endBracket
		}

		if end != -1 {
			jsonStr = jsonStr[:end+1]
		}
	}

	return jsonStr
}

// 解析 AI 响应（单条）
func parseAnalysisResponse(response string) (*LogAnalysis, error) {
	jsonStr := extractJSON(response)

	var analysis LogAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w\n原始响应: %s\n提取的JSON: %s", err, response, jsonStr)
	}

	return &analysis, nil
}

// 获取AI分析结果
func (cm *CacheManager) GetAIAnalysis(logHash string) (*AIAnalysisCache, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	item, exists := cm.aiCache[logHash]
	if !exists {
		cm.stats.MissCount++
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		cm.stats.MissCount++
		return nil, false
	}

	cm.stats.HitCount++
	return item, true
}

// 设置AI分析结果
func (cm *CacheManager) SetAIAnalysis(logHash string, result *AIAnalysisCache) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 检查是否需要清理空间
	if cm.needsEviction() {
		cm.evictOldest()
	}

	cm.aiCache[logHash] = result
	cm.updateStats()
}
