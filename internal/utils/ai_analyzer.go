package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/xurenlu/aipipe/internal/config"
)

// 日志分析结果
type LogAnalysis struct {
	Line         string  `json:"line"`          // 日志行内容
	Important    bool    `json:"important"`     // 是否重要
	ShouldFilter bool    `json:"should_filter"` // 是否应该过滤
	Summary      string  `json:"summary"`       // 摘要
	Reason       string  `json:"reason"`        // 原因
	Confidence   float64 `json:"confidence"`    // 置信度
}

// AI API 请求和响应结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// 分析日志内容
func AnalyzeLog(logLine string, format string, cfg *config.Config) (*LogAnalysis, error) {
	// 本地预过滤：对于明确的低级别日志，直接过滤，不调用 AI
	if localAnalysis := tryLocalFilter(logLine); localAnalysis != nil {
		return localAnalysis, nil
	}

	// 构建系统提示词和用户提示词
	systemPrompt := buildSystemPrompt(format, cfg)
	userPrompt := buildUserPrompt(logLine)

	// 调用 AI API
	response, err := callAIAPI(systemPrompt, userPrompt, cfg)
	if err != nil {
		return nil, fmt.Errorf("调用 AI API 失败: %w", err)
	}

	// 解析响应
	analysis, err := parseAnalysisResponse(response)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 后处理：保守策略，当 AI 无法确定时，默认过滤
	analysis = applyConservativeFilter(analysis)

	return analysis, nil
}

// 本地预过滤
func tryLocalFilter(logLine string) *LogAnalysis {
	upperLine := strings.ToUpper(logLine)

	// 定义低级别日志的正则模式
	lowLevelPatterns := []struct {
		level   string
		pattern string
		summary string
	}{
		{"TRACE", `\b(TRACE|TRC)\b`, "TRACE 级别日志"},
		{"DEBUG", `\b(DEBUG|DBG|D)\b`, "DEBUG 级别日志"},
		{"INFO", `\b(INFO|INF|I)\b`, "INFO 级别日志"},
		{"VERBOSE", `\bVERBOSE\b`, "VERBOSE 级别日志"},
	}

	for _, pattern := range lowLevelPatterns {
		matched, err := regexp.MatchString(pattern.pattern, upperLine)
		if err == nil && matched {
			// 额外检查：确保不包含明显的错误关键词
			hasErrorKeywords := strings.Contains(upperLine, "ERROR") ||
				strings.Contains(upperLine, "EXCEPTION") ||
				strings.Contains(upperLine, "FATAL") ||
				strings.Contains(upperLine, "CRITICAL") ||
				strings.Contains(upperLine, "FAILED") ||
				strings.Contains(upperLine, "FAILURE")

			// 如果日志级别是低级别，但包含错误关键词，还是交给 AI 判断
			if hasErrorKeywords {
				continue
			}

			return &LogAnalysis{
				ShouldFilter: true,
				Summary:      pattern.summary,
				Reason:       fmt.Sprintf("本地过滤：%s 级别的日志通常无需关注", pattern.level),
			}
		}
	}

	// 无法本地判断，返回 nil，需要调用 AI
	return nil
}

// 构建系统提示词
func buildSystemPrompt(format string, cfg *config.Config) string {
	// 如果指定了提示词文件，尝试从文件加载
	if cfg.PromptFile != "" {
		if prompt, err := loadPromptFromFile(cfg.PromptFile, format); err == nil {
			return prompt
		}
		// 如果文件加载失败，继续使用内置提示词
	}
	
	// 如果配置了自定义提示词，使用自定义提示词
	if cfg.CustomPrompt != "" {
		return fmt.Sprintf("%s\n\n请分析以下 %s 格式的日志行：", cfg.CustomPrompt, format)
	}
	
	// 使用默认内置提示词
	return fmt.Sprintf(`你是一个专业的日志分析专家。请分析以下 %s 格式的日志行，判断其重要性。

分析规则：
1. 重要日志：错误、异常、警告、安全事件、性能问题、系统故障等
2. 不重要日志：调试信息、正常启动/停止、健康检查、常规操作等

请返回JSON格式：
{
  "should_filter": true/false,
  "summary": "简要摘要",
  "reason": "判断原因",
  "confidence": 0.0-1.0
}

如果should_filter为true，表示这是不重要的日志，应该被过滤掉。
如果should_filter为false，表示这是重要的日志，需要关注。`, format)
}

// 构建用户提示词
func buildUserPrompt(logLine string) string {
	return fmt.Sprintf("请分析这条日志行：\n%s", logLine)
}

// 从文件加载提示词
func loadPromptFromFile(filePath, format string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取提示词文件失败: %w", err)
	}
	
	prompt := string(content)
	
	// 如果提示词中包含 {format} 占位符，替换为实际的日志格式
	if strings.Contains(prompt, "{format}") {
		prompt = strings.ReplaceAll(prompt, "{format}", format)
	}
	
	// 如果提示词中包含 {log_line} 占位符，说明这是完整的提示词模板
	// 这种情况下，我们只需要返回提示词，不需要额外的格式说明
	if strings.Contains(prompt, "{log_line}") {
		return prompt, nil
	}
	
	// 否则，添加格式说明
	return fmt.Sprintf("%s\n\n请分析以下 %s 格式的日志行：", prompt, format), nil
}

// 调用 AI API
func callAIAPI(systemPrompt, userPrompt string, cfg *config.Config) (string, error) {
	request := ChatRequest{
		Model: cfg.Model,
		Messages: []ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", cfg.AIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Token)

	client := &http.Client{Timeout: time.Duration(cfg.Timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("AI API 返回空响应")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// 解析分析响应
func parseAnalysisResponse(response string) (*LogAnalysis, error) {
	// 尝试解析JSON响应
	var analysis LogAnalysis
	if err := json.Unmarshal([]byte(response), &analysis); err == nil {
		return &analysis, nil
	}

	// 如果JSON解析失败，尝试从文本中提取信息
	analysis = LogAnalysis{
		ShouldFilter: true, // 默认过滤
		Summary:      "AI分析结果",
		Reason:       "无法解析AI响应",
		Confidence:   0.5,
	}

	// 简单的文本分析
	response = strings.ToLower(response)
	if strings.Contains(response, "important") || strings.Contains(response, "error") || strings.Contains(response, "critical") {
		analysis.ShouldFilter = false
		analysis.Summary = "检测到重要日志"
	}

	return &analysis, nil
}

// 应用保守过滤策略
func applyConservativeFilter(analysis *LogAnalysis) *LogAnalysis {
	// 检查的关键词列表（表示 AI 无法确定或日志异常）
	uncertainKeywords := []string{
		"日志内容异常",
		"日志内容不完整",
		"无法判断",
		"日志格式异常",
		"日志内容不符合预期",
		"无法确定",
		"不确定",
		"无法识别",
		"格式不正确",
		"内容异常",
		"无法解析",
	}

	// 检查 summary 和 reason 字段
	checkText := strings.ToLower(analysis.Summary + " " + analysis.Reason)

	for _, keyword := range uncertainKeywords {
		if strings.Contains(checkText, strings.ToLower(keyword)) {
			// 发现不确定的关键词，强制过滤
			analysis.ShouldFilter = true
			if analysis.Reason == "" {
				analysis.Reason = "AI 无法确定日志重要性，采用保守策略过滤"
			} else {
				analysis.Reason = analysis.Reason + "（保守策略：过滤）"
			}
			break
		}
	}

	return analysis
}
