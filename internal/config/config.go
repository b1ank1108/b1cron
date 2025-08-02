package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 全局配置结构
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	JWT         JWTConfig         `yaml:"jwt"`
	DefaultUser DefaultUserConfig `yaml:"default_user"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type  string `yaml:"type"`
	Path  string `yaml:"path"`
	Debug bool   `yaml:"debug"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret               string `yaml:"secret"`
	ExpireHours          int    `yaml:"expire_hours"`
	RefreshExpireHours   int    `yaml:"refresh_expire_hours"`
	CookieName           string `yaml:"cookie_name"`
	HTTPOnly             bool   `yaml:"http_only"`
	Secure               bool   `yaml:"secure"`
}

// DefaultUserConfig 默认用户配置
type DefaultUserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var GlobalConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*Config, error) {
	// 如果配置文件路径为空，使用默认路径
	if configPath == "" {
		configPath = "config.yaml"
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析YAML配置
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// 处理相对路径
	if !filepath.IsAbs(config.Database.Path) {
		config.Database.Path = filepath.Join(filepath.Dir(configPath), config.Database.Path)
	}

	// 设置全局配置
	GlobalConfig = &config

	return &config, nil
}

// validateConfig 验证配置有效性
func validateConfig(config *Config) error {
	// 验证服务器配置
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// 验证JWT配置
	if config.JWT.Secret == "" || config.JWT.Secret == "b1cron-jwt-secret-key-change-in-production" {
		fmt.Println("警告: 正在使用默认的JWT密钥，生产环境中请务必修改!")
	}
	if config.JWT.ExpireHours <= 0 {
		return fmt.Errorf("invalid JWT expire hours: %d", config.JWT.ExpireHours)
	}

	// 验证数据库配置
	if config.Database.Type != "sqlite" {
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}
	if config.Database.Path == "" {
		return fmt.Errorf("database path cannot be empty")
	}

	return nil
}

// GetAddr 获取服务器监听地址
func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// IsDevelopment 是否为开发模式
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "debug"
}

// IsProduction 是否为生产模式
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}