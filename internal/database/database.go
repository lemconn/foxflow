package database

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"foxflow/internal/config"
	"foxflow/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	dbPath := config.GetDBPath()

	// 检查数据库文件是否存在
	exists := false
	if _, err := os.Stat(dbPath); err == nil {
		exists = true
	}

	// 连接数据库
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 如果数据库文件不存在，执行初始化SQL
	if !exists {
		if err := initDatabaseFromSQL(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
	} else {
		// 如果数据库已存在，只进行轻量级的迁移检查
		if err := checkAndMigrateTables(); err != nil {
			return fmt.Errorf("failed to check database schema: %w", err)
		}
	}

	return nil
}

// initDatabaseFromSQL 从SQL文件初始化数据库
func initDatabaseFromSQL() error {
	sqlPath := filepath.Join(config.GlobalConfig.WorkDir, "scripts", "foxflow.sql")

	// 读取SQL文件
	sqlContent, err := ioutil.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// 执行SQL语句
	if err := DB.Exec(string(sqlContent)).Error; err != nil {
		return fmt.Errorf("failed to execute SQL: %w", err)
	}

	return nil
}

// checkAndMigrateTables 检查并迁移表结构
func checkAndMigrateTables() error {
	// 检查必要的表是否存在，如果不存在则创建
	var count int64

	// 检查 fox_users 表
	DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='fox_users'").Scan(&count)
	if count == 0 {
		if err := DB.AutoMigrate(&models.FoxUser{}); err != nil {
			return fmt.Errorf("failed to create fox_users table: %w", err)
		}
	}

	// 检查 fox_symbols 表
	DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='fox_symbols'").Scan(&count)
	if count == 0 {
		if err := DB.AutoMigrate(&models.FoxSymbol{}); err != nil {
			return fmt.Errorf("failed to create fox_symbols table: %w", err)
		}
	}

	// 检查 fox_ss 表
	DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='fox_ss'").Scan(&count)
	if count == 0 {
		if err := DB.AutoMigrate(&models.FoxSS{}); err != nil {
			return fmt.Errorf("failed to create fox_ss table: %w", err)
		}
	}

	// 检查 fox_exchanges 表
	DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='fox_exchanges'").Scan(&count)
	if count == 0 {
		if err := DB.AutoMigrate(&models.FoxExchange{}); err != nil {
			return fmt.Errorf("failed to create fox_exchanges table: %w", err)
		}
		// 插入默认数据
		insertDefaultExchanges()
	}

	// 检查 fox_strategies 表
	DB.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='fox_strategies'").Scan(&count)
	if count == 0 {
		if err := DB.AutoMigrate(&models.FoxStrategy{}); err != nil {
			return fmt.Errorf("failed to create fox_strategies table: %w", err)
		}
		// 插入默认数据
		insertDefaultStrategies()
	}

	return nil
}

// insertDefaultExchanges 插入默认交易所数据
func insertDefaultExchanges() {
	exchanges := []models.FoxExchange{
		{Name: "okx", APIURL: "https://www.okx.com", ProxyURL: "http://127.0.0.1:7890", Status: "active"},
		{Name: "binance", APIURL: "https://api.binance.com", ProxyURL: "http://127.0.0.1:7890", Status: "active"},
	}

	for _, exchange := range exchanges {
		DB.FirstOrCreate(&exchange, models.FoxExchange{Name: exchange.Name})
	}
}

// insertDefaultStrategies 插入默认策略数据
func insertDefaultStrategies() {
	strategies := []models.FoxStrategy{
		{Name: "volume", Description: "成交量策略", Parameters: `{"threshold": 100}`, Status: "active"},
		{Name: "macd", Description: "MACD策略", Parameters: `{"threshold": 50}`, Status: "active"},
		{Name: "rsi", Description: "RSI策略", Parameters: `{"threshold": 10}`, Status: "active"},
	}

	for _, strategy := range strategies {
		DB.FirstOrCreate(&strategy, models.FoxStrategy{Name: strategy.Name})
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
