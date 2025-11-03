package response

import "time"

// BalanceResponse for wallet balances
type BalanceResponse struct {
	Balance       string    `json:"balance"`
	LockedBalance string    `json:"locked_balance"`
	Available     string    `json:"available"` // balance - locked_balance
	UpdatedAt     time.Time `json:"updated_at"`
}
