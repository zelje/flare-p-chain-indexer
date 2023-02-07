package shared

import (
	"flare-indexer/database"
	"flare-indexer/utils"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
)

type InputUpdater interface {
	// Update inputs with addresses. Updater can get outputs from cache, db, chain (indexer, api), ...
	UpdateInputs(inputs []*database.TxInput) error

	// Put outputs of a transaction to cache -- to avoid updating from chain or database
	CacheOutputs(txID string, outs []*database.TxOutput)
}

type IdIndexKey struct {
	string
	uint32
}

type BaseInputUpdater struct {
	cache utils.Cache[string, []*database.TxOutput] // Map from transaction id to its outputs
}

func (iu *BaseInputUpdater) CacheOutputs(txID string, outs []*database.TxOutput) {
	iu.cache.Add(txID, outs[:])
}

// Return map from output tx id to inputs referring to this output which have not been updated yet
func (iu *BaseInputUpdater) UpdateInputsFromCache(inputs []*database.TxInput) map[string][]*database.TxInput {
	notUpdated := make(map[string][]*database.TxInput)

	// Update from cache and fill missing outputs (for inputs)
	for _, in := range inputs {
		if outs, ok := iu.cache.Get(in.OutTxID); ok {
			in.Address = outs[in.OutIdx].Address
		} else {
			ins, ok := notUpdated[in.OutTxID]
			if !ok {
				ins = make([]*database.TxInput, 0, 4)
			}
			notUpdated[in.OutTxID] = append(ins, in)
		}
	}
	return notUpdated
}

// Update input address from outputs
//  - updated inputs will be removed from the map
//  - inputs is map from output tx to inputs referring to this output
func UpdateInputsWithOutputs(inputs map[string][]*database.TxInput, outputs []*database.TxOutput) error {
	txIdToOutput := make(map[IdIndexKey]*database.TxOutput)
	txIds := mapset.NewSet[string]()
	for _, out := range outputs {
		txIdToOutput[IdIndexKey{out.TxID, out.Idx}] = out
		txIds.Add(out.TxID)
	}
	for txID := range txIds.Iterator().C {
		ins, ok := inputs[txID]
		if !ok {
			continue
		}
		for _, in := range ins {
			out, ok := txIdToOutput[IdIndexKey{in.OutTxID, in.OutIdx}]
			if !ok {
				return fmt.Errorf("missing output with index %d for transaction %s", in.OutIdx, in.OutTxID)
			}
			in.Address = out.Address
		}
		delete(inputs, txID)
	}
	return nil
}
