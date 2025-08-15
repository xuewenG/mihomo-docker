package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type config struct {
	LogFile        string   `yaml:"log_file"`
	MihomoApiUrl   string   `yaml:"mihomo_api_url"`
	MihomoApiToken string   `yaml:"mihomo_api_token"`
	SubscribeUrl   string   `yaml:"subscribe_url"`
	BaseConfigPath string   `yaml:"base_config_path"`
	UpdateInterval string   `yaml:"update_interval"`
	UpdateFields   []string `yaml:"update_fields"`
}

var Config = &config{}

// 初始化配置
func InitConfig() error {
	data, err := os.ReadFile("mihomo-updater.yaml")
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := yaml.Unmarshal(data, Config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	if Config.SubscribeUrl == "" {
		return fmt.Errorf("subscribe_url 不能为空")
	}

	if Config.BaseConfigPath == "" {
		return fmt.Errorf("base_config_path 不能为空")
	}

	if Config.UpdateInterval == "" {
		Config.UpdateInterval = "1h"
	}

	return nil
}
