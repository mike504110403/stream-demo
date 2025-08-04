package api

import (
	"net/http"
	"strconv"
	"stream-demo/backend/dto"
	"stream-demo/backend/dto/request"
	"stream-demo/backend/dto/response"
	"stream-demo/backend/services"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var req request.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	// 從 context 獲取使用者 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	// 轉換為 DTO
	createDTO := &dto.PaymentCreateDTO{
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		Description:   req.Description,
	}

	payment, err := h.paymentService.CreatePayment(userID.(uint), createDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.NewSuccessResponse(response.NewPaymentResponse(payment)))
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的支付 ID"))
		return
	}

	payment, err := h.paymentService.GetPaymentByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, response.NewErrorResponse(404, "支付不存在"))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewPaymentResponse(payment)))
}

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	// 從 context 獲取使用者 ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, response.NewErrorResponse(401, "未登入"))
		return
	}

	payments, err := h.paymentService.GetPaymentsByUserID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(payments))
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的支付 ID"))
		return
	}

	payment, err := h.paymentService.CompletePayment(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewPaymentResponse(payment)))
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的支付 ID"))
		return
	}

	var req request.RefundPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, err.Error()))
		return
	}

	payment, err := h.paymentService.RefundPayment(uint(id), &dto.PaymentRefundDTO{Reason: req.Reason})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewPaymentResponse(payment)))
}

func (h *PaymentHandler) CompletePayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的支付 ID"))
		return
	}

	payment, err := h.paymentService.CompletePayment(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(response.NewPaymentResponse(payment)))
}

func (h *PaymentHandler) GetUserPayments(c *gin.Context) {
	// 獲取用戶 ID 參數
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse(400, "無效的用戶 ID"))
		return
	}

	payments, err := h.paymentService.GetPaymentsByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.NewErrorResponse(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.NewSuccessResponse(payments))
}
