package handler

import (
	"strconv"
	"telkom_coin_back_end/internal/dto/request"
	service "telkom_coin_back_end/internal/services"
	"telkom_coin_back_end/pkg/helpers"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	TxService *service.TransactionService
}

func NewTransactionHandler(txService *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{TxService: txService}
}

// GetTransactionHistory gets user's transaction history
func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.GetTransactionHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid query parameters", err)
		return
	}

	// Set default values if not provided
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	transactions, err := h.TxService.GetHistory(userID, &req)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get transaction history", err)
		return
	}

	helpers.SuccessResponse(c, "Transaction history retrieved", transactions)
}

// GetTransactionDetail gets specific transaction detail by hash
func (h *TransactionHandler) GetTransactionDetail(c *gin.Context) {
	txHash := c.Param("hash")

	if txHash == "" {
		helpers.BadRequestResponse(c, "Transaction hash is required", nil)
		return
	}

	transaction, err := h.TxService.GetByHash(txHash)
	if err != nil {
		helpers.NotFoundResponse(c, "Transaction not found")
		return
	}

	helpers.SuccessResponse(c, "Transaction detail retrieved", transaction)
}

// GetTransactionByID gets specific transaction detail by ID
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	txIDStr := c.Param("id")

	if txIDStr == "" {
		helpers.BadRequestResponse(c, "Transaction ID is required", nil)
		return
	}

	txID, err := strconv.ParseInt(txIDStr, 10, 64)
	if err != nil {
		helpers.BadRequestResponse(c, "Invalid transaction ID", err)
		return
	}

	transaction, err := h.TxService.GetByID(txID)
	if err != nil {
		helpers.NotFoundResponse(c, "Transaction not found")
		return
	}

	helpers.SuccessResponse(c, "Transaction detail retrieved", transaction)
}

// GetPendingTransactions gets pending transactions (admin function)
func (h *TransactionHandler) GetPendingTransactions(c *gin.Context) {
	limitStr := c.Query("limit")
	limit := 1000 // default limit

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	transactions, err := h.TxService.GetPendingTransactions(limit)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get pending transactions", err)
		return
	}

	helpers.SuccessResponse(c, "Pending transactions retrieved", transactions)
}
