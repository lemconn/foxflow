package cli

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"
)

// InitExchangeSymbols 初始化交易所交易对数据
func InitExchangeSymbols() {
	// 以协程方式获取每个交易所的交易对信息
	go func() {
		log.Println("开始初始化交易所交易对数据...")
		exchangeManager := exchange.GetManager()
		availableExchanges := exchangeManager.GetAvailableExchanges()

		config.ExchangeSymbolList = make(map[string][]config.SymbolInfo)

		for _, exchangeName := range availableExchanges {

			go func(name string) {

				ex, err := exchangeManager.GetExchange(name)
				if err != nil {
					log.Printf("获取交易所 %s 实例失败: %v", name, err)
					return
				}

				// 获取交易对信息，优先获取永续合约
				symbols, err := ex.GetAllSymbols(context.Background(), "SWAP")
				if err != nil {
					log.Printf("获取交易所 %s 交易对信息失败: %v", name, err)
					return
				}

				// 存储到全局变量中
				symbolList := make([]config.SymbolInfo, 0)
				for _, symbol := range symbols {

					// 非 `-USDT-SWAP` 结尾的直接过滤掉
					if !strings.HasSuffix(symbol.Name, "-USDT-SWAP") {
						continue
					}

					symbolInfo := config.SymbolInfo{
						Type:  symbol.Type,
						Name:  symbol.Name,
						Base:  symbol.Base,
						Quote: symbol.Quote,
					}

					if symbol.MaxLever != "" {
						symbolInfo.MaxLever, err = strconv.ParseFloat(symbol.MaxLever, 64)
						if err != nil {
							log.Printf("获取交易所 %s 交易对最大杠杆失败: %v", name, err)
							continue
						}
					}

					if symbol.MaxLever != "" {
						symbolInfo.MinSize, err = strconv.ParseFloat(symbol.MinSize, 64)
						if err != nil {
							log.Printf("获取交易所 %s 交易对最小下单量失败: %v", name, err)
							continue
						}
					}

					if symbol.ContractValue != "" {
						symbolInfo.Contract, err = strconv.ParseFloat(symbol.ContractValue, 64)
						if err != nil {
							log.Printf("获取交易所 %s 交易对合约面值失败: %v", name, err)
							continue
						}
					}

					symbolList = append(symbolList, symbolInfo)
				}

				config.ExchangeSymbolList[name] = symbolList

			}(exchangeName)
		}
	}()
}
