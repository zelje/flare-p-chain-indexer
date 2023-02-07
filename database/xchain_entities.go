package database

import (
	"time"
)

// Table with indexed data for an X-chain transaction
type XChainTx struct {
	BaseEntity
	Type      XChainTxType `gorm:"type:varchar(20)"`                 // Transaction type
	TxID      string       `gorm:"type:varchar(50);unique;not null"` // Transaction ID
	TxIndex   uint64       `gorm:"unique"`                           // Transaction index
	Timestamp time.Time
	Memo      string `gorm:"type:varchar(256)"`
	Bytes     []byte `gorm:"type:mediumblob"`
}

type XChainTxInput struct {
	TxInput
}

type XChainTxOutput struct {
	TxOutput
}
