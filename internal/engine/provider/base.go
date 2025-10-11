package provider

import (
	"fmt"
	"time"
)

// BaseProvider 基础提供者，提供通用的参数处理方法
type BaseProvider struct {
	name string
}

// NewBaseProvider 创建基础提供者
func NewBaseProvider(name string) *BaseProvider {
	return &BaseProvider{name: name}
}

// GetName 获取提供者名称
func (b *BaseProvider) GetName() string {
	return b.name
}

// GetParam 获取指定类型的参数值
func (b *BaseProvider) GetParam(params []DataParam, name string, expectedType string) (interface{}, error) {
	for _, param := range params {
		if param.Name == name {
			// 直接进行类型转换
			switch expectedType {
			case "int":
				if intValue, ok := param.Value.(int); ok {
					return intValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid int", name)
			case "float64":
				if floatValue, ok := param.Value.(float64); ok {
					return floatValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid float64", name)
			case "string":
				if stringValue, ok := param.Value.(string); ok {
					return stringValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid string", name)
			case "bool":
				if boolValue, ok := param.Value.(bool); ok {
					return boolValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid bool", name)
			case "time":
				if timeValue, ok := param.Value.(time.Time); ok {
					return timeValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid time.Time", name)
			case "array":
				if arrayValue, ok := param.Value.([]interface{}); ok {
					return arrayValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid array", name)
			case "map":
				if mapValue, ok := param.Value.(map[string]interface{}); ok {
					return mapValue, nil
				}
				return nil, fmt.Errorf("parameter '%s' is not a valid map", name)
			default:
				return nil, fmt.Errorf("unsupported parameter type: %s", expectedType)
			}
		}
	}
	return nil, fmt.Errorf("parameter '%s' not found", name)
}

// GetIntParam 获取 int 类型参数
func (b *BaseProvider) GetIntParam(params []DataParam, name string) (int, error) {
	value, err := b.GetParam(params, name, "int")
	if err != nil {
		return 0, err
	}
	return value.(int), nil
}


// 便利函数：创建参数值
func NewParam(name string, value interface{}) DataParam {
	return DataParam{Name: name, Value: value}
}
