package database

import (
	"gorm.io/gorm"
)

func FetchXChainTxOutputs(db *gorm.DB, ids []string) ([]XChainTxOutput, error) {
	var txs []XChainTxOutput
	err := db.Where("tx_id IN ?", ids).Find(&txs).Error
	return txs, err
}

func CreateXChainEntities(db *gorm.DB, vertices []*XChainVtx, txs []*XChainTx, ins []*XChainTxInput, outs []*XChainTxOutput) error {
	if len(vertices) > 0 { // attempt to create from an empty slice returns error
		err := db.Create(vertices).Error
		if err != nil {
			return err
		}
	}
	if len(txs) > 0 {
		err := db.Create(txs).Error
		if err != nil {
			return err
		}
	}
	if len(ins) > 0 {
		err := db.Create(ins).Error
		if err != nil {
			return err
		}
	}
	if len(outs) > 0 {
		return db.Create(outs).Error
	}
	return nil
}
