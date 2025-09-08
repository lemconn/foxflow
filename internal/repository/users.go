package repository

import (
	"foxflow/internal/database"
	"foxflow/internal/models"
)

// ListUsers 列出所有用户
func ListUsers() ([]models.FoxUser, error) {
	db := database.GetDB()
	var users []models.FoxUser
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindUserByUsername 根据用户名查找用户
func FindUserByUsername(username string) (*models.FoxUser, error) {
	db := database.GetDB()
	var user models.FoxUser
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func CreateUser(user *models.FoxUser) error {
	db := database.GetDB()
	return db.Create(user).Error
}

// DeleteUserByUsername 删除用户
func DeleteUserByUsername(username string) error {
	db := database.GetDB()
	return db.Where("username = ?", username).Delete(&models.FoxUser{}).Error
}

// SetAllUsersInactive 将所有用户置为未激活
func SetAllUsersInactive() error {
	db := database.GetDB()
	return db.Model(&models.FoxUser{}).Where("1 = 1").Update("is_active", false).Error
}

// ActivateUserByUsername 激活指定用户
func ActivateUserByUsername(username string) error {
	db := database.GetDB()
	return db.Model(&models.FoxUser{}).Where("username = ?", username).Update("is_active", true).Error
}
