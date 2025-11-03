package request

type WithdrawRequest struct {
	Amount             string                 `json:"amount" binding:"required"`
	DestinationType    string                 `json:"destination_type" binding:"required,oneof=wallet bank ewallet"`
	DestinationAddress string                 `json:"destination_address" binding:"required"` // Wallet address or account
	BankDetails        map[string]interface{} `json:"bank_details,omitempty"`                 // For fiat withdrawal
	Pin                string                 `json:"pin" binding:"required,len=6,numeric"`
}
