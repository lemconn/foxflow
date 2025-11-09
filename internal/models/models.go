package models

import (
	"gorm.io/gorm"
)

// FoxAccount 用户表
type FoxAccount struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null;default:''" json:"name"`
	Exchange   string `gorm:"not null;default:'okx';check:exchange IN ('okx', 'binance', 'gate')" json:"exchange"`
	AccessKey  string `gorm:"not null;default:''" json:"access_key"`
	SecretKey  string `gorm:"not null;default:''" json:"secret_key"`
	Passphrase string `gorm:"not null;default:''" json:"passphrase"`
	IsActive   int    `gorm:"not null;default:0" json:"is_active"`
	TradeType  string `gorm:"not null;default:'';check:trade_type IN ('mock', 'live')" json:"trade_type"`
	CreatedAt  string `gorm:"not null;default:''" json:"created_at"`
	UpdatedAt  string `gorm:"not null;default:''" json:"updated_at"`
}

func (FoxAccount) TableName() string {
	return "fox_accounts"
}

// FoxSymbol 标的表
type FoxSymbol struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"not null;default:''" json:"name"`
	AccountID  uint   `gorm:"not null;default:0" json:"account_id"`
	Exchange   string `gorm:"not null;default:'okx';check:exchange IN ('okx', 'binance', 'gate')" json:"exchange"`
	Leverage   int    `gorm:"not null;default:1" json:"leverage"`
	MarginType string `gorm:"not null;default:'isolated';check:margin_type IN ('isolated', 'cross')" json:"margin_type"`
	CreatedAt  string `gorm:"not null;default:''" json:"created_at"`
	UpdatedAt  string `gorm:"not null;default:''" json:"updated_at"`
}

func (FoxSymbol) TableName() string {
	return "fox_symbols"
}

// FoxOrder 策略订单表
type FoxOrder struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Exchange   string `gorm:"not null;default:'okx';check:exchange IN ('okx', 'binance', 'gate')" json:"exchange"` // 交易所
	AccountID  uint   `gorm:"not null;default:0" json:"account_id"`
	Symbol     string `gorm:"not null;default:''" json:"symbol"`
	Side       string `gorm:"not null;default:'';check:side IN ('buy', 'sell')" json:"side"`
	PosSide    string `gorm:"not null;default:'';check:pos_side IN ('long', 'short')" json:"pos_side"`
	MarginType string `gorm:"not null;default:'';check:margin_type IN ('isolated', 'cross')" json:"margin_type"`
	Price      string `gorm:"not null;default:0" json:"price"`      // 限价单时的金额
	Size       string `gorm:"not null;default:0" json:"size"`       // 用户填写购买数量/金额
	SizeType   string `gorm:"not null;default:''" json:"size_type"` // 用户购买的单位，USDT表示size是购买金额，而不是标的数量。空表示size是标的数量
	OrderType  string `gorm:"not null;default:'limit';check:order_type IN ('limit', 'market')" json:"order_type"`
	Strategy   string `gorm:"not null;default:''" json:"strategy"`
	OrderID    string `gorm:"not null;default:''" json:"order_id"`
	Type       string `gorm:"not null;default:'open';check:type IN ('open', 'close')" json:"type"`
	Status     string `gorm:"not null;default:'waiting';check:status IN ('waiting', 'opened', 'closed', 'failed', 'cancelled')" json:"status"`
	Msg        string `gorm:"not null;default:''" json:"msg"` // 订单描述（引擎处理结果）
	CreatedAt  string `gorm:"not null;default:''" json:"created_at"`
	UpdatedAt  string `gorm:"not null;default:''" json:"updated_at"`
}

func (FoxOrder) TableName() string {
	return "fox_orders"
}

// FoxExchange 交易所配置表
type FoxExchange struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Name      string `gorm:"not null;default:'';unique" json:"name"`
	APIURL    string `gorm:"not null;default:''" json:"api_url"`
	ProxyURL  string `gorm:"not null;default:''" json:"proxy_url"`
	IsActive  int    `gorm:"not null;default:0" json:"is_active"`
	CreatedAt string `gorm:"not null;default:''" json:"created_at"`
	UpdatedAt string `gorm:"not null;default:''" json:"updated_at"`
}

func (FoxExchange) TableName() string {
	return "fox_exchanges"
}

// 初始化数据库表
func InitDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&FoxAccount{},
		&FoxOrder{},
		&FoxExchange{},
		&FoxSymbol{},
	)
}
