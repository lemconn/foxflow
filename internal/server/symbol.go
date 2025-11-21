package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/database"
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

// UpdateSymbol 更新标的杠杆/保证金配置
func (s *SymbolServer) UpdateSymbol(ctx context.Context, req *pb.UpdateSymbolRequest) (*pb.UpdateSymbolResponse, error) {
	if req.AccountId <= 0 {
		return &pb.UpdateSymbolResponse{Success: false, Message: "account_id 是必填参数"}, nil
	}
	if req.Symbol == "" {
		return &pb.UpdateSymbolResponse{Success: false, Message: "symbol 是必填参数"}, nil
	}
	if req.Margin != "isolated" && req.Margin != "cross" {
		return &pb.UpdateSymbolResponse{Success: false, Message: "margin 只能为 isolated 或 cross"}, nil
	}
	if req.Leverage <= 0 {
		return &pb.UpdateSymbolResponse{Success: false, Message: "leverage 必须大于 0"}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).Preload(database.Adapter().FoxAccount.Config).First()
	if err != nil {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}

	exchangeName := req.Exchange
	if exchangeName == "" {
		exchangeName = account.Exchange
	}
	if exchangeName != account.Exchange {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: "请求的交易所与账户所属交易所不匹配",
		}, nil
	}

	exchangeClient, err := exchange.GetManager().GetExchange(exchangeName)
	if err != nil {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	if err := exchangeClient.SetAccount(ctx, account); err != nil {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: fmt.Sprintf("设置账户失败: %v", err),
		}, nil
	}

	symbolList := s.getSymbolList(ctx, exchangeName, exchangeClient)
	var symbolInfo config.SymbolInfo
	for _, symbol := range symbolList {
		if symbol.Name == req.Symbol {
			symbolInfo = symbol
			break
		}
	}
	if symbolInfo.Name == "" {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: fmt.Sprintf("symbol %s 不存在", req.Symbol),
		}, nil
	}
	if req.Leverage > symbolInfo.MaxLever {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: fmt.Sprintf("杠杆倍数过大，最大可用杠杆为 %d", symbolInfo.MaxLever),
		}, nil
	}

	if err := exchangeClient.SetLeverage(ctx, req.Symbol, req.Leverage, req.Margin); err != nil {
		return &pb.UpdateSymbolResponse{
			Success: false,
			Message: fmt.Sprintf("设置杠杆失败: %v", err),
		}, nil
	}

	return &pb.UpdateSymbolResponse{
		Success: true,
		Message: fmt.Sprintf("更新标的杠杆成功: %s %s %d", req.Symbol, req.Margin, req.Leverage),
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
