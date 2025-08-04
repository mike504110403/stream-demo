package repositories

import (
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// PaymentRepository 支付資料庫操作介面
type PaymentRepository interface {
	Create(payment *models.Payment) error
	FindByID(id uint) (*models.Payment, error)
	FindByUserID(userID uint) ([]models.Payment, error)
	FindByTransactionID(transactionID string) (*models.Payment, error)
	Update(payment *models.Payment) error
	Delete(id uint) error
}

// paymentRepository 支付資料庫操作實現
type paymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository 創建支付資料庫操作實例
func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

// Create 創建支付
func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// FindByID 根據 ID 查找支付
func (r *paymentRepository) FindByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// FindByUserID 根據用戶 ID 查找支付
func (r *paymentRepository) FindByUserID(userID uint) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("user_id = ?", userID).Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
}

// FindByTransactionID 根據交易 ID 查找支付
func (r *paymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("transaction_id = ?", transactionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// Update 更新支付
func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

// Delete 刪除支付
func (r *paymentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Payment{}, id).Error
}
