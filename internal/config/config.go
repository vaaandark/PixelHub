package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   ServerConfig   `toml:"server"`
	MCP      MCPConfig      `toml:"mcp"`
	Database DatabaseConfig `toml:"database"`
	Storage  StorageConfig  `toml:"storage"`
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

type MCPConfig struct {
	Host   string `toml:"host"`
	Port   string `toml:"port"`
	APIKey string `toml:"api_key"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type StorageConfig struct {
	Provider   string           `toml:"provider"`
	TencentCOS TencentCOSConfig `toml:"tencent_cos"`
}

type TencentCOSConfig struct {
	SecretID  string `toml:"secret_id"`
	SecretKey string `toml:"secret_key"`
	BucketURL string `toml:"bucket_url"`
	CDNURL    string `toml:"cdn_url"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
