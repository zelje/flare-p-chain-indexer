package shared

import (
	"flare-indexer/database"
	"flare-indexer/utils"
	"fmt"

	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

// Create outputs of BaseTx, provided their type is *secp256k1fx.TransferOutput and the
// number of addresses for each output is 1. Error is returned if these two conditions
// are not met.
func TxOutputsFromBaseTx(txID string, baseTx *txs.BaseTx) ([]*database.TxOutput, error) {
	txOuts := make([]*database.TxOutput, len(baseTx.Outs))
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

		txOuts[outi] = &database.TxOutput{
			TxID:    txID,
			Amount:  to.Amt,
			Address: addr,
			Idx:     uint32(outi),
		}
	}
	return txOuts, nil
}

// Create inputs to BaseTx. Note that addresses of inputs are are not set. They should be updated from
// cached outputs, outputs from the database or outputs from chain
func TxInputsFromBaseTx(txID string, baseTx *txs.BaseTx) []*database.TxInput {
	txIns := make([]*database.TxInput, len(baseTx.Ins))
	for ini, in := range baseTx.Ins {
		txIns[ini] = &database.TxInput{
			TxID:    txID,
			OutTxID: in.TxID.String(),
			OutIdx:  in.OutputIndex,
		}
	}
	return txIns
}
