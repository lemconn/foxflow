package exchange

import (
	"context"
	"fmt"
	"sync"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
)

// Manager 交易所管理器
type Manager struct {
	exchanges map[string]Exchange
	mu        sync.RWMutex
}

var (
	manager *Manager
	once    sync.Once
)

// GetManager 获取交易所管理器单例
func GetManager() *Manager {
	once.Do(func() {
		manager = &Manager{
			exchanges: make(map[string]Exchange),
		}
		manager.initExchanges()
	})
	return manager
}

// initExchanges 初始化交易所
func (m *Manager) initExchanges() {
	db := database.GetDB()
	if db == nil {
		// 如果数据库未初始化，使用默认配置
		m.initDefaultExchanges()
		return
	}

	var exchanges []models.FoxExchange

	// 从数据库加载所有交易所配置
	if err := db.Find(&exchanges).Error; err != nil {
		// 如果数据库中没有配置，使用默认配置
		m.initDefaultExchanges()
		return
	}

	// 根据配置创建交易所实例
	for _, exchange := range exchanges {
		switch exchange.Name {
		case "okx":
			m.exchanges[exchange.Name] = NewOKXExchange(exchange.APIURL, exchange.ProxyURL)
		}
	}
}

// initDefaultExchanges 初始化默认交易所
func (m *Manager) initDefaultExchanges() {
	m.exchanges["okx"] = NewOKXExchange("https://www.okx.com", "")
	//m.exchanges["binance"] = NewBinanceExchange("https://api.binance.com", "")
	//m.exchanges["gate"] = NewGateExchange("https://api.gateio.ws", "")
}

// GetExchange 获取交易所实例
func (m *Manager) GetExchange(name string) (Exchange, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	exchange, exists := m.exchanges[name]
	if !exists {
		return nil, fmt.Errorf("exchange %s not found", name)
	}

	return exchange, nil
}

// GetAvailableExchanges 获取可用的交易所列表
func (m *Manager) GetAvailableExchanges() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var exchanges []string
	for name := range m.exchanges {
		exchanges = append(exchanges, name)
	}

	return exchanges
}

// ConnectAccount 连接用户到指定交易所
func (m *Manager) ConnectAccount(ctx context.Context, exchangeName string, account *models.FoxAccount) error {
	exchange, err := m.GetExchange(exchangeName)
	if err != nil {
		return err
	}

	return exchange.Connect(ctx, account)
}

// DisconnectAccount 断开用户连接
func (m *Manager) DisconnectAccount(exchangeName string) error {
	exchange, err := m.GetExchange(exchangeName)
	if err != nil {
		return err
	}

	return exchange.Disconnect()
}
