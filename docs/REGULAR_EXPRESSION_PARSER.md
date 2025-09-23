### 1. Go 正则解析器实现

```go
package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type Command struct {
	Action    string
	Symbol    string
	Side      string
	Amount    float64
	Condition string
}

func ParseCommand(input string) (*Command, error) {
	// 正则：匹配 open SYMBOL long|short 数字 [with 条件...]
	pattern := `^(open|close)\s+([A-Za-z0-9]+)\s+(long|short)\s+([\d.]+)(?:\s+with\s+(.+))?$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return nil, fmt.Errorf("invalid command: %s", input)
	}

	amount, err := strconv.ParseFloat(matches[4], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %s", matches[4])
	}

	return &Command{
		Action:    matches[1],
		Symbol:    matches[2],
		Side:      matches[3],
		Amount:    amount,
		Condition: matches[5], // 可能为空
	}, nil
}

func main() {
	input := `open SOL long 1.0 with avg(kline.BTC.close, 5) > 100 and has(news.theblockbeats.last_title, "新高")`

	cmd, err := ParseCommand(input)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Action: %s\n", cmd.Action)
	fmt.Printf("Symbol: %s\n", cmd.Symbol)
	fmt.Printf("Side: %s\n", cmd.Side)
	fmt.Printf("Amount: %.2f\n", cmd.Amount)
	fmt.Printf("Condition: %s\n", cmd.Condition)
}
```

### 2. 运行结果
```vbnet
Action: open
Symbol: SOL
Side: long
Amount: 1.00
Condition: avg(kline.BTC.close, 5) > 100 and has(news.theblockbeats.last_title, "新高")
```

### 3. 说明
- `Action` 目前只支持 open 和 close，你可以扩展成 (open|close|cancel|modify)。
- `Symbol` 允许字母和数字（[A-Za-z0-9]+），比如 BTCUSDT。
- `Side` 目前只支持 long|short。
- `Amount` 允许小数。
- `Condition` 是可选的（with ... 部分），提取出来可以直接送进你的 DSL 解析器。
