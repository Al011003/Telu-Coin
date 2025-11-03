package response

import "time"

// BaseResponse is the standard API response format
type BaseResponse struct {
    Status  string      `json:"status"`
    Message interface{} `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// AuthResponse for authentication endpoints
type AuthResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
	ID            int64     `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	WalletAddress string    `json:"wallet_address"`
	CreatedAt     time.Time `json:"created_at"`
}

type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}
