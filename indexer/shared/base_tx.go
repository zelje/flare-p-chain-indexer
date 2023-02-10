package shared

import (
	"flare-indexer/database"
	"flare-indexer/utils"
	"fmt"

	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/fx"
	"github.com/ava-labs/avalanchego/vms/secp256k1fx"
)

// Create chain specific database object from generic TxOutput (TxInput) type, e.g.,
// XChainTxOutput, PChainTxInput
type DbOutputCreator func(out *database.TxOutput) Output
type DbInputCreator func(out *database.TxInput) Input

// Create database outputs from TransferableOutputs, provided their type is *secp256k1fx.TransferOutput and the
// number of addresses for each output is 1. Error is returned if these two conditions
// are not met.
func OutputsFromTxOuts(txID string, outs []*avax.TransferableOutput, creator DbOutputCreator) ([]Output, error) {
	txOuts := make([]Output, len(outs))
	for outi, cout := range outs {
		dbOut := &database.TxOutput{
			TxID: txID,
			Idx:  uint32(outi),
		}
		err := UpdateTransferableOutput(dbOut, cout)
		if err != nil {
			return nil, err
		}
		txOuts[outi] = creator(dbOut)
	}
	return txOuts, nil
}

// Update database output from TransferableOutput provided its type is *secp256k1fx.TransferOutput and the
// number of addresses is 1. Error is returned if these two conditions are not met.
func UpdateTransferableOutput(dbOut *database.TxOutput, out *avax.TransferableOutput) error {
	to, ok := out.Out.(*secp256k1fx.TransferOutput)
	if !ok {
		return fmt.Errorf("TransferableOutput has unsupported type")
	}
	if len(to.Addrs) != 1 {
		return fmt.Errorf("TransferableOutput has 0 or more than one address")
	}

	addr, err := utils.FormatAddressBytes(to.Addrs[0].Bytes())
	if err != nil {
		return err
	}
	dbOut.Amount = to.Amt
	dbOut.Address = addr
	return nil
}

// Return address from Owner interface provided its type is *secp256k1fx.OutputOwners and the
// number of addresses is 1. Error is returned if these two conditions are not met.
func RewardsOwnerAddress(owner fx.Owner) (string, error) {
	oo, ok := owner.(*secp256k1fx.OutputOwners)
	if !ok {
		return "", fmt.Errorf("rewards owner has unsupported type")
	}
	if len(oo.Addrs) != 1 {
		return "", fmt.Errorf("rewards owner has 0 or more than one address")
	}
	return utils.FormatAddressBytes(oo.Addrs[0].Bytes())
}

// Create inputs to BaseTx. Note that addresses of inputs are are not set. They should be updated from
// cached outputs, outputs from the database or outputs from chain
func InputsFromBaseTx(txID string, baseTx *avax.BaseTx, creator DbInputCreator) []Input {
	txIns := make([]Input, len(baseTx.Ins))
	for ini, in := range baseTx.Ins {
		txIns[ini] = creator(&database.TxInput{
			TxID:    txID,
			OutTxID: in.TxID.String(),
			OutIdx:  in.OutputIndex,
		})
	}
	return txIns
}
