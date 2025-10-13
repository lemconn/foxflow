package exchange

import (
	"context"
	"net"
	"testing"
	"time"
)

// checkNetworkConnectivity 检查网络连接性
func checkNetworkConnectivity(host string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func TestOKXExchange_GetMarkPriceCandles_RealAPI(t *testing.T) {
	// 首先检查网络连接性
	okxHost := "www.okx.com:443"
	timeout := 5 * time.Second

	// 检查是否能直接连接OKX
	canConnectDirectly := checkNetworkConnectivity(okxHost, timeout)

	var exchange *OKXExchange
	var apiURL string

	if canConnectDirectly {
		// 可以直接连接，使用官方API
		apiURL = "https://www.okx.com"
		exchange = NewOKXExchange(apiURL, "")
		t.Log("使用直接连接测试OKX API")
	} else {
		// 无法直接连接，尝试使用代理
		proxyURL := "http://127.0.0.1:7890"
		apiURL = "https://www.okx.com"
		exchange = NewOKXExchange(apiURL, proxyURL)
		t.Log("使用代理连接测试OKX API")
	}

	// 创建上下文，设置较短的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 执行测试 - 获取BTC-USDT-SWAP的15分钟K线数据，最多5条
	candles, err := exchange.GetMarkPriceCandles(ctx, "BTC-USDT-SWAP", "15m", 5)

	// 如果是网络连接问题，跳过测试而不是失败
	if err != nil {
		errorMsg := err.Error()
		if contains(errorMsg, "no route to host") ||
			contains(errorMsg, "connection refused") ||
			contains(errorMsg, "timeout") ||
			contains(errorMsg, "network is unreachable") ||
			contains(errorMsg, "i/o timeout") ||
			contains(errorMsg, "connection reset by peer") {
			t.Skipf("网络连接问题，跳过测试: %v", err)
		}
		t.Fatalf("获取K线数据失败: %v", err)
	}

	if len(candles) == 0 {
		t.Fatal("未获取到任何K线数据")
	}

	t.Logf("成功获取到 %d 条K线数据", len(candles))

	// 验证数据质量
	for i, candle := range candles {
		// 验证时间戳
		if candle.Ts == 0 {
			t.Errorf("K线 %d 的时间戳无效", i)
		}

		// 验证价格数据
		if candle.Open <= 0 {
			t.Errorf("K线 %d 的开盘价无效: %f", i, candle.Open)
		}
		if candle.High <= 0 {
			t.Errorf("K线 %d 的最高价无效: %f", i, candle.High)
		}
		if candle.Low <= 0 {
			t.Errorf("K线 %d 的最低价无效: %f", i, candle.Low)
		}
		if candle.Close <= 0 {
			t.Errorf("K线 %d 的收盘价无效: %f", i, candle.Close)
		}

		// 验证价格逻辑关系
		if candle.High < candle.Low {
			t.Errorf("K线 %d 的最高价低于最低价: H=%f, L=%f", i, candle.High, candle.Low)
		}
		if candle.High < candle.Open {
			t.Errorf("K线 %d 的最高价低于开盘价: H=%f, O=%f", i, candle.High, candle.Open)
		}
		if candle.High < candle.Close {
			t.Errorf("K线 %d 的最高价低于收盘价: H=%f, C=%f", i, candle.High, candle.Close)
		}
		if candle.Low > candle.Open {
			t.Errorf("K线 %d 的最低价高于开盘价: L=%f, O=%f", i, candle.Low, candle.Open)
		}
		if candle.Low > candle.Close {
			t.Errorf("K线 %d 的最低价高于收盘价: L=%f, C=%f", i, candle.Low, candle.Close)
		}

		// 验证确认状态
		if candle.Confirm != 0 && candle.Confirm != 1 {
			t.Errorf("K线 %d 的确认状态无效: %d", i, candle.Confirm)
		}

		// 打印第一条K线的详细信息用于验证
		if i == 0 {
			timestamp := time.Unix(candle.Ts/1000, (candle.Ts%1000)*1000000)
			t.Logf("第一条K线数据:")
			t.Logf("  时间: %s", timestamp.Format("2006-01-02 15:04:05"))
			t.Logf("  开盘价: %.2f", candle.Open)
			t.Logf("  最高价: %.2f", candle.High)
			t.Logf("  最低价: %.2f", candle.Low)
			t.Logf("  收盘价: %.2f", candle.Close)
			t.Logf("  状态: %s", getConfirmStatus(candle.Confirm))
		}
	}

	// 验证时间顺序（如果有多条数据）
	if len(candles) > 1 {
		for i := 1; i < len(candles); i++ {
			if candles[i].Ts >= candles[i-1].Ts {
				t.Errorf("K线数据时间顺序错误: 第%d条时间戳(%d)应该小于第%d条时间戳(%d)",
					i, candles[i].Ts, i-1, candles[i-1].Ts)
			}
		}
	}
}

// getConfirmStatus 返回确认状态的中文描述
func getConfirmStatus(confirm int) string {
	if confirm == 0 {
		return "未完结"
	}
	return "已完结"
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
