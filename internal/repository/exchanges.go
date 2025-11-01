package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"gorm.io/gorm"
)

// ListExchanges 列出所有交易所
func ListExchanges() ([]*model.FoxExchange, error) {
	exchanges, err := database.Adapter().FoxExchange.Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return exchanges, nil
}

// SetAllExchangesInactive 将所有交易所置为未激活
func SetAllExchangesInactive() error {
	_, err := database.Adapter().FoxExchange.Where(
		database.Adapter().FoxExchange.IsActive.Eq(1),
	).Update(database.Adapter().FoxExchange.IsActive, 0)
	if err != nil {
		return err
	}
	return nil
}

func GetExchange(name string) (*model.FoxExchange, error) {
	exchange, err := database.Adapter().FoxExchange.Where(
		database.Adapter().FoxExchange.Name.Eq(name),
	).First()
	if (err != nil) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return exchange, nil
}

// ActivateExchange 激活指定交易所
func ActivateExchange(name string) error {
	_, err := database.Adapter().FoxExchange.Where(
		database.Adapter().FoxExchange.Name.Eq(name),
	).Update(database.Adapter().FoxExchange.IsActive, 1)
	return err
}
