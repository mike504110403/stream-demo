package repo

import (
	"errors"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

func (r *MysqlRepo) CreateUser(user *models.User) error {
	return r.MysqlDB.Create(user).Error
}

func (r *MysqlRepo) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.MysqlDB.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *MysqlRepo) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.MysqlDB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *MysqlRepo) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.MysqlDB.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *MysqlRepo) UpdateUser(user *models.User) error {
	return r.MysqlDB.Save(user).Error
}

func (r *MysqlRepo) DeleteUser(id uint) error {
	return r.MysqlDB.Delete(&models.User{}, id).Error
}
