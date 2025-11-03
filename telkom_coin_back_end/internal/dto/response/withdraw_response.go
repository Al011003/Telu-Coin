package response

import "time"

type WithdrawResponse struct {
	ID                 int64      `json:"id"`
	Amount             string     `json:"amount"`
	DestinationAddress string     `json:"destination_address"`
	DestinationType    string     `json:"destination_type"`
	Status             string     `json:"status"`
	TxHash             *string    `json:"tx_hash,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ProcessedAt        *time.Time `json:"processed_at,omitempty"`
}
