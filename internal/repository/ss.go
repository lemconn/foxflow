package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/gorm"
)

// ListWaitingSSOrders 列出等待中的策略订单，可按用户过滤
func ListWaitingSSOrders(accountID uint) ([]models.FoxSS, error) {
	db := database.GetDB()
	var ss []models.FoxSS
	query := db.Where("status = ?", "waiting")
	if accountID != 0 {
		query = query.Where("account_id = ?", accountID)
	}
	if err := query.Find(&ss).Error; err != nil {
		return nil, err
	}
	return ss, nil
}

func ListSSOrders(accountID uint, status []string) ([]*models.FoxSS, error) {
	var ss []*models.FoxSS

	db := database.GetDB()
	query := db.Where("account_id = ?", accountID)
	if len(status) > 0 {
		query = query.Where("status in ?", status)
	}
	err := query.Find(&ss).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return ss, nil
}

// CreateSSOrder 创建策略订单
func CreateSSOrder(order *models.FoxSS) error {
	db := database.GetDB()
	return db.Create(order).Error
}

// FindSSOrderByIDForUser 查找用户策略订单
func FindSSOrderByIDForUser(accountID uint, id uint64) (*models.FoxSS, error) {
	var order models.FoxSS

	db := database.GetDB()
	err := db.Where("id = ? AND account_id = ?", id, accountID).First(&order).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &order, nil
}

// SaveSSOrder 保存策略订单
func SaveSSOrder(order *models.FoxSS) error {
	db := database.GetDB()
	return db.Save(order).Error
}
