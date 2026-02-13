package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Site     SiteConfig     `yaml:"site"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Path string `yaml:"path"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Password string `yaml:"password"`
}

// SiteConfig 网站配置
type SiteConfig struct {
	Title string `yaml:"title"`
}

// Load 从YAML文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
