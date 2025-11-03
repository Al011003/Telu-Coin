package repository

import (
	"errors"
	"math/big"
	"telkom_coin_back_end/internal/models"

	"gorm.io/gorm"
)

// BalanceRepositoryInterface defines the contract for balance repository
type BalanceRepositoryInterface interface {
	Create(balance *models.Balance) error
	GetByUserID(userID int64) (*models.Balance, error)
	GetByWalletAddress(address string) (*models.Balance, error)
	Update(balance *models.Balance) error
	SetBalance(userID int64, newBalance string) error
	AddBalance(userID int64, amount string) error
	SubtractBalance(userID int64, amount string) error
	LockBalance(userID int64, amount string) error
	UnlockBalance(userID int64, amount string) error
	DeductLockedBalance(userID int64, amount string) error
	GetAvailableBalance(userID int64) (string, error)
	UpdateWithTx(tx *gorm.DB, balance *models.Balance) error
}

type BalanceRepository struct {
	db *gorm.DB
}

func NewBalanceRepository(db *gorm.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

// Create new balance
func (r *BalanceRepository) Create(balance *models.Balance) error {
	return r.db.Create(balance).Error
}

// Get balance by user ID
func (r *BalanceRepository) GetByUserID(userID int64) (*models.Balance, error) {
	var balance models.Balance
	err := r.db.Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// Get balance by wallet address
func (r *BalanceRepository) GetByWalletAddress(address string) (*models.Balance, error) {
	var balance models.Balance
	err := r.db.Where("wallet_address = ?", address).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// Update balance
func (r *BalanceRepository) Update(balance *models.Balance) error {
	return r.db.Save(balance).Error
}

// Set balance (direct update)
func (r *BalanceRepository) SetBalance(userID int64, newBalance string) error {
	return r.db.Model(&models.Balance{}).
		Where("user_id = ?", userID).
		Update("balance", newBalance).Error
}

// Add to balance
func (r *BalanceRepository) AddBalance(userID int64, amount string) error {
	// Get current balance
	balance, err := r.GetByUserID(userID)
	if err != nil {
		return err
	}

	// Convert to big.Int
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	addAmount := new(big.Int)
	addAmount.SetString(amount, 10)

	// Add
	newBalance := new(big.Int).Add(currentBalance, addAmount)

	// Update
	return r.SetBalance(userID, newBalance.String())
}

// Subtract from balance
func (r *BalanceRepository) SubtractBalance(userID int64, amount string) error {
	// Get current balance
	balance, err := r.GetByUserID(userID)
	if err != nil {
		return err
	}

	// Convert to big.Int
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	subAmount := new(big.Int)
	subAmount.SetString(amount, 10)

	// Check if sufficient
	if currentBalance.Cmp(subAmount) < 0 {
		return errors.New("insufficient balance")
	}

	// Subtract
	newBalance := new(big.Int).Sub(currentBalance, subAmount)

	// Update
	return r.SetBalance(userID, newBalance.String())
}

// Lock balance (for pending transactions)
func (r *BalanceRepository) LockBalance(userID int64, amount string) error {
	balance, err := r.GetByUserID(userID)
	if err != nil {
		return err
	}

	// Check available balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	lockedBalance := new(big.Int)
	lockedBalance.SetString(balance.LockedBalance, 10)

	lockAmount := new(big.Int)
	lockAmount.SetString(amount, 10)

	availableBalance := new(big.Int).Sub(currentBalance, lockedBalance)

	if availableBalance.Cmp(lockAmount) < 0 {
		return errors.New("insufficient available balance")
	}

	// Add to locked balance
	newLockedBalance := new(big.Int).Add(lockedBalance, lockAmount)

	return r.db.Model(&models.Balance{}).
		Where("user_id = ?", userID).
		Update("locked_balance", newLockedBalance.String()).Error
}

// Unlock balance (transaction confirmed/failed)
func (r *BalanceRepository) UnlockBalance(userID int64, amount string) error {
	balance, err := r.GetByUserID(userID)
	if err != nil {
		return err
	}

	lockedBalance := new(big.Int)
	lockedBalance.SetString(balance.LockedBalance, 10)

	unlockAmount := new(big.Int)
	unlockAmount.SetString(amount, 10)

	// Subtract from locked balance
	newLockedBalance := new(big.Int).Sub(lockedBalance, unlockAmount)

	// Ensure not negative
	if newLockedBalance.Sign() < 0 {
		newLockedBalance = big.NewInt(0)
	}

	return r.db.Model(&models.Balance{}).
		Where("user_id = ?", userID).
		Update("locked_balance", newLockedBalance.String()).Error
}

// Deduct locked balance (transaction confirmed)
func (r *BalanceRepository) DeductLockedBalance(userID int64, amount string) error {
	// First subtract from balance
	if err := r.SubtractBalance(userID, amount); err != nil {
		return err
	}

	// Then unlock
	return r.UnlockBalance(userID, amount)
}

// Get available balance (balance - locked_balance)
func (r *BalanceRepository) GetAvailableBalance(userID int64) (string, error) {
	balance, err := r.GetByUserID(userID)
	if err != nil {
		return "0", err
	}

	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	lockedBalance := new(big.Int)
	lockedBalance.SetString(balance.LockedBalance, 10)

	available := new(big.Int).Sub(currentBalance, lockedBalance)

	return available.String(), nil
}

func (r *BalanceRepository) UpdateWithTx(tx *gorm.DB, balance *models.Balance) error {
	return tx.Save(balance).Error
}
