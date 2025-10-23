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
		&models.FoxSS{},
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

	// 插入默认策略数据
	var strategyCount int64
	DB.Model(&models.FoxStrategy{}).Count(&strategyCount)
	if strategyCount == 0 {
		insertDefaultStrategies()
	}

	// 插入测试用户数据
	var userCount int64
	DB.Model(&models.FoxAccount{}).Count(&userCount)
	if userCount == 0 {
		insertTestUsers()
	}

	//// 插入测试标的数据
	//var symbolCount int64
	//DB.Model(&models.FoxSymbol{}).Count(&symbolCount)
	//if symbolCount == 0 {
	//	insertTestSymbols()
	//}

	// 插入测试策略订单数据
	var orderCount int64
	DB.Model(&models.FoxSS{}).Count(&orderCount)
	if orderCount == 0 {
		insertTestOrders()
	}

	return nil
}

// insertDefaultExchanges 插入默认交易所数据（存在则不做任何处理，不存在则添加）
func insertDefaultExchanges() {
	exchanges := []models.FoxExchange{
		{Name: "okx", APIURL: "https://www.okx.com", ProxyURL: "http://127.0.0.1:7890", IsActive: false},
		{Name: "binance", APIURL: "https://api.binance.com", ProxyURL: "http://127.0.0.1:7890", IsActive: false},
		{Name: "gate", APIURL: "https://api.gateio.ws", ProxyURL: "http://127.0.0.1:7890", IsActive: false},
	}

	for _, exchange := range exchanges {
		DB.FirstOrCreate(&exchange, models.FoxExchange{Name: exchange.Name})
	}
}

// insertDefaultStrategies 插入默认策略数据（存在则不做任何处理，不存在则添加）
func insertDefaultStrategies() {
	strategies := []models.FoxStrategy{
		{Name: "volume", Description: "成交量策略", Parameters: `{"threshold": 100}`, IsActive: true},
		{Name: "macd", Description: "MACD策略", Parameters: `{"threshold": 50}`, IsActive: true},
		{Name: "rsi", Description: "RSI策略", Parameters: `{"threshold": 10}`, IsActive: true},
	}

	for _, strategy := range strategies {
		DB.FirstOrCreate(&strategy, models.FoxStrategy{Name: strategy.Name})
	}
}

// insertTestUsers 插入测试用户数据
func insertTestUsers() {
	users := []models.FoxAccount{
		{Name: "test_user_1", Exchange: "binance", AccessKey: "test_binance_access_key_1", SecretKey: "test_binance_secret_key_1", IsActive: true, TradeType: "mock"},
		{Name: "test_user_2", Exchange: "okx", AccessKey: "test_okx_access_key_2", SecretKey: "test_okx_secret_key_2", IsActive: true, TradeType: "real"},
		{Name: "test_user_3", Exchange: "gate", AccessKey: "test_gate_access_key_3", SecretKey: "test_gate_secret_key_3", IsActive: false, TradeType: "mock"},
		{Name: "demo_trader", Exchange: "binance", AccessKey: "demo_binance_key", SecretKey: "demo_binance_secret", IsActive: true, TradeType: "mock"},
	}

	for _, user := range users {
		DB.FirstOrCreate(&user, models.FoxAccount{Name: user.Name, AccessKey: user.AccessKey, SecretKey: user.SecretKey})
	}
}

//
//// insertTestSymbols 插入测试标的数据
//func insertTestSymbols() {
//	symbols := []models.FoxSymbol{
//		{Name: "BTCUSDT", UserID: 1, Exchange: "binance", Leverage: 10, MarginType: "isolated"},
//		{Name: "ETHUSDT", UserID: 1, Exchange: "binance", Leverage: 5, MarginType: "cross"},
//		{Name: "BTC-USDT-SWAP", UserID: 2, Exchange: "okx", Leverage: 20, MarginType: "isolated"},
//		{Name: "ETH-USDT-SWAP", UserID: 2, Exchange: "okx", Leverage: 15, MarginType: "cross"},
//		{Name: "BTC_USDT", UserID: 3, Exchange: "gate", Leverage: 8, MarginType: "isolated"},
//		{Name: "ADAUSDT", UserID: 4, Exchange: "binance", Leverage: 3, MarginType: "isolated"},
//	}
//
//	for _, symbol := range symbols {
//		DB.FirstOrCreate(&symbol, models.FoxSymbol{UserID: symbol.UserID, Exchange: symbol.Exchange, Name: symbol.Name})
//	}
//}

// insertTestOrders 插入测试策略订单数据
func insertTestOrders() {
	orders := []models.FoxSS{
		// 用户1的订单
		{UserID: 1, Symbol: "BTCUSDT", Side: "buy", PosSide: "long", Px: 45000.50, Sz: 0.01, OrderType: "limit", Strategy: "macd", OrderID: "binance_order_001", Type: "open", Status: "waiting"},
		{UserID: 1, Symbol: "BTCUSDT", Side: "sell", PosSide: "long", Px: 46000.00, Sz: 0.01, OrderType: "limit", Strategy: "macd", OrderID: "binance_order_002", Type: "close", Status: "pending"},
		{UserID: 1, Symbol: "ETHUSDT", Side: "buy", PosSide: "long", Px: 3200.25, Sz: 0.1, OrderType: "market", Strategy: "volume", OrderID: "binance_order_003", Type: "open", Status: "filled"},

		// 用户2的订单
		{UserID: 2, Symbol: "BTC-USDT-SWAP", Side: "buy", PosSide: "long", Px: 45100.00, Sz: 0.02, OrderType: "limit", Strategy: "rsi", OrderID: "okx_order_001", Type: "open", Status: "waiting"},
		{UserID: 2, Symbol: "ETH-USDT-SWAP", Side: "sell", PosSide: "short", Px: 3150.00, Sz: 0.05, OrderType: "limit", Strategy: "volume", OrderID: "okx_order_002", Type: "open", Status: "pending"},
		{UserID: 2, Symbol: "BTC-USDT-SWAP", Side: "sell", PosSide: "long", Px: 46000.00, Sz: 0.02, OrderType: "limit", Strategy: "rsi", OrderID: "okx_order_003", Type: "close", Status: "cancelled"},

		// 用户4的订单
		{UserID: 4, Symbol: "ADAUSDT", Side: "buy", PosSide: "long", Px: 0.45, Sz: 1000, OrderType: "limit", Strategy: "macd", OrderID: "binance_order_004", Type: "open", Status: "waiting"},
		{UserID: 4, Symbol: "ADAUSDT", Side: "sell", PosSide: "long", Px: 0.48, Sz: 1000, OrderType: "limit", Strategy: "macd", OrderID: "binance_order_005", Type: "close", Status: "waiting"},
		{UserID: 4, Symbol: "BTCUSDT", Side: "buy", PosSide: "long", Px: 44800.00, Sz: 0.005, OrderType: "market", Strategy: "volume", OrderID: "binance_order_006", Type: "open", Status: "filled"},
	}

	for _, order := range orders {
		DB.FirstOrCreate(&order, models.FoxSS{OrderID: order.OrderID})
	}
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
