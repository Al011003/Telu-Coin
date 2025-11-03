package response

import "time"

type TopupResponse struct {
	ID             int64                  `json:"id"`
	Amount         string                 `json:"amount"`
	Currency       string                 `json:"currency"`
	PaymentMethod  string                 `json:"payment_method"`
	PaymentDetails map[string]interface{} `json:"payment_details"`
	Status         string                 `json:"status"`
	TxHash         string                 `json:"tx_hash,omitempty"` // Added for transaction tracking
	ExpiredAt      *time.Time             `json:"expired_at,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

type TopupHistoryResponse struct {
	Topups     []TopupResponse `json:"topups"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}
