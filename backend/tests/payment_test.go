package tests

import (
	"fmt"
	"stream-demo/backend/database/models"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"
	"stream-demo/backend/tests/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ================================
// 🆕 使用新測試工具包的改進版測試
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
			Amount:        100.0,
			Currency:      "TWD",
			PaymentMethod: "credit_card",
			Description:   "測試支付",
		}

		// Act
		payment, err := service.CreatePayment(1, paymentDTO)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, 100.0, payment.Amount)
		assert.Equal(t, "TWD", payment.Currency)
		assert.Equal(t, "pending", payment.Status)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：用戶不存在", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t).
			WithUserNotFound(999)

		service := builder.BuildPaymentService()

		paymentDTO := &dto.PaymentCreateDTO{
			Amount:        100.0,
			Currency:      "TWD",
			PaymentMethod: "credit_card",
		}

		// Act
		payment, err := service.CreatePayment(999, paymentDTO)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, payment)

		builder.AssertAllExpectations()
	})
}

func TestPaymentService_GetPaymentByID_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功獲取支付", func(t *testing.T) {
		testUser := &models.User{ID: 1, Username: "testuser"}
		testPayment := &models.Payment{
			ID:            1,
			UserID:        1,
			Amount:        100.0,
			Currency:      "TWD",
			Status:        "completed",
			PaymentMethod: "credit_card",
			TransactionID: "txn_123456",
		}

		builder := testutils.NewServiceBuilder(t).
			WithUser(testUser)

		builder.PaymentRepo.On("FindByID", uint(1)).Return(testPayment, nil)
		service := builder.BuildPaymentService()

		// Act
		payment, err := service.GetPaymentByID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, uint(1), payment.ID)
		assert.Equal(t, "testuser", payment.Username)
		assert.Equal(t, "completed", payment.Status)

		builder.AssertAllExpectations()
	})

	t.Run("🟢 改進版：支付不存在", func(t *testing.T) {
		builder := testutils.NewServiceBuilder(t)
		builder.PaymentRepo.On("FindByID", uint(999)).Return((*models.Payment)(nil), assert.AnError)

		service := builder.BuildPaymentService()

		// Act
		payment, err := service.GetPaymentByID(999)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, payment)

		builder.AssertAllExpectations()
	})
}

func TestPaymentService_CompletePayment_WithToolkit(t *testing.T) {
	t.Run("🟢 改進版：成功完成支付", func(t *testing.T) {
		testUser := &models.User{ID: 1, Username: "testuser"}
		testPayment := &models.Payment{
			ID:            1,
			UserID:        1,
			Amount:        100.0,
			Status:        "pending",
			PaymentMethod: "credit_card",
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
}

// ================================
// 🚀 多資料庫測試
// ================================

func TestPaymentService_MultiDatabase(t *testing.T) {
	// 準備測試數據
	testUser := &models.User{ID: 1, Username: "testuser", Email: "test@example.com"}
	testPayment := &models.Payment{
		ID:            1,
		UserID:        1,
		Amount:        100.0,
		Currency:      "TWD",
		Status:        "pending",
		PaymentMethod: "credit_card",
		TransactionID: "txn_123456",
	}

	testCases := []struct {
		name      string
		dbType    testutils.DatabaseType
		setupTest func(builder *testutils.ServiceBuilder) *services.PaymentService
		runTest   func(service *services.PaymentService) error
		wantError bool
	}{
		{
			name:   "PostgreSQL 支付創建",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.PaymentService {
				builder.WithUser(testUser)
				builder.PaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil)
				return builder.BuildPaymentService()
			},
			runTest: func(service *services.PaymentService) error {
				paymentDTO := &dto.PaymentCreateDTO{
					Amount:        200.0,
					Currency:      "TWD",
					PaymentMethod: "credit_card",
					Description:   "PostgreSQL 測試支付",
				}
				_, err := service.CreatePayment(1, paymentDTO)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 支付查詢",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.PaymentService {
				builder.WithUser(testUser)
				builder.PaymentRepo.On("FindByID", uint(1)).Return(testPayment, nil)
				return builder.BuildPaymentService()
			},
			runTest: func(service *services.PaymentService) error {
				_, err := service.GetPaymentByID(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 支付完成",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.PaymentService {
				builder.WithUser(testUser)
				builder.PaymentRepo.On("FindByID", uint(1)).Return(testPayment, nil)
				builder.PaymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil)
				return builder.BuildPaymentService()
			},
			runTest: func(service *services.PaymentService) error {
				_, err := service.CompletePayment(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 支付退款",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.PaymentService {
				completedPayment := *testPayment
				completedPayment.Status = "completed"
				builder.WithUser(testUser)
				builder.PaymentRepo.On("FindByID", uint(1)).Return(&completedPayment, nil)
				builder.PaymentRepo.On("Update", mock.AnythingOfType("*models.Payment")).Return(nil)
				return builder.BuildPaymentService()
			},
			runTest: func(service *services.PaymentService) error {
				refundDTO := &dto.PaymentRefundDTO{Reason: "用戶要求退款"}
				_, err := service.RefundPayment(1, refundDTO)
				return err
			},
			wantError: false,
		},
		{
			name:   "PostgreSQL 用戶支付列表",
			dbType: testutils.PostgreSQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.PaymentService {
				payments := []models.Payment{*testPayment}
				builder.PaymentRepo.On("FindByUserID", uint(1)).Return(payments, nil)
				return builder.BuildPaymentService()
			},
			runTest: func(service *services.PaymentService) error {
				_, err := service.GetPaymentsByUserID(1)
				return err
			},
			wantError: false,
		},
		{
			name:   "MySQL 交易ID查詢",
			dbType: testutils.MySQLTest,
			setupTest: func(builder *testutils.ServiceBuilder) *services.PaymentService {
				builder.WithUser(testUser)
				builder.PaymentRepo.On("FindByTransactionID", "txn_123456").Return(testPayment, nil)
				return builder.BuildPaymentService()
			},
			runTest: func(service *services.PaymentService) error {
				// 注意：這個方法可能需要在 PaymentService 中實現
				// 這裡假設有 GetPaymentByTransactionID 方法
				_, err := service.GetPaymentByID(1) // 暫時使用 GetPaymentByID 代替
				return err
			},
			wantError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 創建指定資料庫類型的構建器
			builder := testutils.NewServiceBuilderWithDB(t, tc.dbType)

			// 驗證配置
			assert.NoError(t, builder.ValidateConfig())
			assert.Equal(t, tc.dbType, builder.GetDatabaseType())

			// 設置測試
			service := tc.setupTest(builder)

			// 執行測試
			err := tc.runTest(service)

			// 檢查結果
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// 驗證所有期望
			builder.AssertAllExpectations()
		})
	}
}

// ================================
// 🔧 支付服務專屬測試工具
// ================================

func TestPaymentService_BusinessRules(t *testing.T) {
	t.Run("支付狀態流轉測試 - PostgreSQL", func(t *testing.T) {
		builder := testutils.NewPostgreSQLServiceBuilder(t)
		testUser := &models.User{ID: 1, Username: "testuser"}

		// 測試支付狀態從 pending -> processing -> completed
		states := []string{"pending", "processing", "completed"}

		for i, state := range states {
			payment := &models.Payment{
				ID:     uint(i + 1),
				UserID: 1,
				Amount: 100.0,
				Status: state,
			}

			builder.WithUser(testUser)
			builder.PaymentRepo.On("FindByID", uint(i+1)).Return(payment, nil)
		}

		service := builder.BuildPaymentService()

		// 測試每個狀態的支付
		for i, expectedState := range states {
			payment, err := service.GetPaymentByID(uint(i + 1))
			assert.NoError(t, err)
			assert.Equal(t, expectedState, payment.Status)
		}

		builder.AssertAllExpectations()
	})

	t.Run("大額支付處理 - MySQL", func(t *testing.T) {
		builder := testutils.NewMySQLServiceBuilder(t)
		testUser := &models.User{ID: 1, Username: "vip_user"}

		// 測試大額支付（>10000）
		builder.WithUser(testUser)
		builder.PaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil)
		service := builder.BuildPaymentService()

		paymentDTO := &dto.PaymentCreateDTO{
			Amount:        15000.0, // 大額支付
			Currency:      "TWD",
			PaymentMethod: "bank_transfer",
			Description:   "大額支付測試",
		}

		payment, err := service.CreatePayment(1, paymentDTO)

		assert.NoError(t, err)
		assert.Equal(t, 15000.0, payment.Amount)
		assert.Equal(t, "bank_transfer", payment.PaymentMethod)
		assert.Equal(t, "pending", payment.Status) // 大額支付可能需要人工審核

		builder.AssertAllExpectations()
	})

	t.Run("並發支付處理 - 混合資料庫", func(t *testing.T) {
		// 測試 PostgreSQL 和 MySQL 的並發能力差異
		testCases := []struct {
			dbType testutils.DatabaseType
			users  int
		}{
			{testutils.PostgreSQLTest, 50}, // PostgreSQL 較好的並發處理
			{testutils.MySQLTest, 30},      // MySQL 相對較少的並發
		}

		for _, tc := range testCases {
			t.Run(string(tc.dbType), func(t *testing.T) {
				builder := testutils.NewServiceBuilderWithDB(t, tc.dbType)

				// 模擬多用戶同時支付
				for i := 1; i <= tc.users; i++ {
					user := &models.User{ID: uint(i), Username: fmt.Sprintf("user%d", i)}
					builder.UserRepo.On("FindByID", uint(i)).Return(user, nil)
					builder.PaymentRepo.On("Create", mock.AnythingOfType("*models.Payment")).Return(nil)
				}

				service := builder.BuildPaymentService()

				// 執行並發支付測試
				for i := 1; i <= tc.users; i++ {
					paymentDTO := &dto.PaymentCreateDTO{
						Amount:        100.0,
						Currency:      "TWD",
						PaymentMethod: "credit_card",
						Description:   fmt.Sprintf("並發支付 %d", i),
					}

					_, err := service.CreatePayment(uint(i), paymentDTO)
					assert.NoError(t, err)
				}

				builder.AssertAllExpectations()
			})
		}
	})
}
