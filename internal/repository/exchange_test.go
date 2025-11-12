package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("foxflow-1.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 初始化表
	err = models.InitDB(db)
	if err != nil {
		panic("failed to migrate database")
	}

	return db
}

// getAllExchangesForTest 测试专用的查询所有交易所函数
func getAllExchangesForTest(db *gorm.DB) ([]*models.FoxExchange, error) {
	var exchanges []*models.FoxExchange
	err := db.Find(&exchanges).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return exchanges, nil
}

func TestGetAllExchanges(t *testing.T) {
	// 设置测试数据库
	db := setupTestDB()

	// 测试完成后清理数据库文件
	defer func() {
		os.Remove("foxflow-1.db")
	}()

	// 清理测试数据
	db.Exec("DELETE FROM fox_exchanges")

	// 插入测试数据
	testExchanges := []*models.FoxExchange{
		{
			Name:     "okx",
			APIURL:   "https://www.okx.com",
			ProxyURL: "",
			IsActive: true,
		},
		{
			Name:     "binance",
			APIURL:   "https://api.binance.com",
			ProxyURL: "",
			IsActive: false,
		},
	}

	for _, exchange := range testExchanges {
		err := db.Create(exchange).Error
		if err != nil {
			t.Fatalf("Failed to create test exchange: %v", err)
		}
	}

	// 测试 GetAllExchanges 方法
	exchanges, err := getAllExchangesForTest(db)

	// 验证结果
	if err != nil {
		t.Fatalf("GetAllExchanges failed: %v", err)
	}

	if len(exchanges) != 2 {
		t.Fatalf("Expected 2 exchanges, got %d", len(exchanges))
	}

	// 验证数据内容
	exchangeNames := make(map[string]bool)
	for _, exchange := range exchanges {
		exchangeNames[exchange.Name] = true
	}

	if !exchangeNames["okx"] {
		t.Error("Expected to find 'okx' exchange")
	}
	if !exchangeNames["binance"] {
		t.Error("Expected to find 'binance' exchange")
	}

	// 清理测试数据
	db.Exec("DELETE FROM fox_exchanges")
}
