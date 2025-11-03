package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type TransactionMetadata map[string]interface{}

func (m *TransactionMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(TransactionMetadata)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

func (m TransactionMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return json.Marshal(m)
}

type Transaction struct {
	ID          int64               `gorm:"primaryKey;autoIncrement" json:"id"`
	TxHash      string              `gorm:"type:varchar(66);uniqueIndex;not null" json:"tx_hash"`
	FromAddress string              `gorm:"type:varchar(42);not null;index" json:"from_address"`
	ToAddress   string              `gorm:"type:varchar(42);not null;index" json:"to_address"`
	Amount      string              `gorm:"type:varchar(50);not null" json:"amount"`
	TxType      string              `gorm:"type:varchar(20);not null;index" json:"tx_type"` // transfer, topup, withdraw
	Status      string              `gorm:"type:varchar(20);default:'pending';index" json:"status"`
	BlockNumber *int64              `gorm:"type:bigint" json:"block_number,omitempty"`
	GasUsed     *int64              `gorm:"type:bigint" json:"gas_used,omitempty"`
	GasPrice    *int64              `gorm:"type:bigint" json:"gas_price,omitempty"`
	Nonce       *int64              `gorm:"type:bigint" json:"nonce,omitempty"`
	Metadata    TransactionMetadata `gorm:"type:json" json:"metadata,omitempty"`
	CreatedAt   time.Time           `gorm:"autoCreateTime;index" json:"created_at"`
	ConfirmedAt *time.Time          `json:"confirmed_at,omitempty"`
}

func (Transaction) TableName() string {
	return "transactions"
}
