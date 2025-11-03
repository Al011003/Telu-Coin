package response

type BlockchainTransaction struct {
	Hash        string                 `json:"hash"`
	BlockNumber uint64                 `json:"block_number"`
	Type        string                 `json:"type"`
	Address     string                 `json:"address"`
	Timestamp   int64                  `json:"timestamp"`
	Data        map[string]interface{} `json:"data"`
}

type BlockchainTxListResponse struct {
	TotalCount   int                     `json:"total_count"`
	Transactions []BlockchainTransaction `json:"transactions"`
	Page         int                     `json:"page"`
	Limit        int                     `json:"limit"`
}
