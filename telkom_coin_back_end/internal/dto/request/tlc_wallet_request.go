package request

// TLCTransferRequest represents a TLC transfer request
type TLCTransferRequest struct {
	ToAddress string `json:"to_address" binding:"required" validate:"eth_address"`
	Amount    string `json:"amount" binding:"required" validate:"numeric,gt=0"`
	Memo      string `json:"memo,omitempty" validate:"max=100"`
	Pin       string `json:"pin" binding:"required" validate:"len=6,numeric"`
}

type ValidateTransferRequest struct {
	ToAddress string `json:"to_address" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
}

// TLCTopupRequest represents a TLC topup request
type TLCTopupRequest struct {
	Amount       string `json:"amount" binding:"required" validate:"numeric,gte=10000"`
	PaymentProof string `json:"payment_proof,omitempty" validate:"omitempty,min=1,max=500"` // Optional for TLC Wallet
	Pin          string `json:"pin" binding:"required" validate:"len=6,numeric"`
}

// TLCWithdrawRequest represents a TLC withdrawal request
type TLCWithdrawRequest struct {
	Amount      string `json:"amount" binding:"required" validate:"numeric,gte=1000"`
	BankAccount string `json:"bank_account" binding:"required" validate:"min=10,max=200"`
	Pin         string `json:"pin" binding:"required" validate:"len=6,numeric"`
}

// ProcessTopupRequest represents a request to process a pending topup
type ProcessTopupRequest struct {
	RequestID string `json:"request_id" binding:"required" validate:"len=66"` // 0x + 64 hex chars
}

// TLCTransactionHistoryRequest represents a request for transaction history
type TLCTransactionHistoryRequest struct {
	WalletAddress string `json:"wallet_address,omitempty" validate:"omitempty,eth_address"`
	TxType        string `json:"tx_type,omitempty" validate:"omitempty,oneof=transfer topup withdraw"`
	Status        string `json:"status,omitempty" validate:"omitempty,oneof=confirmed pending failed"`
	Page          int    `json:"page,omitempty" validate:"omitempty,min=1"`
	Limit         int    `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	FromDate      string `json:"from_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	ToDate        string `json:"to_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// TLCBalanceRequest represents a request to get balance
type TLCBalanceRequest struct {
	WalletAddress string `json:"wallet_address,omitempty" validate:"omitempty,eth_address"`
}
