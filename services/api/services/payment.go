package services

import (
	"errors"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	dto "stream-demo/backend/dto"
	postgresqlRepo "stream-demo/backend/repositories/postgresql"
	"time"

	"github.com/google/uuid"
)

// PaymentService 支付服務
type PaymentService struct {
	Conf      *config.Config
	Repo      *postgresqlRepo.PostgreSQLRepo
	RepoSlave *postgresqlRepo.PostgreSQLRepo
}

// NewPaymentService 創建支付服務實例
func NewPaymentService(conf *config.Config) *PaymentService {
	return &PaymentService{
		Conf:      conf,
		Repo:      postgresqlRepo.NewPostgreSQLRepo(conf.DB["master"]),
		RepoSlave: postgresqlRepo.NewPostgreSQLRepo(conf.DB["slave"]),
	}
}

// CreatePayment 創建支付
func (s *PaymentService) CreatePayment(userID uint, createDTO *dto.PaymentCreateDTO) (*dto.PaymentDTO, error) {
	// 檢查用戶是否存在
	user, err := s.RepoSlave.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("用戶不存在")
	}

	// 生成交易 ID
	transactionID := uuid.New().String()

	// 創建支付
	payment := &models.Payment{
		UserID:        userID,
		Amount:        createDTO.Amount,
		Currency:      createDTO.Currency,
		Status:        "pending",
		PaymentMethod: createDTO.PaymentMethod,
		TransactionID: transactionID,
		Description:   createDTO.Description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.Repo.CreatePayment(payment); err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.PaymentDTO{
		ID:            payment.ID,
		UserID:        payment.UserID,
		Username:      user.Username,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        payment.Status,
		PaymentMethod: payment.PaymentMethod,
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		RefundReason:  payment.RefundReason,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}

// GetPaymentByID 根據 ID 獲取支付
func (s *PaymentService) GetPaymentByID(id uint) (*dto.PaymentDTO, error) {
	payment, err := s.RepoSlave.FindPaymentByID(id)
	if err != nil {
		return nil, err
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(payment.UserID)
	if err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.PaymentDTO{
		ID:            payment.ID,
		UserID:        payment.UserID,
		Username:      user.Username,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        payment.Status,
		PaymentMethod: payment.PaymentMethod,
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		RefundReason:  payment.RefundReason,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}

// GetPaymentsByUserID 根據用戶 ID 獲取支付列表
func (s *PaymentService) GetPaymentsByUserID(userID uint) (*dto.PaymentListDTO, error) {
	payments, err := s.RepoSlave.FindPaymentByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	// 轉換為 DTO
	paymentDTOs := make([]dto.PaymentDTO, len(payments))
	for i, payment := range payments {
		paymentDTOs[i] = dto.PaymentDTO{
			ID:            payment.ID,
			UserID:        payment.UserID,
			Username:      user.Username,
			Amount:        payment.Amount,
			Currency:      payment.Currency,
			Status:        payment.Status,
			PaymentMethod: payment.PaymentMethod,
			TransactionID: payment.TransactionID,
			Description:   payment.Description,
			RefundReason:  payment.RefundReason,
			CreatedAt:     payment.CreatedAt,
			UpdatedAt:     payment.UpdatedAt,
		}
	}

	return &dto.PaymentListDTO{
		Total:    int64(len(payments)),
		Payments: paymentDTOs,
	}, nil
}

// CompletePayment 完成支付
func (s *PaymentService) CompletePayment(id uint) (*dto.PaymentDTO, error) {
	payment, err := s.Repo.FindPaymentByID(id)
	if err != nil {
		return nil, err
	}

	if payment.Status != "pending" {
		return nil, errors.New("支付狀態不正確")
	}

	payment.Status = "completed"
	payment.UpdatedAt = time.Now()

	if err := s.Repo.UpdatePayment(payment); err != nil {
		return nil, err
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(payment.UserID)
	if err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.PaymentDTO{
		ID:            payment.ID,
		UserID:        payment.UserID,
		Username:      user.Username,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        payment.Status,
		PaymentMethod: payment.PaymentMethod,
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		RefundReason:  payment.RefundReason,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}

// RefundPayment 退款
func (s *PaymentService) RefundPayment(id uint, paymentRefund *dto.PaymentRefundDTO) (*dto.PaymentDTO, error) {
	payment, err := s.Repo.FindPaymentByID(id)
	if err != nil {
		return nil, err
	}

	if payment.Status != "completed" {
		return nil, errors.New("支付狀態不正確")
	}

	payment.Status = "refunded"
	payment.RefundReason = paymentRefund.Reason
	payment.UpdatedAt = time.Now()

	if err := s.Repo.UpdatePayment(payment); err != nil {
		return nil, err
	}

	// 獲取用戶資訊
	user, err := s.RepoSlave.FindUserByID(payment.UserID)
	if err != nil {
		return nil, err
	}

	// 轉換為 DTO
	return &dto.PaymentDTO{
		ID:            payment.ID,
		UserID:        payment.UserID,
		Username:      user.Username,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		Status:        payment.Status,
		PaymentMethod: payment.PaymentMethod,
		TransactionID: payment.TransactionID,
		Description:   payment.Description,
		RefundReason:  payment.RefundReason,
		CreatedAt:     payment.CreatedAt,
		UpdatedAt:     payment.UpdatedAt,
	}, nil
}
