package main

import (
	"testing"
	"time"
)

// 测试配置验证器
func TestConfigValidator(t *testing.T) {
	validator := NewConfigValidator()

	// 测试有效配置
	validConfig := Config{
		AIEndpoint:  "https://api.openai.com/v1/chat/completions",
		Token:       "sk-1234567890abcdef",
		Model:       "gpt-4",
		MaxRetries:  3,
		Timeout:     30,
		RateLimit:   100,
		LocalFilter: true,
	}

	err := validator.Validate(&validConfig)
	if err != nil {
		t.Errorf("有效配置验证失败: %v", err)
	}

	// 测试无效配置
	invalidConfig := Config{
		AIEndpoint: "invalid-url",
		Token:      "short",
		Model:      "",
		MaxRetries: 15,   // 超出范围
		Timeout:    1,    // 超出范围
		RateLimit:  2000, // 超出范围
	}

	err = validator.Validate(&invalidConfig)
	if err == nil {
		t.Error("无效配置应该验证失败")
	}

	errors := validator.GetErrors()
	if len(errors) < 4 {
		t.Errorf("应该至少有4个验证错误，实际: %d", len(errors))
	}
}

// 测试错误处理器
func TestErrorHandler(t *testing.T) {
	handler := NewErrorHandler()

	// 注册恢复策略
	handler.RegisterStrategy(ErrorCategoryNetwork, &NetworkErrorRecovery{
		maxRetries: 3,
		backoff:    time.Second,
	})

	// 测试网络错误处理
	networkErr := &AIPipeError{
		Code:     "NETWORK_ERROR",
		Category: ErrorCategoryNetwork,
		Level:    ErrorLevelWarning,
		Message:  "connection timeout",
		Context:  map[string]interface{}{"operation": "test"},
	}

	// 测试错误分类
	handler.classifyError(networkErr)
	if networkErr.Category != ErrorCategoryNetwork {
		t.Errorf("错误分类错误，期望: %s, 实际: %s", ErrorCategoryNetwork, networkErr.Category)
	}

	if networkErr.Code != "NETWORK_ERROR" {
		t.Errorf("错误代码错误，期望: NETWORK_ERROR, 实际: %s", networkErr.Code)
	}
}

// 测试配置结构
func TestConfigStructure(t *testing.T) {
	config := Config{
		AIEndpoint:   "https://api.openai.com/v1/chat/completions",
		Token:        "sk-1234567890abcdef",
		Model:        "gpt-4",
		CustomPrompt: "测试提示词",
		MaxRetries:   3,
		Timeout:      30,
		RateLimit:    100,
		LocalFilter:  true,
	}

	// 测试字段值
	if config.AIEndpoint == "" {
		t.Error("AIEndpoint 不能为空")
	}

	if config.Token == "" {
		t.Error("Token 不能为空")
	}

	if config.Model == "" {
		t.Error("Model 不能为空")
	}

	if config.MaxRetries < 0 || config.MaxRetries > 10 {
		t.Error("MaxRetries 应该在 0-10 范围内")
	}

	if config.Timeout < 5 || config.Timeout > 300 {
		t.Error("Timeout 应该在 5-300 范围内")
	}

	if config.RateLimit < 1 || config.RateLimit > 1000 {
		t.Error("RateLimit 应该在 1-1000 范围内")
	}
}

// 测试错误级别
func TestErrorLevel(t *testing.T) {
	levels := []ErrorLevel{ErrorLevelInfo, ErrorLevelWarning, ErrorLevelError, ErrorLevelCritical}
	levelNames := []string{"INFO", "WARNING", "ERROR", "CRITICAL"}

	for i, level := range levels {
		expected := levelNames[i]
		actual := []string{"INFO", "WARNING", "ERROR", "CRITICAL"}[level]
		if actual != expected {
			t.Errorf("错误级别 %d 名称错误，期望: %s, 实际: %s", level, expected, actual)
		}
	}
}

// 测试错误分类
func TestErrorCategory(t *testing.T) {
	categories := []ErrorCategory{
		ErrorCategoryConfig,
		ErrorCategoryNetwork,
		ErrorCategoryAI,
		ErrorCategoryProcessing,
		ErrorCategoryOutput,
		ErrorCategoryFile,
	}

	expectedCategories := []string{
		"config",
		"network",
		"ai",
		"processing",
		"output",
		"file",
	}

	for i, category := range categories {
		if string(category) != expectedCategories[i] {
			t.Errorf("错误分类 %d 错误，期望: %s, 实际: %s", i, expectedCategories[i], category)
		}
	}
}

// 测试辅助函数
func TestMinMax(t *testing.T) {
	// 测试 min 函数
	if min(5, 3) != 3 {
		t.Error("min(5, 3) 应该返回 3")
	}

	if min(3, 5) != 3 {
		t.Error("min(3, 5) 应该返回 3")
	}

	// 测试 max 函数
	if max(5, 3) != 5 {
		t.Error("max(5, 3) 应该返回 5")
	}

	if max(3, 5) != 5 {
		t.Error("max(3, 5) 应该返回 5")
	}
}

// 测试配置验证错误
func TestConfigValidationError(t *testing.T) {
	err := ConfigValidationError{
		Field:   "ai_endpoint",
		Message: "必须是有效的 URL 格式",
		Value:   "invalid-url",
	}

	expected := "配置验证失败 [ai_endpoint]: 必须是有效的 URL 格式 (当前值: invalid-url)"
	if err.Error() != expected {
		t.Errorf("ConfigValidationError.Error() 错误，期望: %s, 实际: %s", expected, err.Error())
	}
}

// 测试 AIPipeError
func TestAIPipeError(t *testing.T) {
	err := &AIPipeError{
		Code:     "TEST_ERROR",
		Category: ErrorCategoryProcessing,
		Level:    ErrorLevelError,
		Message:  "测试错误",
	}

	expected := "[processing] TEST_ERROR: 测试错误"
	if err.Error() != expected {
		t.Errorf("AIPipeError.Error() 错误，期望: %s, 实际: %s", expected, err.Error())
	}
}

// 基准测试：配置验证性能
func BenchmarkConfigValidation(b *testing.B) {
	validator := NewConfigValidator()
	config := Config{
		AIEndpoint:  "https://api.openai.com/v1/chat/completions",
		Token:       "sk-1234567890abcdef",
		Model:       "gpt-4",
		MaxRetries:  3,
		Timeout:     30,
		RateLimit:   100,
		LocalFilter: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(&config)
	}
}

// 基准测试：错误处理性能
func BenchmarkErrorHandling(b *testing.B) {
	// 创建一个不输出日志的错误处理器
	handler := &ErrorHandler{
		recovery: &ErrorRecovery{
			strategies: make(map[ErrorCategory]RecoveryStrategy),
			maxRetries: 3,
			backoff:    time.Millisecond,
		},
		logger: nil, // 不设置 logger 避免输出
	}

	handler.RegisterStrategy(ErrorCategoryNetwork, &NetworkErrorRecovery{
		maxRetries: 3,
		backoff:    time.Millisecond,
	})

	err := &AIPipeError{
		Code:     "NETWORK_ERROR",
		Category: ErrorCategoryNetwork,
		Level:    ErrorLevelWarning,
		Message:  "connection timeout",
		Context:  map[string]interface{}{"operation": "test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(err, map[string]interface{}{"test": "value"})
	}
}
