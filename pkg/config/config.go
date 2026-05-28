package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
)

const defaultConfigPath = "configs/configs.yaml"

// Config 描述应用全局配置
type Config struct {
	App    AppConfig    `yaml:"app"`
	Log    LogConfig    `yaml:"log"`
	Agent  AgentConfig  `yaml:"agent"`
	LLM    LLMConfig    `yaml:"llm"`
	RAG    RAGConfig    `yaml:"rag"`
	Tools  ToolsConfig  `yaml:"tools"`
	Server ServerConfig `yaml:"server"`
}

// AppConfig 描述应用基础信息
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Env     string `yaml:"env"`
	Mode    string `yaml:"mode"`
}

// LogConfig 描述日志配置
type LogConfig struct {
	Level      string `yaml:"level"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// AgentConfig 描述 Agent 行为开关
type AgentConfig struct {
	EnableDemo bool `yaml:"enable_demo"`
}

// LLMConfig 描述模型调用配置
type LLMConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	APIKey   string `yaml:"api_key"`
}

// RAGConfig 描述检索增强配置
type RAGConfig struct {
	Enabled bool `yaml:"enabled"`
}

// ToolsConfig 描述工具调用配置
type ToolsConfig struct {
	Enabled bool `yaml:"enabled"`
}

// ServerConfig 描述进程关闭配置
type ServerConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	ShutdownTimeoutSeconds int    `yaml:"shutdown_timeout_seconds"`
}

var globalConfig *Config

// Load 读取配置文件并应用环境变量覆盖
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = getEnv("CONFIG_PATH", defaultConfigPath)
	}

	cfg := Default()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	applyEnv(cfg)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	globalConfig = cfg
	return cfg, nil
}

// MustLoad 加载配置并在失败时 panic
func MustLoad(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		panic(err)
	}
	return cfg
}

// Get 获取全局配置
func Get() *Config {
	if globalConfig == nil {
		panic("配置未初始化，请先调用 Load")
	}
	return globalConfig
}

// Default 返回可直接启动的默认配置
func Default() *Config {
	return &Config{
		App: AppConfig{
			Name:    "solvify-agent",
			Version: "0.1.0",
			Env:     "development",
			Mode:    "release",
		},
		Log: LogConfig{
			Level:      "info",
			Filename:   "logs/solvify-agent.log",
			MaxSize:    100,
			MaxBackups: 7,
			MaxAge:     30,
			Compress:   true,
		},
		Agent: AgentConfig{
			EnableDemo: true,
		},
		LLM: LLMConfig{
			Provider: "mock",
			Model:    "mock-knowledge-assistant",
		},
		RAG: RAGConfig{
			Enabled: true,
		},
		Tools: ToolsConfig{
			Enabled: true,
		},
		Server: ServerConfig{
			Host:                   "",
			Port:                   8080,
			ShutdownTimeoutSeconds: 10,
		},
	}
}

// Validate 校验配置是否满足启动要求
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return errors.New("server.port 必须在 1 到 65535 之间")
	}
	if c.LLM.Provider == "" {
		return errors.New("LLM provider 不能为空")
	}
	if c.LLM.Model == "" {
		return errors.New("LLM model 不能为空")
	}
	if c.Server.ShutdownTimeoutSeconds <= 0 {
		return errors.New("服务关闭超时时间必须大于 0")
	}
	return nil
}

// Addr 返回 HTTP Server 监听地址
func (c *ServerConfig) Addr() string {
	if c.Host == "" {
		return fmt.Sprintf(":%d", c.Port)
	}
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// applyEnv 使用环境变量覆盖配置文件值
func applyEnv(cfg *Config) {
	cfg.App.Env = getEnv("APP_ENV", cfg.App.Env)
	cfg.App.Mode = getEnv("APP_MODE", cfg.App.Mode)
	cfg.Server.Host = getEnv("SERVER_HOST", cfg.Server.Host)
	cfg.Log.Level = getEnv("LOG_LEVEL", cfg.Log.Level)
	cfg.Log.Filename = getEnv("LOG_FILENAME", cfg.Log.Filename)
	cfg.LLM.Provider = getEnv("LLM_PROVIDER", cfg.LLM.Provider)
	cfg.LLM.Model = getEnv("LLM_MODEL", cfg.LLM.Model)
	cfg.LLM.APIKey = getEnv("LLM_API_KEY", cfg.LLM.APIKey)

	if value := os.Getenv("RAG_ENABLED"); value != "" {
		cfg.RAG.Enabled = parseBool(value, cfg.RAG.Enabled)
	}
	if value := os.Getenv("TOOLS_ENABLED"); value != "" {
		cfg.Tools.Enabled = parseBool(value, cfg.Tools.Enabled)
	}
	if value := os.Getenv("SHUTDOWN_TIMEOUT_SECONDS"); value != "" {
		cfg.Server.ShutdownTimeoutSeconds = parseInt(value, cfg.Server.ShutdownTimeoutSeconds)
	}
	if value := os.Getenv("SERVER_PORT"); value != "" {
		cfg.Server.Port = parseInt(value, cfg.Server.Port)
	}
	if value := os.Getenv("LOG_MAX_SIZE"); value != "" {
		cfg.Log.MaxSize = parseInt(value, cfg.Log.MaxSize)
	}
	if value := os.Getenv("LOG_MAX_BACKUPS"); value != "" {
		cfg.Log.MaxBackups = parseInt(value, cfg.Log.MaxBackups)
	}
	if value := os.Getenv("LOG_MAX_AGE"); value != "" {
		cfg.Log.MaxAge = parseInt(value, cfg.Log.MaxAge)
	}
	if value := os.Getenv("LOG_COMPRESS"); value != "" {
		cfg.Log.Compress = parseBool(value, cfg.Log.Compress)
	}
}

// getEnv 读取环境变量并在为空时返回默认值
func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// parseBool 解析布尔环境变量并在失败时保留原值
func parseBool(value string, fallback bool) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// parseInt 解析整数环境变量并在失败时保留原值
func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
