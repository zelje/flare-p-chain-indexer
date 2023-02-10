package migrations

import (
	"flare-indexer/database"
	"flare-indexer/indexer/pchain"
	"flare-indexer/indexer/xchain"

	"gorm.io/gorm"
)

func init() {
	Container.Add("2023-01-27-00-00", "Create initial state for X-Chain transactions", createXChainTxState)
	Container.Add("2023-02-10-00-00", "Create initial state for P-Chain transactions", createPChainTxState)
}

func createXChainTxState(db *gorm.DB) error {
	return database.CreateState(db, &database.State{
		Name:           xchain.StateName,
		NextDBIndex:    0,
		LastChainIndex: 0,
	})
}

func createPChainTxState(db *gorm.DB) error {
	return database.CreateState(db, &database.State{
		Name:           pchain.StateName,
		NextDBIndex:    0,
		LastChainIndex: 0,
	})
}
