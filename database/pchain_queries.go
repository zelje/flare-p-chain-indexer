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
	txType PChainTxType,
	nodeID string,
	address string,
	startTime time.Time,
	endTime time.Time,
	offset int,
	limit int,
) ([]string, error) {
	var validatorTxs []PChainTx

	query := db.Where(&PChainTx{Type: txType, NodeID: nodeID})
	if !startTime.IsZero() {
		query = query.Where("start_time >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("end_time <= ?", endTime)
	}
	if len(address) > 0 {
		query = query.Joins("left join p_chain_tx_inputs as inputs on inputs.tx_id = p_chain_txes.tx_id").
			Where("inputs.address = ?", address)
	}
	err := query.Offset(offset).Limit(limit).Order("p_chain_txes.id").
		Distinct().Select("p_chain_txes.tx_id").Find(&validatorTxs).Error
	if err != nil {
		return nil, err
	}

	return utils.Map(validatorTxs, func(t PChainTx) string { return *t.TxID }), nil
}

func FetchPChainTxFull(db *gorm.DB, txID string) (*PChainTx, []PChainTxInput, []PChainTxOutput, error) {
	var tx PChainTx
	err := db.Where(&PChainTx{TxID: &txID}).First(&tx).Error
	if err != nil {
		return nil, nil, nil, err
	}

	var inputs []PChainTxInput
	err = db.Where(&PChainTxInput{TxInput: TxInput{TxID: txID}}).Find(&inputs).Error
	if err != nil {
		return nil, nil, nil, err
	}

	var outputs []PChainTxOutput
	err = db.Where(&PChainTxOutput{TxOutput: TxOutput{TxID: txID}}).Order("idx").Find(&outputs).Error
	if err != nil {
		return nil, nil, nil, err
	}

	return &tx, inputs, outputs, nil
}

type PChainTxData struct {
	PChainTx
	InputAddress string
}

// Find P-chain transaction in given block height
// Returns transaction and true if found, nil and true if block was found,
// nil and false if block height does not exist.
func FindPChainTxInBlockHeight(db *gorm.DB,
	txID string,
	height uint32,
) (*PChainTxData, bool, error) {
	var txs []PChainTxData
	// err := db.Where(&PChainTx{BlockHeight: height}).Find(&txs).Error
	err := db.Table("p_chain_txes").
		Joins("left join p_chain_tx_inputs as inputs on inputs.tx_id = p_chain_txes.tx_id").
		Where("p_chain_txes.block_height = ?", height).
		Select("p_chain_txes.*, inputs.address as input_address").
		Scan(&txs).Error
	if err != nil {
		return nil, false, err
	}
	if len(txs) == 0 {
		return nil, false, nil
	}
	tx := &txs[0]
	if *tx.TxID != txID {
		return nil, true, nil
	}
	return &txs[0], true, nil
}
