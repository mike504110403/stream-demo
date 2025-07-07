package repo

import (
	"errors"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

func (r *MysqlRepo) CreatePayment(payment *models.Payment) error {
	return r.MysqlDB.Create(payment).Error
}

func (r *MysqlRepo) FindPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	if err := r.MysqlDB.Preload("User").First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *MysqlRepo) FindPaymentByUserID(userID uint) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.MysqlDB.Where("user_id = ?", userID).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *MysqlRepo) FindPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.MysqlDB.Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (r *MysqlRepo) UpdatePayment(payment *models.Payment) error {
	return r.MysqlDB.Save(payment).Error
}

func (r *MysqlRepo) DeletePayment(id uint) error {
	return r.MysqlDB.Delete(&models.Payment{}, id).Error
}

func (r *MysqlRepo) ListPayment(offset, limit int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	if err := r.MysqlDB.Model(&models.Payment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.MysqlDB.Offset(offset).Limit(limit).Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}
