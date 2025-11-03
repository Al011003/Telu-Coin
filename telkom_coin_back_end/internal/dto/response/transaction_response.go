package response

import "time"

type TransactionResponse struct {
	ID          int64                  `json:"id"`
	TxHash      string                 `json:"tx_hash"`
	FromAddress string                 `json:"from_address"`
	ToAddress   string                 `json:"to_address"`
	Amount      string                 `json:"amount"`
	TxType      string                 `json:"tx_type"`
	Status      string                 `json:"status"`
	BlockNumber *int64                 `json:"block_number,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ConfirmedAt *time.Time             `json:"confirmed_at,omitempty"`
}

type TransactionDetailResponse struct {
	ID          int64                  `json:"id"`
	TxHash      string                 `json:"tx_hash"`
	FromAddress string                 `json:"from_address"`
	FromUser    *UserResponse          `json:"from_user,omitempty"` // Jika ada
	ToAddress   string                 `json:"to_address"`
	ToUser      *UserResponse          `json:"to_user,omitempty"` // Jika ada
	Amount      string                 `json:"amount"`
	TxType      string                 `json:"tx_type"`
	Status      string                 `json:"status"`
	BlockNumber *int64                 `json:"block_number,omitempty"`
	GasUsed     *int64                 `json:"gas_used,omitempty"`
	GasPrice    *int64                 `json:"gas_price,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	ConfirmedAt *time.Time             `json:"confirmed_at,omitempty"`
}

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	Limit        int                   `json:"limit"`
}
