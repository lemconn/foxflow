package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const Version = "v0.1.0"

const (
	DefaultExchange = "okx"
)

const DateFormat = "2006-01-02 15:04:05"

type SymbolInfo struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Base     string `json:"base"`
	Quote    string `json:"quote"`
	MaxLever string `json:"max_lever"`
	MinSize  string `json:"min_size"` // 最小下单（合约：张，现货：交易货币）
}

// ExchangeSymbolList 各个交易所交易对数据（内存存储）
var ExchangeSymbolList map[string][]SymbolInfo

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

func SetConfigExchangeSymbolList(config map[string][]SymbolInfo) {
	ExchangeSymbolList = config
}
