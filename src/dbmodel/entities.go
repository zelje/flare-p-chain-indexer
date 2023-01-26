package dbmodel

import (
	"time"
)

type BaseEntity struct {
	ID uint64 `gorm:"primaryKey"`
}

type State struct {
	BaseEntity
	Name           string `gorm:"type:varchar(50);index"`
	NextDBIndex    uint64 // Next item to index, i.e., "last index" + 1
	LastChainIndex uint64
	Updated        time.Time
}

// Table with indexed data for X-chain transaction
type XChainTx struct {
	BaseEntity
	Type      TransactionType `gorm:"type:varchar(20)"`                 // Transaction type
	TxID      string          `gorm:"type:varchar(50);unique;not null"` // Transaction ID
	TxIndex   uint64          `gorm:"unique"`                           // Transaction index
	Timestamp time.Time
	Memo      string `gorm:"type:varchar(256)"`
	Bytes     []byte `gorm:"type:mediumblob"`
}

type XChainTxInput struct {
	BaseEntity
	TxID    string `gorm:"type:varchar(50);not null"` // Transaction ID
	Address string `gorm:"type:varchar(60);index"`
}

type XChainTxOutput struct {
	BaseEntity
	TxID    string `gorm:"type:varchar(50);not null"` // Transaction ID
	Amount  uint64
	Idx     uint32
	Address string `gorm:"type:varchar(60);index"`
}
