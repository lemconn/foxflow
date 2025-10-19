package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/lemconn/foxflow/internal/exchange"
)

// 使用 exchange 包中的 Ticker 类型
type MarketData = exchange.Ticker

// MarketProvider 行情数据模块
type MarketProvider struct {
	*BaseProvider
	exchangeMgr *exchange.Manager
}

// NewMarketProvider 创建行情数据模块
func NewMarketProvider() *MarketProvider {
	module := &MarketProvider{
		BaseProvider: NewBaseProvider("market"),
		exchangeMgr:  exchange.GetManager(),
	}

	return module
}

// GetData 获取数据
// MarketProvider 通过 exchange 实时获取行情数据
// params 参数（可选）：
// - 目前暂未使用，保留用于未来扩展
func (p *MarketProvider) GetData(ctx context.Context, dataSource, field string, params ...interface{}) (interface{}, error) {
	// 解析字段 - 支持多级字段如 "BTC.price"
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("market field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

	// 获取交易所实例
	exchangeInstance, err := p.exchangeMgr.GetExchange(dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange %s: %w", dataSource, err)
	}

	// 使用 GetSwapSymbolByName 转换 symbol 参数
	exchangeSymbol := exchangeInstance.GetSwapSymbolByName(ctx, symbol)

	// 通过 exchange 实时获取行情数据
	ticker, err := exchangeInstance.GetTicker(ctx, exchangeSymbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker data for %s %s: %w", dataSource, exchangeSymbol, err)
	}

	// 验证符号是否匹配
	if ticker.Symbol != exchangeSymbol {
		return nil, fmt.Errorf("symbol mismatch: expected %s, got %s", exchangeSymbol, ticker.Symbol)
	}

	// 提取指定字段
	switch fieldName {
	case "price":
		return ticker.Price, nil
	case "volume":
		return ticker.Volume, nil
	case "high":
		return ticker.High, nil
	case "low":
		return ticker.Low, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", fieldName)
	}
}
