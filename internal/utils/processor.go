package utils

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xurenlu/aipipe/internal/config"
)

// 处理标准输入
func ProcessStdin(cfg *config.Config, showNotImportant bool) {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineCount := 0
	filteredCount := 0
	alertCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if strings.TrimSpace(line) == "" {
			continue
		}

		// 简单的本地过滤逻辑
		if shouldFilter(line) {
			filteredCount++
			if showNotImportant {
				fmt.Printf("🔇 [过滤] %s\n", line)
			}
			continue
		}

		// 重要日志，显示并发送通知
		fmt.Printf("⚠️  [重要] %s\n", line)
		fmt.Printf("   📝 摘要: %s\n", generateSummary(line))

		// 发送通知
		go sendNotification(generateSummary(line), line)
		alertCount++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ 读取输入失败: %v\n", err)
		return
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 统计: 总计 %d 行, 过滤 %d 行, 告警 %d 次\n", lineCount, filteredCount, alertCount)
}

// 简单的过滤逻辑
func shouldFilter(line string) bool {
	upperLine := strings.ToUpper(line)

	// 过滤掉明显的低级别日志
	lowLevelKeywords := []string{
		"DEBUG", "TRACE", "VERBOSE", "INFO",
		"STARTING", "STARTED", "STOPPING", "STOPPED",
		"CONNECTED", "DISCONNECTED", "INITIALIZED",
		"HEALTH CHECK", "PING", "PONG",
	}

	for _, keyword := range lowLevelKeywords {
		if strings.Contains(upperLine, keyword) {
			// 但如果包含错误关键词，不过滤
			if containsErrorKeywords(upperLine) {
				continue
			}
			return true
		}
	}

	return false
}

// 检查是否包含错误关键词
func containsErrorKeywords(line string) bool {
	errorKeywords := []string{
		"ERROR", "EXCEPTION", "FATAL", "CRITICAL",
		"FAILED", "FAILURE", "TIMEOUT", "OUT OF MEMORY",
		"CONNECTION REFUSED", "ACCESS DENIED", "PERMISSION DENIED",
		"NOT FOUND", "UNAVAILABLE", "DOWN", "OFFLINE",
	}

	for _, keyword := range errorKeywords {
		if strings.Contains(line, keyword) {
			return true
		}
	}

	return false
}

// 生成日志摘要
func generateSummary(line string) string {
	upperLine := strings.ToUpper(line)

	// 根据关键词生成摘要
	if strings.Contains(upperLine, "ERROR") {
		return "检测到错误日志"
	}
	if strings.Contains(upperLine, "EXCEPTION") {
		return "检测到异常"
	}
	if strings.Contains(upperLine, "FATAL") {
		return "检测到致命错误"
	}
	if strings.Contains(upperLine, "TIMEOUT") {
		return "检测到超时问题"
	}
	if strings.Contains(upperLine, "CONNECTION") {
		return "检测到连接问题"
	}
	if strings.Contains(upperLine, "MEMORY") {
		return "检测到内存问题"
	}
	if strings.Contains(upperLine, "DATABASE") {
		return "检测到数据库问题"
	}

	return "检测到重要日志"
}

// 发送通知
func sendNotification(summary, content string) {
	// 截断内容
	displayContent := content
	if len(displayContent) > 100 {
		displayContent = displayContent[:100] + "..."
	}

	// 发送系统通知
	sendSystemNotification(summary, displayContent)
}

// 发送系统通知
func sendSystemNotification(summary, content string) {
	// 检测操作系统并发送相应的通知
	if isMacOS() {
		sendMacOSNotification(summary, content)
	} else if isLinux() {
		sendLinuxNotification(summary, content)
	}
}

// 检测是否为 macOS
func isMacOS() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "darwin")
}

// 检测是否为 Linux
func isLinux() bool {
	return strings.Contains(strings.ToLower(os.Getenv("GOOS")), "linux")
}

// 发送 macOS 通知
func sendMacOSNotification(summary, content string) {
	// 使用 osascript 发送通知
	script := fmt.Sprintf(`display notification "%s" with title "⚠️ 重要日志告警" subtitle "%s"`,
		escapeForAppleScript(content),
		escapeForAppleScript(summary))

	cmd := exec.Command("osascript", "-")
	cmd.Stdin = strings.NewReader(script)
	cmd.Env = append(os.Environ(), "LANG=zh_CN.UTF-8")

	err := cmd.Run()
	if err != nil {
		// 静默失败
	}
}

// 发送 Linux 通知
func sendLinuxNotification(summary, content string) {
	cmd := exec.Command("notify-send",
		"⚠️ 重要日志告警",
		fmt.Sprintf("%s\n%s", summary, content),
		"--urgency=critical",
		"--expire-time=10000")

	err := cmd.Run()
	if err != nil {
		// 静默失败
	}
}

// 转义 AppleScript 字符串
func escapeForAppleScript(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
