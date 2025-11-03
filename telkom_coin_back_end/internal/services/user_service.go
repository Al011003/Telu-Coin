package service

import (
	"errors"
	"math/big"
	"telkom_coin_back_end/internal/dto/request"
	"telkom_coin_back_end/internal/dto/response"
	"telkom_coin_back_end/internal/repository"

	"gorm.io/gorm"
)

type UserService struct {
	userRepo    *repository.UserRepository
	balanceRepo *repository.BalanceRepository
}

func NewUserService(userRepo *repository.UserRepository, balanceRepo *repository.BalanceRepository) *UserService {
	return &UserService{
		userRepo:    userRepo,
		balanceRepo: balanceRepo,
	}
}

// Get user profile by ID
func (s *UserService) GetProfile(userID int64) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &response.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		WalletAddress: user.WalletAddress,
		KYCStatus:     user.KYCStatus,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// Get user detail with balance
func (s *UserService) GetDetailProfile(userID int64) (*response.UserDetailResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Get balance
	balance, err := s.balanceRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Calculate available balance
	currentBalance := new(big.Int)
	currentBalance.SetString(balance.Balance, 10)

	lockedBalance := new(big.Int)
	lockedBalance.SetString(balance.LockedBalance, 10)

	availableBalance := new(big.Int).Sub(currentBalance, lockedBalance)

	return &response.UserDetailResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		WalletAddress: user.WalletAddress,
		KYCStatus:     user.KYCStatus,
		Status:        user.Status,
		Balance: response.BalanceResponse{
			Balance:       balance.Balance,
			LockedBalance: balance.LockedBalance,
			Available:     availableBalance.String(),
			UpdatedAt:     balance.UpdatedAt,
		},
		CreatedAt: user.CreatedAt,
	}, nil
}

// Update user profile
func (s *UserService) UpdateProfile(userID int64, req *request.UpdateProfileRequest) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if username is taken (if changed)
	if req.Username != "" && req.Username != user.Username {
		exists, err := s.userRepo.UsernameExists(req.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("username already taken")
		}
		user.Username = req.Username
	}

	// Check if email is taken (if changed)
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.EmailExists(req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already registered")
		}
		user.Email = req.Email
	}

	// Update phone
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	// Save changes
	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &response.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		WalletAddress: user.WalletAddress,
		KYCStatus:     user.KYCStatus,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// Get user by username (for transfer by username)
func (s *UserService) GetByUsername(username string) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &response.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		WalletAddress: user.WalletAddress,
		KYCStatus:     user.KYCStatus,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
	}, nil
}

// Get user by wallet address
func (s *UserService) GetByWalletAddress(address string) (*response.UserResponse, error) {
	user, err := s.userRepo.GetByWalletAddress(address)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &response.UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		WalletAddress: user.WalletAddress,
		KYCStatus:     user.KYCStatus,
		Status:        user.Status,
		CreatedAt:     user.CreatedAt,
	}, nil
}
