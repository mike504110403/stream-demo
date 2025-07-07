package tests

import (
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"
	"stream-demo/backend/tests/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ================================
// 🆕 使用測試工具包的改進版測試
// ================================

func TestPaymentService_CreatePayment_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功建立支付", func(t *testing.T) {
		// 改進後：簡化設置
		testUser := &models.User{ID: 1, Username: "testuser"}
		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		// 設置支付創建期望
		builder.PaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil)
		service := builder.BuildPaymentService()

		paymentDTO := &dto.PaymentCreateDTO{
			Amount:        100.00,
			Currency:      "TWD",
			Description:   "測試支付",
			PaymentMethod: "credit_card",
		}

		// Act
		payment, err := service.CreatePayment(1, paymentDTO)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, uint(1), payment.UserID)
		assert.Equal(t, paymentDTO.Amount, payment.Amount)
		assert.Equal(t, paymentDTO.Currency, payment.Currency)
		assert.Equal(t, "pending", payment.Status)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：用戶不存在", func(t *testing.T) {
		// 改進後：使用工具包便利方法
		builder := testutils.NewServiceBuilder(t).
			WithUserNotFound(999)

		service := builder.BuildPaymentService()

		paymentDTO := &dto.PaymentCreateDTO{
			Amount:        100.00,
			Currency:      "TWD",
			Description:   "測試支付",
			PaymentMethod: "credit_card",
		}

		// Act
		payment, err := service.CreatePayment(999, paymentDTO)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, payment)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：儲存庫錯誤", func(t *testing.T) {
		// 改進後：鏈式設置錯誤情況
		testUser := &models.User{ID: 1, Username: "testuser"}
		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		builder.PaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(assert.AnError)
		service := builder.BuildPaymentService()

		paymentDTO := &dto.PaymentCreateDTO{
			Amount:        100.00,
			Currency:      "TWD",
			Description:   "測試支付",
			PaymentMethod: "credit_card",
		}

		// Act
		payment, err := service.CreatePayment(1, paymentDTO)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, payment)

		builder.AssertAllExpectations()
	})
}

func TestPaymentService_CompletePayment_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功完成支付", func(t *testing.T) {
		// 改進後：預設置支付和用戶
		testUser := &models.User{ID: 1, Username: "testuser"}
		testPayment := &models.Payment{
			ID:     1,
			UserID: 1,
			Status: "pending",
			Amount: 100.00,
		}

		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		builder.PaymentRepo.On("FindByID", uint(1)).Return(testPayment, nil)
		builder.PaymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil)

		service := builder.BuildPaymentService()

		// Act
		payment, err := service.CompletePayment(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, "completed", payment.Status)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：支付不存在", func(t *testing.T) {
		// 改進後：簡化錯誤設置
		builder := testutils.NewServiceBuilder(t)
		builder.PaymentRepo.On("FindByID", uint(999)).Return((*models.Payment)(nil), assert.AnError)

		service := builder.BuildPaymentService()

		// Act
		payment, err := service.CompletePayment(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, payment)

		builder.AssertAllExpectations()
	})
}

func TestPaymentService_RefundPayment_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功退款", func(t *testing.T) {
		// 改進後：預設置完整場景
		testUser := &models.User{ID: 1, Username: "testuser"}
		testPayment := &models.Payment{
			ID:     1,
			UserID: 1,
			Status: "completed",
			Amount: 100.00,
		}

		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		builder.PaymentRepo.On("FindByID", uint(1)).Return(testPayment, nil)
		builder.PaymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil)

		service := builder.BuildPaymentService()

		refundDTO := &dto.PaymentRefundDTO{
			Reason: "用戶要求退款",
		}

		// Act
		payment, err := service.RefundPayment(1, refundDTO)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, "refunded", payment.Status)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：支付不存在", func(t *testing.T) {
		// 改進後：一行設置錯誤
		builder := testutils.NewServiceBuilder(t)
		builder.PaymentRepo.On("FindByID", uint(999)).Return((*models.Payment)(nil), assert.AnError)

		service := builder.BuildPaymentService()

		refundDTO := &dto.PaymentRefundDTO{
			Reason: "用戶要求退款",
		}

		// Act
		payment, err := service.RefundPayment(999, refundDTO)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, payment)

		builder.AssertAllExpectations()
	})
}

// ================================
// 🔄 原有測試保留（向後兼容）
// ================================

// MockPaymentRepository 模擬支付儲存庫
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) FindByID(id uint) (*models.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByUserID(userID uint) ([]models.Payment, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) FindByTransactionID(transactionID string) (*models.Payment, error) {
	args := m.Called(transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(payment *models.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockUserRepositoryForPayment for PaymentService
type MockUserRepositoryForPayment struct {
	mock.Mock
}

func (m *MockUserRepositoryForPayment) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForPayment) FindByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForPayment) FindByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForPayment) FindByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForPayment) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepositoryForPayment) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestPaymentService_CreatePayment(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		paymentDTO *dto.PaymentCreateDTO
		mockSetup  func(*MockPaymentRepository, *MockUserRepositoryForPayment)
		wantErr    bool
	}{
		{
			name:   "成功建立支付",
			userID: 1,
			paymentDTO: &dto.PaymentCreateDTO{
				Amount:        100.00,
				Currency:      "TWD",
				Description:   "測試支付",
				PaymentMethod: "credit_card",
			},
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockPaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "用戶不存在",
			userID: 999,
			paymentDTO: &dto.PaymentCreateDTO{
				Amount:        100.00,
				Currency:      "TWD",
				Description:   "測試支付",
				PaymentMethod: "credit_card",
			},
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockUserRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name:   "儲存庫錯誤",
			userID: 1,
			paymentDTO: &dto.PaymentCreateDTO{
				Amount:        100.00,
				Currency:      "TWD",
				Description:   "測試支付",
				PaymentMethod: "credit_card",
			},
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockPaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(MockPaymentRepository)
			mockUserRepo := new(MockUserRepositoryForPayment)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewPaymentService(cfg)
			tt.mockSetup(mockPaymentRepo, mockUserRepo)

			payment, err := service.CreatePayment(tt.userID, tt.paymentDTO)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, tt.userID, payment.UserID)
				assert.Equal(t, tt.paymentDTO.Amount, payment.Amount)
				assert.Equal(t, tt.paymentDTO.Currency, payment.Currency)
				assert.Equal(t, tt.paymentDTO.Description, payment.Description)
				assert.Equal(t, "pending", payment.Status)
			}
			mockPaymentRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_CompletePayment(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		mockSetup func(*MockPaymentRepository, *MockUserRepositoryForPayment)
		wantErr   bool
	}{
		{
			name: "成功完成支付",
			id:   1,
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockPaymentRepo.On("FindByID", uint(1)).Return(&models.Payment{
					ID:     1,
					UserID: 1,
					Status: "pending",
				}, nil)
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockPaymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "支付不存在",
			id:   999,
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockPaymentRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(MockPaymentRepository)
			mockUserRepo := new(MockUserRepositoryForPayment)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewPaymentService(cfg)
			tt.mockSetup(mockPaymentRepo, mockUserRepo)

			payment, err := service.CompletePayment(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, "completed", payment.Status)
			}
			mockPaymentRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestPaymentService_RefundPayment(t *testing.T) {
	tests := []struct {
		name      string
		id        uint
		refundDTO *dto.PaymentRefundDTO
		mockSetup func(*MockPaymentRepository, *MockUserRepositoryForPayment)
		wantErr   bool
	}{
		{
			name: "成功退款",
			id:   1,
			refundDTO: &dto.PaymentRefundDTO{
				Reason: "用戶要求退款",
			},
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockPaymentRepo.On("FindByID", uint(1)).Return(&models.Payment{
					ID:     1,
					UserID: 1,
					Status: "completed",
				}, nil)
				mockUserRepo.On("FindByID", uint(1)).Return(&models.User{ID: 1, Username: "testuser"}, nil)
				mockPaymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "支付不存在",
			id:   999,
			refundDTO: &dto.PaymentRefundDTO{
				Reason: "用戶要求退款",
			},
			mockSetup: func(mockPaymentRepo *MockPaymentRepository, mockUserRepo *MockUserRepositoryForPayment) {
				mockPaymentRepo.On("FindByID", uint(999)).Return(nil, assert.AnError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPaymentRepo := new(MockPaymentRepository)
			mockUserRepo := new(MockUserRepositoryForPayment)
			cfg := config.NewPostgreSQLConfig("config.yaml", "local")
			service := services.NewPaymentService(cfg)
			tt.mockSetup(mockPaymentRepo, mockUserRepo)

			payment, err := service.RefundPayment(tt.id, tt.refundDTO)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, payment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, payment)
				assert.Equal(t, "refunded", payment.Status)
			}
			mockPaymentRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
		})
	}
}
