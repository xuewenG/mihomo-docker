package main

import (
	"log"
	"mihomo-docker/pkg/config"
	"mihomo-docker/pkg/logger"
	"mihomo-docker/pkg/mihomo"
	"mihomo-docker/pkg/updater"
	"strings"
)

var version string
var commitId string

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := logger.SetLogFilePath(config.Config.LogFile); err != nil {
		log.Fatalf("设置日志输出失败: %v", err)
	}

	log.Printf("version: %s", version)
	log.Printf("commitId: %s", commitId)
	log.Printf("Base Config Path: %s", config.Config.BaseConfigPath)
	log.Printf("Update Interval: %s", config.Config.UpdateInterval)
	log.Printf("Update Fields: %s", strings.Join(config.Config.UpdateFields, ", "))

	if config.Config.MihomoApiUrl != "" {
		log.Printf("Mihomo API Url: 已配置")
	}

	if config.Config.MihomoApiToken != "" {
		log.Printf("Mihomo API Token: 已配置")
	}

	if err := mihomo.StartMihomo(); err != nil {
		log.Fatalf("启动 Mihomo 失败: %v", err)
	}

	updater := updater.NewUpdater()
	if err := updater.StartScheduler(); err != nil {
		log.Fatalf("启动定时任务失败: %v", err)
	}
}
