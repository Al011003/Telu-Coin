package service

import (
	"errors"
	"math/big"
	"telkom_coin_back_end/internal/blockchain"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/repository"
	"telkom_coin_back_end/internal/web3"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

type BalanceService struct {
	balanceRepo       *repository.BalanceRepository
	userRepo          *repository.UserRepository
	blockchainService *blockchain.BlockchainService
	contractService   *web3.ContractService
}

func NewBalanceService(balanceRepo *repository.BalanceRepository, userRepo *repository.UserRepository) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		userRepo:    userRepo,
	}
}

// NewEnhancedBalanceService creates a balance service with blockchain integration
func NewEnhancedBalanceService(balanceRepo *repository.BalanceRepository, blockchainService *blockchain.BlockchainService) *BalanceService {
	var contractService *web3.ContractService
	if blockchainService.IsWeb3Enabled() {
		web3Client, err := web3.NewWeb3Client()
		if err == nil {
			contractService, _ = web3.NewContractService(web3Client)
		}
	}

	return &BalanceService{
		balanceRepo:       balanceRepo,
		blockchainService: blockchainService,
		contractService:   contractService,
	}
}

// Get balance by user ID from blockchain only
func (s *BalanceService) GetBalance(userID int64) (*response.BalanceResponse, error) {
	// Get user to get wallet address
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Create blockchain service
	blockchainService, err := blockchain.NewBlockchainService("ganache-local")
	if err != nil {
		return nil, errors.New("blockchain service unavailable")
	}

	if !blockchainService.IsWeb3Enabled() {
		return nil, errors.New("blockchain not available - pure blockchain mode required")
	}

	// Get balance from blockchain
	balanceWei, err := blockchainService.GetTokenBalance(user.WalletAddress)
	if err != nil {
		return nil, errors.New("failed to get balance from blockchain: " + err.Error())
	}

	// Convert from wei to TLC
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	balanceTLC := new(big.Int).Div(balanceWei, divisor)

	return &response.BalanceResponse{
		Balance:       balanceTLC.String(),
		LockedBalance: "0", // No locked balance in pure blockchain mode
		Available:     balanceTLC.String(),
		UpdatedAt:     time.Now(),
	}, nil
}

// Get balance by wallet address
func (s *BalanceService) GetBalanceByAddress(address string) (*response.BalanceResponse, error) {
	balance, err := s.balanceRepo.GetByWalletAddress(address)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("balance not found")
		}
		return nil, err
	}

	// Calculate available balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	lockedBalance := new(big.Int)
	lockedBalance.SetString(balance.LockedBalance, 10)

	availableBalance := new(big.Int).Sub(currentBalance, lockedBalance)

	return &response.BalanceResponse{
		Balance:       balance.Balance,
		LockedBalance: balance.LockedBalance,
		Available:     availableBalance.String(),
		UpdatedAt:     balance.UpdatedAt,
	}, nil
}

// Check if balance is sufficient
func (s *BalanceService) IsSufficientBalance(userID int64, amount string) (bool, error) {
	balance, err := s.balanceRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}

	// Get available balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	lockedBalance := new(big.Int)
	lockedBalance.SetString(balance.LockedBalance, 10)

	availableBalance := new(big.Int).Sub(currentBalance, lockedBalance)

	// Compare with amount
	requiredAmount := new(big.Int)
	_, ok := requiredAmount.SetString(amount, 10)
	if !ok {
		return false, errors.New("invalid amount format")
	}

	return availableBalance.Cmp(requiredAmount) >= 0, nil
}

// GetBlockchainBalance gets real-time balance from blockchain
func (s *BalanceService) GetBlockchainBalance(walletAddress string) (*response.TLCBalanceResponse, error) {
	if s.blockchainService == nil || !s.blockchainService.IsWeb3Enabled() || s.contractService == nil {
		return &response.TLCBalanceResponse{
			WalletAddress: walletAddress,
			Balance:       "0",
			BalanceWei:    "0",
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

// CheckBlockchainBalance checks if user has sufficient balance on blockchain
func (s *BalanceService) CheckBlockchainBalance(walletAddress, amount string) (bool, error) {
	if s.blockchainService == nil || !s.blockchainService.IsWeb3Enabled() || s.contractService == nil {
		return false, errors.New("blockchain not available")
	}

	// Get balance from smart contract
	balance, err := s.contractService.GetBalance(common.HexToAddress(walletAddress))
	if err != nil {
		return false, errors.New("failed to get balance from blockchain: " + err.Error())
	}

	// Convert amount to wei
	amountBigInt := new(big.Int)
	amountBigInt.SetString(amount, 10)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amountWei := new(big.Int).Mul(amountBigInt, multiplier)

	// Compare balance with required amount
	return balance.Cmp(amountWei) >= 0, nil
}
