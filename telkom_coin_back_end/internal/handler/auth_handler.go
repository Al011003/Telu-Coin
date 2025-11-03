package handler

import (
	"net/http"
	"telkom_coin_back_end/internal/dto/request"
	"telkom_coin_back_end/internal/dto/response"
	service "telkom_coin_back_end/internal/services"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
)

type AuthHandler struct {
	AuthService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Register user
func (h *AuthHandler) Register(c *gin.Context) {
    var req request.RegisterRequest

    // 1️⃣ Bind JSON minimal / langsung cek duplicate
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.BaseResponse{
            Status:  "error",
            Message: err.Error(),
        })
        return
    }

    // 2️⃣ Duplicate check dulu
    if err := h.AuthService.CheckDuplicate(req.Username, req.Email); err != nil {
        c.JSON(http.StatusBadRequest, response.BaseResponse{
            Status: "error",
            Message: map[string]string{
                "username": "username already taken",
                "email":    "email already registered",
            },
        })
        return
    }

    // 3️⃣ Ozzo validate semua field
    if err := req.Validate(); err != nil {
        if errs, ok := err.(validation.Errors); ok {
            validationErrors := map[string]string{}
            for field, e := range errs {
                validationErrors[field] = e.Error()
            }

            c.JSON(http.StatusBadRequest, response.BaseResponse{
                Status:  "error",
                Message: validationErrors,
            })
            return
        }

        // fallback error umum
        c.JSON(http.StatusBadRequest, response.BaseResponse{
            Status:  "error",
            Message: err.Error(),
        })
        return
    }

    // 4️⃣ Kalau semua clear, panggil service register
    user, err := h.AuthService.Register(req.Username, req.Email, req.Phone, req.Password, req.Pin)
    if err != nil {
        c.JSON(http.StatusBadRequest, response.BaseResponse{
            Status:  "error",
            Message: err.Error(),
        })
        return
    }

    // 5️⃣ Response sukses
    c.JSON(http.StatusOK, response.BaseResponse{
        Status:  "success",
        Message: "user registered",
        Data: response.RegisterResponse{
            ID:            user.ID,
            Username:      user.Username,
            Email:         user.Email,
            WalletAddress: user.WalletAddress,
            CreatedAt:     user.CreatedAt,
        },
    })
}




// Login user
func (h *AuthHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BaseResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	token, err := h.AuthService.Login(req.EmailOrUsername, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.BaseResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.BaseResponse{
		Status:  "success",
		Message: "login successful",
		Data:    response.AuthResponse{Token: token},
	})
}
