package xchain

import (
	"flare-indexer/database"
	"flare-indexer/utils"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

type XChainTxInputBase struct {
	TxID        string
	OutputIndex uint32
}

type XChainTxData struct {
	Tx     *database.XChainTx         // db entity for transaction
	TxIns  []*XChainTxInputBase       // non-db entities for input (should be filled with additional indexer call or from db)
	TxOuts []*database.XChainTxOutput // db entities for outputs
}

// Fetch all outputs of transaction, provided their type is *secp256k1fx.TransferOutput and the
// number of addresses for each output is 1
func XChainTxOutputsFromBaseTx(txID string, baseTx *txs.BaseTx) ([]*database.XChainTxOutput, error) {
	txOuts := make([]*database.XChainTxOutput, len(baseTx.Outs))
	for outi, cout := range baseTx.Outs {
		to, ok := cout.Out.(*secp256k1fx.TransferOutput)
		if !ok {
			return nil, fmt.Errorf("output of BaseTx has unsupported type")
		}
		if len(to.Addrs) != 1 {
			return nil, fmt.Errorf("output of BaseTx has 0 or more than one address")
		}

		addr, err := utils.FormatAddressBytes(to.Addrs[0].Bytes())
		if err != nil {
			return nil, err
		}
		txOuts[outi] = &database.XChainTxOutput{
			TxID:    txID,
			Amount:  to.Amt,
			Address: addr,
			Idx:     uint32(outi),
		}
	}
	return txOuts, nil
}

// Return xChainTxData from from BaseTx, Container, index
// We expect that outputs are of type *secp256k1fx.TransferOutput and the number of addresses for each output is 1
// Note that inputs are not db entities but "placehorders"
func XChainTxDataFromBaseTx(container *indexer.Container, baseTx *txs.BaseTx, txType database.TransactionType, index uint64) (*XChainTxData, error) {
	// TODO: check for asset?

	tx := &database.XChainTx{}
	tx.TxID = container.ID.String()
	tx.TxIndex = index
	tx.Timestamp = time.Unix(container.Timestamp/1e9, container.Timestamp%1e9)
	tx.Memo = string(baseTx.Memo)
	tx.Bytes = container.Bytes

	var err error
	d := XChainTxData{}

	d.Tx = tx
	d.TxOuts, err = XChainTxOutputsFromBaseTx(tx.TxID, baseTx)
	if err != nil {
		return nil, err
	}

	d.TxIns = make([]*XChainTxInputBase, len(baseTx.Ins))
	for ini, in := range baseTx.Ins {
		// logger.Debug("       tx %s has input %s #%d", container.ID.String(), in.TxID.String(), in.OutputIndex)
		d.TxIns[ini] = &XChainTxInputBase{
			TxID:        in.TxID.String(),
			OutputIndex: in.OutputIndex,
		}
	}

	return &d, nil
}
