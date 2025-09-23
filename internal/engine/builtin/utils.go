package builtin

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// toFloat64 转换为float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// toBool 转换为bool
func toBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return val == "true", nil
	case float64:
		return val != 0, nil
	case int:
		return val != 0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// toString 转换为string
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case bool:
		return strconv.FormatBool(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toTime 转换为time.Time
func toTime(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		// 尝试解析常见的时间格式
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t, nil
			}
		}

		return time.Time{}, fmt.Errorf("cannot parse time string: %s", val)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", v)
	}
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// equals 比较两个值是否相等
func equals(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// toStringArray 转换为字符串数组
func toStringArray(v interface{}) ([]string, error) {
	switch val := v.(type) {
	case []string:
		return val, nil
	case []interface{}:
		result := make([]string, len(val))
		for i, item := range val {
			result[i] = toString(item)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to []string", v)
	}
}
