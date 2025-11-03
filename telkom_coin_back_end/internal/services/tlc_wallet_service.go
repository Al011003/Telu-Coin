package service

import (
	"errors"
	"log"
	"math/big" // 👈 1. DITAMBAHKAN import "strings"
	"strconv"
	"strings"
	"telkom_coin_back_end/internal/blockchain"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/models"
	"telkom_coin_back_end/internal/repository"
	"telkom_coin_back_end/internal/web3"
	"telkom_coin_back_end/pkg/crypto"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type TLCWalletService struct {
	userRepo          *repository.UserRepository
	balanceRepo       *repository.BalanceRepository
	txRepo            *repository.TransactionRepository
	blockchainService *blockchain.BlockchainService
	contractService   *web3.ContractService
	explorerService   *BlockchainExplorerService // 👈 TAMBAHKAN INI
}

func NewTLCWalletService(
	userRepo *repository.UserRepository,
	balanceRepo *repository.BalanceRepository,
	txRepo *repository.TransactionRepository,
	blockchainService *blockchain.BlockchainService,
	explorerService *BlockchainExplorerService, // 👈 TAMBAHKAN PARAMETER INI
) *TLCWalletService {
	// Initialize contract service if blockchain is enabled
	var contractService *web3.ContractService
	if blockchainService.IsWeb3Enabled() {
		web3Client, err := web3.NewWeb3Client()
		if err == nil {
			contractService, _ = web3.NewContractService(web3Client)
		}
	}

	return &TLCWalletService{
		userRepo:          userRepo,
		balanceRepo:       balanceRepo,
		txRepo:            txRepo,
		blockchainService: blockchainService,
		contractService:   contractService,
		explorerService:   explorerService, // 👈 ASSIGN DI SINI
	}
}

// GetBlockchainBalance gets real-time TLC balance from smart contract
func (s *TLCWalletService) GetBlockchainBalance(walletAddress string) (*response.TLCBalanceResponse, error) {
	if !s.blockchainService.IsWeb3Enabled() || s.contractService == nil {
		return &response.TLCBalanceResponse{
			WalletAddress: walletAddress,
			Balance:       "0",
			TokenSymbol:   "TLC",
			TokenName:     "Telkom Coin",
			Decimals:      18,
			UpdatedAt:     time.Now(),
		}, nil
	}

	// Get balance from smart contract
	balance, err := s.contractService.GetBalance(common.HexToAddress(walletAddress))
	if err != nil {
		return nil, errors.New("failed to get balance from blockchain: " + err.Error())
	}

	// Convert from wei to TLC (divide by 10^18)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	balanceInTLC := new(big.Int).Div(balance, divisor)

	return &response.TLCBalanceResponse{
		WalletAddress: walletAddress,
		Balance:       balanceInTLC.String(),
		BalanceWei:    balance.String(),
		TokenSymbol:   "TLC",
		TokenName:     "Telkom Coin",
		Decimals:      18,
		UpdatedAt:     time.Now(),
	}, nil
}

func (s *TLCWalletService) ValidateTransfer(userID int64, toAddress, amount string) (*response.ValidateTransferResponse, error) {
	// Validasi user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// Validasi format alamat
	if !common.IsHexAddress(toAddress) {
		return nil, errors.New("invalid recipient address format")
	}
	if toAddress == user.WalletAddress {
		return nil, errors.New("cannot transfer to yourself")
	}

	// Validasi amount format
	amountToken := new(big.Int)
	amountToken, ok := amountToken.SetString(amount, 10)
	if !ok {
		return nil, errors.New("invalid amount format")
	}
	if amountToken.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	// Cek saldo on-chain
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountWei := new(big.Int).Mul(amountToken, multiplier)

	if s.contractService == nil || !s.blockchainService.IsWeb3Enabled() {
		return nil, errors.New("blockchain not available")
	}

	senderOnChainWei, err := s.contractService.GetBalance(common.HexToAddress(user.WalletAddress))
	if err != nil {
		return nil, errors.New("failed to check on-chain balance: " + err.Error())
	}
	if senderOnChainWei.Cmp(amountWei) < 0 {
		return nil, errors.New("insufficient on-chain balance")
	}

	// 🔑 AMBIL NAMA PENERIMA DARI DATABASE
	recipientUser, err := s.userRepo.GetByWalletAddress(toAddress)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("recipient address not found in system")
		}
		return nil, err
	}

	// Response untuk popup konfirmasi
	return &response.ValidateTransferResponse{
		FromAddress:   user.WalletAddress,
		FromUsername:  user.Username,
		ToAddress:     toAddress,
		ToUsername:    recipientUser.Username, // 👈 Nama penerima
		Amount:        amount,
		AmountWei:     amountWei.String(),
		SenderBalance: senderOnChainWei.String(),
		IsValid:       true,
	}, nil
}

// ============================================================================
// ENDPOINT 2: TransferTLC - Execute transfer setelah user confirm + input PIN
// ============================================================================
func (s *TLCWalletService) TransferTLC(userID int64, toAddress, amount, memo, pin string) (*response.TLCTransferResponse, error) {
	// Step 1: Validasi user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// Step 2: Verifikasi PIN
	if !crypto.VerifyPin(pin, user.PinHash) {
		return nil, errors.New("invalid PIN")
	}

	// Step 3: Validasi format alamat
	if !common.IsHexAddress(toAddress) {
		return nil, errors.New("invalid recipient address format")
	}
	if toAddress == user.WalletAddress {
		return nil, errors.New("cannot transfer to yourself")
	}

	// Step 4: Validasi amount
	amountToken := new(big.Int)
	amountToken, ok := amountToken.SetString(amount, 10)
	if !ok {
		return nil, errors.New("invalid amount format")
	}
	if amountToken.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.New("amount must be greater than 0")
	}

	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountWei := new(big.Int).Mul(amountToken, multiplier)

	// Step 5: Cek blockchain availability
	if s.contractService == nil || !s.blockchainService.IsWeb3Enabled() {
		return nil, errors.New("blockchain not available")
	}

	// Step 6: Double-check saldo on-chain
	senderOnChainWei, err := s.contractService.GetBalance(common.HexToAddress(user.WalletAddress))
	if err != nil {
		return nil, errors.New("failed to check on-chain balance: " + err.Error())
	}
	if senderOnChainWei.Cmp(amountWei) < 0 {
		return nil, errors.New("insufficient on-chain balance")
	}

	// Step 7: Decrypt private key
	privateKey, err := crypto.DecryptPrivateKey(user.PrivateKeyEncrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt private key")
	}

	// Step 8: Debug logging
	log.Println("====================== [ DEBUG INFO TLCWalletService ] ======================")
	log.Printf("[DEBUG] User ID: %d", userID)
	log.Printf("[DEBUG] From Address: %s", user.WalletAddress)
	log.Printf("[DEBUG] To Address: %s", toAddress)
	log.Printf("[DEBUG] Amount to send (Wei): %s", amountWei.String())
	log.Printf("[DEBUG] Memo: %s", memo)
	log.Println("============================================================================")

	// Step 9: Execute transfer on-chain
	txHash, err := s.contractService.Transfer(privateKey, common.HexToAddress(toAddress), amountWei)
	if err != nil {
		log.Printf("[FATAL] Error from contractService.Transfer: %v", err)
		return nil, errors.New("blockchain transfer failed: " + err.Error())
	}

	// Step 10: Wait for receipt
	log.Printf("⏳ Waiting for tx %s to be mined...", txHash)
	receipt, err := s.contractService.WaitForReceipt(txHash, 60)
	if err != nil {
		log.Printf("⚠️ Transaction %s sent but not confirmed or failed: %v", txHash, err)
		return nil, errors.New("transaction not confirmed or failed: " + err.Error())
	}

	log.Printf("✅ Transaction confirmed in block %d", receipt.BlockNumber.Uint64())

	// Step 11: Save transaction record
	blockNum := int64(receipt.BlockNumber.Uint64())
	txRecord := &models.Transaction{
		TxHash:      txHash,
		FromAddress: user.WalletAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		TxType:      "transfer",
		Status:      "confirmed",
		BlockNumber: &blockNum,
		CreatedAt:   time.Now(),
		ConfirmedAt: func() *time.Time {
			t := time.Now()
			return &t
		}(),
	}

	if err := s.txRepo.Create(txRecord); err != nil {
		log.Printf("[ERROR] Failed to save transaction record: %v", err)
	}

	// Step 12: Return response
	return &response.TLCTransferResponse{
		TxHash:      txHash,
		FromAddress: user.WalletAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		AmountWei:   amountWei.String(),
		Memo:        memo,
		Status:      "confirmed",
		Timestamp:   time.Now(),
	}, nil
}

func (s *TLCWalletService) GetTransactionHistory(userID int64, page, limit int) (*response.TLCTransactionHistoryResponse, error) {
	// 1. Dapatkan user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// ✅ 2. AMBIL DARI BLOCKCHAIN menggunakan explorerService
	result, err := s.explorerService.GetTransactionsByAddress(user.WalletAddress, page, limit)
	if err != nil {
		log.Printf("[WARN] Failed to get transactions from blockchain: %v", err)
		// Fallback ke DB jika blockchain error
		return s.getTransactionHistoryFromDB(user.WalletAddress, page, limit)
	}

	// ✅ 3. TAMBAHKAN DIRECTION untuk setiap transaksi
	for i := range result.Transactions {
		tx := &result.Transactions[i]
		
		// Tentukan direction based on user wallet
		if tx.Type == "transfer" {
			if strings.EqualFold(tx.FromAddress, user.WalletAddress) {
				tx.Direction = "outgoing"
			} else if strings.EqualFold(tx.ToAddress, user.WalletAddress) {
				tx.Direction = "incoming"
			}
		} else {
			tx.Direction = "" // topup/mint/burn tidak ada direction
		}

		log.Printf("[DEBUG] TxHash=%s | Type=%s | From=%s | To=%s | User=%s | Direction=%s",
			tx.TxHash, tx.Type, tx.FromAddress, tx.ToAddress, user.WalletAddress, tx.Direction)
	}

	log.Printf("[SUCCESS] Retrieved %d transactions from blockchain for user %d", len(result.Transactions), userID)
	return result, nil
}

// ✅ 4. HELPER METHOD: Fallback ke DB jika blockchain error
func (s *TLCWalletService) getTransactionHistoryFromDB(walletAddress string, page, limit int) (*response.TLCTransactionHistoryResponse, error) {
	log.Printf("[INFO] Falling back to database for wallet %s", walletAddress)
	
	transactions, totalCount, err := s.txRepo.FindHistoryByWalletAddress(walletAddress, page, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to get transaction history from DB: %v", err)
		return nil, errors.New("failed to retrieve transaction history")
	}

	var historyItems []response.TLCTransactionItem
	for _, tx := range transactions {
		// Tentukan arah transaksi
		var direction string
		if tx.TxType == "transfer" {
			direction = "outgoing"
			if tx.ToAddress == walletAddress {
				direction = "incoming"
			}
		} else {
			direction = ""
		}

		// Konversi Amount ke Wei
		amountToken, _ := new(big.Int).SetString(tx.Amount, 10)
		multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
		amountWei := new(big.Int).Mul(amountToken, multiplier)

		// Ambil data dari pointer dengan aman
		var blockNumber, gasUsed uint64
		var gasPrice string
		if tx.BlockNumber != nil {
			blockNumber = uint64(*tx.BlockNumber)
		}
		if tx.GasUsed != nil {
			gasUsed = uint64(*tx.GasUsed)
		}
		if tx.GasPrice != nil {
			gasPrice = strconv.FormatInt(*tx.GasPrice, 10)
		}

		// Ambil memo dari metadata dengan aman
		var memo string
		if memoVal, ok := tx.Metadata["memo"].(string); ok {
			memo = memoVal
		}

		// Buat item response
		item := response.TLCTransactionItem{
			TxHash:      tx.TxHash,
			BlockNumber: blockNumber,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			Amount:      tx.Amount,
			AmountWei:   amountWei.String(),
			Type:        tx.TxType,
			Status:      tx.Status,
			GasUsed:     gasUsed,
			GasPrice:    gasPrice,
			Memo:        memo,
			Timestamp:   tx.CreatedAt,
			Direction:   direction,
		}

		historyItems = append(historyItems, item)
	}

	return &response.TLCTransactionHistoryResponse{
		Transactions: historyItems,
		TotalCount:   int(totalCount),
		Page:         page,
		Limit:        limit,
	}, nil
}