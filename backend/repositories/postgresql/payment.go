package postgresql

import (
	"errors"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// CreatePayment 創建支付記錄
func (r *PostgreSQLRepo) CreatePayment(payment *models.Payment) error {
	return r.PostgreSQLDB.Create(payment).Error
}

// FindPaymentByID 根據ID查找支付記錄
func (r *PostgreSQLRepo) FindPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	if err := r.PostgreSQLDB.Preload("User").First(&payment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// FindPaymentByUserID 根據用戶ID查找支付記錄列表
func (r *PostgreSQLRepo) FindPaymentByUserID(userID uint) ([]models.Payment, error) {
	var payments []models.Payment
	if err := r.PostgreSQLDB.Where("user_id = ?", userID).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// FindPaymentByTransactionID 根據交易ID查找支付記錄
func (r *PostgreSQLRepo) FindPaymentByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	if err := r.PostgreSQLDB.Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

// UpdatePayment 更新支付記錄
func (r *PostgreSQLRepo) UpdatePayment(payment *models.Payment) error {
	return r.PostgreSQLDB.Save(payment).Error
}

// DeletePayment 刪除支付記錄
func (r *PostgreSQLRepo) DeletePayment(id uint) error {
	return r.PostgreSQLDB.Delete(&models.Payment{}, id).Error
}

// ListPayment 分頁列出支付記錄
func (r *PostgreSQLRepo) ListPayment(offset, limit int) ([]*models.Payment, int64, error) {
	var payments []*models.Payment
	var total int64

	if err := r.PostgreSQLDB.Model(&models.Payment{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.PostgreSQLDB.Offset(offset).Limit(limit).Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, total, nil
}
