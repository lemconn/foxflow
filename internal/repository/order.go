package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/pkg/dao/model"
	"gorm.io/gorm"
)

func ListSSOrders(accountID int64, status []string) ([]*model.FoxOrder, error) {
	tx := database.Adapter().FoxOrder.Where(database.Adapter().FoxOrder.AccountID.Eq(accountID))

	if len(status) > 0 {
		tx = tx.Where(database.Adapter().FoxOrder.Status.In(status...))
	}

	orders, err := tx.Order(database.Adapter().FoxOrder.ID.Desc()).Find()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return orders, nil
}
