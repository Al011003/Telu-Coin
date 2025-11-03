package models

import (
	"time"
)

type Balance struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64     `gorm:"not null;index" json:"user_id"`
	WalletAddress string    `gorm:"type:varchar(42);not null;index" json:"wallet_address"`
	TokenAddress  *string   `gorm:"type:varchar(42)" json:"token_address,omitempty"`
	Balance       string    `gorm:"type:varchar(50);not null;default:'0'" json:"balance"`
	LockedBalance string    `gorm:"type:varchar(50);default:'0'" json:"locked_balance"`
	LastSyncBlock *int64    `gorm:"type:bigint" json:"last_sync_block,omitempty"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Balance) TableName() string {
	return "balances"
}
