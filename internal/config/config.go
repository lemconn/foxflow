package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const Version = "v0.1.0"

type Config struct {
	DBPath   string
	ProxyURL string
	LogLevel string
	LogFile  string
	WorkDir  string
}

var GlobalConfig *Config

func LoadConfig() error {
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 尝试加载.env文件
	envPath := filepath.Join(workDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		godotenv.Load(envPath)
	}

	GlobalConfig = &Config{
		DBPath:   getEnv("DB_PATH", ".foxflow.db"),
		ProxyURL: getEnv("PROXY_URL", "http://127.0.0.1:7890"),
		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogFile:  getEnv("LOG_FILE", "logs/foxflow.log"),
		WorkDir:  workDir,
	}

	// 确保日志目录存在
	logDir := filepath.Dir(GlobalConfig.LogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetDBPath() string {
	if GlobalConfig == nil {
		return ".foxflow.db"
	}
	return filepath.Join(GlobalConfig.WorkDir, GlobalConfig.DBPath)
}
