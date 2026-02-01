package updater

import (
	"fmt"
	"io"
	"log"
	"mihomo-docker/pkg/config"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type MihomoConfig struct {
	Other map[string]any `yaml:",inline"`
}

type Updater struct {
	httpClient *http.Client
}

// 创建新的 Mihomo 配置更新器
func NewUpdater() *Updater {
	return &Updater{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// 从订阅地址获取 Mihomo 配置
func (u *Updater) fetchSubscribe() (*MihomoConfig, error) {
	resp, err := u.httpClient.Get(config.Config.SubscribeUrl)
	if err != nil {
		return nil, fmt.Errorf("请求订阅地址失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("订阅地址返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var mihomoConfig MihomoConfig
	if err := yaml.Unmarshal(body, &mihomoConfig); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败: %w", err)
	}

	return &mihomoConfig, nil
}

// 读取源 Mihomo 配置文件
func (u *Updater) readSourceConfig() (*MihomoConfig, error) {
	data, err := os.ReadFile(config.Config.BaseConfigPath)
	if err != nil {
		return nil, fmt.Errorf("读取源 Mihomo 配置文件失败: %w", err)
	}

	var mihomoConfig MihomoConfig
	if err := yaml.Unmarshal(data, &mihomoConfig); err != nil {
		return nil, fmt.Errorf("解析源 Mihomo 配置文件失败: %w", err)
	}

	return &mihomoConfig, nil
}

// 合并 Mihomo 配置
func (u *Updater) mergeMihomoConfig(sourceConfig, newConfig *MihomoConfig) *MihomoConfig {
	mergedConfig := &MihomoConfig{
		Other: make(map[string]any),
	}

	for k, v := range sourceConfig.Other {
		mergedConfig.Other[k] = v
	}

	for _, field := range config.Config.UpdateFields {
		if value, exists := newConfig.Other[field]; exists {
			mergedConfig.Other[field] = value
		}
	}

	return mergedConfig
}

// 保存 Mihomo 配置到文件
func (u *Updater) saveMihomoConfig(mihomoConfig *MihomoConfig) error {
	saveConfigPath := "/root/.config/mihomo/config.yaml"
	dir := filepath.Dir(saveConfigPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	file, err := os.OpenFile(saveConfigPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("创建 Mihomo 配置文件失败: %w", err)
	}
	defer file.Close()

	// 过滤出其他 Mihomo 配置
	baseConfig := make(map[string]any)
	for k, v := range mihomoConfig.Other {
		if !slices.Contains(config.Config.UpdateFields, k) {
			baseConfig[k] = v
		}
	}

	// 写入其他 Mihomo 配置
	if len(baseConfig) > 0 {
		otherData, err := yaml.Marshal(baseConfig)
		if err != nil {
			return fmt.Errorf("序列化其他 Mihomo 配置失败: %w", err)
		}
		_, err = file.Write(otherData)
		if err != nil {
			return fmt.Errorf("写入其他 Mihomo 配置失败: %w", err)
		}
	}

	// 写入需要更新的 Mihomo 配置字段
	for _, field := range config.Config.UpdateFields {
		if value, exists := mihomoConfig.Other[field]; exists {
			fieldData, err := yaml.Marshal(map[string]any{field: value})
			if err != nil {
				return fmt.Errorf("序列化 Mihomo 配置字段 %s 失败: %w", field, err)
			}
			_, err = file.Write(fieldData)
			if err != nil {
				return fmt.Errorf("写入 Mihomo 配置字段 %s 失败: %w", field, err)
			}
		}
	}

	return nil
}

// 通知 Mihomo 重新加载配置
func (u *Updater) notifyMihomoReload() error {
	if config.Config.MihomoApiUrl == "" {
		log.Println("未配置 Mihomo API Url，跳过通知 Mihomo 重新加载")
		return nil
	}

	url := config.Config.MihomoApiUrl + "/configs"
	body := `{"path":"/root/.config/mihomo/config.yaml"}`
	req, err := http.NewRequest("PUT", url, strings.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建通知 Mihomo 重新加载配置请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if config.Config.MihomoApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.Config.MihomoApiToken)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("通知 Mihomo 重新加载配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("重新加载 Mihomo 配置失败，错误状态码: %d，错误信息: %s", resp.StatusCode, string(body))
	}

	return nil
}

// 执行 Mihomo 配置更新
func (u *Updater) updateConfig() error {
	log.Println("开始更新 Mihomo 配置...")

	// 获取新的 Mihomo 配置
	newConfig, err := u.fetchSubscribe()
	if err != nil {
		return fmt.Errorf("获取新 Mihomo 配置失败: %w", err)
	}

	// 读取源 Mihomo 配置
	sourceConfig, err := u.readSourceConfig()
	if err != nil {
		return fmt.Errorf("读取源 Mihomo 配置失败: %w", err)
	}

	// 合并 Mihomo 配置
	mergedConfig := u.mergeMihomoConfig(sourceConfig, newConfig)

	// 保存 Mihomo 配置
	if err := u.saveMihomoConfig(mergedConfig); err != nil {
		return fmt.Errorf("保存 Mihomo 配置失败: %w", err)
	}

	// 重新加载 Mihomo 配置
	if err := u.notifyMihomoReload(); err != nil {
		log.Printf("重新加载 Mihomo 配置失败: %v", err)
	}

	log.Printf("Mihomo 配置更新成功")
	return nil
}

// 解析时间间隔
func parseDuration(duration string) (time.Duration, error) {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0, fmt.Errorf("解析时间间隔失败: %w", err)
	}
	return d, nil
}

// 启动定时任务
func (u *Updater) StartScheduler() error {
	interval, err := parseDuration(config.Config.UpdateInterval)
	if err != nil {
		return fmt.Errorf("解析更新 Mihomo 配置间隔失败: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Mihomo 配置更新器已启动，每 %s 更新一次...", config.Config.UpdateInterval)

	if err := u.updateConfig(); err != nil {
		fmt.Errorf("初始更新 Mihomo 配置失败: %v", err)
	}

	for range ticker.C {
		if err := u.updateConfig(); err != nil {
			log.Printf("定时更新 Mihomo 配置失败: %v", err)
		}
	}

	return nil
}
