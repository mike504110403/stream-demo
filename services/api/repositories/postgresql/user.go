package postgresql

import (
	"errors"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// CreateUser 創建用戶
func (r *PostgreSQLRepo) CreateUser(user *models.User) error {
	return r.PostgreSQLDB.Create(user).Error
}

// FindUserByID 根據ID查找用戶
func (r *PostgreSQLRepo) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.PostgreSQLDB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByUsername 根據用戶名查找用戶
func (r *PostgreSQLRepo) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.PostgreSQLDB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// FindUserByEmail 根據郵箱查找用戶
func (r *PostgreSQLRepo) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.PostgreSQLDB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用戶
func (r *PostgreSQLRepo) UpdateUser(user *models.User) error {
	return r.PostgreSQLDB.Save(user).Error
}

// DeleteUser 刪除用戶
func (r *PostgreSQLRepo) DeleteUser(id uint) error {
	return r.PostgreSQLDB.Delete(&models.User{}, id).Error
}
