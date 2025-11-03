package request

type TransferRequest struct {
	ToAddress string `json:"to_address" binding:"required,len=42"` // Ethereum address
	Amount    string `json:"amount" binding:"required"`
	Pin       string `json:"pin" binding:"required,len=6,numeric"`
	Memo      string `json:"memo" binding:"omitempty,max=200"` // Optional memo
}

type TransferByUsernameRequest struct {
	ToUsername string `json:"to_username" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	Pin        string `json:"pin" binding:"required,len=6,numeric"`
	Memo       string `json:"memo" binding:"omitempty,max=200"`
}
