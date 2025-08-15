package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// 设置日志输出路径
func SetLogFilePath(logFile string) error {
	if logFile == "" {
		return nil
	}

	dir := filepath.Dir(logFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	log.SetOutput(file)
	return nil
}
