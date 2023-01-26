package dbmodel

import "gorm.io/gorm"

func FetchState(db *gorm.DB, name string) State {
	var currentState State
	db.Where(&State{Name: name}).First(&currentState)
	return currentState
}

func FetchXChainTxOutputs(db *gorm.DB, ids []string) []XChainTxOutput {
	var txs []XChainTxOutput
	db.Where("tx_id IN ?", ids).Find(&txs)
	return txs
}
