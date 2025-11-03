package response

import "time"

type UserResponse struct {
	ID            int64     `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone,omitempty"`
	WalletAddress string    `json:"wallet_address"`
	KYCStatus     string    `json:"kyc_status"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserDetailResponse struct {
	ID            int64           `json:"id"`
	Username      string          `json:"username"`
	Email         string          `json:"email"`
	Phone         string          `json:"phone,omitempty"`
	WalletAddress string          `json:"wallet_address"`
	KYCStatus     string          `json:"kyc_status"`
	Status        string          `json:"status"`
	Balance       BalanceResponse `json:"balance"`
	CreatedAt     time.Time       `json:"created_at"`
}
