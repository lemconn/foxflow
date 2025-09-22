package provider

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

// DataParam 单个数据参数
type DataParam struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"` // 实际值
}

// ParamType 参数类型
type ParamType string

const (
	ParamTypeInt    ParamType = "int"
	ParamTypeFloat  ParamType = "float"
	ParamTypeString ParamType = "string"
	ParamTypeBool   ParamType = "bool"
	ParamTypeAny    ParamType = "any"
)

// FunctionParam 单个函数参数信息
type FunctionParam struct {
	ParamIndex int         `json:"param_index"` // 函数中第几个参数（从0开始）
	ParamName  string      `json:"param_name"`  // 参数名称
	ParamType  ParamType   `json:"param_type"`  // 参数类型
	Required   bool        `json:"required"`    // 是否必需
	Default    interface{} `json:"default"`     // 默认值（可选）
}

// FunctionParamInfo 函数参数信息
type FunctionParamInfo struct {
	FunctionName string          `json:"function_name"` // 函数名称
	Params       []FunctionParam `json:"params"`        // 参数列表
}

// Provider 数据提供者接口
type Provider interface {
	// GetName 获取提供者名称
	GetName() string

	// GetData 获取数据
	// params 为可选参数，使用 ...DataParam 的方式传递
	// 如果不传递参数，则使用默认行为
	GetData(ctx context.Context, entity, field string, params ...DataParam) (interface{}, error)

	// GetFunctionParamMapping 获取函数参数映射
	// 返回该数据提供者支持的函数及其参数映射关系
	GetFunctionParamMapping() map[string]FunctionParamInfo
}

// Manager 数据提供者管理器
type Manager struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewManager 创建数据提供者管理器
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider 注册数据提供者
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.GetName()] = provider
}

// GetProvider 获取数据提供者
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[name]
	if !exists {
		return nil, fmt.Errorf("data provider not found: %s", name)
	}

	return provider, nil
}

// GetData 获取数据
func (m *Manager) GetData(ctx context.Context, providerName, entity, field string, params ...DataParam) (interface{}, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetData(ctx, entity, field, params...)
}

// ListProviders 列出所有注册的提供者
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}

	return providers
}

// InitDefaultProviders 初始化默认数据提供者
func InitDefaultProviders() *Manager {
	manager := NewManager()

	// 注册所有默认提供者
	manager.RegisterProvider(NewKlineProvider())
	manager.RegisterProvider(NewMarketProvider())
	manager.RegisterProvider(NewNewsProvider())
	manager.RegisterProvider(NewIndicatorsProvider())

	return manager
}

// ConvertParamValue 根据参数类型转换值
func ConvertParamValue(value interface{}, paramType ParamType) (interface{}, error) {
	switch paramType {
	case ParamTypeInt:
		switch v := value.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		case string:
			// 尝试解析字符串为整数
			if intVal, err := strconv.Atoi(v); err == nil {
				return intVal, nil
			}
			return nil, fmt.Errorf("cannot convert string '%s' to int", v)
		default:
			return nil, fmt.Errorf("cannot convert %T to int", value)
		}
	case ParamTypeFloat:
		switch v := value.(type) {
		case float64:
			return v, nil
		case int:
			return float64(v), nil
		case string:
			// 尝试解析字符串为浮点数
			if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
				return floatVal, nil
			}
			return nil, fmt.Errorf("cannot convert string '%s' to float", v)
		default:
			return nil, fmt.Errorf("cannot convert %T to float", value)
		}
	case ParamTypeString:
		switch v := value.(type) {
		case string:
			return v, nil
		default:
			return fmt.Sprintf("%v", v), nil
		}
	case ParamTypeBool:
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			// 尝试解析字符串为布尔值
			if boolVal, err := strconv.ParseBool(v); err == nil {
				return boolVal, nil
			}
			return nil, fmt.Errorf("cannot convert string '%s' to bool", v)
		default:
			return nil, fmt.Errorf("cannot convert %T to bool", value)
		}
	case ParamTypeAny:
		return value, nil
	default:
		return nil, fmt.Errorf("unknown parameter type: %s", paramType)
	}
}
