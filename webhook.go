package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 构建企业微信webhook payload
func buildWeChatPayload(summary, logLine string) map[string]interface{} {
	content := fmt.Sprintf("⚠️ 重要日志告警\n\n📋 摘要: %s\n\n📝 日志内容:\n%s\n\n📁 文件: %s\n\n⏰ 时间: %s",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
}

// 构建飞书webhook payload
func buildFeishuPayload(summary, logLine string) map[string]interface{} {
	// 构建更详细的飞书通知内容
	content := fmt.Sprintf("⚠️ 重要日志告警\n\n📋 摘要: %s\n\n📝 日志内容:\n%s\n\n📁 文件: %s\n\n⏰ 时间: %s\n\n🔍 来源: AIPipe 日志监控系统",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
		},
	}
}

// 构建Slack webhook payload
func buildSlackPayload(summary, logLine string) map[string]interface{} {
	text := fmt.Sprintf("⚠️ 重要日志告警\n\n*摘要:* %s\n\n*日志内容:*\n```\n%s\n```\n\n*文件:* `%s`\n\n*时间:* %s",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"text":       text,
		"username":   "AIPipe",
		"icon_emoji": ":warning:",
	}
}

// 构建通用webhook payload
func buildGenericPayload(summary, logLine string) map[string]interface{} {
	return map[string]interface{}{
		"summary":   summary,
		"log_line":  logLine,
		"log_file":  currentLogFile,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"source":    "AIPipe",
		"level":     "warning",
	}
}

// 添加webhook签名
func addWebhookSignature(req *http.Request, body []byte, secret, webhookType string) {
	// 这里可以实现不同webhook平台的签名算法
	// 目前只是占位符实现
	switch webhookType {
	case "dingtalk":
		// 钉钉签名实现
		// req.Header.Set("X-DingTalk-Signature", signature)
	case "wechat":
		// 企业微信签名实现
		// req.Header.Set("X-WeChat-Signature", signature)
	case "feishu":
		// 飞书签名实现
		// req.Header.Set("X-Feishu-Signature", signature)
	case "slack":
		// Slack签名实现
		// req.Header.Set("X-Slack-Signature", signature)
	default:
		// 通用签名
		// req.Header.Set("X-Webhook-Signature", signature)
	}
}

// 智能识别webhook类型
func detectWebhookType(webhookURL string) string {
	u, err := url.Parse(webhookURL)
	if err != nil {
		return "custom"
	}

	host := strings.ToLower(u.Host)
	path := strings.ToLower(u.Path)

	// 钉钉
	if strings.Contains(host, "dingtalk") || strings.Contains(path, "dingtalk") {
		return "dingtalk"
	}

	// 企业微信
	if strings.Contains(host, "qyapi.weixin.qq.com") || strings.Contains(path, "wechat") {
		return "wechat"
	}

	// 飞书
	if strings.Contains(host, "feishu") || strings.Contains(path, "feishu") {
		return "feishu"
	}

	// Slack
	if strings.Contains(host, "slack.com") || strings.Contains(path, "slack") {
		return "slack"
	}

	return "custom"
}

// 安全发送webhook通知（带panic恢复和超时控制）
func safeSendWebhookNotification(config WebhookConfig, summary, logLine, webhookType string) {
	defer func() {
		if r := recover(); r != nil {
			if *verbose || *debug {
				log.Printf("❌ %s webhook通知panic恢复: %v", webhookType, r)
			}
		}
	}()

	// 使用context控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用channel控制并发
	done := make(chan error, 1)
	go func() {
		done <- sendWebhookNotificationWithContext(ctx, config, summary, logLine, webhookType)
	}()

	select {
	case err := <-done:
		if err != nil && (*verbose || *debug) {
			log.Printf("❌ %s webhook发送失败: %v", webhookType, err)
		}
	case <-ctx.Done():
		if *verbose || *debug {
			log.Printf("❌ %s webhook发送超时: %v", webhookType, ctx.Err())
		}
	}
}

// 带context的webhook发送函数
func sendWebhookNotificationWithContext(ctx context.Context, config WebhookConfig, summary, logLine, webhookType string) error {
	if !config.Enabled || config.URL == "" {
		return nil
	}

	var payload interface{}

	// 根据webhook类型构建不同的payload
	switch webhookType {
	case "dingtalk":
		payload = buildDingTalkPayload(summary, logLine)
	case "wechat":
		payload = buildWeChatPayload(summary, logLine)
	case "feishu":
		payload = buildFeishuPayload(summary, logLine)
	case "slack":
		payload = buildSlackPayload(summary, logLine)
	default:
		payload = buildGenericPayload(summary, logLine)
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("构建webhook payload失败: %w", err)
	}

	req, err := http.NewRequest("POST", config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建webhook请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 如果配置了签名密钥，添加签名
	if config.Secret != "" {
		addWebhookSignature(req, jsonData, config.Secret, webhookType)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("发送webhook失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook响应错误 %d: %s", resp.StatusCode, string(body))
	}

	if *verbose || *debug {
		log.Printf("✅ %s webhook已发送: %s", webhookType, summary)
	}
	return nil
}

// 发送webhook通知（兼容旧接口）
func sendWebhookNotification(config WebhookConfig, summary, logLine, webhookType string) {
	ctx := context.Background()
	if err := sendWebhookNotificationWithContext(ctx, config, summary, logLine, webhookType); err != nil {
		if *verbose || *debug {
			log.Printf("❌ %s webhook发送失败: %v", webhookType, err)
		}
	}
}

// 构建钉钉webhook payload
func buildDingTalkPayload(summary, logLine string) map[string]interface{} {
	content := fmt.Sprintf("⚠️ 重要日志告警\n\n📋 摘要: %s\n\n📝 日志内容:\n%s\n\n📁 文件: %s\n\n⏰ 时间: %s",
		summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
}
