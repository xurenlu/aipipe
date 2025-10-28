package notification

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os/exec"
	"strings"
	"time"

	"github.com/xurenlu/aipipe/internal/config"
)

// 通知消息
type NotificationMessage struct {
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Level       string            `json:"level"`
	Timestamp   time.Time         `json:"timestamp"`
	Source      string            `json:"source"`
	Metadata    map[string]string `json:"metadata"`
	Attachments []string          `json:"attachments,omitempty"`
}

// 通知器接口
type Notifier interface {
	Send(message *NotificationMessage) error
	IsEnabled() bool
	GetName() string
}

// 邮件通知器
type EmailNotifier struct {
	config config.EmailConfig
}

func NewEmailNotifier(cfg config.EmailConfig) *EmailNotifier {
	return &EmailNotifier{config: cfg}
}

func (e *EmailNotifier) Send(message *NotificationMessage) error {
	if !e.config.Enabled {
		return fmt.Errorf("邮件通知未启用")
	}

	// 构建邮件内容
	subject := fmt.Sprintf("[%s] %s", message.Level, message.Title)
	body := fmt.Sprintf(`
时间: %s
来源: %s
级别: %s

内容:
%s

---
AIPipe 日志监控系统
`, message.Timestamp.Format("2006-01-02 15:04:05"), message.Source, message.Level, message.Content)

	// 发送邮件
	if e.config.Provider == "smtp" {
		return e.sendSMTP(subject, body)
	} else if e.config.Provider == "resend" {
		return e.sendResend(subject, body)
	}

	return fmt.Errorf("不支持的邮件提供商: %s", e.config.Provider)
}

func (e *EmailNotifier) sendSMTP(subject, body string) error {
	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = e.config.FromEmail
	headers["To"] = strings.Join(e.config.ToEmails, ", ")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"

	// 构建邮件内容
	var emailContent bytes.Buffer
	for k, v := range headers {
		emailContent.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	emailContent.WriteString("\r\n")
	emailContent.WriteString(body)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)
	
	return smtp.SendMail(addr, auth, e.config.FromEmail, e.config.ToEmails, emailContent.Bytes())
}

func (e *EmailNotifier) sendResend(subject, body string) error {
	// Resend API 实现
	// 这里需要实现 Resend API 调用
	return fmt.Errorf("Resend API 暂未实现")
}

func (e *EmailNotifier) IsEnabled() bool {
	return e.config.Enabled
}

func (e *EmailNotifier) GetName() string {
	return "email"
}

// Webhook 通知器
type WebhookNotifier struct {
	config config.WebhookConfig
	name   string
}

func NewWebhookNotifier(cfg config.WebhookConfig, name string) *WebhookNotifier {
	return &WebhookNotifier{config: cfg, name: name}
}

func (w *WebhookNotifier) Send(message *NotificationMessage) error {
	if !w.config.Enabled {
		return fmt.Errorf("Webhook 通知未启用")
	}

	// 构建请求数据
	data := map[string]interface{}{
		"title":     message.Title,
		"content":   message.Content,
		"level":     message.Level,
		"timestamp": message.Timestamp.Format(time.RFC3339),
		"source":    message.Source,
		"metadata":  message.Metadata,
	}

	// 添加特定平台的格式
	switch w.name {
	case "dingtalk":
		data = w.formatDingTalk(data)
	case "wechat":
		data = w.formatWeChat(data)
	case "feishu":
		data = w.formatFeishu(data)
	case "slack":
		data = w.formatSlack(data)
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", w.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// 添加签名（如果配置了密钥）
	if w.config.Secret != "" {
		signature := w.generateSignature(jsonData, w.config.Secret)
		req.Header.Set("X-Signature", signature)
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func (w *WebhookNotifier) formatDingTalk(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("[%s] %s\n%s", 
				data["level"], data["title"], data["content"]),
		},
	}
}

func (w *WebhookNotifier) formatWeChat(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("[%s] %s\n%s", 
				data["level"], data["title"], data["content"]),
		},
	}
}

func (w *WebhookNotifier) formatFeishu(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("[%s] %s\n%s", 
				data["level"], data["title"], data["content"]),
		},
	}
}

func (w *WebhookNotifier) formatSlack(data map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"text": fmt.Sprintf("[%s] %s\n%s", 
			data["level"], data["title"], data["content"]),
	}
}

func (w *WebhookNotifier) generateSignature(data []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (w *WebhookNotifier) IsEnabled() bool {
	return w.config.Enabled
}

func (w *WebhookNotifier) GetName() string {
	return w.name
}

// 系统通知器
type SystemNotifier struct {
	enabled bool
}

func NewSystemNotifier() *SystemNotifier {
	return &SystemNotifier{enabled: true}
}

func (s *SystemNotifier) Send(message *NotificationMessage) error {
	if !s.enabled {
		return fmt.Errorf("系统通知未启用")
	}

	// 检测操作系统并使用相应的通知命令
	if s.isMacOS() {
		return s.sendMacOSNotification(message)
	} else if s.isLinux() {
		return s.sendLinuxNotification(message)
	}

	return fmt.Errorf("不支持的操作系统")
}

func (s *SystemNotifier) isMacOS() bool {
	_, err := exec.LookPath("osascript")
	return err == nil
}

func (s *SystemNotifier) isLinux() bool {
	_, err := exec.LookPath("notify-send")
	return err == nil
}

func (s *SystemNotifier) sendMacOSNotification(message *NotificationMessage) error {
	title := fmt.Sprintf("[%s] %s", message.Level, message.Title)
	content := message.Content
	
	// 使用 osascript 发送 macOS 通知
	cmd := exec.Command("osascript", "-e", 
		fmt.Sprintf("display notification \"%s\" with title \"%s\"", content, title))
	
	return cmd.Run()
}

func (s *SystemNotifier) sendLinuxNotification(message *NotificationMessage) error {
	title := fmt.Sprintf("[%s] %s", message.Level, message.Title)
	content := message.Content
	
	// 使用 notify-send 发送 Linux 通知
	cmd := exec.Command("notify-send", title, content)
	
	return cmd.Run()
}

func (s *SystemNotifier) IsEnabled() bool {
	return s.enabled
}

func (s *SystemNotifier) GetName() string {
	return "system"
}

// 通知管理器
type NotificationManager struct {
	notifiers []Notifier
}

func NewNotificationManager(cfg *config.Config) *NotificationManager {
	nm := &NotificationManager{
		notifiers: make([]Notifier, 0),
	}

	// 添加邮件通知器
	if cfg.Notifiers.Email.Enabled {
		nm.notifiers = append(nm.notifiers, NewEmailNotifier(cfg.Notifiers.Email))
	}

	// 添加 Webhook 通知器
	if cfg.Notifiers.DingTalk.Enabled {
		nm.notifiers = append(nm.notifiers, NewWebhookNotifier(cfg.Notifiers.DingTalk, "dingtalk"))
	}
	if cfg.Notifiers.WeChat.Enabled {
		nm.notifiers = append(nm.notifiers, NewWebhookNotifier(cfg.Notifiers.WeChat, "wechat"))
	}
	if cfg.Notifiers.Feishu.Enabled {
		nm.notifiers = append(nm.notifiers, NewWebhookNotifier(cfg.Notifiers.Feishu, "feishu"))
	}
	if cfg.Notifiers.Slack.Enabled {
		nm.notifiers = append(nm.notifiers, NewWebhookNotifier(cfg.Notifiers.Slack, "slack"))
	}

	// 添加自定义 Webhook 通知器
	for _, webhook := range cfg.Notifiers.CustomWebhooks {
		if webhook.Enabled {
			nm.notifiers = append(nm.notifiers, NewWebhookNotifier(webhook, "custom"))
		}
	}

	// 添加系统通知器
	nm.notifiers = append(nm.notifiers, NewSystemNotifier())

	return nm
}

// 发送通知
func (nm *NotificationManager) Send(message *NotificationMessage) error {
	var errors []string

	for _, notifier := range nm.notifiers {
		if notifier.IsEnabled() {
			if err := notifier.Send(message); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", notifier.GetName(), err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("通知发送失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// 发送简单通知
func (nm *NotificationManager) SendSimple(title, content, level string) error {
	message := &NotificationMessage{
		Title:     title,
		Content:   content,
		Level:     level,
		Timestamp: time.Now(),
		Source:    "AIPipe",
		Metadata:  make(map[string]string),
	}

	return nm.Send(message)
}

// 获取启用的通知器数量
func (nm *NotificationManager) GetEnabledCount() int {
	count := 0
	for _, notifier := range nm.notifiers {
		if notifier.IsEnabled() {
			count++
		}
	}
	return count
}
