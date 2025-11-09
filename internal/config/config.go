package config

import (
	"os"
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
	MaxLever int64  `json:"max_lever"`
	MinSize  string `json:"min_size"` // 最小下单（合约：张，现货：交易货币）
	Contract string `json:"contract"` // 张数
}

// ExchangeSymbolList 各个交易所交易对数据（内存存储）
var ExchangeSymbolList map[string][]SymbolInfo

type Config struct {
	DBFile  string
	WorkDir string
}

var GlobalConfig *Config

func LoadConfig() error {
	// 获取工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	GlobalConfig = &Config{
		DBFile:  "./foxflow.db",
		WorkDir: workDir,
	}

	return nil
}
