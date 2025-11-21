package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"gorm.io/gorm"
)

func ActiveAccount() (*model.FoxAccount, error) {
	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.IsActive.Eq(1),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return account, nil
}

func ExchangeAccountList(exchangeName string) ([]*model.FoxAccount, error) {
	accounts, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Exchange.Eq(exchangeName),
	).Preload(database.Adapter().FoxAccount.Config).
		Preload(database.Adapter().FoxAccount.TradeConfigs).
		Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return accounts, nil
}

// FindAccountByName 根据用户名查找用户
func FindAccountByName(name string) (*model.FoxAccount, error) {
	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Name.Eq(name),
	).Preload(database.Adapter().FoxAccount.TradeConfigs).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return account, nil
}

// FindAccountByExchangeName 根据交易所+用户名查找用户
func FindAccountByExchangeName(exchange, name string) (*model.FoxAccount, error) {
	account, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Exchange.Eq(exchange),
		database.Adapter().FoxAccount.Name.Eq(name),
	).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return account, nil
}

// CreateAccount 创建用户
func CreateAccount(account *model.FoxAccount) error {
	return database.Adapter().FoxAccount.Create(account)
}

// DeleteAccountByName 删除用户
func DeleteAccountByName(name string) error {
	_, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Name.Eq(name),
	).Delete()
	return err
}

// SetAllAccountInactive 将所有用户置为未激活
func SetAllAccountInactive() error {
	_, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.IsActive.Eq(1),
	).Update(database.Adapter().FoxAccount.IsActive, 0)
	return err
}

// ActivateAccountByName 激活指定用户
func ActivateAccountByName(name string) error {
	_, err := database.Adapter().FoxAccount.Where(
		database.Adapter().FoxAccount.Name.Eq(name),
	).Update(database.Adapter().FoxAccount.IsActive, 1)
	return err
}

// UpdateAccount 更行账户
func UpdateAccount(account *model.FoxAccount) error {
	if err := database.Adapter().FoxAccount.Save(account); err != nil {
		return err
	}
	return nil
}
