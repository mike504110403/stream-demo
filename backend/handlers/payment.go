package handlers

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付處理器
type PaymentHandler struct {
	paymentService *services.PaymentService
}

// NewPaymentHandler 創建支付處理器實例
func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// CreatePayment 創建支付
func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := c.GetUint("user_id")

	var createDTO dto.PaymentCreateDTO
	if err := c.ShouldBindJSON(&createDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.CreatePayment(userID, &createDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// GetPayment 獲取支付資訊
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的支付 ID"})
		return
	}

	payment, err := h.paymentService.GetPaymentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "支付不存在"})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// GetUserPayments 獲取用戶的支付列表
func (h *PaymentHandler) GetUserPayments(c *gin.Context) {
	userID := c.GetUint("user_id")

	payments, err := h.paymentService.GetPaymentsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// ProcessPayment 處理支付
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的支付 ID"})
		return
	}

	payment, err := h.paymentService.CompletePayment(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

// RefundPayment 退款
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的支付 ID"})
		return
	}

	var refundDTO dto.PaymentRefundDTO
	if err := c.ShouldBindJSON(&refundDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentService.RefundPayment(uint(id), &refundDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}
