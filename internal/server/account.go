package server

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/lemconn/foxflow/internal/config"
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/exchange"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"github.com/lemconn/foxflow/internal/repository"
	pb "github.com/lemconn/foxflow/proto/generated"
	"gorm.io/gorm"
)

type AccountServer struct{}

func NewAccountServer() *AccountServer {
	return &AccountServer{}
}

func (s *AccountServer) GetAccounts(ctx context.Context, req *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	// 从数据库获取账户列表
	accounts, err := database.Adapter().FoxAccount.
		Preload(database.Adapter().FoxAccount.Config).
		Preload(database.Adapter().FoxAccount.TradeConfigs).
		Find()
	if err != nil && !errors.Is(gorm.ErrRecordNotFound, err) {
		log.Printf("获取账户列表失败: %v", err)
		return &pb.GetAccountsResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户列表失败: %v", err),
		}, nil
	}

	// 转换为 protobuf 格式
	var pbAccounts []*pb.AccountsItem
	for _, account := range accounts {
		pbAccounts = append(pbAccounts, buildPBAccountItem(account))
	}

	return &pb.GetAccountsResponse{
		Success:  true,
		Message:  fmt.Sprintf("成功获取 %d 个账户", len(pbAccounts)),
		Accounts: pbAccounts,
	}, nil
}

// UseAccount 激活账户
func (s *AccountServer) UseAccount(ctx context.Context, req *pb.UseAccountRequest) (*pb.UseAccountResponse, error) {
	if req.Account == "" {
		return &pb.UseAccountResponse{
			Success: false,
			Message: "account 是必填参数",
		}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Name.Eq(req.Account),
	).Preload(database.Adapter().FoxAccount.TradeConfigs).First()
	if err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}
	if account == nil || account.ID == 0 {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("account `%s` not found", req.Account),
		}, nil
	}

	if err := repository.SetAllAccountInactive(); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("重置账户状态失败: %v", err),
		}, nil
	}

	if err := repository.SetAllExchangesInactive(); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("重置交易所状态失败: %v", err),
		}, nil
	}

	ex, err := repository.GetExchange(account.Exchange)
	if err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所失败: %v", err),
		}, nil
	}
	if ex == nil || ex.ID == 0 {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("exchange `%s` not found", account.Exchange),
		}, nil
	}

	if err := repository.ActivateExchange(account.Exchange); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("激活交易所失败: %v", err),
		}, nil
	}

	if err := repository.ActivateAccountByName(account.Name); err != nil {
		return &pb.UseAccountResponse{
			Success: false,
			Message: fmt.Sprintf("激活账户失败: %v", err),
		}, nil
	}

	if exchangeClient, err := exchange.GetManager().GetExchange(account.Exchange); err == nil {
		if err := exchangeClient.SetAccount(ctx, account); err != nil {
			log.Printf("设置账户失败: %v", err)
		}
	} else {
		log.Printf("获取交易所客户端失败: %v", err)
	}

	return &pb.UseAccountResponse{
		Success: true,
		Message: fmt.Sprintf("已激活用户: %s", account.Name),
		Account: buildPBAccountItem(account),
		Exchange: &pb.ExchangesItem{
			Name:        ex.Name,
			ApiUrl:      ex.APIURL,
			ProxyUrl:    ex.ProxyURL,
			StatusValue: ex.IsActive,
		},
	}, nil
}

// UpdateTradeConfig 更新账户杠杆配置
func (s *AccountServer) UpdateTradeConfig(ctx context.Context, req *pb.UpdateTradeConfigRequest) (*pb.UpdateTradeConfigResponse, error) {
	if req.AccountId <= 0 {
		return &pb.UpdateTradeConfigResponse{Success: false, Message: "account_id 是必填参数"}, nil
	}
	if req.Margin != "isolated" && req.Margin != "cross" {
		return &pb.UpdateTradeConfigResponse{Success: false, Message: "margin 只能为 isolated 或 cross"}, nil
	}
	if req.Leverage <= 0 {
		return &pb.UpdateTradeConfigResponse{Success: false, Message: "leverage 必须大于 0"}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.UpdateTradeConfigResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}

	tradeConfigDAO := database.Adapter().FoxTradeConfig
	tradeConfig, err := tradeConfigDAO.Where(
		tradeConfigDAO.AccountID.Eq(req.AccountId),
		tradeConfigDAO.Margin.Eq(req.Margin),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.UpdateTradeConfigResponse{
			Success: false,
			Message: fmt.Sprintf("查询杠杆配置失败: %v", err),
		}, nil
	}

	if tradeConfig == nil {
		if err := tradeConfigDAO.Create(&model.FoxTradeConfig{
			AccountID: req.AccountId,
			Margin:    req.Margin,
			Leverage:  req.Leverage,
		}); err != nil {
			return &pb.UpdateTradeConfigResponse{
				Success: false,
				Message: fmt.Sprintf("创建杠杆配置失败: %v", err),
			}, nil
		}
	} else {
		tradeConfig.Leverage = req.Leverage
		if err := tradeConfigDAO.Save(tradeConfig); err != nil {
			return &pb.UpdateTradeConfigResponse{
				Success: false,
				Message: fmt.Sprintf("更新杠杆配置失败: %v", err),
			}, nil
		}
	}

	tradeConfigs, err := tradeConfigDAO.Where(
		tradeConfigDAO.AccountID.Eq(req.AccountId),
	).Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.UpdateTradeConfigResponse{
			Success: false,
			Message: fmt.Sprintf("获取用户杠杆配置失败: %v", err),
		}, nil
	}

	for _, newTradeConfig := range tradeConfigs {
		account.TradeConfigs = append(account.TradeConfigs, *newTradeConfig)
	}

	return &pb.UpdateTradeConfigResponse{
		Success: true,
		Message: fmt.Sprintf("已将 %s 杠杆设置为 %d", req.Margin, req.Leverage),
		Account: buildPBAccountItem(account),
	}, nil
}

// UpdateProxyConfig 更新账户代理设置
func (s *AccountServer) UpdateProxyConfig(ctx context.Context, req *pb.UpdateProxyConfigRequest) (*pb.UpdateProxyConfigResponse, error) {
	if req.AccountId <= 0 {
		return &pb.UpdateProxyConfigResponse{Success: false, Message: "account_id 是必填参数"}, nil
	}

	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(req.AccountId),
	).Preload(database.Adapter().FoxAccount.TradeConfigs).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.UpdateProxyConfigResponse{
			Success: false,
			Message: fmt.Sprintf("获取账户失败: %v", err),
		}, nil
	}

	configDAO := database.Adapter().FoxConfig
	config, err := configDAO.Where(
		configDAO.AccountID.Eq(req.AccountId),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.UpdateProxyConfigResponse{
			Success: false,
			Message: fmt.Sprintf("查询代理配置失败: %v", err),
		}, nil
	}

	if config == nil {
		if req.ProxyUrl != "" {
			if err := configDAO.Create(&model.FoxConfig{
				AccountID: req.AccountId,
				ProxyURL:  req.ProxyUrl,
			}); err != nil {
				return &pb.UpdateProxyConfigResponse{
					Success: false,
					Message: fmt.Sprintf("创建代理配置失败: %v", err),
				}, nil
			}
			account.Config.ProxyURL = req.ProxyUrl
		}
	} else {
		config.ProxyURL = req.ProxyUrl
		if err := configDAO.Save(config); err != nil {
			return &pb.UpdateProxyConfigResponse{
				Success: false,
				Message: fmt.Sprintf("更新代理配置失败: %v", err),
			}, nil
		}
		account.Config.ProxyURL = req.ProxyUrl
	}

	message := "网络代理设置已更新"
	if req.ProxyUrl == "" {
		message = "已清空网络代理设置"
	}

	return &pb.UpdateProxyConfigResponse{
		Success: true,
		Message: message,
		Account: buildPBAccountItem(account),
	}, nil
}

// UpdateAccount 更新账户信息
func (s *AccountServer) UpdateAccount(ctx context.Context, req *pb.UpdateAccountRequest) (*pb.UpdateAccountResponse, error) {
	if req.Exchange == "" {
		return &pb.UpdateAccountResponse{Success: false, Message: "exchange 是必填参数"}, nil
	}
	if req.TargetAccount == "" {
		return &pb.UpdateAccountResponse{Success: false, Message: "target_account 是必填参数"}, nil
	}
	if req.TradeType == "" || req.Name == "" || req.ApiKey == "" || req.SecretKey == "" {
		return &pb.UpdateAccountResponse{Success: false, Message: "缺少必要的账户参数"}, nil
	}
	if req.Exchange == config.DefaultExchange && req.Passphrase == "" {
		return &pb.UpdateAccountResponse{Success: false, Message: "okx 账户必须提供 passphrase"}, nil
	}

	accountInfo, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Exchange.Eq(req.Exchange),
		database.Adapter().FoxAccount.Name.Eq(req.TargetAccount),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err != nil {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("查询账户失败: %v", err),
		}, nil
	}
	if accountInfo == nil || accountInfo.ID == 0 {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: "账户不存在",
		}, nil
	}

	account := &model.FoxAccount{
		ID:         accountInfo.ID,
		Exchange:   accountInfo.Exchange,
		TradeType:  req.TradeType,
		Name:       req.Name,
		AccessKey:  req.ApiKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		IsActive:   accountInfo.IsActive,
	}

	exchangeClient, err := exchange.GetManager().GetExchange(account.Exchange)
	if err != nil {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所客户端失败: %v", err),
		}, nil
	}

	exchangeAccount, err := exchangeClient.GetAccount(ctx)
	if err != nil {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取交易所账户失败: %v", err),
		}, nil
	}

	if err := exchangeClient.Connect(ctx, account); err != nil {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("连接交易所失败: %v", err),
		}, nil
	}

	if err := exchangeClient.SetAccount(ctx, exchangeAccount); err != nil {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("恢复交易所账户失败: %v", err),
		}, nil
	}

	if err := repository.UpdateAccount(account); err != nil {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("更新账户失败: %v", err),
		}, nil
	}

	updatedAccount, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.ID.Eq(account.ID),
	).Preload(database.Adapter().FoxAccount.Config).
		Preload(database.Adapter().FoxAccount.TradeConfigs).
		First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return &pb.UpdateAccountResponse{
			Success: false,
			Message: fmt.Sprintf("获取最新账户信息失败: %v", err),
		}, nil
	}

	return &pb.UpdateAccountResponse{
		Success: true,
		Message: "更新账户成功",
		Account: buildPBAccountItem(updatedAccount),
	}, nil
}

func buildPBAccountItem(account *model.FoxAccount) *pb.AccountsItem {
	// 处理杠杆倍数
	tradeConfigs := make([]model.FoxTradeConfig, 0)
	if len(account.TradeConfigs) == 0 {
		newTradeConfigs, _ := database.Adapter().FoxTradeConfig.Where(
			database.Adapter().FoxTradeConfig.AccountID.Eq(account.ID),
		).Find()
		if newTradeConfigs != nil {
			for _, newTradeConfig := range newTradeConfigs {
				tradeConfigs = append(tradeConfigs, *newTradeConfig)
			}
		}
	} else {
		for _, accountTradeConfig := range account.TradeConfigs {
			tradeConfigs = append(tradeConfigs, accountTradeConfig)
		}
	}

	var crossLeverage, isolatedLeverage int64
	for _, tradeConfig := range tradeConfigs {
		if tradeConfig.Margin == "cross" {
			crossLeverage = tradeConfig.Leverage
		} else if tradeConfig.Margin == "isolated" {
			isolatedLeverage = tradeConfig.Leverage
		}
	}

	// 处理代理地址
	proxyURL := account.Config.ProxyURL
	if proxyURL == "" {
		if cfg, err := database.Adapter().FoxConfig.Where(
			database.Adapter().FoxConfig.AccountID.Eq(account.ID),
		).First(); err == nil && cfg != nil {
			proxyURL = cfg.ProxyURL
		}
	}

	return &pb.AccountsItem{
		Id:               account.ID,
		Name:             account.Name,
		Exchange:         account.Exchange,
		TradeTypeValue:   account.TradeType,
		StatusValue:      account.IsActive,
		CrossLeverage:    crossLeverage,
		IsolatedLeverage: isolatedLeverage,
		ProxyUrl:         proxyURL,
	}
}
