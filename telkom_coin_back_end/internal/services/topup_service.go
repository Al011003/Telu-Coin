package service

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"telkom_coin_back_end/internal/blockchain"
	"telkom_coin_back_end/internal/dto/request"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/models"
	"telkom_coin_back_end/internal/repository"
	"telkom_coin_back_end/internal/web3"
	"telkom_coin_back_end/pkg/crypto"
	"time"

	"gorm.io/gorm"
)

type TopupService struct {
	userRepo        *repository.UserRepository
	balanceRepo     *repository.BalanceRepository
	txRepo          *repository.TransactionRepository
	contractService *web3.ContractService
}

func NewTopupService(
	userRepo *repository.UserRepository,
	balanceRepo *repository.BalanceRepository,
	txRepo *repository.TransactionRepository,
	blockchainService *blockchain.BlockchainService,
) *TopupService {
	var contractService *web3.ContractService
	if blockchainService.IsWeb3Enabled() {
		web3Client, err := web3.NewWeb3Client()
		if err == nil {
			contractService, _ = web3.NewContractService(web3Client)
		}
	}

	return &TopupService{
		userRepo:        userRepo,
		balanceRepo:     balanceRepo,
		txRepo:          txRepo,
		contractService: contractService, // <-- Pastikan ini diisi
	}
}

// RequestTopup - Pure blockchain topup (no MySQL storage, only blockchain)
func (s *TopupService) RequestTopup(userID int64, req *request.TopupRequest) (*response.TopupResponse, error) {
	// === 1️⃣ Ambil user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// === 2️⃣ Verifikasi PIN
	if !crypto.VerifyPin(req.Pin, user.PinHash) {
		return nil, errors.New("invalid PIN")
	}

	// === 3️⃣ Validasi nominal
	amountBigInt := new(big.Int)
	amountBigInt.SetString(req.Amount, 10)
	if amountBigInt.Cmp(big.NewInt(10000)) < 0 {
		return nil, errors.New("minimum topup amount is 10,000")
	}

	// Konversi ke wei (18 desimal)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountWei := new(big.Int).Mul(amountBigInt, multiplier)

	// === 4️⃣ Siapkan koneksi blockchain
	blockchainService, err := blockchain.NewBlockchainService("ganache-local")
	if err != nil {
		return nil, errors.New("blockchain service unavailable")
	}
	if !blockchainService.IsWeb3Enabled() {
		return nil, errors.New("blockchain not available")
	}

	// === 5️⃣ Dekripsi private key user
	privateKey, err := crypto.DecryptPrivateKey(user.PrivateKeyEncrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt private key")
	}

	// === 6️⃣ Eksekusi transaksi ke blockchain
	web3Client, err := web3.NewWeb3Client()
	if err != nil {
		return nil, errors.New("failed to connect to blockchain")
	}

	contractService, err := web3.NewContractService(web3Client)
	if err != nil {
		return nil, errors.New("failed to initialize contract service")
	}

	paymentProof := fmt.Sprintf("AUTO_TOPUP_%s_%s", req.PaymentMethod, req.Amount)
	txHash, err := contractService.InstantTopup(privateKey, amountWei, paymentProof)
	if err != nil {
		return nil, fmt.Errorf("blockchain topup failed: %v", err)
	}

	// === 7️⃣ Update balance user di DB
	err = s.balanceRepo.AddBalance(user.ID, req.Amount)
	if err != nil {
		log.Printf("⚠️ failed to update balance after topup: %v", err)
	}

	// === 8️⃣ Simpan transaksi ke DB
	now := time.Now()
	tx := &models.Transaction{
		TxHash:      txHash,
		FromAddress: "0x0000000000000000000000000000000000000000", // dianggap dari sistem
		ToAddress:   user.WalletAddress,
		Amount:      req.Amount,
		TxType:      "topup",
		Status:      "confirmed",
		BlockNumber: nil,
		GasUsed:     nil,
		GasPrice:    nil,
		Nonce:       nil,
		Metadata: models.TransactionMetadata{
			"payment_method": req.PaymentMethod,
			"note":           "Blockchain topup synced to DB",
			"auto_confirmed": true,
		},
		CreatedAt:   now,
		ConfirmedAt: &now,
	}

	if err := s.txRepo.Create(tx); err != nil {
		log.Printf("⚠️ failed to save transaction: %v", err)
	}

	// === 9️⃣ Return response
	return &response.TopupResponse{
		ID:            tx.ID,
		Amount:        req.Amount,
		Currency:      "TLC",
		PaymentMethod: req.PaymentMethod,
		Status:        "confirmed",
		TxHash:        txHash,
		PaymentDetails: map[string]interface{}{
			"method":          req.PaymentMethod,
			"blockchain_only": false,
			"synced_to_db":    true,
			"note":            "Topup confirmed & recorded",
		},
		CreatedAt: now,
	}, nil
}

func (s *TopupService) GetTopupHistory(userID int64, page, limit int) (*response.TopupHistoryResponse, error) {
	// 1. Get user untuk mendapatkan alamat wallet-nya
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 2. Panggil fungsi repository untuk mengambil data dari DATABASE
	transactions, totalCount, err := s.txRepo.FindTopupHistoryByWalletAddress(user.WalletAddress, page, limit)
	if err != nil {
		log.Printf("[ERROR] Failed to get topup history from DB: %v", err)
		return nil, errors.New("failed to retrieve topup history")
	}

	// 3. Konversi dari model database ke DTO response
	var topupResponses []response.TopupResponse
	for _, tx := range transactions {
		var paymentMethod string
		if method, ok := tx.Metadata["payment_method"].(string); ok {
			paymentMethod = method
		}

		topup := response.TopupResponse{
			ID:            tx.ID, // Gunakan TxHash sebagai ID
			Amount:        tx.Amount,
			Status:        tx.Status,
			TxHash:        tx.TxHash,
			PaymentMethod: paymentMethod,
			CreatedAt:     tx.CreatedAt,
		}
		topupResponses = append(topupResponses, topup)
	}

	// 4. Hitung total halaman untuk pagination
	totalPages := (int(totalCount) + limit - 1) / limit

	// 5. Kembalikan response yang sudah jadi
	return &response.TopupHistoryResponse{
		Topups:     topupResponses,
		TotalCount: int(totalCount),
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
