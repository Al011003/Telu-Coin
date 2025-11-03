package request

type TopupRequest struct {
	Amount        string `json:"amount" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=bank_transfer va e-wallet"`
	Pin           string `json:"pin" binding:"required" validate:"len=6,numeric"` // Added PIN for verification
}
