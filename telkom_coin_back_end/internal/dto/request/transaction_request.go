package request

type GetTransactionHistoryRequest struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	TxType string `form:"type" binding:"omitempty,oneof=transfer topup withdraw"`
	Status string `form:"status" binding:"omitempty,oneof=pending confirmed failed"`
}
