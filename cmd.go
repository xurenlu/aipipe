package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// 根命令
var rootCmd = &cobra.Command{
	Use:   "aipipe",
	Short: "智能日志监控工具",
	Long: `AIPipe 是一个智能日志过滤和监控工具，使用可配置的 AI 服务自动分析日志内容，
过滤不重要的日志，并对重要事件发送通知。

支持多种日志格式：Java、PHP、Nginx、Ruby、Python、FastAPI、journald、syslog等。
支持多源监控：同时监控多个日志文件、journalctl、标准输入等。

快速开始：
  1. 添加日志源：
     aipipe config add "应用名称" file "/var/log/app.log" java
  
  2. 启动监控：
     aipipe
  
  3. 查看配置：
     aipipe config list

常用命令：
  aipipe config add      - 添加日志监控源
  aipipe config list    - 列出所有日志源
  aipipe config test    - 测试配置文件
  aipipe config edit    - 编辑配置文件
  aipipe config remove  - 删除日志源

更多帮助：
  aipipe config --help  - 查看配置管理命令详情
  aipipe --verbose      - 显示详细输出
  aipipe --debug        - 调试模式

文档：
  https://github.com/xurenlu/aipipe/blob/main/README.md`,
	Run: runMain,
}

// 配置文件管理命令
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "配置文件管理",
	Long: `管理AIPipe的配置文件，包括创建、编辑、测试等功能。

配置文件位置：
  - 主配置: ~/.config/aipipe.{json,yaml,toml}
  - 多源配置: ~/.config/aipipe-sources.{json,yaml,toml}

支持的操作：
  add      - 添加日志监控源
  list     - 列出所有日志源
  remove   - 删除日志源
  test     - 测试配置文件
  edit     - 编辑配置文件

配置文件格式：
  - JSON: 自动检测 .json 文件
  - YAML: 自动检测 .yaml 或 .yml 文件
  - TOML: 自动检测 .toml 文件

示例：
  # 添加文件日志源
  aipipe config add "Java应用" file "/var/log/java.log" java
  
  # 添加journalctl日志源
  aipipe config add "系统服务" journalctl "nginx,docker" journald
  
  # 列出所有日志源
  aipipe config list
  
  # 测试配置文件
  aipipe config test
  
  # 编辑配置文件
  aipipe config edit
  
  # 删除日志源
  aipipe config remove "旧服务"`,
}

// 添加日志源命令
var addCmd = &cobra.Command{
	Use:   "add [name] [type] [path] [format]",
	Short: "添加日志监控源",
	Long: `添加新的日志监控源到配置文件中。

参数:
  name    源名称 (如: "Java应用日志")
  type    源类型 (file, journalctl, stdin)
  path    文件路径或journalctl参数
  format  日志格式 (java, php, nginx, journald等)

源类型说明:
  file        - 监控日志文件
  journalctl  - 监控系统日志（systemd journal）
  stdin       - 监控标准输入流

支持的日志格式:
  java, php, nginx, ruby, python, fastapi, go, rust, nodejs, 
  typescript, docker, kubernetes, postgresql, mysql, redis,
  elasticsearch, git, jenkins, github, journald, syslog

示例:
  # 添加Java应用日志
  aipipe config add "Java应用" file "/var/log/java.log" java
  
  # 添加PHP应用日志
  aipipe config add "PHP应用" file "/var/log/php.log" php
  
  # 添加Nginx访问日志
  aipipe config add "Nginx访问" file "/var/log/nginx/access.log" nginx
  
  # 添加系统服务监控（journalctl）
  aipipe config add "系统服务" journalctl "nginx,docker,postgresql" journald
  
  # 添加数据库服务监控
  aipipe config add "数据库" journalctl "postgresql,mysql" journald
  
  # 添加Docker日志
  aipipe config add redis "Docker容器" file "/var/log/docker.log" docker

注意事项:
  - 源名称是唯一的，不能重复
  - 文件路径必须存在或将要存在
  - journalctl类型需要root权限`,
	Args: cobra.ExactArgs(4),
	Run:  runAddSource,
}

// 删除日志源命令
var removeCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "删除日志监控源",
	Long: `从配置文件中删除指定的日志监控源。

参数:
  name    要删除的源名称（必须完全匹配）

操作说明:
  - 会永久删除配置中的日志源
  - 删除前建议使用 'list' 命令查看源名称
  - 删除操作不可恢复

示例:
  # 查看所有日志源
  aipipe config list
  
  # 删除指定的日志源
  aipipe config remove "Java应用日志"
  aipipe config remove "系统服务监控"
  aipipe config remove "旧服务"

注意事项:
  - 源名称必须与配置中的完全一致
  - 删除后需要重新运行监控程序才能生效`,
	Args: cobra.ExactArgs(1),
	Run:  runRemoveSource,
}

// 列出日志源命令
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有日志监控源",
	Long: `显示当前配置文件中所有的日志监控源。

输出信息:
  - 源名称和类型
  - 文件路径或服务列表
  - 日志格式
  - 启用状态
  - 优先级

示例:
  # 列出所有日志源
  aipipe config list
  
  # 输出示例:
  # 📋 日志监控源列表:
  # 1. Java应用日志 (file) - ✅ 启用
  #    路径: /var/log/java.log
  #    格式: java
  #    描述: 监控Java应用程序日志
  # 2. 系统服务监控 (journalctl) - ✅ 启用
  #    服务: nginx, docker, postgresql`,
	Run:   runListSources,
}

// 测试配置文件命令
var testCmd = &cobra.Command{
	Use:   "test [config-file]",
	Short: "测试配置文件",
	Long: `测试配置文件的格式和内容是否正确。

参数:
  config-file  配置文件路径 (可选，默认自动检测)

功能:
  - 检测配置文件格式 (JSON/YAML/TOML)
  - 验证配置文件语法
  - 解析并显示配置内容
  - 检查必要字段是否存在

示例:
  # 测试默认配置文件
  aipipe config test
  
  # 测试指定配置文件
  aipipe config test ~/.config/aipipe.yaml
  aipipe config test ~/.config/aipipe.json
  aipipe config test ~/.config/aipipe.toml
  
  # 输出示例:
  # 🔍 测试配置文件: /Users/user/.config/aipipe.json
  # 📄 检测到格式: json
  # ✅ 主配置解析成功
  #    AI端点: https://api.openai.com/v1/chat/completions
  #    模型: gpt-4

常见问题:
  - 如果配置文件格式错误，会显示具体错误信息
  - 如果文件不存在，会提示创建默认配置
  - 如果必要字段缺失，会提示使用默认值`,
	Args: cobra.MaximumNArgs(1),
	Run:  runTestConfig,
}

// 编辑配置文件命令
var editCmd = &cobra.Command{
	Use:   "edit [config-file]",
	Short: "编辑配置文件",
	Long: `使用默认编辑器打开配置文件进行编辑。

参数:
  config-file  配置文件路径 (可选，默认自动检测)

编辑器选择:
  1. 优先使用环境变量 EDITOR 指定的编辑器
  2. 默认使用 vim 编辑器

设置编辑器:
  export EDITOR=nano      # 使用nano编辑器
  export EDITOR=code      # 使用VS Code编辑器
  export EDITOR=vim       # 使用vim编辑器
  export EDITOR=emacs     # 使用emacs编辑器

示例:
  # 编辑默认配置文件
  aipipe config edit
  
  # 编辑指定配置文件
  aipipe config edit ~/.config/aipipe.yaml
  aipipe config edit ~/.config/aipipe.json
  
  # 使用特定编辑器
  export EDITOR=nano
  aipipe config edit

保存配置:
  - 编辑完成后保存并退出编辑器
  - 建议使用 'test' 命令验证配置正确性
  - 配置修改后需要重启监控程序才能生效`,
	Args: cobra.MaximumNArgs(1),
	Run:  runEditConfig,
}

// 初始化命令
func init() {
	// 添加全局标志
	rootCmd.PersistentFlags().StringVar(configFile, "config", "", "指定配置文件路径")
	rootCmd.PersistentFlags().BoolVar(verbose, "verbose", false, "显示详细输出")
	rootCmd.PersistentFlags().BoolVar(debug, "debug", false, "调试模式")

	// 添加子命令
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(addCmd)
	configCmd.AddCommand(removeCmd)
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(testCmd)
	configCmd.AddCommand(editCmd)

	// 设置配置文件自动检测
	cobra.OnInitialize(initConfig)
}

// 初始化配置
func initConfig() {
	if *configFile != "" {
		viper.SetConfigFile(*configFile)
	} else {
		// 自动检测配置文件
		configPath, err := findDefaultConfig()
		if err == nil {
			viper.SetConfigFile(configPath)
		} else {
			viper.SetConfigName("aipipe")
			viper.SetConfigType("yaml")
			viper.AddConfigPath("$HOME/.config")
		}
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("❌ 读取配置文件失败: %v\n", err)
		}
	}
}

// 运行主程序
func runMain(cmd *cobra.Command, args []string) {
	// 检查是否使用多源监控
	if shouldUseMultiSource() {
		processMultiSource()
		return
	}

	// 加载配置文件
	if err := loadConfig(); err != nil {
		fmt.Printf("⚠️  加载配置文件失败，使用默认配置: %v\n", err)
		globalConfig = defaultConfig
	}

	fmt.Printf("🚀 AIPipe 启动 - 监控 %s 格式日志\n", *logFormat)

	// 显示模式提示
	if !*showNotImportant {
		fmt.Println("💡 只显示重要日志（过滤的日志不显示）")
		if !*verbose {
			fmt.Println("   使用 --show-not-important 显示所有日志")
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if *filePath != "" {
		// 文件监控模式
		fmt.Printf("📁 监控文件: %s\n", *filePath)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if err := watchFile(*filePath); err != nil {
			fmt.Printf("❌ 监控文件失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		// 标准输入模式
		fmt.Println("📥 从标准输入读取日志...")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		if *noBatch {
			processStdin()
		} else {
			processStdinWithBatch()
		}
	}
}

// 添加日志源
func runAddSource(cmd *cobra.Command, args []string) {
	name := args[0]
	sourceType := args[1]
	path := args[2]
	format := args[3]

	// 加载多源配置
	configPath, err := findMultiSourceConfig()
	if err != nil {
		fmt.Printf("❌ 查找多源配置文件失败: %v\n", err)
		return
	}

	config, err := loadMultiSourceConfig(configPath)
	if err != nil {
		fmt.Printf("❌ 加载多源配置文件失败: %v\n", err)
		return
	}

	// 检查源是否已存在
	for _, source := range config.Sources {
		if source.Name == name {
			fmt.Printf("❌ 源 '%s' 已存在\n", name)
			return
		}
	}

	// 创建新源
	newSource := SourceConfig{
		Name:        name,
		Type:        sourceType,
		Path:        path,
		Format:      format,
		Enabled:     true,
		Priority:    len(config.Sources) + 1,
		Description: fmt.Sprintf("监控%s日志", name),
	}

	// 如果是journalctl类型，解析服务参数
	if sourceType == "journalctl" {
		services := strings.Split(path, ",")
		for i, service := range services {
			services[i] = strings.TrimSpace(service)
		}
		newSource.Journal = &JournalConfig{
			Services: services,
			Priority: "err",
		}
	}

	// 添加新源
	config.Sources = append(config.Sources, newSource)

	// 保存配置
	if err := saveMultiSourceConfig(configPath, config); err != nil {
		fmt.Printf("❌ 保存配置文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 成功添加日志源: %s (%s)\n", name, sourceType)
}

// 删除日志源
func runRemoveSource(cmd *cobra.Command, args []string) {
	name := args[0]

	// 加载多源配置
	configPath, err := findMultiSourceConfig()
	if err != nil {
		fmt.Printf("❌ 查找多源配置文件失败: %v\n", err)
		return
	}

	config, err := loadMultiSourceConfig(configPath)
	if err != nil {
		fmt.Printf("❌ 加载多源配置文件失败: %v\n", err)
		return
	}

	// 查找并删除源
	found := false
	for i, source := range config.Sources {
		if source.Name == name {
			config.Sources = append(config.Sources[:i], config.Sources[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ 未找到源: %s\n", name)
		return
	}

	// 保存配置
	if err := saveMultiSourceConfig(configPath, config); err != nil {
		fmt.Printf("❌ 保存配置文件失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 成功删除日志源: %s\n", name)
}

// 列出日志源
func runListSources(cmd *cobra.Command, args []string) {
	// 加载多源配置
	configPath, err := findMultiSourceConfig()
	if err != nil {
		fmt.Printf("❌ 查找多源配置文件失败: %v\n", err)
		return
	}

	config, err := loadMultiSourceConfig(configPath)
	if err != nil {
		fmt.Printf("❌ 加载多源配置文件失败: %v\n", err)
		return
	}

	if len(config.Sources) == 0 {
		fmt.Println("📋 没有配置的日志源")
		return
	}

	fmt.Printf("📋 日志监控源列表 (%s):\n", configPath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, source := range config.Sources {
		status := "❌ 禁用"
		if source.Enabled {
			status = "✅ 启用"
		}

		fmt.Printf("%d. %s (%s) - %s\n", i+1, source.Name, source.Type, status)
		fmt.Printf("   路径: %s\n", source.Path)
		fmt.Printf("   格式: %s\n", source.Format)
		fmt.Printf("   描述: %s\n", source.Description)
		if source.Journal != nil {
			fmt.Printf("   服务: %s\n", strings.Join(source.Journal.Services, ", "))
		}
		fmt.Println()
	}
}

// 测试配置文件
func runTestConfig(cmd *cobra.Command, args []string) {
	var configPath string
	var err error

	if len(args) > 0 {
		configPath = args[0]
	} else {
		configPath, err = findDefaultConfig()
		if err != nil {
			fmt.Printf("❌ 查找配置文件失败: %v\n", err)
			return
		}
	}

	fmt.Printf("🔍 测试配置文件: %s\n", configPath)

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ 配置文件不存在: %s\n", configPath)
		return
	}

	// 检测文件格式
	format := detectConfigFormat(configPath)
	fmt.Printf("📄 检测到格式: %s\n", format)

	// 测试解析
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("❌ 读取文件失败: %v\n", err)
		return
	}

	// 测试主配置
	var mainConfig Config
	if err := parseConfigFile(data, format, &mainConfig); err != nil {
		fmt.Printf("❌ 解析主配置失败: %v\n", err)
	} else {
		fmt.Printf("✅ 主配置解析成功\n")
		fmt.Printf("   AI端点: %s\n", mainConfig.AIEndpoint)
		fmt.Printf("   模型: %s\n", mainConfig.Model)
	}

	// 测试多源配置
	var multiConfig MultiSourceConfig
	if err := parseConfigFile(data, format, &multiConfig); err == nil && len(multiConfig.Sources) > 0 {
		fmt.Printf("✅ 多源配置解析成功\n")
		fmt.Printf("   源数量: %d\n", len(multiConfig.Sources))
		for _, source := range multiConfig.Sources {
			fmt.Printf("   - %s (%s)\n", source.Name, source.Type)
		}
	}

	fmt.Println("✅ 配置文件测试完成")
}

// 编辑配置文件
func runEditConfig(cmd *cobra.Command, args []string) {
	var configPath string
	var err error

	if len(args) > 0 {
		configPath = args[0]
	} else {
		configPath, err = findDefaultConfig()
		if err != nil {
			fmt.Printf("❌ 查找配置文件失败: %v\n", err)
			return
		}
	}

	fmt.Printf("📝 编辑配置文件: %s\n", configPath)

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ 配置文件不存在: %s\n", configPath)
		return
	}

	// 获取编辑器
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// 打开编辑器
	cmd_exec := exec.Command(editor, configPath)
	cmd_exec.Stdin = os.Stdin
	cmd_exec.Stdout = os.Stdout
	cmd_exec.Stderr = os.Stderr

	if err := cmd_exec.Run(); err != nil {
		fmt.Printf("❌ 打开编辑器失败: %v\n", err)
		return
	}

	fmt.Println("✅ 配置文件编辑完成")
}

// 保存多源配置文件
func saveMultiSourceConfig(configPath string, config *MultiSourceConfig) error {
	// 检测文件格式
	format := detectConfigFormat(configPath)

	// 根据格式保存
	switch format {
	case "json":
		return saveJSONConfig(configPath, config)
	case "yaml":
		return saveYAMLConfig(configPath, config)
	case "toml":
		return saveTOMLConfig(configPath, config)
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", format)
	}
}

// 保存JSON配置
func saveJSONConfig(configPath string, config *MultiSourceConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// 保存YAML配置
func saveYAMLConfig(configPath string, config *MultiSourceConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// 保存TOML配置
func saveTOMLConfig(configPath string, config *MultiSourceConfig) error {
	var buf strings.Builder
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return os.WriteFile(configPath, []byte(buf.String()), 0644)
}
