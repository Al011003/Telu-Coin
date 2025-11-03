package validator

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	validate.RegisterValidation("wallet_address", validateWalletAddress)
	validate.RegisterValidation("pin", validatePin)
	validate.RegisterValidation("amount", validateAmount)
	validate.RegisterValidation("tx_type", validateTxType)
}

// ValidateStruct validates a struct and returns validation errors
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()

		switch tag {
		case "required":
			errors[field] = field + " is required"
		case "email":
			errors[field] = field + " must be a valid email address"
		case "min":
			errors[field] = field + " must be at least " + err.Param() + " characters"
		case "max":
			errors[field] = field + " must be at most " + err.Param() + " characters"
		case "wallet_address":
			errors[field] = field + " must be a valid wallet address"
		case "pin":
			errors[field] = field + " must be a 6-digit PIN"
		case "amount":
			errors[field] = field + " must be a valid positive amount"
		case "tx_type":
			errors[field] = field + " must be a valid transaction type (transfer, topup, withdraw)"
		default:
			errors[field] = field + " is invalid"
		}
	}

	return errors
}

// Custom validator for wallet address
func validateWalletAddress(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return false
	}

	// Check if rest are valid hex characters
	for _, char := range address[2:] {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

// Custom validator for PIN
func validatePin(fl validator.FieldLevel) bool {
	pin := fl.Field().String()
	if len(pin) != 6 {
		return false
	}

	// Check if all characters are digits
	for _, char := range pin {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

// Custom validator for amount
func validateAmount(fl validator.FieldLevel) bool {
	amount := fl.Field().String()
	if amount == "" {
		return false
	}

	// Check if it's a valid positive number
	for i, char := range amount {
		if char < '0' || char > '9' {
			// Allow decimal point
			if char == '.' && i > 0 && i < len(amount)-1 {
				continue
			}
			return false
		}
	}

	return true
}

// Custom validator for transaction type
func validateTxType(fl validator.FieldLevel) bool {
	txType := fl.Field().String()
	validTypes := []string{"transfer", "topup", "withdraw"}

	for _, validType := range validTypes {
		if txType == validType {
			return true
		}
	}

	return false
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	return validate.Var(email, "required,email") == nil
}

// ValidateRequired validates required field
func ValidateRequired(value interface{}) bool {
	return validate.Var(value, "required") == nil
}

// ValidateMin validates minimum length
func ValidateMin(value string, min int) bool {
	tag := "min=" + strconv.Itoa(min)
	return validate.Var(value, tag) == nil
}

// ValidateMax validates maximum length
func ValidateMax(value string, max int) bool {
	tag := "max=" + strconv.Itoa(max)
	return validate.Var(value, tag) == nil
}

// ValidateLength validates exact length
func ValidateLength(value string, length int) bool {
	tag := "len=" + strconv.Itoa(length)
	return validate.Var(value, tag) == nil
}

// ValidateNumeric validates numeric value
func ValidateNumeric(value string) bool {
	return validate.Var(value, "numeric") == nil
}

// ValidateAlphaNumeric validates alphanumeric value
func ValidateAlphaNumeric(value string) bool {
	return validate.Var(value, "alphanum") == nil
}

// ValidateURL validates URL format
func ValidateURL(url string) bool {
	return validate.Var(url, "url") == nil
}

// ValidateJSON validates JSON format
func ValidateJSON(jsonStr string) bool {
	return validate.Var(jsonStr, "json") == nil
}

// ValidateUUID validates UUID format
func ValidateUUID(uuid string) bool {
	return validate.Var(uuid, "uuid") == nil
}

// ValidateBase64 validates base64 format
func ValidateBase64(base64Str string) bool {
	return validate.Var(base64Str, "base64") == nil
}

// ValidateHexadecimal validates hexadecimal format
func ValidateHexadecimal(hex string) bool {
	return validate.Var(hex, "hexadecimal") == nil
}

// ValidateIP validates IP address format
func ValidateIP(ip string) bool {
	return validate.Var(ip, "ip") == nil
}

// ValidateMAC validates MAC address format
func ValidateMAC(mac string) bool {
	return validate.Var(mac, "mac") == nil
}

// ValidateLatitude validates latitude coordinate
func ValidateLatitude(lat string) bool {
	return validate.Var(lat, "latitude") == nil
}

// ValidateLongitude validates longitude coordinate
func ValidateLongitude(lng string) bool {
	return validate.Var(lng, "longitude") == nil
}

// ValidateISBN validates ISBN format
func ValidateISBN(isbn string) bool {
	return validate.Var(isbn, "isbn") == nil
}

// ValidateISBN10 validates ISBN10 format
func ValidateISBN10(isbn string) bool {
	return validate.Var(isbn, "isbn10") == nil
}

// ValidateISBN13 validates ISBN13 format
func ValidateISBN13(isbn string) bool {
	return validate.Var(isbn, "isbn13") == nil
}

// ValidateCreditCard validates credit card number format
func ValidateCreditCard(cc string) bool {
	return validate.Var(cc, "credit_card") == nil
}
