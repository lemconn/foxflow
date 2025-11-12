package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/models"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/pkg/dao/query"
	"gorm.io/driver/sqlite"
	"gorm.io/gen/field"
	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

var _query *query.Query
var db *gorm.DB

// InitDB Initialize database connection
func InitDB() error {
	if config.GlobalConfig == nil {
		return fmt.Errorf("global config is nil")
	}

	dbDir, dbName := filepath.Split(config.GlobalConfig.DBFile)
	if dbName == "" {
		return fmt.Errorf("db file name is empty")
	}
	if !strings.HasSuffix(dbName, ".db") {
		return fmt.Errorf("db file name must end with .db")
	}

	if _, err := os.Stat(config.GlobalConfig.DBFile); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("failed to create db directory: %v", err)
		}
		file, err := os.Create(dbName)
		if err != nil {
			return fmt.Errorf("failed to create db file: %v", err)
		}
		file.Close()
	}
	if err := os.Chmod(config.GlobalConfig.DBFile, 0755); err != nil {
		return fmt.Errorf("failed to chmod db file: %v", err)
	}

	// Connection database
	var err error
	db, err = gorm.Open(sqlite.Open(config.GlobalConfig.DBFile), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	_query = query.Use(db)
	if _query.Available() == false {
		return errors.New("query is not available")
	}

	// Migrate table structure and data
	if err := migrateTables(); err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	return nil
}

// migrateTables 使用 GORM AutoMigrate 创建和迁移表结构
func migrateTables() error {

	// 这里需要根据系统版本进行迁移数据库
	if err := db.AutoMigrate(
		&models.FoxConfig{},
		&models.FoxAccount{},
		&models.FoxSymbol{},
		&models.FoxOrder{},
		&models.FoxExchange{},
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
	count, err := _query.FoxExchange.Count()
	if err != nil {
		return err
	}
	if count == 0 {
		insertDefaultExchanges()
	}

	return nil
}

// insertDefaultExchanges 插入默认交易所数据（存在则不做任何处理，不存在则添加）
func insertDefaultExchanges() {
	exchanges := []model.FoxExchange{
		{Name: "okx", APIURL: "https://www.okx.com", ProxyURL: "http://127.0.0.1:7890", IsActive: 0},
	}

	for _, exchange := range exchanges {
		_query.FoxExchange.Where(
			_query.FoxExchange.Name.Eq(exchange.Name),
		).Attrs(field.Attrs(exchange)).FirstOrCreate()
	}
}

func Adapter() *query.Query {
	return _query
}
