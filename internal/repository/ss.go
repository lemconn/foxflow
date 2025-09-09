package repository

import (
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
)

// ListWaitingSSOrders 列出等待中的策略订单，可按用户过滤
func ListWaitingSSOrders(userID *uint) ([]models.FoxSS, error) {
	db := database.GetDB()
	var ss []models.FoxSS
	query := db.Where("status = ?", "waiting")
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if err := query.Find(&ss).Error; err != nil {
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
func FindSSOrderByIDForUser(userID uint, id uint64) (*models.FoxSS, error) {
	db := database.GetDB()
	var order models.FoxSS
	if err := db.Where("id = ? AND user_id = ?", id, userID).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// SaveSSOrder 保存策略订单
func SaveSSOrder(order *models.FoxSS) error {
	db := database.GetDB()
	return db.Save(order).Error
}
