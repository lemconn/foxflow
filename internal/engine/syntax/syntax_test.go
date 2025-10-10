package syntax

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lemconn/foxflow/internal/engine/provider"
	"github.com/lemconn/foxflow/internal/engine/registry"
)

// MockDataProvider 模拟数据提供者
type MockDataProvider struct {
	klineData  map[string]map[string][]float64
	marketData map[string]map[string]float64
	newsData   map[string]map[string]interface{}
}

// MockKlineDataSource 模拟K线数据源
type MockKlineDataSource struct {
	provider *MockDataProvider
}

func (m *MockKlineDataSource) GetName() string {
	return "kline"
}

func (m *MockKlineDataSource) GetData(ctx context.Context, dataSource, field string, params ...provider.DataParam) (interface{}, error) {
	// 如果请求历史数据
	if len(params) > 0 {
		for _, param := range params {
			if param.Name == "period" {
				if period, ok := param.Value.(int); ok && period > 0 {
					return m.provider.GetHistoricalData(ctx, "kline", dataSource, field, period)
				}
			}
		}
	}
	// 返回单个数据值
	return m.provider.GetKlineField(ctx, dataSource, field)
}

func (m *MockKlineDataSource) GetFunctionParamMapping() map[string]provider.FunctionParamInfo {
	return map[string]provider.FunctionParamInfo{
		"avg": {
			FunctionName: "avg",
			Params: []provider.FunctionParam{
				{
					ParamIndex: 1,
					ParamName:  "period",
					ParamType:  provider.ParamTypeInt,
					Required:   true,
				},
			},
		},
		"max": {
			FunctionName: "max",
			Params: []provider.FunctionParam{
				{
					ParamIndex: 1,
					ParamName:  "period",
					ParamType:  provider.ParamTypeInt,
					Required:   true,
				},
			},
		},
		"min": {
			FunctionName: "min",
			Params: []provider.FunctionParam{
				{
					ParamIndex: 1,
					ParamName:  "period",
					ParamType:  provider.ParamTypeInt,
					Required:   true,
				},
			},
		},
		"sum": {
			FunctionName: "sum",
			Params: []provider.FunctionParam{
				{
					ParamIndex: 1,
					ParamName:  "period",
					ParamType:  provider.ParamTypeInt,
					Required:   true,
				},
			},
		},
	}
}

// MockMarketDataSource 模拟行情数据源
type MockMarketDataSource struct {
	provider *MockDataProvider
}

func (m *MockMarketDataSource) GetName() string {
	return "market"
}

func (m *MockMarketDataSource) GetData(ctx context.Context, dataSource, field string, params ...provider.DataParam) (interface{}, error) {
	return m.provider.GetMarketField(ctx, dataSource, field)
}

func (m *MockMarketDataSource) GetFunctionParamMapping() map[string]provider.FunctionParamInfo {
	return map[string]provider.FunctionParamInfo{}
}

// MockNewsDataSource 模拟新闻数据源
type MockNewsDataSource struct {
	provider *MockDataProvider
}

func (m *MockNewsDataSource) GetName() string {
	return "news"
}

func (m *MockNewsDataSource) GetData(ctx context.Context, dataSource, field string, params ...provider.DataParam) (interface{}, error) {
	return m.provider.GetNewsField(ctx, dataSource, field)
}

func (m *MockNewsDataSource) GetFunctionParamMapping() map[string]provider.FunctionParamInfo {
	return map[string]provider.FunctionParamInfo{}
}

// MockIndicatorsDataSource 模拟指标数据源
type MockIndicatorsDataSource struct {
	provider *MockDataProvider
}

func (m *MockIndicatorsDataSource) GetName() string {
	return "indicators"
}

func (m *MockIndicatorsDataSource) GetData(ctx context.Context, dataSource, field string, params ...provider.DataParam) (interface{}, error) {
	return m.provider.GetIndicatorField(ctx, dataSource, field)
}

func (m *MockIndicatorsDataSource) GetFunctionParamMapping() map[string]provider.FunctionParamInfo {
	return map[string]provider.FunctionParamInfo{}
}

func NewMockDataProvider() *MockDataProvider {
	return &MockDataProvider{
		klineData:  make(map[string]map[string][]float64),
		marketData: make(map[string]map[string]float64),
		newsData:   make(map[string]map[string]interface{}),
	}
}

func (m *MockDataProvider) GetKline(ctx context.Context, symbol, field string) ([]float64, error) {
	if symbolData, exists := m.klineData[symbol]; exists {
		if fieldData, exists := symbolData[field]; exists {
			return fieldData, nil
		}
	}
	return []float64{100, 101, 102, 103, 104}, nil // 默认数据
}

func (m *MockDataProvider) GetKlineField(ctx context.Context, dataSource, field string) (interface{}, error) {
	// 解析字段 - 支持多级字段如 "BTC.close"
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("kline field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

	if symbolData, exists := m.klineData[symbol]; exists {
		if fieldData, exists := symbolData[fieldName]; exists {
			if len(fieldData) > 0 {
				return fieldData[len(fieldData)-1], nil
			}
		}
	}

	// 默认价格
	if symbol == "BTC" {
		return 104.0, nil
	} else if symbol == "SOL" {
		return 204.0, nil
	}
	return 104.0, nil
}

func (m *MockDataProvider) GetMarketField(ctx context.Context, dataSource, field string) (interface{}, error) {
	// 解析字段 - 支持多级字段如 "BTC.last_px"
	fieldParts := strings.Split(field, ".")
	if len(fieldParts) < 2 {
		return nil, fmt.Errorf("market field must be in format 'SYMBOL.FIELD', got: %s", field)
	}

	symbol := fieldParts[0]
	fieldName := fieldParts[1]

	if symbolData, exists := m.marketData[symbol]; exists {
		if fieldData, exists := symbolData[fieldName]; exists {
			return fieldData, nil
		}
	}

	// 默认行情数据
	switch fieldName {
	case "last_px":
		if symbol == "BTC" {
			return 104.0, nil
		} else if symbol == "SOL" {
			return 204.0, nil
		}
		return 104.0, nil
	case "last_volume":
		return 1000.0, nil
	case "bid":
		if symbol == "BTC" {
			return 103.5, nil
		} else if symbol == "SOL" {
			return 203.5, nil
		}
		return 103.5, nil
	case "ask":
		if symbol == "BTC" {
			return 104.5, nil
		} else if symbol == "SOL" {
			return 204.5, nil
		}
		return 104.5, nil
	}
	return nil, fmt.Errorf("unknown field: %s", fieldName)
}

func (m *MockDataProvider) GetNewsField(ctx context.Context, dataSource, field string) (interface{}, error) {
	// News 模块支持简单字段名，不需要多级字段
	if sourceData, exists := m.newsData[dataSource]; exists {
		if fieldData, exists := sourceData[field]; exists {
			return fieldData, nil
		}
		// 处理last_title字段，等同于title
		if field == "last_title" {
			if fieldData, exists := sourceData["title"]; exists {
				return fieldData, nil
			}
		}
	}

	// 默认新闻数据
	switch field {
	case "title", "last_title":
		if dataSource == "coindesk" {
			return "比特币价格创新高突破历史记录", nil
		} else if dataSource == "theblockbeats" {
			return "加密货币市场迎来重大突破创新高", nil
		}
		return "比特币价格创新高突破历史记录", nil
	case "last_update_time":
		return time.Now().Add(-300 * time.Second), nil // 5分钟前
	case "sentiment":
		return "positive", nil
	default:
		return "", nil
	}
}

func (m *MockDataProvider) GetIndicatorField(ctx context.Context, dataSource, field string) (interface{}, error) {
	// 默认指标数据
	switch field {
	case "rsi":
		return 65.5, nil
	case "macd":
		return 1.2, nil
	default:
		return 0.0, nil
	}
}

// 实现 functions.Evaluator 接口
func (m *MockDataProvider) GetFieldValue(ctx context.Context, module, dataSource, field string) (interface{}, error) {
	switch module {
	case "kline":
		return m.GetKlineField(ctx, dataSource, field)
	case "market":
		return m.GetMarketField(ctx, dataSource, field)
	case "news":
		return m.GetNewsField(ctx, dataSource, field)
	case "indicators":
		return m.GetIndicatorField(ctx, dataSource, field)
	default:
		return nil, fmt.Errorf("unknown module: %s", module)
	}
}

func (m *MockDataProvider) CallFunction(ctx context.Context, name string, args []interface{}) (interface{}, error) {
	// 这里需要实现函数调用逻辑
	// 为了简化测试，暂时返回错误
	return nil, fmt.Errorf("function call not implemented in mock")
}

func (m *MockDataProvider) GetDataSource(name string) (interface{}, bool) {
	switch name {
	case "kline":
		return &MockKlineDataSource{provider: m}, true
	case "market":
		return &MockMarketDataSource{provider: m}, true
	case "news":
		return &MockNewsDataSource{provider: m}, true
	case "indicators":
		return &MockIndicatorsDataSource{provider: m}, true
	default:
		return nil, false
	}
}

func (m *MockDataProvider) GetHistoricalData(ctx context.Context, source, dataSource, field string, period int) ([]interface{}, error) {
	switch source {
	case "kline":
		// 解析字段获取符号
		fieldParts := strings.Split(field, ".")
		if len(fieldParts) < 2 {
			return nil, fmt.Errorf("kline field must be in format 'SYMBOL.FIELD', got: %s", field)
		}
		symbol := fieldParts[0]
		fieldName := fieldParts[1]

		data, err := m.GetKline(ctx, symbol, fieldName)
		if err != nil {
			return nil, err
		}
		// 转换为 []interface{}
		result := make([]interface{}, len(data))
		for i, v := range data {
			result[i] = v
		}
		return result, nil
	default:
		return nil, fmt.Errorf("historical data not supported for source: %s", source)
	}
}

func TestParser(t *testing.T) {
	parser := NewParser()

	testCases := []struct {
		name       string
		expression string
		shouldErr  bool
	}{
		{
			name:       "简单比较",
			expression: "kline.okx.BTC.close > 100",
			shouldErr:  false,
		},
		{
			name:       "逻辑表达式",
			expression: "kline.okx.BTC.close > 100 and kline.binance.ETH.close < 200",
			shouldErr:  false,
		},
		{
			name:       "函数调用",
			expression: "avg(kline.okx.BTC.close, 5)",
			shouldErr:  false,
		},
		{
			name:       "括号表达式",
			expression: "(kline.okx.BTC.close > 100) and (kline.binance.ETH.close < 200)",
			shouldErr:  false,
		},
		{
			name:       "字符串数组",
			expression: "contains(news.coindesk.title, [\"新高\", \"突破\"])",
			shouldErr:  false,
		},
		{
			name:       "复杂表达式",
			expression: "(avg(kline.okx.BTC.close, 5) > 100 and ago(news.coindesk.last_update_time) < 600) or (kline.gate.SOL.close >= 200 and contains(news.theblockbeats.title, [\"新高\"]))",
			shouldErr:  false,
		},
		{
			name:       "语法错误 - 缺少括号",
			expression: "kline.okx.BTC.close > 100 and",
			shouldErr:  true,
		},
		{
			name:       "语法错误 - 无效操作符",
			expression: "kline.okx.BTC.close >> 100",
			shouldErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic: %v", r)
					}
				}()
				_, err = parser.Parse(tc.expression)
			}()

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none for expression: %s", tc.expression)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for expression %s: %v", tc.expression, err)
				}
			}
		})
	}
}

func TestTokenizer(t *testing.T) {
	tokenizer := NewTokenizer("kline.okx.BTC.close > 100 and avg(kline.binance.ETH.close, 5) < 200")

	expectedTokens := []Token{
		{Type: TokenIdent, Value: "kline"},
		{Type: TokenDot, Value: "."},
		{Type: TokenIdent, Value: "okx"},
		{Type: TokenDot, Value: "."},
		{Type: TokenIdent, Value: "BTC"},
		{Type: TokenDot, Value: "."},
		{Type: TokenIdent, Value: "close"},
		{Type: TokenOp, Value: ">"},
		{Type: TokenNumber, Value: "100"},
		{Type: TokenAnd, Value: "and"},
		{Type: TokenIdent, Value: "avg"},
		{Type: TokenLParen, Value: "("},
		{Type: TokenIdent, Value: "kline"},
		{Type: TokenDot, Value: "."},
		{Type: TokenIdent, Value: "binance"},
		{Type: TokenDot, Value: "."},
		{Type: TokenIdent, Value: "ETH"},
		{Type: TokenDot, Value: "."},
		{Type: TokenIdent, Value: "close"},
		{Type: TokenComma, Value: ","},
		{Type: TokenNumber, Value: "5"},
		{Type: TokenRParen, Value: ")"},
		{Type: TokenOp, Value: "<"},
		{Type: TokenNumber, Value: "200"},
		{Type: TokenEOF, Value: ""},
	}

	for i, expected := range expectedTokens {
		token := tokenizer.NextToken()
		if token.Type != expected.Type || token.Value != expected.Value {
			t.Errorf("Token %d: expected %v but got %v", i, expected, token)
		}
	}
}

func TestSyntaxEngine(t *testing.T) {
	// 创建模拟数据提供者
	mockProvider := NewMockDataProvider()

	// 设置测试数据
	mockProvider.klineData["BTC"] = map[string][]float64{
		"close": {100, 101, 102, 103, 104},
		"open":  {99, 100, 101, 102, 103},
		"high":  {101, 102, 103, 104, 105},
		"low":   {98, 99, 100, 101, 102},
	}

	mockProvider.klineData["SOL"] = map[string][]float64{
		"close": {200, 201, 202, 203, 204},
		"open":  {199, 200, 201, 202, 203},
		"high":  {201, 202, 203, 204, 205},
		"low":   {198, 199, 200, 201, 202},
	}

	mockProvider.marketData["BTC"] = map[string]float64{
		"last_px":     104.0,
		"last_volume": 1000.0,
		"bid":         103.5,
		"ask":         104.5,
	}

	mockProvider.marketData["SOL"] = map[string]float64{
		"last_px":     204.0,
		"last_volume": 2000.0,
		"bid":         203.5,
		"ask":         204.5,
	}

	mockProvider.newsData["coindesk"] = map[string]interface{}{
		"title":            "比特币价格创新高突破历史记录",
		"last_update_time": time.Now().Add(-300 * time.Second),
		"sentiment":        "positive",
	}

	mockProvider.newsData["theblockbeats"] = map[string]interface{}{
		"title":            "加密货币市场迎来重大突破创新高",
		"last_update_time": time.Now().Add(-600 * time.Second),
	}

	// 创建语法引擎
	registry := registry.DefaultRegistry()

	// 将 MockDataProvider 注册为数据源
	klineDataSource := &MockKlineDataSource{provider: mockProvider}
	marketDataSource := &MockMarketDataSource{provider: mockProvider}
	newsDataSource := &MockNewsDataSource{provider: mockProvider}
	indicatorsDataSource := &MockIndicatorsDataSource{provider: mockProvider}

	registry.RegisterProvider(klineDataSource)
	registry.RegisterProvider(marketDataSource)
	registry.RegisterProvider(newsDataSource)
	registry.RegisterProvider(indicatorsDataSource)

	evaluator := NewEvaluator(registry)
	engine := &Engine{
		parser:    NewParser(),
		evaluator: evaluator,
		registry:  registry,
	}

	ctx := context.Background()

	// 测试用例
	testCases := []struct {
		name       string
		expression string
		expected   bool
		shouldErr  bool
	}{
		{
			name:       "简单比较",
			expression: "kline.okx.BTC.close > 100",
			expected:   true,
			shouldErr:  false,
		},
		{
			name:       "逻辑AND",
			expression: "kline.okx.BTC.close > 100 and kline.okx.BTC.close < 200",
			expected:   true,
			shouldErr:  false,
		},
		{
			name:       "逻辑OR",
			expression: "kline.okx.BTC.close < 50 or kline.okx.BTC.close > 100",
			expected:   true,
			shouldErr:  false,
		},
		{
			name:       "函数调用 - avg",
			expression: "avg(kline.okx.BTC.close, 5) > 100",
			expected:   true,
			shouldErr:  false,
		},
		{
			name:       "函数调用 - ago",
			expression: "ago(news.coindesk.last_update_time) < 600",
			expected:   true,
			shouldErr:  false,
		},
		{
			name:       "函数调用 - has",
			expression: "has(news.coindesk.title, \"新高\")",
			expected:   true,
			shouldErr:  false,
		},
		{
			name:       "复杂表达式",
			expression: "(avg(kline.okx.BTC.close, 5) > market.okx.BTC.last_px and ago(news.coindesk.last_update_time) < 600) or (market.gate.SOL.last_px >= 200 and has(news.theblockbeats.last_title, \"新高\"))",
			expected:   true,
			shouldErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := engine.ExecuteExpressionToBool(ctx, tc.expression)

			if tc.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tc.expected {
				t.Errorf("Expected %v but got %v for expression: %s", tc.expected, result, tc.expression)
			}
		})
	}
}
