package repository

import (
	"telkom_coin_back_end/internal/models"

	"gorm.io/gorm"
)

// TransactionRepositoryInterface defines the contract for transaction repository
type TransactionRepositoryInterface interface {
	Create(tx *models.Transaction) error
	GetByID(id int64) (*models.Transaction, error)
	GetByHash(hash string) (*models.Transaction, error)
	GetByAddress(address string, limit, offset int) ([]models.Transaction, error)
	GetByUserAddress(address string, txType string, status string, limit, offset int) ([]models.Transaction, error)
	CountByUserAddress(address string, txType string, status string) (int64, error)
	GetByStatus(status string, limit, offset int) ([]models.Transaction, error)
	GetPending(limit int) ([]models.Transaction, error)
	Update(tx *models.Transaction) error
	UpdateStatus(hash, status string) error
	UpdateBlockInfo(hash string, blockNumber, gasUsed, gasPrice int64) error
	HashExists(hash string) (bool, error)
	GetTopupHistory(address string, limit, offset int) ([]models.Transaction, int64, error)
	CreateWithTx(tx *gorm.DB, txModel *models.Transaction) error
}

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create new transaction
func (r *TransactionRepository) Create(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

// Get transaction by ID
func (r *TransactionRepository) GetByID(id int64) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// Get transaction by hash
func (r *TransactionRepository) GetByHash(hash string) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Where("tx_hash = ?", hash).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

// Get transactions by address (from or to)
func (r *TransactionRepository) GetByAddress(address string, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("from_address = ? OR to_address = ?", address, address).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// Get transactions by user (via wallet address)
func (r *TransactionRepository) GetByUserAddress(address string, txType string, status string, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	query := r.db.Where("from_address = ? OR to_address = ?", address, address)

	if txType != "" {
		query = query.Where("tx_type = ?", txType)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, err
}

// Count transactions by user
func (r *TransactionRepository) CountByUserAddress(address string, txType string, status string) (int64, error) {
	var count int64
	query := r.db.Model(&models.Transaction{}).Where("from_address = ? OR to_address = ?", address, address)

	if txType != "" {
		query = query.Where("tx_type = ?", txType)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return count, err
}

// Get transactions by status
func (r *TransactionRepository) GetByStatus(status string, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// Get pending transactions
func (r *TransactionRepository) GetPending(limit int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Where("status = ?", "pending").
		Order("created_at ASC").
		Limit(limit).
		Find(&transactions).Error
	return transactions, err
}

// Update transaction
func (r *TransactionRepository) Update(tx *models.Transaction) error {
	return r.db.Save(tx).Error
}

// Update transaction status
func (r *TransactionRepository) UpdateStatus(hash, status string) error {
	return r.db.Model(&models.Transaction{}).
		Where("tx_hash = ?", hash).
		Update("status", status).Error
}

// Update transaction with block info
func (r *TransactionRepository) UpdateBlockInfo(hash string, blockNumber, gasUsed, gasPrice int64) error {
	return r.db.Model(&models.Transaction{}).
		Where("tx_hash = ?", hash).
		Updates(map[string]interface{}{
			"block_number": blockNumber,
			"gas_used":     gasUsed,
			"gas_price":    gasPrice,
			"status":       "confirmed",
		}).Error
}

// Check if transaction hash exists
func (r *TransactionRepository) HashExists(hash string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Transaction{}).Where("tx_hash = ?", hash).Count(&count).Error
	return count > 0, err
}

// Get topup history for user
func (r *TransactionRepository) GetTopupHistory(address string, limit, offset int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var totalCount int64

	// Get topup transactions for this address
	query := r.db.Where("tx_type = ? AND to_address = ?", "topup", address)

	// Get total count
	err := query.Model(&models.Transaction{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Get transactions with pagination
	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error

	return transactions, totalCount, err
}

func (r *TransactionRepository) CreateWithTx(tx *gorm.DB, txModel *models.Transaction) error {
	return tx.Create(txModel).Error
}

func (r *TransactionRepository) FindTopupHistoryByWalletAddress(
	walletAddress string,
	page int,
	limit int,
) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var totalCount int64

	// Alamat nol (zero address) sebagai penanda asal top-up
	zeroAddress := "0x0000000000000000000000000000000000000000"

	// Buat query dasar sesuai logikamu
	query := r.db.Where(
		"to_address = ? AND from_address = ? AND tx_type = ?",
		walletAddress,
		zeroAddress,
		"topup",
	)

	// Hitung total data terlebih dahulu untuk pagination
	err := query.Model(&models.Transaction{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Hitung offset untuk pagination
	offset := (page - 1) * limit

	// Ambil data sesuai halaman dengan urutan terbaru dulu
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error

	return transactions, totalCount, err
}

func (r *TransactionRepository) FindHistoryByWalletAddress(
	walletAddress string,
	page int,
	limit int,
) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var totalCount int64

	// Buat query dasar: cari di mana walletAddress adalah PENGIRIM ATAU PENERIMA
	query := r.db.Where("from_address = ? OR to_address = ?", walletAddress, walletAddress)

	// Hitung total data terlebih dahulu untuk pagination
	err := query.Model(&models.Transaction{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Hitung offset untuk pagination
	offset := (page - 1) * limit

	// Ambil data sesuai halaman dengan urutan terbaru dulu
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error

	return transactions, totalCount, err
}
