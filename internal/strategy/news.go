package strategy

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lemconn/foxflow/internal/exchange"
)

// NewsData 新闻数据
type NewsData struct {
	Source         string    `json:"source"`           // 新闻源，如 theblockbeats
	Title          string    `json:"title"`            // 新闻标题
	Content        string    `json:"content"`          // 新闻内容
	UpdateTime     time.Time `json:"update_time"`      // 更新时间
	LastTitle      string    `json:"last_title"`       // 最新标题
	LastUpdateTime time.Time `json:"last_update_time"` // 最新更新时间
	Keywords       []string  `json:"keywords"`         // 关键词
	Sentiment      string    `json:"sentiment"`        // 情感分析：positive, negative, neutral
}

// NewsStrategy 新闻策略
type NewsStrategy struct {
	name        string
	description string
	mockData    map[string]*NewsData
	mu          sync.RWMutex
}

// NewNewsStrategy 创建新闻策略
func NewNewsStrategy() *NewsStrategy {
	strategy := &NewsStrategy{
		name:        "news",
		description: "新闻数据策略：基于新闻数据进行条件判断",
		mockData:    make(map[string]*NewsData),
	}

	// 初始化Mock数据
	strategy.initMockData()

	return strategy
}

// initMockData 初始化Mock数据
func (s *NewsStrategy) initMockData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TheBlockBeats Mock数据
	s.mockData["theblockbeats"] = &NewsData{
		Source:         "theblockbeats",
		Title:          "SOL突破新高，市值创新纪录",
		Content:        "Solana代币SOL价格突破200美元大关，创下历史新高...",
		UpdateTime:     time.Now().Add(-5 * time.Minute),
		LastTitle:      "SOL突破新高，市值创新纪录",
		LastUpdateTime: time.Now().Add(-5 * time.Minute),
		Keywords:       []string{"SOL", "突破", "新高", "市值"},
		Sentiment:      "positive",
	}

	// CoinDesk Mock数据
	s.mockData["coindesk"] = &NewsData{
		Source:         "coindesk",
		Title:          "比特币ETF获批，市场反应积极",
		Content:        "美国证券交易委员会批准了首个比特币ETF...",
		UpdateTime:     time.Now().Add(-10 * time.Minute),
		LastTitle:      "比特币ETF获批，市场反应积极",
		LastUpdateTime: time.Now().Add(-10 * time.Minute),
		Keywords:       []string{"比特币", "ETF", "获批", "市场"},
		Sentiment:      "positive",
	}

	// CoinTelegraph Mock数据
	s.mockData["cointelegraph"] = &NewsData{
		Source:         "cointelegraph",
		Title:          "以太坊2.0升级进展顺利",
		Content:        "以太坊网络升级进展顺利，交易费用显著降低...",
		UpdateTime:     time.Now().Add(-15 * time.Minute),
		LastTitle:      "以太坊2.0升级进展顺利",
		LastUpdateTime: time.Now().Add(-15 * time.Minute),
		Keywords:       []string{"以太坊", "升级", "交易费用", "网络"},
		Sentiment:      "positive",
	}
}

func (s *NewsStrategy) GetName() string {
	return s.name
}

func (s *NewsStrategy) GetDescription() string {
	return s.description
}

func (s *NewsStrategy) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"source":   "新闻源，如 theblockbeats, coindesk, cointelegraph",
		"field":    "字段名：last_title, last_update_time, sentiment, keywords",
		"operator": "操作符：>, <, >=, <=, ==, !=, in, not_in, contains",
		"value":    "比较值或字符串",
		"values":   "值数组（用于 in/not_in 操作）",
	}
}

func (s *NewsStrategy) ValidateParameters(params map[string]interface{}) error {
	// 验证必需参数
	requiredParams := []string{"source", "field", "operator"}
	for _, param := range requiredParams {
		if _, exists := params[param]; !exists {
			return fmt.Errorf("missing required parameter: %s", param)
		}
	}

	// 验证新闻源
	source, ok := params["source"].(string)
	if !ok {
		return fmt.Errorf("source must be a string")
	}

	// 检查新闻源是否在Mock数据中存在
	s.mu.RLock()
	_, exists := s.mockData[source]
	s.mu.RUnlock()
	if !exists {
		return fmt.Errorf("invalid source: %s", source)
	}

	// 验证字段名
	field, ok := params["field"].(string)
	if !ok {
		return fmt.Errorf("field must be a string")
	}

	validFields := []string{"last_title", "last_update_time", "sentiment", "keywords"}
	validField := false
	for _, validF := range validFields {
		if field == validF {
			validField = true
			break
		}
	}
	if !validField {
		return fmt.Errorf("invalid field: %s, valid fields are: %v", field, validFields)
	}

	// 验证操作符
	operator, ok := params["operator"].(string)
	if !ok {
		return fmt.Errorf("operator must be a string")
	}

	validOperators := []string{">", "<", ">=", "<=", "==", "!=", "in", "not_in", "contains"}
	validOp := false
	for _, validO := range validOperators {
		if operator == validO {
			validOp = true
			break
		}
	}
	if !validOp {
		return fmt.Errorf("invalid operator: %s, valid operators are: %v", operator, validOperators)
	}

	// 验证值
	value := params["value"]
	values := params["values"]

	// 对于 in/not_in 操作，需要 values 参数
	if operator == "in" || operator == "not_in" {
		if values == nil {
			return fmt.Errorf("values parameter is required for %s operator", operator)
		}
	} else {
		// 对于其他操作，需要 value 参数
		if value == nil {
			return fmt.Errorf("value parameter is required for %s operator", operator)
		}
	}

	// 验证时间字段的特殊处理
	if field == "last_update_time" {
		if operator != ">" && operator != "<" && operator != ">=" && operator != "<=" && operator != "==" && operator != "!=" {
			return fmt.Errorf("time field only supports comparison operators")
		}
	}

	return nil
}

func (s *NewsStrategy) Evaluate(ctx context.Context, exchange exchange.Exchange, symbol string, params map[string]interface{}) (bool, error) {
	// 获取参数
	source, ok := params["source"].(string)
	if !ok {
		return false, fmt.Errorf("source parameter must be a string")
	}

	field, ok := params["field"].(string)
	if !ok {
		return false, fmt.Errorf("field parameter must be a string")
	}

	operator, ok := params["operator"].(string)
	if !ok {
		return false, fmt.Errorf("operator parameter must be a string")
	}

	value := params["value"]
	values := params["values"]

	// 获取新闻数据
	newsData, err := s.getNewsData(source)
	if err != nil {
		return false, fmt.Errorf("failed to get news data for %s: %w", source, err)
	}

	// 执行比较
	var result bool

	switch operator {
	case "in", "not_in":
		if values == nil {
			return false, fmt.Errorf("values parameter is required for %s operator", operator)
		}
		result, err = s.evaluateInOperator(newsData, field, operator, values)
	case "contains":
		if value == nil {
			return false, fmt.Errorf("value parameter is required for contains operator")
		}
		result, err = s.evaluateContainsOperator(newsData, field, value)
	default:
		if value == nil {
			return false, fmt.Errorf("value parameter is required for %s operator", operator)
		}
		result, err = s.evaluateComparisonOperator(newsData, field, operator, value)
	}

	if err != nil {
		return false, fmt.Errorf("failed to evaluate: %w", err)
	}

	return result, nil
}

// getNewsData 获取新闻数据
func (s *NewsStrategy) getNewsData(source string) (*NewsData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.mockData[source]
	if !exists {
		return nil, fmt.Errorf("no news data found for source: %s", source)
	}

	return data, nil
}

// evaluateInOperator 评估 in/not_in 操作
func (s *NewsStrategy) evaluateInOperator(data *NewsData, field, operator string, values interface{}) (bool, error) {
	// 获取字段值
	fieldValue, err := s.getFieldValue(data, field)
	if err != nil {
		return false, err
	}

	// 转换值数组
	valueList, err := s.convertValueList(values)
	if err != nil {
		return false, err
	}

	// 检查是否包含
	contains := false
	for _, val := range valueList {
		if s.compareFieldValue(fieldValue, val) {
			contains = true
			break
		}
	}

	// 根据操作符返回结果
	if operator == "in" {
		return contains, nil
	} else { // not_in
		return !contains, nil
	}
}

// evaluateContainsOperator 评估 contains 操作
func (s *NewsStrategy) evaluateContainsOperator(data *NewsData, field string, value interface{}) (bool, error) {
	// 获取字段值
	fieldValue, err := s.getFieldValue(data, field)
	if err != nil {
		return false, err
	}

	// 转换比较值
	compareValue, err := s.convertValue(value)
	if err != nil {
		return false, err
	}

	// 字符串包含检查
	fieldStr := fmt.Sprintf("%v", fieldValue)
	compareStr := fmt.Sprintf("%v", compareValue)

	return strings.Contains(fieldStr, compareStr), nil
}

// evaluateComparisonOperator 评估比较操作
func (s *NewsStrategy) evaluateComparisonOperator(data *NewsData, field, operator string, value interface{}) (bool, error) {
	// 获取字段值
	fieldValue, err := s.getFieldValue(data, field)
	if err != nil {
		return false, err
	}

	// 转换比较值
	compareValue, err := s.convertValue(value)
	if err != nil {
		return false, err
	}

	// 执行比较
	return s.compareValues(fieldValue, operator, compareValue)
}

// getFieldValue 获取字段值
func (s *NewsStrategy) getFieldValue(data *NewsData, field string) (interface{}, error) {
	switch field {
	case "last_title":
		return data.LastTitle, nil
	case "last_update_time":
		return data.LastUpdateTime, nil
	case "sentiment":
		return data.Sentiment, nil
	case "keywords":
		return data.Keywords, nil
	default:
		return nil, fmt.Errorf("unknown field: %s", field)
	}
}

// convertValue 转换值
func (s *NewsStrategy) convertValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// 尝试解析为数字
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, nil
		}
		// 尝试解析为时间
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		// 作为字符串返回
		return v, nil
	case float64, int, int64, time.Time:
		return v, nil
	default:
		return v, nil
	}
}

// convertValueList 转换值列表
func (s *NewsStrategy) convertValueList(values interface{}) ([]interface{}, error) {
	switch v := values.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = val
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported values type: %T", values)
	}
}

// compareFieldValue 比较字段值
func (s *NewsStrategy) compareFieldValue(fieldValue, compareValue interface{}) bool {
	// 字符串比较
	fieldStr := fmt.Sprintf("%v", fieldValue)
	compareStr := fmt.Sprintf("%v", compareValue)

	// 对于字符串字段，检查是否包含
	if fieldStr != "" && compareStr != "" {
		return strings.Contains(fieldStr, compareStr)
	}

	return fieldStr == compareStr
}

// compareValues 比较值
func (s *NewsStrategy) compareValues(fieldValue interface{}, operator string, compareValue interface{}) (bool, error) {
	// 时间比较
	if fieldTime, ok := fieldValue.(time.Time); ok {
		if compareTime, ok := compareValue.(time.Time); ok {
			return s.compareTimes(fieldTime, operator, compareTime), nil
		}
		// 尝试将比较值转换为时间
		if compareStr, ok := compareValue.(string); ok {
			if compareTime, err := time.Parse(time.RFC3339, compareStr); err == nil {
				return s.compareTimes(fieldTime, operator, compareTime), nil
			}
		}
		// 尝试将比较值转换为数字（秒数）
		if compareNum, ok := compareValue.(float64); ok {
			compareTime := time.Now().Add(-time.Duration(compareNum) * time.Second)
			return s.compareTimes(fieldTime, operator, compareTime), nil
		}
	}

	// 数字比较
	if fieldNum, ok := fieldValue.(float64); ok {
		if compareNum, ok := compareValue.(float64); ok {
			return s.compareNumbers(fieldNum, operator, compareNum), nil
		}
	}

	// 字符串比较
	fieldStr := fmt.Sprintf("%v", fieldValue)
	compareStr := fmt.Sprintf("%v", compareValue)
	return s.compareStrings(fieldStr, operator, compareStr), nil
}

// compareTimes 比较时间
func (s *NewsStrategy) compareTimes(fieldTime time.Time, operator string, compareTime time.Time) bool {
	switch operator {
	case ">":
		return fieldTime.After(compareTime)
	case "<":
		return fieldTime.Before(compareTime)
	case ">=":
		return fieldTime.After(compareTime) || fieldTime.Equal(compareTime)
	case "<=":
		return fieldTime.Before(compareTime) || fieldTime.Equal(compareTime)
	case "==":
		return fieldTime.Equal(compareTime)
	case "!=":
		return !fieldTime.Equal(compareTime)
	default:
		return false
	}
}

// compareNumbers 比较数字
func (s *NewsStrategy) compareNumbers(fieldNum float64, operator string, compareNum float64) bool {
	switch operator {
	case ">":
		return fieldNum > compareNum
	case "<":
		return fieldNum < compareNum
	case ">=":
		return fieldNum >= compareNum
	case "<=":
		return fieldNum <= compareNum
	case "==":
		return fieldNum == compareNum
	case "!=":
		return fieldNum != compareNum
	default:
		return false
	}
}

// compareStrings 比较字符串
func (s *NewsStrategy) compareStrings(fieldStr string, operator string, compareStr string) bool {
	switch operator {
	case ">":
		return fieldStr > compareStr
	case "<":
		return fieldStr < compareStr
	case ">=":
		return fieldStr >= compareStr
	case "<=":
		return fieldStr <= compareStr
	case "==":
		return fieldStr == compareStr
	case "!=":
		return fieldStr != compareStr
	default:
		return false
	}
}

// UpdateMockData 更新Mock数据（用于测试）
func (s *NewsStrategy) UpdateMockData(source string, data *NewsData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mockData[source] = data
}

// GetMockData 获取Mock数据（用于测试）
func (s *NewsStrategy) GetMockData(source string) (*NewsData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, exists := s.mockData[source]
	return data, exists
}
