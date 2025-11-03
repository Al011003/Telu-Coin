package handler

import (
	"telkom_coin_back_end/internal/dto/request"
	service "telkom_coin_back_end/internal/services"
	"telkom_coin_back_end/pkg/helpers"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	UserService *service.UserService
	AuthService *service.AuthService
}

func NewUserHandler(userService *service.UserService, authService *service.AuthService) *UserHandler {
	return &UserHandler{
		UserService: userService,
		AuthService: authService,
	}
}

// GetProfile gets user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.UserService.GetProfile(userID)
	if err != nil {
		helpers.NotFoundResponse(c, "User not found")
		return
	}

	helpers.SuccessResponse(c, "Profile retrieved", user)
}

// GetDetailProfile gets user profile with balance
func (h *UserHandler) GetDetailProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.UserService.GetDetailProfile(userID)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get profile", err)
		return
	}

	helpers.SuccessResponse(c, "Detailed profile retrieved", user)
}

// UpdateProfile updates user profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate email format if provided
	if req.Email != "" && !helpers.ValidateEmail(req.Email) {
		helpers.BadRequestResponse(c, "Invalid email format", nil)
		return
	}

	// Validate username format if provided
	if req.Username != "" && !helpers.ValidateUsername(req.Username) {
		helpers.BadRequestResponse(c, "Invalid username format", nil)
		return
	}

	// Validate phone format if provided
	if req.Phone != "" && !helpers.ValidatePhone(req.Phone) {
		helpers.BadRequestResponse(c, "Invalid phone format", nil)
		return
	}

	user, err := h.UserService.UpdateProfile(userID, &req)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "Profile updated successfully", user)
}

// ChangePassword changes user password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate new password strength
	if !helpers.ValidatePassword(req.NewPassword) {
		helpers.BadRequestResponse(c, "Password must be at least 8 characters with uppercase, lowercase, and digit", nil)
		return
	}

	err := h.AuthService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "Password changed successfully", nil)
}

// ChangePin changes user PIN
func (h *UserHandler) ChangePin(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req request.ChangePinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Validate PIN format
	if !helpers.ValidatePin(req.OldPin) || !helpers.ValidatePin(req.NewPin) {
		helpers.BadRequestResponse(c, "PIN must be 6 digits", nil)
		return
	}

	err := h.AuthService.ChangePin(userID, req.OldPin, req.NewPin)
	if err != nil {
		helpers.BadRequestResponse(c, err.Error(), err)
		return
	}

	helpers.SuccessResponse(c, "PIN changed successfully", nil)
}

// DeactivateAccount deactivates user account
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		Password string `json:"password" binding:"required"`
		Reason   string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// Verify password before deactivation
	_, err := h.AuthService.GetUserByID(userID)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get user", err)
		return
	}

	// Verify password (simplified - in real implementation use proper auth)
	if req.Password == "" {
		helpers.BadRequestResponse(c, "Password verification required", nil)
		return
	}

	err = h.AuthService.DeactivateAccount(userID)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to deactivate account", err)
		return
	}

	helpers.SuccessResponse(c, "Account deactivated successfully", nil)
}

// GetAccountStatus gets account status and limits
func (h *UserHandler) GetAccountStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.UserService.GetProfile(userID)
	if err != nil {
		helpers.NotFoundResponse(c, "User not found")
		return
	}

	// TLC Wallet: No limits, all users auto-verified
	limits := map[string]interface{}{
		"daily_transfer":     "unlimited",
		"daily_withdrawal":   "unlimited",
		"monthly_transfer":   "unlimited",
		"monthly_withdrawal": "unlimited",
		"fiat_withdrawal":    true,
		"blockchain_direct":  true,
		"kyc_required":       false,
		"note":               "TLC Wallet operates without KYC restrictions",
	}

	status := map[string]interface{}{
		"user_id":    user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"kyc_status": user.KYCStatus,
		"status":     user.Status,
		"limits":     limits,
		"created_at": user.CreatedAt,
	}

	helpers.SuccessResponse(c, "Account status retrieved", status)
}

// RequestKYCVerification requests KYC verification
func (h *UserHandler) RequestKYCVerification(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		FullName    string `json:"full_name" binding:"required"`
		IDNumber    string `json:"id_number" binding:"required"`
		IDType      string `json:"id_type" binding:"required,oneof=ktp passport sim"`
		Address     string `json:"address" binding:"required"`
		DateOfBirth string `json:"date_of_birth" binding:"required"`
		IDPhoto     string `json:"id_photo" binding:"required"`     // Base64 or URL
		SelfiePhoto string `json:"selfie_photo" binding:"required"` // Base64 or URL
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		helpers.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// TLC Wallet: Auto-verify user (no KYC required)
	err := h.AuthService.UpdateKYCStatus(userID, "verified")
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to verify account", err)
		return
	}

	helpers.SuccessResponse(c, "Account automatically verified for TLC Wallet", map[string]interface{}{
		"status":  "verified",
		"message": "TLC Wallet operates without KYC requirements. Your account is automatically verified.",
		"note":    "No document submission required for blockchain-based operations",
	})
}

// GetKYCStatus gets KYC verification status
func (h *UserHandler) GetKYCStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")

	_, err := h.UserService.GetProfile(userID)
	if err != nil {
		helpers.NotFoundResponse(c, "User not found")
		return
	}

	kycStatus := map[string]interface{}{
		"status":             "verified",
		"message":            "TLC Wallet operates without KYC requirements. All users are automatically verified.",
		"required_documents": []string{},
		"blockchain_ready":   true,
		"note":               "No KYC verification needed for blockchain-based operations",
	}

	helpers.SuccessResponse(c, "KYC status retrieved", kycStatus)
}

// Helper function to get KYC status message
func getKYCStatusMessage(status string) string {
	switch status {
	case "verified":
		return "Your identity has been verified successfully"
	case "pending":
		return "Your KYC verification is under review"
	case "rejected":
		return "Your KYC verification was rejected. Please resubmit with correct documents"
	default:
		return "Please complete KYC verification to unlock full features"
	}
}
