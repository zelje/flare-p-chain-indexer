package pchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/migrations"
	"time"

	"gorm.io/gorm"
)

func init() {
	migrations.Container.Add("2023-02-10-00-00", "Create initial state for P-Chain transactions", createPChainTxState)
}

func createPChainTxState(db *gorm.DB) error {
	return database.CreateState(db, &database.State{
		Name:           StateName,
		NextDBIndex:    0,
		LastChainIndex: 0,
		Updated:        time.Now(),
	})
}
