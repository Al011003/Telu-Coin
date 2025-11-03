package service

import (
	"errors"
	"telkom_coin_back_end/internal/models"
	"telkom_coin_back_end/internal/repository"
	"telkom_coin_back_end/pkg/crypto"
	"time"

	"gorm.io/gorm"
)

type AuthService struct {
	userRepo    repository.UserRepositoryInterface
	balanceRepo repository.BalanceRepositoryInterface
}

func NewAuthService(userRepo repository.UserRepositoryInterface) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

// SetBalanceRepo sets the balance repository (for dependency injection)
func (s *AuthService) SetBalanceRepo(balanceRepo repository.BalanceRepositoryInterface) {
	s.balanceRepo = balanceRepo
}
// AuthService.go
func (s *AuthService) CheckDuplicate(username, email string) error {
    // Cek email
    exists, err := s.userRepo.EmailExists(email)
    if err != nil {
        return err
    }
    if exists {
        return errors.New("email already registered")
    }

    // Cek username
    exists, err = s.userRepo.UsernameExists(username)
    if err != nil {
        return err
    }
    if exists {
        return errors.New("username already taken")
    }

    return nil
}

// Register creates a new user account
func (s *AuthService) Register(username, email, phone, password, pin string) (*models.User, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// Check if username already exists
	exists, err = s.userRepo.UsernameExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already taken")
	}

	// Generate wallet keypair
	walletAddress, privateKey, err := crypto.GenerateWalletKeypair()
	if err != nil {
		return nil, errors.New("failed to generate wallet")
	}

	// Encrypt private key with master secret (JWT_SECRET)
	encryptedPrivateKey, err := crypto.EncryptPrivateKey(privateKey)
	if err != nil {
		return nil, errors.New("failed to encrypt private key")
	}

	// Hash password
	passwordHash := crypto.HashPassword(password)

	// Hash PIN
	pinHash := crypto.HashPin(pin)

	// Create user
	user := &models.User{
		Username:            username,
		Email:               email,
		Phone:               phone,
		PasswordHash:        passwordHash,
		WalletAddress:       walletAddress,
		PrivateKeyEncrypted: encryptedPrivateKey,
		PinHash:             pinHash,
		KYCStatus:           "verified", // Auto-verified for TLC Wallet (no KYC required)
		Status:              "active",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Create initial balance record
	if s.balanceRepo != nil {
		balance := models.Balance{
			UserID:        user.ID,
			WalletAddress: walletAddress,
			Balance:       "0",
			LockedBalance: "0",
			UpdatedAt:     time.Now(),
		}
		s.balanceRepo.Create(&balance)
	}

	return user, nil
}

// Login authenticates user and returns JWT token
func (s *AuthService) Login(emailOrUsername, password string) (string, error) {
	var user *models.User
	var err error

	// Try to find user by email first
	if user, err = s.userRepo.GetByEmail(emailOrUsername); err != nil {
		// If not found by email, try username
		if err == gorm.ErrRecordNotFound {
			if user, err = s.userRepo.GetByUsername(emailOrUsername); err != nil {
				if err == gorm.ErrRecordNotFound {
					return "", errors.New("invalid credentials")
				}
				return "", err
			}
		} else {
			return "", err
		}
	}

	// Check if account is active
	if user.Status != "active" {
		return "", errors.New("account is not active")
	}

	// Verify password
	if !crypto.CheckPasswordHash(password, user.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := crypto.GenerateJWT(user.ID)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user not found")
		}
		return err
	}

	// Verify old password
	if !crypto.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return errors.New("invalid old password")
	}

	// Hash new password
	newPasswordHash := crypto.HashPassword(newPassword)

	// Re-encrypt private key with new password
	// First decrypt with old password
	// Decrypt private key pakai JWT_SECRET
	privateKey, err := crypto.DecryptPrivateKey(user.PrivateKeyEncrypted)
	if err != nil {
		return errors.New("failed to decrypt private key")
	}

	// Encrypt lagi pakai JWT_SECRET (opsional, kalau mau re-encrypt ulang)
	newEncryptedPrivateKey, err := crypto.EncryptPrivateKey(privateKey)
	if err != nil {
		return errors.New("failed to encrypt private key")
	}

	// Update user
	fields := map[string]interface{}{
		"password_hash":         newPasswordHash,
		"private_key_encrypted": newEncryptedPrivateKey,
		"updated_at":            time.Now(),
	}

	return s.userRepo.UpdateFields(userID, fields)
}

// SetPin sets or updates user PIN
func (s *AuthService) SetPin(userID int64, pin string) error {
	// Hash PIN
	pinHash := crypto.HashPin(pin)

	// Update user PIN
	fields := map[string]interface{}{
		"pin_hash":   pinHash,
		"updated_at": time.Now(),
	}

	return s.userRepo.UpdateFields(userID, fields)
}

// VerifyPin verifies user PIN
func (s *AuthService) VerifyPin(userID int64, pin string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user not found")
		}
		return err
	}

	// Check if PIN is set
	if user.PinHash == "" {
		return errors.New("PIN not set")
	}

	// Verify PIN
	if !crypto.VerifyPin(pin, user.PinHash) {
		return errors.New("invalid PIN")
	}

	return nil
}

// ChangePin changes user PIN
func (s *AuthService) ChangePin(userID int64, oldPin, newPin string) error {
	// Verify old PIN first
	if err := s.VerifyPin(userID, oldPin); err != nil {
		return errors.New("invalid old PIN")
	}

	// Set new PIN
	return s.SetPin(userID, newPin)
}

// GetUserByID gets user by ID
func (s *AuthService) GetUserByID(userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// ValidateToken validates JWT token and returns user ID
func (s *AuthService) ValidateToken(tokenString string) (int64, error) {
	return crypto.ValidateJWT(tokenString)
}

// GetPrivateKey decrypts and returns user's private key
func (s *AuthService) GetPrivateKey(userID int64, password string) (string, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("user not found")
		}
		return "", err
	}

	// Decrypt private key
	privateKey, err := crypto.DecryptPrivateKey(user.PrivateKeyEncrypted)

	if err != nil {
		return "", errors.New("invalid password or corrupted private key")
	}

	return privateKey, nil
}

// DeactivateAccount deactivates user account
func (s *AuthService) DeactivateAccount(userID int64) error {
	fields := map[string]interface{}{
		"status":     "inactive",
		"updated_at": time.Now(),
	}
	return s.userRepo.UpdateFields(userID, fields)
}

// ActivateAccount activates user account
func (s *AuthService) ActivateAccount(userID int64) error {
	fields := map[string]interface{}{
		"status":     "active",
		"updated_at": time.Now(),
	}
	return s.userRepo.UpdateFields(userID, fields)
}

// UpdateKYCStatus updates user KYC status
func (s *AuthService) UpdateKYCStatus(userID int64, status string) error {
	validStatuses := []string{"pending", "verified", "rejected"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("invalid KYC status")
	}

	fields := map[string]interface{}{
		"kyc_status": status,
		"updated_at": time.Now(),
	}
	return s.userRepo.UpdateFields(userID, fields)
}
