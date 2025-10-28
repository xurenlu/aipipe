package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// 带context的邮件发送函数
func sendEmailNotificationWithContext(ctx context.Context, summary, logLine string) error {
	emailConfig := globalConfig.Notifiers.Email

	if !emailConfig.Enabled || len(emailConfig.ToEmails) == 0 {
		return nil
	}

	subject := fmt.Sprintf("⚠️ 重要日志告警: %s", summary)
	body := fmt.Sprintf(`
重要日志告警

摘要: %s

日志内容:
%s

文件: %s

时间: %s
来源: AIPipe 日志监控系统
`, summary, logLine, currentLogFile, time.Now().Format("2006-01-02 15:04:05"))

	var err error
	if emailConfig.Provider == "resend" {
		err = sendResendEmailWithContext(ctx, emailConfig, subject, body)
	} else {
		err = sendSMTPEmailWithContext(ctx, emailConfig, subject, body)
	}

	if err != nil {
		return fmt.Errorf("邮件发送失败: %w", err)
	}

	if *verbose || *debug {
		log.Printf("✅ 邮件已发送: %s", subject)
	}
	return nil
}

// 发送邮件通知（兼容旧接口）
func sendEmailNotification(summary, logLine string) {
	ctx := context.Background()
	if err := sendEmailNotificationWithContext(ctx, summary, logLine); err != nil {
		if *verbose || *debug {
			log.Printf("❌ 邮件发送失败: %v", err)
		}
	}
}

// 带context的SMTP邮件发送
func sendSMTPEmailWithContext(ctx context.Context, config EmailConfig, subject, body string) error {
	// 检查context是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 构建邮件内容
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		config.FromEmail, strings.Join(config.ToEmails, ","), subject, body)

	// 构建SMTP地址
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// 创建TLS配置
	tlsConfig := &tls.Config{
		ServerName: config.Host,
	}

	// 建立连接
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Quit()

	// 认证
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	// 发送邮件
	if err := client.Mail(config.FromEmail); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	for _, to := range config.ToEmails {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取数据写入器失败: %w", err)
	}
	defer writer.Close()

	if _, err := writer.Write([]byte(msg)); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	return nil
}

// 带context的Resend邮件发送
func sendResendEmailWithContext(ctx context.Context, config EmailConfig, subject, body string) error {
	// 检查context是否已取消
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 构建请求
	payload := map[string]interface{}{
		"from":    config.FromEmail,
		"to":      config.ToEmails,
		"subject": subject,
		"html":    body,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Password) // 使用password字段存储API key

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API错误 %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 通过SMTP发送邮件
func sendSMTPEmail(config EmailConfig, subject, body string) error {
	// 构建邮件内容
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		config.FromEmail, strings.Join(config.ToEmails, ","), subject, body)

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)

	// 使用统一的SMTP发送方式
	err := smtp.SendMail(addr, auth, config.FromEmail, config.ToEmails, []byte(message))

	return err
}

// SSL邮件发送

// 通过Resend API发送邮件
func sendResendEmail(config EmailConfig, subject, body string) error {
	type ResendRequest struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		Text    string   `json:"text"`
	}

	reqBody := ResendRequest{
		From:    config.FromEmail,
		To:      config.ToEmails,
		Subject: subject,
		Text:    body,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.Password)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error: %s", string(body))
	}

	return nil
}

// 安全发送邮件通知（带panic恢复和超时控制）
func safeSendEmailNotification(summary, logLine string) {
	defer func() {
		if r := recover(); r != nil {
			if *verbose || *debug {
				log.Printf("❌ 邮件通知panic恢复: %v", r)
			}
		}
	}()

	// 使用context控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用channel控制并发
	done := make(chan error, 1)
	go func() {
		done <- sendEmailNotificationWithContext(ctx, summary, logLine)
	}()

	select {
	case err := <-done:
		if err != nil && (*verbose || *debug) {
			log.Printf("❌ 邮件发送失败: %v", err)
		}
	case <-ctx.Done():
		if *verbose || *debug {
			log.Printf("❌ 邮件发送超时: %v", ctx.Err())
		}
	}
}
