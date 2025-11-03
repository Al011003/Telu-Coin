package service

import (
	"errors"
	"telkom_coin_back_end/internal/dto/request"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/repository"

	"gorm.io/gorm"
)

type TransactionService struct {
	txRepo   *repository.TransactionRepository
	userRepo *repository.UserRepository
}

func NewTransactionService(txRepo *repository.TransactionRepository, userRepo *repository.UserRepository) *TransactionService {
	return &TransactionService{
		txRepo:   txRepo,
		userRepo: userRepo,
	}
}

// Get transaction history by user ID
func (s *TransactionService) GetHistory(userID int64, req *request.GetTransactionHistoryRequest) (*response.TransactionListResponse, error) {
	// Get user to get wallet address
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Set default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	offset := (req.Page - 1) * req.Limit

	// Get transactions
	transactions, err := s.txRepo.GetByUserAddress(user.WalletAddress, req.TxType, req.Status, req.Limit, offset)
	if err != nil {
		return nil, err
	}

	// Count total
	total, err := s.txRepo.CountByUserAddress(user.WalletAddress, req.TxType, req.Status)
	if err != nil {
		return nil, err
	}

	// Convert to response
	var txResponses []response.TransactionResponse
	for _, tx := range transactions {
		txResponses = append(txResponses, response.TransactionResponse{
			ID:          tx.ID,
			TxHash:      tx.TxHash,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			Amount:      tx.Amount,
			TxType:      tx.TxType,
			Status:      tx.Status,
			BlockNumber: tx.BlockNumber,
			Metadata:    tx.Metadata,
			CreatedAt:   tx.CreatedAt,
			ConfirmedAt: tx.ConfirmedAt,
		})
	}

	return &response.TransactionListResponse{
		Transactions: txResponses,
		Total:        total,
		Page:         req.Page,
		Limit:        req.Limit,
	}, nil
}

// Get transaction detail by hash
func (s *TransactionService) GetByHash(txHash string) (*response.TransactionDetailResponse, error) {
	tx, err := s.txRepo.GetByHash(txHash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Get from user (if exists in our system)
	var fromUser *response.UserResponse
	fromUserData, err := s.userRepo.GetByWalletAddress(tx.FromAddress)
	if err == nil {
		fromUser = &response.UserResponse{
			ID:            fromUserData.ID,
			Username:      fromUserData.Username,
			Email:         fromUserData.Email,
			WalletAddress: fromUserData.WalletAddress,
		}
	}

	// Get to user (if exists in our system)
	var toUser *response.UserResponse
	toUserData, err := s.userRepo.GetByWalletAddress(tx.ToAddress)
	if err == nil {
		toUser = &response.UserResponse{
			ID:            toUserData.ID,
			Username:      toUserData.Username,
			Email:         toUserData.Email,
			WalletAddress: toUserData.WalletAddress,
		}
	}

	return &response.TransactionDetailResponse{
		ID:          tx.ID,
		TxHash:      tx.TxHash,
		FromAddress: tx.FromAddress,
		FromUser:    fromUser,
		ToAddress:   tx.ToAddress,
		ToUser:      toUser,
		Amount:      tx.Amount,
		TxType:      tx.TxType,
		Status:      tx.Status,
		BlockNumber: tx.BlockNumber,
		GasUsed:     tx.GasUsed,
		GasPrice:    tx.GasPrice,
		Metadata:    tx.Metadata,
		CreatedAt:   tx.CreatedAt,
		ConfirmedAt: tx.ConfirmedAt,
	}, nil
}

// Get transaction detail by ID
func (s *TransactionService) GetByID(txID int64) (*response.TransactionDetailResponse, error) {
	tx, err := s.txRepo.GetByID(txID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Get from user (if exists in our system)
	var fromUser *response.UserResponse
	fromUserData, err := s.userRepo.GetByWalletAddress(tx.FromAddress)
	if err == nil {
		fromUser = &response.UserResponse{
			ID:            fromUserData.ID,
			Username:      fromUserData.Username,
			Email:         fromUserData.Email,
			WalletAddress: fromUserData.WalletAddress,
		}
	}

	// Get to user (if exists in our system)
	var toUser *response.UserResponse
	toUserData, err := s.userRepo.GetByWalletAddress(tx.ToAddress)
	if err == nil {
		toUser = &response.UserResponse{
			ID:            toUserData.ID,
			Username:      toUserData.Username,
			Email:         toUserData.Email,
			WalletAddress: toUserData.WalletAddress,
		}
	}

	return &response.TransactionDetailResponse{
		ID:          tx.ID,
		TxHash:      tx.TxHash,
		FromAddress: tx.FromAddress,
		FromUser:    fromUser,
		ToAddress:   tx.ToAddress,
		ToUser:      toUser,
		Amount:      tx.Amount,
		TxType:      tx.TxType,
		Status:      tx.Status,
		BlockNumber: tx.BlockNumber,
		GasUsed:     tx.GasUsed,
		GasPrice:    tx.GasPrice,
		Metadata:    tx.Metadata,
		CreatedAt:   tx.CreatedAt,
		ConfirmedAt: tx.ConfirmedAt,
	}, nil
}

// Get pending transactions (for processing)
func (s *TransactionService) GetPendingTransactions(limit int) ([]response.TransactionResponse, error) {
	if limit == 0 {
		limit = 100
	}

	transactions, err := s.txRepo.GetPending(limit)
	if err != nil {
		return nil, err
	}

	var txResponses []response.TransactionResponse
	for _, tx := range transactions {
		txResponses = append(txResponses, response.TransactionResponse{
			ID:          tx.ID,
			TxHash:      tx.TxHash,
			FromAddress: tx.FromAddress,
			ToAddress:   tx.ToAddress,
			Amount:      tx.Amount,
			TxType:      tx.TxType,
			Status:      tx.Status,
			BlockNumber: tx.BlockNumber,
			Metadata:    tx.Metadata,
			CreatedAt:   tx.CreatedAt,
			ConfirmedAt: tx.ConfirmedAt,
		})
	}

	return txResponses, nil
}
