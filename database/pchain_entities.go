package database

import (
	"time"
)

// Table with indexed data for a P-chain transaction
type PChainTx struct {
	BaseEntity
	Type         PChainTxType `gorm:"type:varchar(20)"`                 // Transaction type
	TxID         string       `gorm:"type:varchar(50);unique;not null"` // Transaction ID
	BlockID      string       `gorm:"type:varchar(50);not null"`        // Block ID
	BlockIndex   uint64       // Block index
	Timestamp    time.Time
	ChainID      string    `gorm:"type:varchar(50)"` // Filled in case of export or import transaction
	NodeID       string    `gorm:"type:varchar(50)"` // Filled in case of add delegator or validator transaction
	StartTime    time.Time // Start time of validator or delegator (when NodeID is not null)
	EndTime      time.Time // End time of validator or delegator (when NodeID is not null)
	Weight       uint64    // Weight (stake amount) (when NodeID is not null)
	RewardsOwner string    `gorm:"type:varchar(60)"` // Rewards owner address (in case of add delegator or validator transaction)
	Memo         string    `gorm:"type:varchar(256)"`
	Bytes        []byte    `gorm:"type:mediumblob"`
}

type PChainTxInput struct {
	TxInput
}

type PChainTxOutput struct {
	TxOutput
	Type PChainOutputType `gorm:"type:varchar(20)"` // Transaction output type (default or "stake" output)
}
