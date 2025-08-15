package mihomo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// 启动 Mihomo
func StartMihomo() error {
	log.Println("正在启动 Mihomo...")
	cmd := exec.Command("/mihomo")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 Mihomo 失败: %v", err)
	}

	log.Printf("Mihomo 已启动，进程 ID: %d", cmd.Process.Pid)
	return nil
}
