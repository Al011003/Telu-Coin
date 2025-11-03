// File: internal/dto/response/blockchain_stats_response.go
package response

type BlockchainStatsResponse struct {
	TotalMinted       string `json:"total_minted"`
	TotalBurned       string `json:"total_burned"`
	CirculatingSupply string `json:"circulating_supply"`
	MaxSupply         string `json:"max_supply"`
	TotalTransactions int    `json:"total_transactions"`
	TotalAddresses    int    `json:"total_addresses"`
	LatestBlock       uint64 `json:"latest_block"`
}