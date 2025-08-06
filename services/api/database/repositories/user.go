package repositories

import (
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// UserRepository 用戶資料庫操作介面
type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
}

// userRepository 用戶資料庫操作實現
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 創建用戶資料庫操作實例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create 創建用戶
func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByID 根據 ID 查找用戶
func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根據 Email 查找用戶
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根據用戶名查找用戶
func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用戶
func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// Delete 刪除用戶
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}
