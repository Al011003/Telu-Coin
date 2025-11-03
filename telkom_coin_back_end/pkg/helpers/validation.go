package helpers

import (
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateUsername validates username format
func ValidateUsername(username string) bool {
	// Username should be 3-30 characters, alphanumeric and underscore only
	if len(username) < 3 || len(username) > 30 {
		return false
	}
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return usernameRegex.MatchString(username)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) bool {
	// Password should be at least 8 characters
	if len(password) < 8 {
		return false
	}
	
	// Should contain at least one uppercase, one lowercase, one digit
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	
	return hasUpper && hasLower && hasDigit
}

// ValidatePhone validates phone number format
func ValidatePhone(phone string) bool {
	// Remove spaces and dashes
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	
	// Should be 10-15 digits, optionally starting with +
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

// ValidateWalletAddress validates Ethereum-like wallet address
func ValidateWalletAddress(address string) bool {
	// Should be 42 characters starting with 0x
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}
	
	// Rest should be valid hex
	addressRegex := regexp.MustCompile(`^0x[a-fA-F0-9]{40}$`)
	return addressRegex.MatchString(address)
}

// ValidateAmount validates amount string (should be positive number)
func ValidateAmount(amount string) bool {
	if amount == "" {
		return false
	}
	
	// Try to parse as big.Int
	amountBig := new(big.Int)
	_, ok := amountBig.SetString(amount, 10)
	if !ok {
		return false
	}
	
	// Should be positive
	return amountBig.Sign() > 0
}

// ValidatePin validates PIN format (6 digits)
func ValidatePin(pin string) bool {
	if len(pin) != 6 {
		return false
	}
	
	// Should be all digits
	pinRegex := regexp.MustCompile(`^[0-9]{6}$`)
	return pinRegex.MatchString(pin)
}

// ValidateTransactionType validates transaction type
func ValidateTransactionType(txType string) bool {
	validTypes := []string{"transfer", "topup", "withdraw"}
	for _, validType := range validTypes {
		if txType == validType {
			return true
		}
	}
	return false
}

// ValidateTransactionStatus validates transaction status
func ValidateTransactionStatus(status string) bool {
	validStatuses := []string{"pending", "confirmed", "failed", "cancelled"}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit string) (int, int, error) {
	pageInt := 1
	limitInt := 10
	
	if page != "" {
		p, err := strconv.Atoi(page)
		if err != nil || p < 1 {
			pageInt = 1
		} else {
			pageInt = p
		}
	}
	
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil || l < 1 || l > 100 {
			limitInt = 10
		} else {
			limitInt = l
		}
	}
	
	return pageInt, limitInt, nil
}

// SanitizeString removes potentially harmful characters
func SanitizeString(input string) string {
	// Remove HTML tags and scripts
	input = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(input, "")
	
	// Remove SQL injection patterns
	input = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script)`).ReplaceAllString(input, "")
	
	// Trim whitespace
	return strings.TrimSpace(input)
}

// ValidateIDRAmount validates IDR amount (should be positive and reasonable)
func ValidateIDRAmount(amount string) bool {
	if !ValidateAmount(amount) {
		return false
	}
	
	amountBig := new(big.Int)
	amountBig.SetString(amount, 10)
	
	// Minimum 1000 IDR (1000 rupiah)
	minAmount := big.NewInt(1000)
	// Maximum 1 billion IDR
	maxAmount := big.NewInt(1000000000)
	
	return amountBig.Cmp(minAmount) >= 0 && amountBig.Cmp(maxAmount) <= 0
}

// ValidateCoinAmount validates coin amount
func ValidateCoinAmount(amount string) bool {
	if !ValidateAmount(amount) {
		return false
	}
	
	amountBig := new(big.Int)
	amountBig.SetString(amount, 10)
	
	// Minimum 1 coin unit
	minAmount := big.NewInt(1)
	// Maximum 1 trillion coin units
	maxAmount := big.NewInt(1000000000000)
	
	return amountBig.Cmp(minAmount) >= 0 && amountBig.Cmp(maxAmount) <= 0
}
