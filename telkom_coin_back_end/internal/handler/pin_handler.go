package handler

import (
	"net/http"
	"telkom_coin_back_end/internal/dto/request"
	"telkom_coin_back_end/internal/dto/response"
	service "telkom_coin_back_end/internal/services"

	"github.com/gin-gonic/gin"
)

type PinHandler struct {
	PinService *service.PinService
}

func NewPinHandler(pinService *service.PinService) *PinHandler {
	return &PinHandler{PinService: pinService}
}

// Set or update PIN
func (h *PinHandler) SetPin(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req request.SetPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BaseResponse{Status: "error", Message: err.Error()})
		return
	}

	// Validate PIN confirmation
	if req.Pin != req.PinConfirm {
		c.JSON(http.StatusBadRequest, response.BaseResponse{Status: "error", Message: "PIN and PIN confirmation do not match"})
		return
	}

	err := h.PinService.SetPin(userID, req.Pin)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BaseResponse{Status: "error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.BaseResponse{Status: "success", Message: "PIN updated"})
}

// Verify PIN
func (h *PinHandler) VerifyPin(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var req request.VerifyPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BaseResponse{Status: "error", Message: err.Error()})
		return
	}

	err := h.PinService.VerifyPin(userID, req.Pin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.BaseResponse{Status: "error", Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, response.BaseResponse{Status: "success", Message: "PIN verified"})
}
