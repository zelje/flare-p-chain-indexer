package shared

import (
	"flare-indexer/database"
	"flare-indexer/utils"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
)

type InputUpdater interface {
	// Update inputs with addresses. Updater can get outputs from cache, db, chain (indexer, api), ...
	// Parameter is a map from output tx id to inputs referring to this output which have not been updated yet
	// Updated inputs should be removed from the map
	UpdateInputs(inputs map[string][]*database.TxInput) error

	// Put outputs of a transaction to cache -- to avoid updating from chain or database
	CacheOutputs(txID string, outs []*database.TxOutput)
}

type IdIndexKey struct {
	string
	uint32
}

type BaseInputUpdater struct {
	Cache utils.Cache[string, []*database.TxOutput] // Map from transaction id to its outputs
}

func (iu *BaseInputUpdater) InitCache(maxSize int) {
	iu.Cache = utils.NewCache[string, []*database.TxOutput](maxSize)
}

func (iu *BaseInputUpdater) CacheOutputs(txID string, outs []*database.TxOutput) {
	iu.Cache.Add(txID, outs[:])
}

// Update inputs with addresses from outputs in cache
func (iu *BaseInputUpdater) UpdateInputsFromCache(notUpdated map[string][]*database.TxInput) error {
	cachedOutputs := make([]*database.TxOutput, 0)
	for k := range notUpdated {
		if outs, ok := iu.Cache.Get(k); ok {
			cachedOutputs = append(cachedOutputs, outs...)
		}
	}
	return UpdateInputsWithOutputs(notUpdated, cachedOutputs)
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
