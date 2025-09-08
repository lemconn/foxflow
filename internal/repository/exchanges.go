package repository

import (
	"foxflow/internal/database"
	"foxflow/internal/models"
)

// ListExchanges 列出所有交易所
func ListExchanges() ([]models.FoxExchange, error) {
	db := database.GetDB()
	var exchanges []models.FoxExchange
	if err := db.Find(&exchanges).Error; err != nil {
		return nil, err
	}
	return exchanges, nil
}

// SetAllExchangesInactive 将所有交易所置为未激活
func SetAllExchangesInactive() error {
	db := database.GetDB()
	return db.Model(&models.FoxExchange{}).Where("1 = 1").Update("is_active", false).Error
}

// ActivateExchange 激活指定交易所
func ActivateExchange(name string) error {
	db := database.GetDB()
	return db.Model(&models.FoxExchange{}).Where("name = ?", name).Update("is_active", true).Error
}
