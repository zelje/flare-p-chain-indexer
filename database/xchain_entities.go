package database

import (
	"time"
)

// Table with indexed data for an X-chain transaction
type XChainTx struct {
	BaseEntity
	Type      XChainTxType `gorm:"type:varchar(20)"`                 // Transaction type
	TxID      string       `gorm:"type:varchar(50);unique;not null"` // Transaction ID
	VtxHeight uint64
	Memo      string `gorm:"type:varchar(256)"`
	Bytes     []byte `gorm:"type:mediumblob"`
}

type XChainTxInput struct {
	TxInput
}

type XChainTxOutput struct {
	TxOutput
}

// Table with indexed data for an X-chain vertex (block)
type XChainVtx struct {
	BaseEntity
	VtxID     string    `gorm:"type:varchar(50);unique;not null"`
	ParentID  string    `gorm:"type:varchar(50)"`
	VtxIndex  uint64    `gorm:"unique"` // Vertex index - from indexer
	Height    uint64    `gorm:"unique"` // Vertex height
	Timestamp time.Time // Time indexed, not when accepted by the consensus
}
