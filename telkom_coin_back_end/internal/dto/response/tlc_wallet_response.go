package response

import "time"

// TLCBalanceResponse represents blockchain-based TLC balance
type TLCBalanceResponse struct {
	WalletAddress string    `json:"wallet_address"`
	Balance       string    `json:"balance"`      // Balance in TLC (human readable)
	BalanceWei    string    `json:"balance_wei"`  // Balance in wei (raw blockchain value)
	TokenSymbol   string    `json:"token_symbol"` // "TLC"
	TokenName     string    `json:"token_name"`   // "Telkom Coin"
	Decimals      int       `json:"decimals"`     // 18
	UpdatedAt     time.Time `json:"updated_at"`
}

// TLCTransferResponse represents a blockchain transfer result

// TLCTopupResponse represents a topup request result
type TLCTopupResponse struct {
	RequestID     string    `json:"request_id"`    // Transaction hash as request ID
	Amount        string    `json:"amount"`        // Amount in TLC
	AmountWei     string    `json:"amount_wei"`    // Amount in wei
	PaymentProof  string    `json:"payment_proof"` // IPFS hash or URL
	Status        string    `json:"status"`        // "pending", "processed"
	TxHash        string    `json:"tx_hash"`       // Transaction hash
	WalletAddress string    `json:"wallet_address"`
	ProcessAfter  time.Time `json:"process_after"` // Can be processed after this time
	CreatedAt     time.Time `json:"created_at"`
}

// TLCWithdrawResponse represents a withdrawal request result
type TLCWithdrawResponse struct {
	RequestID     string    `json:"request_id"`   // Transaction hash as request ID
	Amount        string    `json:"amount"`       // Amount in TLC
	AmountWei     string    `json:"amount_wei"`   // Amount in wei
	BankAccount   string    `json:"bank_account"` // Bank account details
	Status        string    `json:"status"`       // "processing", "completed"
	TxHash        string    `json:"tx_hash"`      // Transaction hash
	WalletAddress string    `json:"wallet_address"`
	CreatedAt     time.Time `json:"created_at"`
}

// ProcessTopupResponse represents the result of processing a topup request
type ProcessTopupResponse struct {
	RequestID     string    `json:"request_id"`      // Original request ID
	ProcessTxHash string    `json:"process_tx_hash"` // Transaction hash of processing
	Status        string    `json:"status"`          // "processed"
	ProcessedAt   time.Time `json:"processed_at"`
}

// TLCTransactionHistoryResponse represents transaction history from blockchain
type TLCTransactionHistoryResponse struct {
	Transactions []TLCTransactionItem `json:"transactions"`
	TotalCount   int                  `json:"total_count"`
	Page         int                  `json:"page"`
	Limit        int                  `json:"limit"`
}

// TLCTransactionItem represents a single transaction from blockchain
type TLCTransactionItem struct {
	TxHash      string    `json:"tx_hash"`
	BlockNumber uint64    `json:"block_number"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      string    `json:"amount"`     // Amount in TLC
	AmountWei   string    `json:"amount_wei"` // Amount in wei
	Type        string    `json:"type"`       // "transfer", "topup", "withdraw"
	Status      string    `json:"status"`     // "confirmed", "pending", "failed"
	GasUsed     uint64    `json:"gas_used"`
	GasPrice    string    `json:"gas_price"`
	Memo        string    `json:"memo,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Direction   string    `json:"direction"`
}

// TLCContractInfoResponse represents smart contract information
type TLCContractInfoResponse struct {
	ContractAddress string `json:"contract_address"`
	TokenName       string `json:"token_name"`       // "Telkom Token"
	TokenSymbol     string `json:"token_symbol"`     // "TELKOM"
	Decimals        uint8  `json:"decimals"`         // 18
	TotalSupply     string `json:"total_supply"`     // Total supply in TLC
	TotalSupplyWei  string `json:"total_supply_wei"` // Total supply in wei
	ExchangeRate    string `json:"exchange_rate"`    // 1 TLC = 1 IDR
	MinMintAmount   string `json:"min_mint_amount"`  // 10,000 TLC
	MinBurnAmount   string `json:"min_burn_amount"`  // 1,000 TLC
	ChainID         string `json:"chain_id"`         // 1337 for Ganache
	NetworkName     string `json:"network_name"`     // "Ganache Local"
}

// TLCWalletInfoResponse represents wallet information
type TLCWalletInfoResponse struct {
	WalletAddress string                  `json:"wallet_address"`
	Balance       TLCBalanceResponse      `json:"balance"`
	ContractInfo  TLCContractInfoResponse `json:"contract_info"`
	IsConnected   bool                    `json:"is_connected"` // Is blockchain connected
	LastSyncAt    time.Time               `json:"last_sync_at"`
}

type ValidateTransferResponse struct {
	FromAddress   string `json:"from_address"`
	FromUsername  string `json:"from_username"`
	ToAddress     string `json:"to_address"`
	ToUsername    string `json:"to_username"` // 👈 Nama penerima untuk popup
	Amount        string `json:"amount"`
	AmountWei     string `json:"amount_wei"`
	SenderBalance string `json:"sender_balance"`
	IsValid       bool   `json:"is_valid"`
}

// TLCTransferResponse - Response setelah transfer sukses
type TLCTransferResponse struct {
	TxHash      string    `json:"tx_hash"`
	FromAddress string    `json:"from_address"`
	ToAddress   string    `json:"to_address"`
	Amount      string    `json:"amount"`
	AmountWei   string    `json:"amount_wei"`
	Memo        string    `json:"memo"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}
