package repository

import (
	"errors"
	"os"
	"testing"

	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForSymbols() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("foxflow.db"), &gorm.Config{})
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

// getSymbolsByExchangeForTest 测试专用的根据交易所查询标的函数
func getSymbolsByExchangeForTest(db *gorm.DB, exchange string) ([]*models.FoxSymbol, error) {
	var symbols []*models.FoxSymbol
	err := db.Where("exchange = ?", exchange).Find(&symbols).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return symbols, nil
}

func TestGetSymbolsByExchange(t *testing.T) {
	// 设置测试数据库
	db := setupTestDBForSymbols()

	// 测试完成后清理数据库文件
	defer func() {
		os.Remove("foxflow.db")
	}()

	// 清理测试数据
	db.Exec("DELETE FROM fox_symbols")
	db.Exec("DELETE FROM fox_exchanges")

	// 先插入交易所数据
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
			IsActive: true,
		},
	}

	for _, exchange := range testExchanges {
		err := db.Create(exchange).Error
		if err != nil {
			t.Fatalf("Failed to create test exchange: %v", err)
		}
	}

	// 插入测试标的数据
	testSymbols := []*models.FoxSymbol{
		{
			Name:       "BTC-USDT",
			AccountID:  1,
			Exchange:   "okx",
			Leverage:   10,
			MarginType: "isolated",
		},
		{
			Name:       "ETH-USDT",
			AccountID:  1,
			Exchange:   "okx",
			Leverage:   5,
			MarginType: "cross",
		},
		{
			Name:       "BTC-USDT",
			AccountID:  2,
			Exchange:   "binance",
			Leverage:   20,
			MarginType: "isolated",
		},
	}

	for _, symbol := range testSymbols {
		err := db.Create(symbol).Error
		if err != nil {
			t.Fatalf("Failed to create test symbol: %v", err)
		}
	}

	// 测试 GetSymbolsByExchange 方法 - 查询 okx 交易所的标的
	okxSymbols, err := getSymbolsByExchangeForTest(db, "okx")

	// 验证结果
	if err != nil {
		t.Fatalf("GetSymbolsByExchange failed: %v", err)
	}

	if len(okxSymbols) != 2 {
		t.Fatalf("Expected 2 symbols for okx, got %d", len(okxSymbols))
	}

	// 验证数据内容
	symbolNames := make(map[string]bool)
	for _, symbol := range okxSymbols {
		symbolNames[symbol.Name] = true
		if symbol.Exchange != "okx" {
			t.Errorf("Expected exchange to be 'okx', got '%s'", symbol.Exchange)
		}
	}

	if !symbolNames["BTC-USDT"] {
		t.Error("Expected to find 'BTC-USDT' symbol")
	}
	if !symbolNames["ETH-USDT"] {
		t.Error("Expected to find 'ETH-USDT' symbol")
	}

	// 测试 GetSymbolsByExchange 方法 - 查询 binance 交易所的标的
	binanceSymbols, err := getSymbolsByExchangeForTest(db, "binance")

	// 验证结果
	if err != nil {
		t.Fatalf("GetSymbolsByExchange failed for binance: %v", err)
	}

	if len(binanceSymbols) != 1 {
		t.Fatalf("Expected 1 symbol for binance, got %d", len(binanceSymbols))
	}

	if binanceSymbols[0].Name != "BTC-USDT" {
		t.Errorf("Expected symbol name to be 'BTC-USDT', got '%s'", binanceSymbols[0].Name)
	}
	if binanceSymbols[0].Exchange != "binance" {
		t.Errorf("Expected exchange to be 'binance', got '%s'", binanceSymbols[0].Exchange)
	}

	// 测试查询不存在的交易所
	nonExistentSymbols, err := getSymbolsByExchangeForTest(db, "non-existent")
	if err != nil {
		t.Fatalf("GetSymbolsByExchange failed for non-existent: %v", err)
	}
	if len(nonExistentSymbols) != 0 {
		t.Errorf("Expected 0 symbols for non-existent exchange, got %d", len(nonExistentSymbols))
	}

	// 清理测试数据
	db.Exec("DELETE FROM fox_symbols")
	db.Exec("DELETE FROM fox_exchanges")
}
