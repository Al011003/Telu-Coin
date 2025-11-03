package handler

import (
	"net/http"
	"telkom_coin_back_end/internal/dto/response"
	service "telkom_coin_back_end/internal/services"

	"github.com/gin-gonic/gin"
)

type BalanceHandler struct {
	BalanceService *service.BalanceService
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
	return &BalanceHandler{BalanceService: balanceService}
}

// Get user's wallet balance
func (h *BalanceHandler) GetBalance(c *gin.Context) {
	userID := c.GetInt64("user_id") // set by JWT middleware

	balance, err := h.BalanceService.GetBalance(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.BaseResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.BaseResponse{
		Status:  "success",
		Message: "balance retrieved",
		Data:    balance,
	})
}
