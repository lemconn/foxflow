package utils

import (
	"log"
	"os"
	"path/filepath"
)

// SetupLogger 设置日志
func SetupLogger(logFile string) error {
	// 确保日志目录存在
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// 设置日志输出
	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}
