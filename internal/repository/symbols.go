package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/gorm"
)

// CreateSymbol 创建交易标的
func CreateSymbol(symbol *models.FoxSymbol) error {
	db := database.GetDB()
	return db.Create(symbol).Error
}

// GetSymbolByNameUser 根据交易多和用户ID获取交易对信息
func GetSymbolByNameUser(name string, userID uint) (*models.FoxSymbol, error) {
	db := database.GetDB()

	symbol := &models.FoxSymbol{}
	err := db.Where("name = ? AND user_id = ?", name, userID).First(symbol).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return symbol, nil
}

// GetSymbolByUser 根据用户查询标的列表
func GetSymbolByUser(userID uint) ([]*models.FoxSymbol, error) {
	db := database.GetDB()
	symbolList := make([]*models.FoxSymbol, 0)
	err := db.Where("user_id = ?", userID).Find(&symbolList).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return symbolList, nil
}

// DeleteSymbolByNameForUser 删除用户下的标的
func DeleteSymbolByNameForUser(userID uint, name string) error {
	db := database.GetDB()
	return db.Where("name = ? AND user_id = ?", name, userID).Delete(&models.FoxSymbol{}).Error
}

// UpdateSymbol 更新标的
func UpdateSymbol(symbol *models.FoxSymbol) error {
	db := database.GetDB()
	return db.Save(symbol).Error
}
