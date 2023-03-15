package database

import (
	"flare-indexer/utils"
	"time"

	"gorm.io/gorm"
)

func FetchPChainTxOutputs(db *gorm.DB, ids []string) ([]PChainTxOutput, error) {
	var txs []PChainTxOutput
	err := db.Where("tx_id IN ?", ids).Find(&txs).Error
	return txs, err
}

func CreatePChainEntities(db *gorm.DB, txs []*PChainTx, ins []*PChainTxInput, outs []*PChainTxOutput) error {
	if len(txs) > 0 { // attempt to create from an empty slice returns error
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

func FetchPChainValidators(
	db *gorm.DB,
	nodeID string,
	address string,
	startTime time.Time,
	endTime time.Time,
	offset int,
	limit int,
) ([]string, error) {
	var validatorTxs []PChainTx

	query := db.Where(&PChainTx{Type: PChainAddValidatorTx, NodeID: nodeID})
	if !startTime.IsZero() {
		query = query.Where("start_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("end_time <= ?", endTime)
	}
	err := query.Select("tx_id").Find(&validatorTxs).Error
	if err != nil {
		return nil, err
	}

	return utils.Map(validatorTxs, func(t PChainTx) string { return t.TxID }), nil
}
