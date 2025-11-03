package service

import (
	"errors"
	"log"
	"math/big"

	// "telkom_coin_back_end/internal/dto/request" // Tidak dipakai lagi di service ini
	"telkom_coin_back_end/internal/blockchain"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/models"
	"telkom_coin_back_end/internal/repository"
	"telkom_coin_back_end/internal/web3"
	"telkom_coin_back_end/pkg/crypto"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
	// Pastikan semua import lain yang dibutuhkan ada
)

type WithdrawService struct {
	userRepo        *repository.UserRepository
	balanceRepo     *repository.BalanceRepository
	txRepo          *repository.TransactionRepository
	contractService *web3.ContractService
	// Hapus authService jika hanya untuk verifikasi PIN, karena kita bisa lakukan di sini
}

// Constructor disederhanakan
func NewWithdrawService(
	userRepo *repository.UserRepository,
	balanceRepo *repository.BalanceRepository,
	txRepo *repository.TransactionRepository,
	blockchainService *blockchain.BlockchainService, // Diperlukan untuk contractService
) *WithdrawService {
	var contractService *web3.ContractService
	if blockchainService.IsWeb3Enabled() {
		web3Client, err := web3.NewWeb3Client()
		if err == nil {
			contractService, _ = web3.NewContractService(web3Client)
		}
	}

	return &WithdrawService{
		userRepo:        userRepo,
		balanceRepo:     balanceRepo,
		txRepo:          txRepo,
		contractService: contractService,
	}
}

/*
// =====================================================================================
//      👇 FUNGSI-FUNGSI UNTUK JALUR TERPUSAT KITA NONAKTIFKAN SEMENTARA 👇
// =====================================================================================
// Request withdrawal (Jalur Terpusat)
func (s *WithdrawService) RequestWithdraw(userID int64, req *request.WithdrawRequest) (*response.WithdrawResponse, error) {
    // ... dinonaktifkan
    return nil, errors.New("this withdrawal path is currently disabled")
}
// ProcessWithdraw (Jalur Terpusat)
func (s *WithdrawService) ProcessWithdraw(withdrawID int64) error {
    // ... dinonaktifkan
    return nil, errors.New("this withdrawal path is currently disabled")
}
// CancelWithdraw (Jalur Terpusat)
func (s *WithdrawService) CancelWithdraw(userID, withdrawID int64) error {
    // ... dinonaktifkan
    return nil, errors.New("this withdrawal path is currently disabled")
}
*/

// CreateWithdrawal adalah satu-satunya fungsi withdraw yang aktif.
// Sebelumnya bernama WithdrawTLCDirect.
func (s *WithdrawService) CreateWithdrawal(userID int64, amount, bankAccount, pin string) (*response.TLCWithdrawResponse, error) {
	// 1. Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 2. Check if user is active
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// 3. Verify PIN
	if !crypto.VerifyPin(pin, user.PinHash) {
		return nil, errors.New("invalid PIN")
	}

	// 4. Validate amount (minimum 1,000 TLC) - bisa diubah sesuai kebutuhan
	// Anda bisa memanggil minBurnAmount dari smart contract untuk nilai dinamis
	amountBigInt, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return nil, errors.New("invalid amount format")
	}

	minWithdraw := big.NewInt(1000) // Contoh nilai statis
	if amountBigInt.Cmp(minWithdraw) < 0 {
		return nil, errors.New("minimum withdrawal amount is 1,000 TLC")
	}

	// 5. Check blockchain availability
	if s.contractService == nil {
		return nil, errors.New("blockchain service not available")
	}

	// 6. Konversi ke wei
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountWei := new(big.Int).Mul(amountBigInt, multiplier)

	// 7. Check balance on blockchain
	currentBalance, err := s.contractService.GetBalance(common.HexToAddress(user.WalletAddress))
	if err != nil {
		return nil, errors.New("failed to check on-chain balance")
	}

	if currentBalance.Cmp(amountWei) < 0 {
		return nil, errors.New("insufficient on-chain balance")
	}

	// 8. Decrypt private key
	privateKey, err := crypto.DecryptPrivateKey(user.PrivateKeyEncrypted)
	if err != nil {
		return nil, errors.New("failed to decrypt private key")
	}

	// 9. Kirim transaksi ke blockchain (burn tokens)
	txHash, err := s.contractService.RequestWithdraw(privateKey, amountWei, bankAccount)
	if err != nil {
		return nil, errors.New("blockchain withdrawal request failed: " + err.Error())
	}

	txRecord := &models.Transaction{
		TxHash:      txHash,
		FromAddress: user.WalletAddress,
		ToAddress:   "", // kosong karena withdraw gak kirim ke address lain
		Amount:      amount,
		TxType:      "withdraw",
		Status:      "processing",
		CreatedAt:   time.Now(),
	}

	if err := s.txRepo.Create(txRecord); err != nil {
		log.Printf("[ERROR] Failed to save withdraw record: %v", err)
		// Bisa return error kalau mau strict
	}

	// 10. Return response sukses
	return &response.TLCWithdrawResponse{
		TxHash:        txHash, // ID unik dari transaksi blockchain
		Amount:        amount,
		AmountWei:     amountWei.String(),
		BankAccount:   bankAccount,
		Status:        "processing", // Status awal, karena token sudah di-burn & tinggal menunggu transfer fiat
		WalletAddress: user.WalletAddress,
		CreatedAt:     time.Now(),
	}, nil
}
