package migrations

import (
	"flare-indexer/src/dbmodel"
	"flare-indexer/src/indexer/xchain"

	"gorm.io/gorm"
)

func init() {
	Container.Add("2023-01-27-00-00", "Create initial state for X-Chain transactions", createXChainTxState)
}

func createXChainTxState(db *gorm.DB) error {
	return dbmodel.CreateState(db, &dbmodel.State{
		Name:           xchain.StateName,
		NextDBIndex:    0,
		LastChainIndex: 0,
	})
}
