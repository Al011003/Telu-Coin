package models

import (
	"time"
)

type User struct {
	ID                  int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username            string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email               string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Phone               string    `gorm:"type:varchar(20)" json:"phone,omitempty"`
	PasswordHash        string    `gorm:"type:varchar(255);not null" json:"-"`
	WalletAddress       string    `gorm:"type:varchar(42);uniqueIndex;not null" json:"wallet_address"`
	PrivateKeyEncrypted string    `gorm:"type:text;not null" json:"-"`
	PinHash             string    `gorm:"type:varchar(255)" json:"-"`
	KYCStatus           string    `gorm:"type:varchar(20);default:'verified'" json:"kyc_status"` // Auto-verified for TLC Wallet
	Status              string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
