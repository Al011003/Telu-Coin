package handler

import (
	"telkom_coin_back_end/internal/dto/request"
	service "telkom_coin_back_end/internal/services"
	"telkom_coin_back_end/pkg/helpers"

	"github.com/gin-gonic/gin"
)

type TopupHandler struct {
	TopupService *service.TopupService
}

func NewTopupHandler(topupService *service.TopupService) *TopupHandler {
	return &TopupHandler{TopupService: topupService}
}

// RequestTopup handles topup request
func (h *TopupHandler) RequestTopup(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.TopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate amount
	if !helpers.ValidateIDRAmount(req.Amount) {
		helpers.BadRequestResponse(c, "Invalid amount", nil)
		return
	}

	topupResponse, err := h.TopupService.RequestTopup(userID, &req)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "Topup request created successfully", topupResponse)
}

// GetTopupHistory gets user's topup history
func (h *TopupHandler) GetTopupHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get pagination parameters
	page, limit, err := helpers.ValidatePagination(c.Query("page"), c.Query("limit"))
	if err != nil {
		helpers.BadRequestResponse(c, "Invalid pagination parameters", err)
		return
	}

	// Get topup history from service
	history, err := h.TopupService.GetTopupHistory(userID, page, limit)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "Topup history retrieved", history)
}

// GetTopupDetail gets specific topup detail
func (h *TopupHandler) GetTopupDetail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	topupID := c.Param("id")

	if topupID == "" {
		helpers.BadRequestResponse(c, "Topup ID is required", nil)
		return
	}

	// TODO: Implement GetTopupDetail in service
	// For now, return empty response
	helpers.SuccessResponse(c, "Topup detail retrieved", map[string]interface{}{
		"id":      topupID,
		"user_id": userID,
	})
}

// CancelTopup cancels a pending topup request
func (h *TopupHandler) CancelTopup(c *gin.Context) {
	_ = c.GetInt64("user_id") // TODO: Use userID when implementing service
	topupID := c.Param("id")

	if topupID == "" {
		helpers.BadRequestResponse(c, "Topup ID is required", nil)
		return
	}

	// TODO: Implement CancelTopup in service
	// For now, return success response
	helpers.SuccessResponse(c, "Topup cancelled successfully", nil)
}

// GetPaymentMethods returns available payment methods
func (h *TopupHandler) GetPaymentMethods(c *gin.Context) {
	paymentMethods := []map[string]interface{}{
		{
			"method":      "bank_transfer",
			"name":        "Bank Transfer",
			"description": "Transfer to our bank account",
			"min_amount":  "10000",
			"max_amount":  "100000000",
			"fee":         "0",
		},
		{
			"method":      "va",
			"name":        "Virtual Account",
			"description": "Pay using virtual account",
			"min_amount":  "10000",
			"max_amount":  "50000000",
			"fee":         "2500",
		},
		{
			"method":      "qris",
			"name":        "QRIS",
			"description": "Scan QR code to pay",
			"min_amount":  "10000",
			"max_amount":  "2000000",
			"fee":         "1000",
		},
	}

	helpers.SuccessResponse(c, "Payment methods retrieved", paymentMethods)
}

// GetExchangeRate returns IDR to Telkom Coin exchange rate
func (h *TopupHandler) GetExchangeRate(c *gin.Context) {
	// In a real implementation, this would be dynamic
	exchangeRate := map[string]interface{}{
		"from":       "IDR",
		"to":         "TELKOM_COIN",
		"rate":       "1", // 1 IDR = 1 TELKOM_COIN
		"min_idr":    "10000",
		"max_idr":    "100000000",
		"updated_at": "2024-01-01T00:00:00Z",
	}

	helpers.SuccessResponse(c, "Exchange rate retrieved", exchangeRate)
}

// CalculateTopup calculates topup amount and fees
func (h *TopupHandler) CalculateTopup(c *gin.Context) {
	var req struct {
		Amount        string `json:"amount" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate amount
	if !helpers.ValidateIDRAmount(req.Amount) {
		helpers.BadRequestResponse(c, "Invalid amount", nil)
		return
	}

	// Calculate fees based on payment method
	var fee string
	switch req.PaymentMethod {
	case "bank_transfer":
		fee = "0"
	case "va":
		fee = "2500"
	case "qris":
		fee = "1000"
	default:
		helpers.BadRequestResponse(c, "Invalid payment method", nil)
		return
	}

	calculation := map[string]interface{}{
		"amount":         req.Amount,
		"payment_method": req.PaymentMethod,
		"fee":            fee,
		"total_idr":      req.Amount, // In real implementation, add fee to amount
		"telkom_coin":    req.Amount, // 1:1 exchange rate
	}

	helpers.SuccessResponse(c, "Topup calculation completed", calculation)
}
