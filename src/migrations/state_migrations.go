package migrations

import "gorm.io/gorm"

func init() {
	Container.Add("2023-01-27-00-00", "Create initial state for X-Chain transactions", createXChainTxState)
}

func createXChainTxState(db *gorm.DB) error {
	return nil
}
