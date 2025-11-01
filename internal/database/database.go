package database

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/pkg/dao/query"
	"gorm.io/driver/mysql"
	"gorm.io/gen/field"

	"gorm.io/gorm"
)

var _query *query.Query
var db *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	if config.GlobalConfig == nil {
		return fmt.Errorf("global config is nil")
	}

	dbs := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.GlobalConfig.DBConfig.Username,
		config.GlobalConfig.DBConfig.Password,
		config.GlobalConfig.DBConfig.Host,
		config.GlobalConfig.DBConfig.Port,
		config.GlobalConfig.DBConfig.DbName,
		config.GlobalConfig.DBConfig.Config)

	// Connection database
	var err error
	db, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dbs,   // DSN data source name
		DefaultStringSize:         256,   // string Default length of type fields
		DisableDatetimePrecision:  true,  // Disable datetime precision, not supported by databases before MySQL 5.6
		DontSupportRenameIndex:    true,  // When renaming the index, delete and create a new one. Databases before MySQL 5.7 and MariaDB do not support renaming indexes.
		DontSupportRenameColumn:   true,  // Use `change` to rename columns. Databases prior to MySQL 8 and MariaDB do not support renaming columns.
		SkipInitializeWithVersion: false, // Automatically configured based on the current MySQL version
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Set up connection pool
	maxIdleConns, _ := strconv.Atoi(config.GlobalConfig.DBConfig.MaxIdleConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	maxOpenConns, _ := strconv.Atoi(config.GlobalConfig.DBConfig.MaxOpenConns)
	sqlDB.SetMaxOpenConns(maxOpenConns)

	_query = query.Use(db)
	if _query.Available() == false {
		return errors.New("query is not available")
	}

	err = insertDefaultData()
	if err != nil {
		return err
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
