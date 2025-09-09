package repository

import (
	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
)

// CreateSymbol 创建交易标的
func CreateSymbol(symbol *models.FoxSymbol) error {
	db := database.GetDB()
	return db.Create(symbol).Error
}

// DeleteSymbolByNameForUser 删除用户下的标的
func DeleteSymbolByNameForUser(userID uint, name string) error {
	db := database.GetDB()
	return db.Where("name = ? AND user_id = ?", name, userID).Delete(&models.FoxSymbol{}).Error
}
