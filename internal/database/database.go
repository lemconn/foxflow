package database

import (
	"fmt"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	dbPath := config.GetDBPath()

	// 连接数据库
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 使用 GORM AutoMigrate 创建和迁移表结构
	if err := migrateTables(); err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	return nil
}

// migrateTables 使用 GORM AutoMigrate 创建和迁移表结构
func migrateTables() error {
	// 使用 AutoMigrate 创建和迁移所有表
	if err := DB.AutoMigrate(
		&models.FoxAccount{},
		&models.FoxSymbol{},
		&models.FoxContractMultiplier{},
		&models.FoxOrder{},
		&models.FoxExchange{},
		&models.FoxStrategy{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// 插入默认数据
	if err := insertDefaultData(); err != nil {
		return fmt.Errorf("failed to insert default data: %w", err)
	}

	return nil
}

// insertDefaultData 插入默认数据
func insertDefaultData() error {
	// 插入默认交易所数据
	var exchangeCount int64
	DB.Model(&models.FoxExchange{}).Count(&exchangeCount)
	if exchangeCount == 0 {
		insertDefaultExchanges()
	}

	return nil
}

// insertDefaultExchanges 插入默认交易所数据（存在则不做任何处理，不存在则添加）
func insertDefaultExchanges() {
	exchanges := []models.FoxExchange{
		{Name: "okx", APIURL: "https://www.okx.com", ProxyURL: "http://127.0.0.1:7890", IsActive: false},
	}

	for _, exchange := range exchanges {
		DB.FirstOrCreate(&exchange, models.FoxExchange{Name: exchange.Name})
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
