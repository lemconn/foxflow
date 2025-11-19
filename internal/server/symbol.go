package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/exchange"
	pb "github.com/lemconn/foxflow/proto/generated"
)

type SymbolServer struct{}

func NewSymbolServer() *SymbolServer {
	return &SymbolServer{}
}

// GetSymbols 获取交易对列表
func (s *SymbolServer) GetSymbols(ctx context.Context, req *pb.GetSymbolsRequest) (*pb.GetSymbolsResponse, error) {
	if req.Exchange == "" {
		return &pb.GetSymbolsResponse{
			Success: false,
			Message: "exchange 是必填参数",
		}, nil
	}

	exchangeName := req.Exchange
	exchangeClient, err := exchange.GetManager().GetExchange(exchangeName)
	if err != nil {
		log.Printf("获取交易所 %s 客户端失败: %v", exchangeName, err)
		return &pb.GetSymbolsResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	symbolList := s.getSymbolList(ctx, exchangeName, exchangeClient)
	if len(symbolList) == 0 {
		return &pb.GetSymbolsResponse{
			Success: false,
			Message: fmt.Sprintf("交易所 %s 暂无交易对数据", exchangeName),
		}, nil
	}

	tickers, err := exchangeClient.GetTickers(ctx)
	if err != nil {
		log.Printf("获取交易所 %s Ticker 数据失败: %v", exchangeName, err)
		return &pb.GetSymbolsResponse{
			Success: false,
			Message: fmt.Sprintf("获取行情数据失败: %v", err),
		}, nil
	}

	tickerMap := make(map[string]exchange.Ticker, len(tickers))
	for _, ticker := range tickers {
		tickerMap[ticker.Symbol] = ticker
	}

	keyword := strings.ToUpper(strings.TrimSpace(req.Keyword))
	var pbSymbols []*pb.SymbolItem
	for _, symbolInfo := range symbolList {
		if keyword != "" && !strings.Contains(symbolInfo.Name, keyword) {
			continue
		}

		pbSymbol := &pb.SymbolItem{
			Exchange:    exchangeName,
			Type:        symbolInfo.Type,
			Name:        symbolInfo.Name,
			Base:        symbolInfo.Base,
			Quote:       symbolInfo.Quote,
			MaxLeverage: symbolInfo.MaxLever,
			MinSize:     symbolInfo.MinSize,
			Contract:    symbolInfo.Contract,
		}

		if ticker, ok := tickerMap[symbolInfo.Name]; ok {
			pbSymbol.Price = ticker.Price
			pbSymbol.Volume = ticker.Volume
			pbSymbol.High = ticker.High
			pbSymbol.Low = ticker.Low
		}

		pbSymbols = append(pbSymbols, pbSymbol)
	}

	return &pb.GetSymbolsResponse{
		Success: true,
		Message: fmt.Sprintf("成功获取 %d 个交易对", len(pbSymbols)),
		Symbols: pbSymbols,
	}, nil
}

func (s *SymbolServer) getSymbolList(ctx context.Context, exchangeName string, exchangeClient exchange.Exchange) []config.SymbolInfo {
	if config.ExchangeSymbolList != nil {
		if symbols, ok := config.ExchangeSymbolList[exchangeName]; ok && len(symbols) > 0 {
			return symbols
		}
	}

	symbols, err := exchangeClient.GetAllSymbols(ctx, "SWAP")
	if err != nil {
		log.Printf("从交易所 %s 获取交易对失败: %v", exchangeName, err)
		return nil
	}

	result := make([]config.SymbolInfo, 0, len(symbols))
	for _, symbol := range symbols {
		if !strings.HasSuffix(symbol.Name, "-USDT-SWAP") {
			continue
		}
		result = append(result, config.SymbolInfo{
			Type:     symbol.Type,
			Name:     symbol.Name,
			Base:     symbol.Base,
			Quote:    symbol.Quote,
			MaxLever: symbol.MaxLever,
			MinSize:  symbol.MinSize,
			Contract: symbol.ContractValue,
		})
	}

	if config.ExchangeSymbolList == nil {
		config.ExchangeSymbolList = make(map[string][]config.SymbolInfo)
	}
	config.ExchangeSymbolList[exchangeName] = result

	return result
}
