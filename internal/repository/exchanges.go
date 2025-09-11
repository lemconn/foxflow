package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/gorm"
)

// ListExchanges 列出所有交易所
func ListExchanges() ([]*models.FoxExchange, error) {
	var exchanges []*models.FoxExchange

	db := database.GetDB()
	err := db.Find(&exchanges).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return exchanges, nil
}

// SetAllExchangesInactive 将所有交易所置为未激活
func SetAllExchangesInactive() error {
	db := database.GetDB()
	return db.Model(&models.FoxExchange{}).Update("is_active", false).Error
}

func GetExchange(name string) (*models.FoxExchange, error) {
	var exchange *models.FoxExchange

	db := database.GetDB()
	err := db.Where("name = ?", name).First(&exchange).Error
	if (err != nil) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return exchange, nil
}

// ActivateExchange 激活指定交易所
func ActivateExchange(name string) error {
	db := database.GetDB()
	return db.Model(&models.FoxExchange{}).Where("name = ?", name).Update("is_active", true).Error
}
