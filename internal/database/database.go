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
		Logger: logger.Default.LogMode(logger.Silent),
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
	// 使用 AutoMigrate 对所有模型进行无损迁移（新增字段/索引等），不影响既有数据
	if err := DB.AutoMigrate(
		&models.FoxUser{},
		&models.FoxSymbol{},
		&models.FoxSS{},
		&models.FoxExchange{},
		&models.FoxStrategy{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// 表为空时插入默认数据
	var count int64
	DB.Model(&models.FoxExchange{}).Count(&count)
	if count == 0 {
		insertDefaultExchanges()
	}
	DB.Model(&models.FoxStrategy{}).Count(&count)
	if count == 0 {
		insertDefaultStrategies()
	}

	// 执行需要的手工数据迁移（若有）。该过程应幂等
	if err := runDataMigrations(); err != nil {
		return err
	}

	return nil
}

// runDataMigrations 执行需要的手工数据迁移（如字段拆分、数据回填等），保证可重复执行且不影响已有数据
func runDataMigrations() error {
	return nil
}

// insertDefaultExchanges 插入默认交易所数据（存在则不做任何处理，不存在则添加）
func insertDefaultExchanges() {
	exchanges := []models.FoxExchange{
		{Name: "okx", APIURL: "https://www.okx.com", ProxyURL: "http://127.0.0.1:7890", Status: "inactive", IsActive: false},
		{Name: "binance", APIURL: "https://api.binance.com", ProxyURL: "http://127.0.0.1:7890", Status: "inactive", IsActive: false},
		{Name: "gate", APIURL: "https://api.gateio.ws", ProxyURL: "http://127.0.0.1:7890", Status: "inactive", IsActive: false},
	}

	for _, exchange := range exchanges {
		DB.FirstOrCreate(&exchange, models.FoxExchange{Name: exchange.Name})
	}
}

// insertDefaultStrategies 插入默认策略数据（存在则不做任何处理，不存在则添加）
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
