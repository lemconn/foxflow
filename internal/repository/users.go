package repository

import (
	"errors"

	"github.com/lemconn/foxflow/internal/database"
	"github.com/lemconn/foxflow/internal/models"
	"gorm.io/gorm"
)

// ListAccount 列出所有用户
func ListAccount() ([]models.FoxAccount, error) {
	db := database.GetDB()
	var users []models.FoxAccount

	err := db.Find(&users).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return users, nil
}

func ExchangeAccountList(exchangeName string) ([]models.FoxAccount, error) {
	db := database.GetDB()
	var users []models.FoxAccount

	err := db.Where("exchange = ?", exchangeName).Find(&users).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return users, nil
}

// FindAccountByName 根据用户名查找用户
func FindAccountByName(name string) (*models.FoxAccount, error) {
	var user models.FoxAccount

	err := database.GetDB().Where("name = ?", name).Find(&user).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &user, nil
}

// FindAccountByExchangeName 根据交易所+用户名查找用户
func FindAccountByExchangeName(exchange, name string) (*models.FoxAccount, error) {
	var user models.FoxAccount

	err := database.GetDB().Where("exchange = ?", exchange).Where("name = ?", name).Find(&user).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return &user, nil
}

// CreateAccount 创建用户
func CreateAccount(user *models.FoxAccount) error {
	return database.GetDB().Create(user).Error
}

// DeleteAccountByName 删除用户
func DeleteAccountByName(name string) error {
	return database.GetDB().Where("name = ?", name).Delete(&models.FoxAccount{}).Error
}

// SetAllAccountInactive 将所有用户置为未激活
func SetAllAccountInactive() error {
	db := database.GetDB()
	return db.Model(&models.FoxAccount{}).Where("1 = 1").Update("is_active", false).Error
}

// ActivateAccountByName 激活指定用户
func ActivateAccountByName(name string) error {
	db := database.GetDB()
	return db.Model(&models.FoxAccount{}).Where("name = ?", name).Update("is_active", true).Error
}

// UpdateAccount 更行账户
func UpdateAccount(account *models.FoxAccount) error {
	db := database.GetDB()
	return db.Save(account).Error
}
