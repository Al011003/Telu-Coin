package service

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/web3"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainExplorerService struct {
	client          *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
}

func NewBlockchainExplorerService() (*BlockchainExplorerService, error) {
	web3Client, err := web3.NewWeb3Client()
	if err != nil {
		return nil, err
	}

	// Contract ABI untuk parsing events
	contractABIString := `[
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "to", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"},
				{"indexed": true, "name": "requestId", "type": "bytes32"}
			],
			"name": "TokensMinted",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "from", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"},
				{"indexed": true, "name": "requestId", "type": "bytes32"}
			],
			"name": "TokensBurned",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "from", "type": "address"},
				{"indexed": true, "name": "to", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"},
				{"indexed": false, "name": "note", "type": "string"}
			],
			"name": "PaymentProcessed",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "user", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"},
				{"indexed": true, "name": "requestId", "type": "bytes32"},
				{"indexed": false, "name": "paymentProof", "type": "string"}
			],
			"name": "TopupRequested",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "user", "type": "address"},
				{"indexed": false, "name": "amount", "type": "uint256"},
				{"indexed": true, "name": "requestId", "type": "bytes32"},
				{"indexed": false, "name": "bankAccount", "type": "string"}
			],
			"name": "WithdrawRequested",
			"type": "event"
		},
		{
			"anonymous": false,
			"inputs": [
				{"indexed": true, "name": "from", "type": "address"},
				{"indexed": true, "name": "to", "type": "address"},
				{"indexed": false, "name": "value", "type": "uint256"}
			],
			"name": "Transfer",
			"type": "event"
		}
	]`

	contractABI, err := abi.JSON(strings.NewReader(contractABIString))
	if err != nil {
		return nil, err
	}

	return &BlockchainExplorerService{
		client:          web3Client.GetClient(),
		contractAddress: web3Client.GetContractAddress(),
		contractABI:     contractABI,
	}, nil
}

// GetAllTransactions gets all transactions with PROPER pagination and deduplication
func (s *BlockchainExplorerService) GetAllTransactions(fromBlock, toBlock *big.Int, page, limit int) (*response.TLCTransactionHistoryResponse, error) {
	if fromBlock == nil {
		fromBlock = big.NewInt(0) // Start from genesis block
	}
	if toBlock == nil {
		toBlock = nil // Latest block
	}

	// Default pagination values
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10 // Default 10 items per page
	}

	// Query all events from contract
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{s.contractAddress},
	}

	logs, err := s.client.FilterLogs(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to filter logs: %v", err)
	}

	var allTransactions []response.TLCTransactionItem
	processedTx := make(map[string]bool)

	// Group logs by transaction hash untuk deteksi duplikat
	txLogs := make(map[string][]types.Log)
	for _, vLog := range logs {
		txHash := vLog.TxHash.Hex()
		txLogs[txHash] = append(txLogs[txHash], vLog)
	}

	// Process each transaction (deduplication)
	for txHash, logsInTx := range txLogs {
		// Skip if already processed
		if processedTx[txHash] {
			continue
		}

		// Prioritas: ambil event custom dulu, skip Transfer
		var selectedLog *types.Log
		hasCustomEvent := false

		// Cari custom event terlebih dahulu (prioritas tinggi)
		for i := range logsInTx {
			eventSig := logsInTx[i].Topics[0].Hex()
			
			// Cek apakah ada custom event
			if eventSig == s.contractABI.Events["TokensMinted"].ID.Hex() ||
				eventSig == s.contractABI.Events["TokensBurned"].ID.Hex() ||
				eventSig == s.contractABI.Events["PaymentProcessed"].ID.Hex() ||
				eventSig == s.contractABI.Events["TopupRequested"].ID.Hex() ||
				eventSig == s.contractABI.Events["WithdrawRequested"].ID.Hex() {
				selectedLog = &logsInTx[i]
				hasCustomEvent = true
				break
			}
		}

		// Jika tidak ada custom event, gunakan Transfer event saja
		if !hasCustomEvent {
			for i := range logsInTx {
				if logsInTx[i].Topics[0].Hex() == s.contractABI.Events["Transfer"].ID.Hex() {
					selectedLog = &logsInTx[i]
					break
				}
			}
		}

		// Parse selected log (hanya 1 per transaksi)
		if selectedLog != nil {
			tx, err := s.parseLogToTransaction(*selectedLog)
			if err == nil && tx != nil {
				allTransactions = append(allTransactions, *tx)
				processedTx[txHash] = true
			}
		}
	}

	// Sort by block number and transaction hash (newest first)
	sort.Slice(allTransactions, func(i, j int) bool {
		if allTransactions[i].BlockNumber == allTransactions[j].BlockNumber {
			return allTransactions[i].TxHash > allTransactions[j].TxHash
		}
		return allTransactions[i].BlockNumber > allTransactions[j].BlockNumber
	})

	// Calculate pagination
	totalCount := len(allTransactions)

	// Apply pagination
	startIdx := (page - 1) * limit
	endIdx := startIdx + limit

	if startIdx >= totalCount {
		// Page out of range
		return &response.TLCTransactionHistoryResponse{
			Transactions: []response.TLCTransactionItem{},
			TotalCount:   totalCount,
			Page:         page,
			Limit:        limit,
		}, nil
	}

	if endIdx > totalCount {
		endIdx = totalCount
	}

	paginatedTransactions := allTransactions[startIdx:endIdx]

	return &response.TLCTransactionHistoryResponse{
		Transactions: paginatedTransactions,
		TotalCount:   totalCount,
		Page:         page,
		Limit:        limit,
	}, nil
}

// GetTransactionsByAddress gets all transactions for a specific address with pagination
func (s *BlockchainExplorerService) GetTransactionsByAddress(address string, page, limit int) (*response.TLCTransactionHistoryResponse, error) {
	// Default pagination values
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// Get all transactions first (without pagination)
	allTx, err := s.GetAllTransactions(nil, nil, 1, 999999) // Get all data
	if err != nil {
		return nil, err
	}

	var userTransactions []response.TLCTransactionItem

	// Filter transactions involving this address
	for _, tx := range allTx.Transactions {
		if strings.EqualFold(tx.FromAddress, address) || strings.EqualFold(tx.ToAddress, address) {
			userTransactions = append(userTransactions, tx)
		}
	}

	// Calculate pagination
	totalCount := len(userTransactions)

	// Apply pagination
	startIdx := (page - 1) * limit
	endIdx := startIdx + limit

	if startIdx >= totalCount {
		return &response.TLCTransactionHistoryResponse{
			Transactions: []response.TLCTransactionItem{},
			TotalCount:   totalCount,
			Page:         page,
			Limit:        limit,
		}, nil
	}

	if endIdx > totalCount {
		endIdx = totalCount
	}

	paginatedTransactions := userTransactions[startIdx:endIdx]

	return &response.TLCTransactionHistoryResponse{
		Transactions: paginatedTransactions,
		TotalCount:   totalCount,
		Page:         page,
		Limit:        limit,
	}, nil
}

// GetTransactionByHash gets specific transaction by hash
func (s *BlockchainExplorerService) GetTransactionByHash(txHash string) (*response.TLCTransactionItem, error) {
	hash := common.HexToHash(txHash)

	// Get transaction receipt
	receipt, err := s.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %v", err)
	}

	// Parse logs in this transaction (prioritas custom event)
	var selectedLog *types.Log
	hasCustomEvent := false

	// Cari custom event dulu
	for i := range receipt.Logs {
		if receipt.Logs[i].Address == s.contractAddress {
			eventSig := receipt.Logs[i].Topics[0].Hex()
			
			if eventSig == s.contractABI.Events["TokensMinted"].ID.Hex() ||
				eventSig == s.contractABI.Events["TokensBurned"].ID.Hex() ||
				eventSig == s.contractABI.Events["PaymentProcessed"].ID.Hex() ||
				eventSig == s.contractABI.Events["TopupRequested"].ID.Hex() ||
				eventSig == s.contractABI.Events["WithdrawRequested"].ID.Hex() {
				selectedLog = receipt.Logs[i]
				hasCustomEvent = true
				break
			}
		}
	}

	// Kalau tidak ada custom event, cari Transfer
	if !hasCustomEvent {
		for i := range receipt.Logs {
			if receipt.Logs[i].Address == s.contractAddress {
				if receipt.Logs[i].Topics[0].Hex() == s.contractABI.Events["Transfer"].ID.Hex() {
					selectedLog = receipt.Logs[i]
					break
				}
			}
		}
	}

	// Parse log terpilih
	if selectedLog != nil {
		tx, err := s.parseLogToTransaction(*selectedLog)
		if err == nil && tx != nil {
			return tx, nil
		}
	}

	return nil, fmt.Errorf("no TLC transaction found in hash %s", txHash)
}

// parseLogToTransaction converts blockchain log to transaction item
func (s *BlockchainExplorerService) parseLogToTransaction(vLog types.Log) (*response.TLCTransactionItem, error) {
	// Get block info for timestamp
	block, err := s.client.BlockByNumber(context.Background(), big.NewInt(int64(vLog.BlockNumber)))
	if err != nil {
		return nil, err
	}

	timestamp := time.Unix(int64(block.Time()), 0)

	// Parse different event types
	switch vLog.Topics[0].Hex() {
	case s.contractABI.Events["TokensMinted"].ID.Hex():
		return s.parseTokensMinted(vLog, timestamp)
	case s.contractABI.Events["TokensBurned"].ID.Hex():
		return s.parseTokensBurned(vLog, timestamp)
	case s.contractABI.Events["PaymentProcessed"].ID.Hex():
		return s.parsePaymentProcessed(vLog, timestamp)
	case s.contractABI.Events["TopupRequested"].ID.Hex():
		return s.parseTopupRequested(vLog, timestamp)
	case s.contractABI.Events["WithdrawRequested"].ID.Hex():
		return s.parseWithdrawRequested(vLog, timestamp)
	case s.contractABI.Events["Transfer"].ID.Hex():
		return s.parseTransfer(vLog, timestamp)
	}

	return nil, fmt.Errorf("unknown event type")
}

// Parse TokensMinted event
func (s *BlockchainExplorerService) parseTokensMinted(vLog types.Log, timestamp time.Time) (*response.TLCTransactionItem, error) {
	event := struct {
		To        common.Address
		Amount    *big.Int
		RequestId [32]byte
	}{}

	err := s.contractABI.UnpackIntoInterface(&event, "TokensMinted", vLog.Data)
	if err != nil {
		return nil, err
	}

	// Get indexed parameter (to address) from topics
	if len(vLog.Topics) > 1 {
		event.To = common.HexToAddress(vLog.Topics[1].Hex())
	}

	// Convert amount from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountTLC := new(big.Int).Div(event.Amount, divisor)

	return &response.TLCTransactionItem{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		FromAddress: "0x0000000000000000000000000000000000000000",
		ToAddress:   event.To.Hex(),
		Amount:      amountTLC.String(),
		AmountWei:   event.Amount.String(),
		Type:        "topup",
		Status:      "confirmed",
		GasUsed:     0,
		GasPrice:    "0",
		Memo:        fmt.Sprintf("Tokens minted - Request ID: %x", event.RequestId),
		Timestamp:   timestamp,
	}, nil
}

// Parse TokensBurned event
func (s *BlockchainExplorerService) parseTokensBurned(vLog types.Log, timestamp time.Time) (*response.TLCTransactionItem, error) {
	event := struct {
		From      common.Address
		Amount    *big.Int
		RequestId [32]byte
	}{}

	err := s.contractABI.UnpackIntoInterface(&event, "TokensBurned", vLog.Data)
	if err != nil {
		return nil, err
	}

	// Get indexed parameter (from address) from topics
	if len(vLog.Topics) > 1 {
		event.From = common.HexToAddress(vLog.Topics[1].Hex())
	}

	// Convert amount from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountTLC := new(big.Int).Div(event.Amount, divisor)

	return &response.TLCTransactionItem{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		FromAddress: event.From.Hex(),
		ToAddress:   "0x0000000000000000000000000000000000000000",
		Amount:      amountTLC.String(),
		AmountWei:   event.Amount.String(),
		Type:        "withdraw",
		Status:      "confirmed",
		GasUsed:     0,
		GasPrice:    "0",
		Memo:        fmt.Sprintf("Tokens burned - Request ID: %x", event.RequestId),
		Timestamp:   timestamp,
	}, nil
}

// Parse PaymentProcessed event
func (s *BlockchainExplorerService) parsePaymentProcessed(vLog types.Log, timestamp time.Time) (*response.TLCTransactionItem, error) {
	event := struct {
		From   common.Address
		To     common.Address
		Amount *big.Int
		Note   string
	}{}

	err := s.contractABI.UnpackIntoInterface(&event, "PaymentProcessed", vLog.Data)
	if err != nil {
		return nil, err
	}

	// Get indexed parameters from topics
	if len(vLog.Topics) > 1 {
		event.From = common.HexToAddress(vLog.Topics[1].Hex())
	}
	if len(vLog.Topics) > 2 {
		event.To = common.HexToAddress(vLog.Topics[2].Hex())
	}

	// Convert amount from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountTLC := new(big.Int).Div(event.Amount, divisor)

	return &response.TLCTransactionItem{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		FromAddress: event.From.Hex(),
		ToAddress:   event.To.Hex(),
		Amount:      amountTLC.String(),
		AmountWei:   event.Amount.String(),
		Type:        "transfer",
		Status:      "confirmed",
		GasUsed:     0,
		GasPrice:    "0",
		Memo:        event.Note,
		Timestamp:   timestamp,
	}, nil
}

// Parse TopupRequested event
func (s *BlockchainExplorerService) parseTopupRequested(vLog types.Log, timestamp time.Time) (*response.TLCTransactionItem, error) {
	event := struct {
		User         common.Address
		Amount       *big.Int
		RequestId    [32]byte
		PaymentProof string
	}{}

	err := s.contractABI.UnpackIntoInterface(&event, "TopupRequested", vLog.Data)
	if err != nil {
		return nil, err
	}

	// Get indexed parameter (user address) from topics
	if len(vLog.Topics) > 1 {
		event.User = common.HexToAddress(vLog.Topics[1].Hex())
	}

	// Convert amount from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountTLC := new(big.Int).Div(event.Amount, divisor)

	return &response.TLCTransactionItem{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		FromAddress: "0x0000000000000000000000000000000000000000",
		ToAddress:   event.User.Hex(),
		Amount:      amountTLC.String(),
		AmountWei:   event.Amount.String(),
		Type:        "topup",
		Status:      "pending",
		GasUsed:     0,
		GasPrice:    "0",
		Memo:        fmt.Sprintf("Topup request - Proof: %s", event.PaymentProof),
		Timestamp:   timestamp,
	}, nil
}

// Parse WithdrawRequested event
func (s *BlockchainExplorerService) parseWithdrawRequested(vLog types.Log, timestamp time.Time) (*response.TLCTransactionItem, error) {
	event := struct {
		User        common.Address
		Amount      *big.Int
		RequestId   [32]byte
		BankAccount string
	}{}

	err := s.contractABI.UnpackIntoInterface(&event, "WithdrawRequested", vLog.Data)
	if err != nil {
		return nil, err
	}

	// Get indexed parameter (user address) from topics
	if len(vLog.Topics) > 1 {
		event.User = common.HexToAddress(vLog.Topics[1].Hex())
	}

	// Convert amount from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountTLC := new(big.Int).Div(event.Amount, divisor)

	return &response.TLCTransactionItem{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		FromAddress: event.User.Hex(),
		ToAddress:   "0x0000000000000000000000000000000000000000",
		Amount:      amountTLC.String(),
		AmountWei:   event.Amount.String(),
		Type:        "withdraw",
		Status:      "confirmed",
		GasUsed:     0,
		GasPrice:    "0",
		Memo:        fmt.Sprintf("Withdrawal to: %s", event.BankAccount),
		Timestamp:   timestamp,
	}, nil
}

// Parse Transfer event (standard ERC20)
func (s *BlockchainExplorerService) parseTransfer(vLog types.Log, timestamp time.Time) (*response.TLCTransactionItem, error) {
	if len(vLog.Topics) < 3 {
		return nil, fmt.Errorf("invalid transfer event")
	}

	from := common.HexToAddress(vLog.Topics[1].Hex())
	to := common.HexToAddress(vLog.Topics[2].Hex())

	// Amount is in data
	amount := new(big.Int).SetBytes(vLog.Data)

	// Convert amount from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountTLC := new(big.Int).Div(amount, divisor)

	// Skip zero transfers
	if amount.Cmp(big.NewInt(0)) == 0 {
		return nil, fmt.Errorf("zero transfer")
	}

	txType := "transfer"
	if from.Hex() == "0x0000000000000000000000000000000000000000" {
		txType = "mint"
	} else if to.Hex() == "0x0000000000000000000000000000000000000000" {
		txType = "burn"
	}

	return &response.TLCTransactionItem{
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: vLog.BlockNumber,
		FromAddress: from.Hex(),
		ToAddress:   to.Hex(),
		Amount:      amountTLC.String(),
		AmountWei:   amount.String(),
		Type:        txType,
		Status:      "confirmed",
		GasUsed:     0,
		GasPrice:    "0",
		Memo:        "Standard transfer",
		Timestamp:   timestamp,
	}, nil
}

func (s *BlockchainExplorerService) GetBlockchainStats() (*response.BlockchainStatsResponse, error) {
	// Get all transactions (no pagination untuk akurasi stats)
	allTx, err := s.GetAllTransactions(nil, nil, 1, 999999)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	var totalMinted big.Int
	var totalBurned big.Int
	uniqueAddresses := make(map[string]bool)

	// Calculate stats from all transactions
	for _, tx := range allTx.Transactions {
		// Parse amount to big.Int
		amount := new(big.Int)
		amount.SetString(tx.Amount, 10)

		switch tx.Type {
		case "mint":
			totalMinted.Add(&totalMinted, amount)
		case "burn":
			totalBurned.Add(&totalBurned, amount)
		}

		// Track unique addresses (exclude zero address)
		zeroAddress := "0x0000000000000000000000000000000000000000"
		if tx.FromAddress != zeroAddress {
			uniqueAddresses[strings.ToLower(tx.FromAddress)] = true
		}
		if tx.ToAddress != zeroAddress {
			uniqueAddresses[strings.ToLower(tx.ToAddress)] = true
		}
	}

	// Calculate circulating supply
	circulatingSupply := new(big.Int).Sub(&totalMinted, &totalBurned)

	// Get latest block number
	latestBlock, err := s.client.BlockNumber(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block: %v", err)
	}

	return &response.BlockchainStatsResponse{
		TotalMinted:       totalMinted.String(),
		TotalBurned:       totalBurned.String(),
		CirculatingSupply: circulatingSupply.String(),
		MaxSupply:         "Unlimited",
		TotalTransactions: allTx.TotalCount,
		TotalAddresses:    len(uniqueAddresses),
		LatestBlock:       latestBlock,
	}, nil
}