package handler

import (
	"strconv"
	"telkom_coin_back_end/internal/dto/request"
	service "telkom_coin_back_end/internal/services"
	"telkom_coin_back_end/pkg/helpers"

	"github.com/gin-gonic/gin"
)

type TLCWalletHandler struct {
	TLCWalletService *service.TLCWalletService
	UserService      *service.UserService
}

func NewTLCWalletHandler(tlcWalletService *service.TLCWalletService, userService *service.UserService) *TLCWalletHandler {
	return &TLCWalletHandler{
		TLCWalletService: tlcWalletService,
		UserService:      userService,
	}
}

// GetTLCBalance gets real-time TLC balance from blockchain
func (h *TLCWalletHandler) GetTLCBalance(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get user to get wallet address
	user, err := h.UserService.GetProfile(userID)
	if err != nil {
		helpers.NotFoundResponse(c, "User not found")
		return
	}

	// Get balance from blockchain
	balance, err := h.TLCWalletService.GetBlockchainBalance(user.WalletAddress)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get balance", err)
		return
	}

	helpers.SuccessResponse(c, "TLC balance retrieved successfully", balance)
}

// TransferTLC transfers TLC tokens on blockchain (no KYC required)
func (h *TLCWalletHandler) ValidateTransfer(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.ValidateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Call validate service (TIDAK ada PIN, cuma validasi data)
	validateResult, err := h.TLCWalletService.ValidateTransfer(userID, req.ToAddress, req.Amount)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "Transfer validation successful", validateResult)
}

// ============================================================================
// HANDLER 2: TransferTLC - Untuk execute transfer (setelah confirm + PIN)
// ============================================================================
func (h *TLCWalletHandler) TransferTLC(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.TLCTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Execute transfer (dengan PIN verification)
	transferResult, err := h.TLCWalletService.TransferTLC(userID, req.ToAddress, req.Amount, req.Memo, req.Pin)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "TLC transfer completed successfully", transferResult)
}

// GetTLCTransactionHistory gets transaction history from blockchain events (placeholder)
func (h *TLCWalletHandler) GetTransactionHistory(c *gin.Context) {
	// 1. Dapatkan userID dari konteks (misal: dari token JWT)
	userID := c.GetInt64("user_id")

	// 2. Parse parameter pagination dari query URL (misal: /history?page=1&limit=20)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 { // Batasi limit untuk performa
		limit = 10
	}

	// 3. Panggil service untuk melakukan tugasnya
	historyResponse, err := h.TLCWalletService.GetTransactionHistory(userID, page, limit)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	// 4. Kirim response sukses ke pengguna
	helpers.SuccessResponse(c, "Transaction history retrieved successfully", historyResponse)
}

// ValidateTLCAddress validates if an address is a valid Ethereum address
func (h *TLCWalletHandler) ValidateTLCAddress(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		helpers.BadRequestResponse(c, "Address parameter is required", nil)
		return
	}

	// Basic validation
	isValid := len(address) == 42 && address[:2] == "0x"

	validation := map[string]interface{}{
		"address":  address,
		"is_valid": isValid,
		"format":   "Ethereum address",
		"network":  "Ganache Local",
	}

	if isValid {
		helpers.SuccessResponse(c, "Address is valid", validation)
	} else {
		helpers.BadRequestResponse(c, "Invalid address format", nil)
	}
}
