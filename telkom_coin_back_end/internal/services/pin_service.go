package service

import (
	"errors"
	"telkom_coin_back_end/internal/repository"
	"telkom_coin_back_end/pkg/crypto"
	"time"

	"gorm.io/gorm"
)

type PinService struct {
	userRepo repository.UserRepositoryInterface
}

func NewPinService(userRepo repository.UserRepositoryInterface) *PinService {
	return &PinService{
		userRepo: userRepo,
	}
}

// SetPin sets or updates user PIN
func (s *PinService) SetPin(userID int64, pin string) error {
	// Validate PIN format (should be 6 digits)
	if len(pin) != 6 {
		return errors.New("PIN must be 6 digits")
	}

	// Check if all characters are digits
	for _, char := range pin {
		if char < '0' || char > '9' {
			return errors.New("PIN must contain only digits")
		}
	}

	// Check if user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user not found")
		}
		return err
	}

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
func (s *PinService) VerifyPin(userID int64, pin string) error {
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
func (s *PinService) ChangePin(userID int64, oldPin, newPin string) error {
	// Verify old PIN first
	if err := s.VerifyPin(userID, oldPin); err != nil {
		return errors.New("invalid old PIN")
	}

	// Set new PIN
	return s.SetPin(userID, newPin)
}

// IsPinSet checks if user has set a PIN
func (s *PinService) IsPinSet(userID int64) (bool, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, errors.New("user not found")
		}
		return false, err
	}

	return user.PinHash != "", nil
}

// ResetPin resets user PIN (admin function)
func (s *PinService) ResetPin(userID int64) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("user not found")
		}
		return err
	}

	// Clear PIN hash
	fields := map[string]interface{}{
		"pin_hash":   "",
		"updated_at": time.Now(),
	}

	return s.userRepo.UpdateFields(userID, fields)
}

// ValidatePinFormat validates PIN format without checking against database
func (s *PinService) ValidatePinFormat(pin string) error {
	if len(pin) != 6 {
		return errors.New("PIN must be 6 digits")
	}

	for _, char := range pin {
		if char < '0' || char > '9' {
			return errors.New("PIN must contain only digits")
		}
	}

	return nil
}

// GetPinAttempts gets number of failed PIN attempts (placeholder for rate limiting)
func (s *PinService) GetPinAttempts(userID int64) (int, error) {
	// In a real implementation, this would track failed attempts
	// For now, return 0
	return 0, nil
}

// IncrementPinAttempts increments failed PIN attempts (placeholder for rate limiting)
func (s *PinService) IncrementPinAttempts(userID int64) error {
	// In a real implementation, this would increment failed attempts
	// and possibly lock account after too many attempts
	return nil
}

// ResetPinAttempts resets failed PIN attempts counter
func (s *PinService) ResetPinAttempts(userID int64) error {
	// In a real implementation, this would reset the failed attempts counter
	return nil
}

// IsAccountLocked checks if account is locked due to too many failed PIN attempts
func (s *PinService) IsAccountLocked(userID int64) (bool, error) {
	// In a real implementation, this would check if account is locked
	// For now, return false
	return false, nil
}
