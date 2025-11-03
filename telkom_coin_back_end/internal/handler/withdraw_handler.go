package handler

import (
	"telkom_coin_back_end/internal/dto/request"
	service "telkom_coin_back_end/internal/services"
	"telkom_coin_back_end/pkg/helpers"

	"github.com/gin-gonic/gin"
)

type WithdrawHandler struct {
	WithdrawService *service.WithdrawService
	UserService     *service.UserService
}

func NewWithdrawHandler(withdrawService *service.WithdrawService, userService *service.UserService) *WithdrawHandler {
	return &WithdrawHandler{
		WithdrawService: withdrawService,
		UserService:     userService,
	}
}

// RequestWithdraw handles withdrawal request
func (h *WithdrawHandler) RequestWithdraw(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Lakukan validasi dasar di handler
	if !helpers.ValidateCoinAmount(req.Amount) {
		helpers.BadRequestResponse(c, "Invalid amount", nil)
		return
	}
	if !helpers.ValidatePin(req.Pin) {
		helpers.BadRequestResponse(c, "Invalid PIN format", nil)
		return
	}
	if req.DestinationAddress == "" {
		helpers.BadRequestResponse(c, "Destination address is required", nil)
		return
	}

	// Untuk bank/ewallet, formatnya bisa lebih spesifik
	// Contoh: "BCA:1234567890:BUDI SETIAWAN"
	bankAccountInfo := req.DestinationAddress

	// 👇 PANGGIL FUNGSI SERVICE YANG SUDAH DIREVISI 👇
	withdrawResponse, err := h.WithdrawService.CreateWithdrawal(userID, req.Amount, bankAccountInfo, req.Pin)
	if err != nil {
		// Error dari service sudah cukup deskriptif
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "Withdrawal request has been sent to the blockchain", withdrawResponse)
}

// GetWithdrawHistory gets user's withdrawal history
func (h *WithdrawHandler) GetWithdrawHistory(c *gin.Context) {
	_ = c.GetInt64("user_id") // TODO: Use userID when implementing service

	// Get pagination parameters
	page, limit, err := helpers.ValidatePagination(c.Query("page"), c.Query("limit"))
	if err != nil {
		helpers.BadRequestResponse(c, "Invalid pagination parameters", err)
		return
	}

	// Get filter parameters
	status := c.Query("status")
	_ = c.Query("destination_type") // TODO: Use destinationType when implementing service

	// Validate status if provided
	if status != "" {
		validStatuses := []string{"pending", "processing", "completed", "failed", "cancelled"}
		isValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			helpers.BadRequestResponse(c, "Invalid status filter", nil)
			return
		}
	}

	// TODO: Implement GetWithdrawHistory in service
	// For now, return empty response
	helpers.SuccessResponse(c, "Withdrawal history retrieved", map[string]interface{}{
		"withdrawals": []interface{}{},
		"total":       0,
		"page":        page,
		"limit":       limit,
	})
}

// GetWithdrawDetail gets specific withdrawal detail
func (h *WithdrawHandler) GetWithdrawDetail(c *gin.Context) {
	userID := c.GetInt64("user_id")
	withdrawID := c.Param("id")

	if withdrawID == "" {
		helpers.BadRequestResponse(c, "Withdrawal ID is required", nil)
		return
	}

	// TODO: Implement GetWithdrawDetail in service
	// For now, return empty response
	helpers.SuccessResponse(c, "Withdrawal detail retrieved", map[string]interface{}{
		"id":      withdrawID,
		"user_id": userID,
	})
}

// CancelWithdraw cancels a pending withdrawal request
func (h *WithdrawHandler) CancelWithdraw(c *gin.Context) {
	_ = c.GetInt64("user_id") // TODO: Use userID when implementing service
	withdrawID := c.Param("id")

	if withdrawID == "" {
		helpers.BadRequestResponse(c, "Withdrawal ID is required", nil)
		return
	}

	// TODO: Implement CancelWithdraw in service
	// For now, return success response
	helpers.SuccessResponse(c, "Withdrawal cancelled successfully", nil)
}

// GetWithdrawMethods returns available withdrawal methods
func (h *WithdrawHandler) GetWithdrawMethods(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get user info to determine available methods based on KYC status
	user, err := h.UserService.GetProfile(userID)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get user info", err)
		return
	}

	methods := []map[string]interface{}{
		{
			"type":            "wallet",
			"name":            "External Wallet",
			"description":     "Withdraw to external wallet address",
			"min_amount":      "100000",
			"max_amount":      "100000000",
			"fee":             "50000",
			"processing_time": "1-24 hours",
			"available":       true,
		},
	}

	// Add fiat withdrawal methods only for verified users
	if user.KYCStatus == "verified" {
		methods = append(methods, map[string]interface{}{
			"type":            "bank",
			"name":            "Bank Transfer",
			"description":     "Withdraw to bank account",
			"min_amount":      "100000",
			"max_amount":      "500000000",
			"fee":             "5000",
			"processing_time": "1-3 business days",
			"available":       true,
		})

		methods = append(methods, map[string]interface{}{
			"type":            "ewallet",
			"name":            "E-Wallet",
			"description":     "Withdraw to e-wallet (OVO, GoPay, DANA)",
			"min_amount":      "50000",
			"max_amount":      "20000000",
			"fee":             "2500",
			"processing_time": "instant",
			"available":       true,
		})
	}

	helpers.SuccessResponse(c, "Withdrawal methods retrieved", methods)
}

// GetExchangeRate returns Telkom Coin to IDR exchange rate
func (h *WithdrawHandler) GetExchangeRate(c *gin.Context) {
	// In a real implementation, this would be dynamic
	exchangeRate := map[string]interface{}{
		"from":       "TELKOM_COIN",
		"to":         "IDR",
		"rate":       "1", // 1 TELKOM_COIN = 1 IDR
		"min_coin":   "100000",
		"max_coin":   "100000000",
		"updated_at": "2024-01-01T00:00:00Z",
	}

	helpers.SuccessResponse(c, "Exchange rate retrieved", exchangeRate)
}

// CalculateWithdraw calculates withdrawal amount and fees
func (h *WithdrawHandler) CalculateWithdraw(c *gin.Context) {
	var req struct {
		Amount          string `json:"amount" binding:"required"`
		DestinationType string `json:"destination_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate amount
	if !helpers.ValidateCoinAmount(req.Amount) {
		helpers.BadRequestResponse(c, "Invalid amount", nil)
		return
	}

	// Calculate fees based on destination type
	var fee string
	var processingTime string
	switch req.DestinationType {
	case "wallet":
		fee = "50000"
		processingTime = "1-24 hours"
	case "bank":
		fee = "5000"
		processingTime = "1-3 business days"
	case "ewallet":
		fee = "2500"
		processingTime = "instant"
	default:
		helpers.BadRequestResponse(c, "Invalid destination type", nil)
		return
	}

	calculation := map[string]interface{}{
		"amount":           req.Amount,
		"destination_type": req.DestinationType,
		"fee":              fee,
		"net_amount":       req.Amount, // In real implementation, subtract fee
		"idr_amount":       req.Amount, // 1:1 exchange rate
		"processing_time":  processingTime,
	}

	helpers.SuccessResponse(c, "Withdrawal calculation completed", calculation)
}

// GetWithdrawLimits returns withdrawal limits for user
func (h *WithdrawHandler) GetWithdrawLimits(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get user info to determine limits based on KYC status
	user, err := h.UserService.GetProfile(userID)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get user info", err)
		return
	}

	var limits map[string]interface{}

	switch user.KYCStatus {
	case "verified":
		limits = map[string]interface{}{
			"daily_limit":     "500000000",  // 500M coins
			"monthly_limit":   "2000000000", // 2B coins
			"per_transaction": "100000000",  // 100M coins
			"kyc_status":      "verified",
			"fiat_withdrawal": true,
		}
	case "pending":
		limits = map[string]interface{}{
			"daily_limit":     "50000000",  // 50M coins
			"monthly_limit":   "200000000", // 200M coins
			"per_transaction": "10000000",  // 10M coins
			"kyc_status":      "pending",
			"fiat_withdrawal": false,
		}
	default:
		limits = map[string]interface{}{
			"daily_limit":     "5000000",  // 5M coins
			"monthly_limit":   "20000000", // 20M coins
			"per_transaction": "1000000",  // 1M coins
			"kyc_status":      "unverified",
			"fiat_withdrawal": false,
		}
	}

	helpers.SuccessResponse(c, "Withdrawal limits retrieved", limits)
}
