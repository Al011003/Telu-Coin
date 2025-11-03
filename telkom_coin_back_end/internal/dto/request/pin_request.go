package request

// SetPinRequest is used to set or update user's wallet PIN
type SetPinRequest struct {
	Pin        string `json:"pin" validate:"required,len=6,numeric"`
	PinConfirm string `json:"pin_confirm" validate:"required,eqfield=Pin"`
}

// VerifyPinRequest used for verifying PIN in sensitive actions
type VerifyPinRequest struct {
	Pin string `json:"pin" validate:"required,len=6,numeric"`
}
