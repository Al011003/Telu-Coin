package handler

import (
	"math/big"
	"strconv"
	service "telkom_coin_back_end/internal/services"
	"telkom_coin_back_end/pkg/helpers"

	"github.com/gin-gonic/gin"
)

type BlockchainExplorerHandler struct {
	ExplorerService *service.BlockchainExplorerService
}

func NewBlockchainExplorerHandler(explorerService *service.BlockchainExplorerService) *BlockchainExplorerHandler {
	return &BlockchainExplorerHandler{
		ExplorerService: explorerService,
	}
}

// GetAllTransactions gets all TLC transactions from blockchain with pagination
func (h *BlockchainExplorerHandler) GetAllTransactions(c *gin.Context) {
	// Parse query parameters
	fromBlockStr := c.Query("from_block")
	toBlockStr := c.Query("to_block")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	var fromBlock, toBlock *big.Int
	page := 1
	limit := 10 // Default limit per page

	// Parse from_block
	if fromBlockStr != "" {
		if num, err := strconv.ParseInt(fromBlockStr, 10, 64); err == nil {
			fromBlock = big.NewInt(num)
		}
	}

	// Parse to_block
	if toBlockStr != "" {
		if num, err := strconv.ParseInt(toBlockStr, 10, 64); err == nil {
			toBlock = big.NewInt(num)
		}
	}

	// Parse page
	if num, err := strconv.Atoi(pageStr); err == nil && num > 0 {
		page = num
	}

	// Parse limit (max 100)
	if num, err := strconv.Atoi(limitStr); err == nil && num > 0 && num <= 100 {
		limit = num
	}

	// Get transactions from blockchain with pagination
	transactions, err := h.ExplorerService.GetAllTransactions(fromBlock, toBlock, page, limit)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get transactions", err)
		return
	}

	helpers.SuccessResponse(c, "All TLC transactions retrieved", transactions)
}

// GetTransactionsByAddress gets all transactions for a specific address with pagination
func (h *BlockchainExplorerHandler) GetTransactionsByAddress(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		helpers.BadRequestResponse(c, "Address parameter is required", nil)
		return
	}

	// Validate address format
	if len(address) != 42 || address[:2] != "0x" {
		helpers.BadRequestResponse(c, "Invalid address format", nil)
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	
	page := 1
	limit := 10

	// Parse page
	if num, err := strconv.Atoi(pageStr); err == nil && num > 0 {
		page = num
	}

	// Parse limit (max 100)
	if num, err := strconv.Atoi(limitStr); err == nil && num > 0 && num <= 100 {
		limit = num
	}

	// Get transactions for this address
	transactions, err := h.ExplorerService.GetTransactionsByAddress(address, page, limit)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get transactions", err)
		return
	}

	helpers.SuccessResponse(c, "Address transactions retrieved", transactions)
}

// GetTransactionByHash gets specific transaction by hash
func (h *BlockchainExplorerHandler) GetTransactionByHash(c *gin.Context) {
	txHash := c.Param("hash")
	if txHash == "" {
		helpers.BadRequestResponse(c, "Transaction hash parameter is required", nil)
		return
	}

	// Validate hash format
	if len(txHash) != 66 || txHash[:2] != "0x" {
		helpers.BadRequestResponse(c, "Invalid transaction hash format", nil)
		return
	}

	// Get transaction details
	transaction, err := h.ExplorerService.GetTransactionByHash(txHash)
	if err != nil {
		helpers.NotFoundResponse(c, "Transaction not found")
		return
	}

	helpers.SuccessResponse(c, "Transaction details retrieved", transaction)
}

// GetBlockchainStats gets overall blockchain statistics
func (h *BlockchainExplorerHandler) GetBlockchainStats(c *gin.Context) {
	// Get ALL transactions to calculate stats (without pagination)
	allTx, err := h.ExplorerService.GetAllTransactions(nil, nil, 1, 999999)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get blockchain stats", err)
		return
	}

	// Calculate statistics
	totalTransactions := allTx.TotalCount
	transferCount := 0
	topupCount := 0
	withdrawCount := 0
	mintCount := 0
	burnCount := 0
	
	// Variables untuk total minted & burned (gunakan big.Int untuk akurasi)
	totalMinted := big.NewInt(0)
	totalBurned := big.NewInt(0)

		for _, tx := range allTx.Transactions {
		// Count by type
		switch tx.Type {
		case "transfer":
			transferCount++
		case "topup": // 👈 UBAH DARI "mint" JADI "topup"
			topupCount++
			// Tambahkan amount ke total minted
			amount := new(big.Int)
			amount.SetString(tx.Amount, 10)
			totalMinted.Add(totalMinted, amount)
		case "withdraw": // 👈 UBAH DARI "burn" JADI "withdraw"
			withdrawCount++
			// Tambahkan amount ke total burned
			amount := new(big.Int)
			amount.SetString(tx.Amount, 10)
			totalBurned.Add(totalBurned, amount)
		}
	}

	// Calculate circulating supply (minted - burned)
	circulatingSupply := new(big.Int).Sub(totalMinted, totalBurned)

	stats := map[string]interface{}{
		"total_transactions": totalTransactions,
		"transaction_types": map[string]int{
			"transfers":   transferCount,
			"topups":      topupCount,
			"withdrawals": withdrawCount,
			"mints":       mintCount,
			"burns":       burnCount,
		},
		"coin_statistics": map[string]string{
			"total_minted":       totalMinted.String(),
			"total_burned":       totalBurned.String(),
			"circulating_supply": circulatingSupply.String(),
			"max_supply":         "Unlimited",
		},
		"network": map[string]interface{}{
			"name":         "Ganache Local",
			"chain_id":     "1337",
			"token_name":   "Telkom Token",
			"token_symbol": "TELKOM",
			"decimals":     18,
		},
		"features": map[string]bool{
			"real_time_tracking": true,
			"full_transparency":  true,
			"no_kyc_required":    true,
			"decentralized":      true,
		},
	}

	helpers.SuccessResponse(c, "Blockchain statistics retrieved", stats)
}
// SearchTransactions searches transactions by various criteria with pagination
func (h *BlockchainExplorerHandler) SearchTransactions(c *gin.Context) {
	query := c.Query("q")
	txType := c.Query("type")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	if query == "" {
		helpers.BadRequestResponse(c, "Search query is required", nil)
		return
	}

	page := 1
	limit := 10
	
	// Parse page
	if num, err := strconv.Atoi(pageStr); err == nil && num > 0 {
		page = num
	}
	
	// Parse limit (max 100)
	if num, err := strconv.Atoi(limitStr); err == nil && num > 0 && num <= 100 {
		limit = num
	}

	// Check if query is an address
	if len(query) == 42 && query[:2] == "0x" {
		transactions, err := h.ExplorerService.GetTransactionsByAddress(query, page, limit)
		if err != nil {
			helpers.InternalServerErrorResponse(c, "Search failed", err)
			return
		}

		// Filter by type if specified
		if txType != "" {
			var filtered []interface{}
			for _, tx := range transactions.Transactions {
				if tx.Type == txType {
					filtered = append(filtered, tx)
				}
			}
			
			helpers.SuccessResponse(c, "Search results", map[string]interface{}{
				"transactions": filtered,
				"total_count":  len(filtered),
				"page":         page,
				"limit":        limit,
				"search_type":  "address_filtered",
				"query":        query,
				"filter":       txType,
			})
			return
		}

		helpers.SuccessResponse(c, "Search results", map[string]interface{}{
			"transactions": transactions.Transactions,
			"total_count":  transactions.TotalCount,
			"page":         transactions.Page,
			"limit":        transactions.Limit,
			"search_type":  "address",
			"query":        query,
		})
		return
	}

	// Check if query is a transaction hash
	if len(query) == 66 && query[:2] == "0x" {
		transaction, err := h.ExplorerService.GetTransactionByHash(query)
		if err != nil {
			helpers.NotFoundResponse(c, "Transaction not found")
			return
		}

		helpers.SuccessResponse(c, "Search results", map[string]interface{}{
			"transaction": transaction,
			"search_type": "transaction_hash",
			"query":       query,
		})
		return
	}

	helpers.BadRequestResponse(c, "Invalid search query. Use address (0x...) or transaction hash (0x...)", nil)
}

// GetLatestTransactions gets the most recent transactions
func (h *BlockchainExplorerHandler) GetLatestTransactions(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	
	page := 1
	limit := 20 // Default limit for latest

	// Parse page
	if num, err := strconv.Atoi(pageStr); err == nil && num > 0 {
		page = num
	}

	// Parse limit (max 100)
	if num, err := strconv.Atoi(limitStr); err == nil && num > 0 && num <= 100 {
		limit = num
	}

	// Get latest transactions with pagination
	transactions, err := h.ExplorerService.GetAllTransactions(nil, nil, page, limit)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get latest transactions", err)
		return
	}

	helpers.SuccessResponse(c, "Latest transactions retrieved", transactions)
}

// GetTransactionTypes gets available transaction types
func (h *BlockchainExplorerHandler) GetTransactionTypes(c *gin.Context) {
	types := map[string]interface{}{
		"available_types": []map[string]string{
			{"type": "transfer", "description": "TLC token transfers between addresses"},
			{"type": "topup", "description": "Topup requests with payment proof"},
			{"type": "withdraw", "description": "Withdrawal requests to bank accounts"},
			{"type": "mint", "description": "New tokens created (topup processing)"},
			{"type": "burn", "description": "Tokens destroyed (withdrawal processing)"},
		},
		"note": "All transactions are recorded on blockchain and publicly viewable",
	}

	helpers.SuccessResponse(c, "Transaction types retrieved", types)
}

// GetTransactionsByType gets transactions filtered by type with pagination
func (h *BlockchainExplorerHandler) GetTransactionsByType(c *gin.Context) {
	txType := c.Param("type")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	// Validate transaction type
	validTypes := map[string]bool{
		"transfer": true,
		"topup":    true,
		"withdraw": true,
		"mint":     true,
		"burn":     true,
	}

	if !validTypes[txType] {
		helpers.BadRequestResponse(c, "Invalid transaction type", nil)
		return
	}

	page := 1
	limit := 10

	// Parse page
	if num, err := strconv.Atoi(pageStr); err == nil && num > 0 {
		page = num
	}

	// Parse limit (max 100)
	if num, err := strconv.Atoi(limitStr); err == nil && num > 0 && num <= 100 {
		limit = num
	}

	// Get all transactions first
	allTx, err := h.ExplorerService.GetAllTransactions(nil, nil, 1, 999999)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to get transactions", err)
		return
	}

	// Filter by type
	var filtered []interface{}
	for _, tx := range allTx.Transactions {
		if tx.Type == txType {
			filtered = append(filtered, tx)
		}
	}

	// Manual pagination for filtered results
	totalCount := len(filtered)
	totalPages := (totalCount + limit - 1) / limit
	startIdx := (page - 1) * limit
	endIdx := startIdx + limit

	if startIdx >= totalCount {
		filtered = []interface{}{}
	} else {
		if endIdx > totalCount {
			endIdx = totalCount
		}
		filtered = filtered[startIdx:endIdx]
	}

	helpers.SuccessResponse(c, "Transactions by type retrieved", map[string]interface{}{
		"transactions": filtered,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"page":         page,
		"limit":        limit,
		"type":         txType,
	})
}

// GetAddressBalance gets the balance of a specific address (bonus feature)
func (h *BlockchainExplorerHandler) GetAddressBalance(c *gin.Context) {
	address := c.Param("address")
	
	// Validate address format
	if len(address) != 42 || address[:2] != "0x" {
		helpers.BadRequestResponse(c, "Invalid address format", nil)
		return
	}

	// Get all transactions for this address
	transactions, err := h.ExplorerService.GetTransactionsByAddress(address, 1, 999999)
	if err != nil {
		helpers.InternalServerErrorResponse(c, "Failed to calculate balance", err)
		return
	}

	// Calculate balance from transactions
	balance := big.NewInt(0)
	received := big.NewInt(0)
	sent := big.NewInt(0)

	for _, tx := range transactions.Transactions {
		amount := new(big.Int)
		amount.SetString(tx.AmountWei, 10)

		// Received
		if tx.ToAddress == address {
			balance.Add(balance, amount)
			received.Add(received, amount)
		}
		
		// Sent
		if tx.FromAddress == address {
			balance.Sub(balance, amount)
			sent.Add(sent, amount)
		}
	}

	// Convert to TLC (divide by 10^18)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	balanceTLC := new(big.Int).Div(balance, divisor)
	receivedTLC := new(big.Int).Div(received, divisor)
	sentTLC := new(big.Int).Div(sent, divisor)

	helpers.SuccessResponse(c, "Address balance calculated", map[string]interface{}{
		"address":             address,
		"balance":             balanceTLC.String(),
		"balance_wei":         balance.String(),
		"total_received":      receivedTLC.String(),
		"total_received_wei":  received.String(),
		"total_sent":          sentTLC.String(),
		"total_sent_wei":      sent.String(),
		"transaction_count":   transactions.TotalCount,
		"note":                "Balance calculated from blockchain transactions",
	})
}

