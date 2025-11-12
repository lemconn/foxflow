package models

import (
	"time"

	"gorm.io/gorm"
)

// FoxConfig Config table
type FoxConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AccountID uint      `gorm:"not null;default:0" json:"account_id"`
	ProxyUrl  string    `gorm:"not null;default:''" json:"proxy_url"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
}

func (FoxConfig) TableName() string {
	return "fox_configs"
}

// FoxTradeConfig Trade config table
type FoxTradeConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AccountID uint      `gorm:"not null;default:0" json:"account_id"`
	Margin    string    `gorm:"not null;default:'';check:margin IN ('isolated', 'cross')" json:"margin"`
	Leverage  int       `gorm:"not null;default:0" json:"leverage"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
}

func (FoxTradeConfig) TableName() string {
	return "fox_trade_configs"
}

// FoxAccount Account table
type FoxAccount struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"not null;default:''" json:"name"`
	Exchange   string    `gorm:"not null;default:''" json:"exchange"`
	AccessKey  string    `gorm:"not null;default:''" json:"access_key"`
	SecretKey  string    `gorm:"not null;default:''" json:"secret_key"`
	Passphrase string    `gorm:"not null;default:''" json:"passphrase"`
	IsActive   int       `gorm:"not null;default:0" json:"is_active"`
	TradeType  string    `gorm:"not null;default:'';check:trade_type IN ('mock', 'live')" json:"trade_type"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
}

func (FoxAccount) TableName() string {
	return "fox_accounts"
}

// FoxSymbol Symbol table
type FoxSymbol struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"not null;default:''" json:"name"`
	AccountID  uint      `gorm:"not null;default:0" json:"account_id"`
	Exchange   string    `gorm:"not null;default:''" json:"exchange"`
	Leverage   int       `gorm:"not null;default:0" json:"leverage"`
	MarginType string    `gorm:"not null;default:'isolated';check:margin_type IN ('isolated', 'cross')" json:"margin_type"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
}

func (FoxSymbol) TableName() string {
	return "fox_symbols"
}

// FoxOrder Order table
type FoxOrder struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Exchange   string    `gorm:"not null;default:'okx'" json:"exchange"`
	AccountID  uint      `gorm:"not null;default:0" json:"account_id"`
	Symbol     string    `gorm:"not null;default:''" json:"symbol"`
	Side       string    `gorm:"not null;default:'';check:side IN ('buy', 'sell')" json:"side"`
	PosSide    string    `gorm:"not null;default:'';check:pos_side IN ('long', 'short')" json:"pos_side"`
	MarginType string    `gorm:"not null;default:'';check:margin_type IN ('isolated', 'cross')" json:"margin_type"`
	Price      string    `gorm:"not null;default:0" json:"price"`
	Size       string    `gorm:"not null;default:0" json:"size"`
	SizeType   string    `gorm:"not null;default:''" json:"size_type"`
	OrderType  string    `gorm:"not null;default:'limit';check:order_type IN ('limit', 'market')" json:"order_type"`
	Strategy   string    `gorm:"not null;default:''" json:"strategy"`
	OrderID    string    `gorm:"not null;default:''" json:"order_id"`
	Type       string    `gorm:"not null;default:'open';check:type IN ('open', 'close')" json:"type"`
	Status     string    `gorm:"not null;default:'waiting';check:status IN ('waiting', 'opened', 'closed', 'failed', 'cancelled')" json:"status"`
	Msg        string    `gorm:"not null;default:''" json:"msg"` // 订单描述（引擎处理结果）
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
}

func (FoxOrder) TableName() string {
	return "fox_orders"
}

// FoxExchange 交易所配置表
type FoxExchange struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null;default:'';unique" json:"name"`
	APIURL    string    `gorm:"not null;default:''" json:"api_url"`
	ProxyURL  string    `gorm:"not null;default:''" json:"proxy_url"`
	IsActive  int       `gorm:"not null;default:0" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
}

func (FoxExchange) TableName() string {
	return "fox_exchanges"
}

// 初始化数据库表
func InitDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&FoxConfig{},
		&FoxTradeConfig{},
		&FoxAccount{},
		&FoxOrder{},
		&FoxExchange{},
		&FoxSymbol{},
	)
}
