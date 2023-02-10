package shared

import (
	"flare-indexer/utils"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
)

type InputUpdater interface {
	// Update inputs with addresses. Updater can get outputs from cache, db, chain (indexer, api), ...
	// Parameter is a map from output tx id to inputs referring to this output which have not been updated yet
	// Updated inputs should be removed from the map
	UpdateInputs(inputs map[string][]Input) error

	// Put outputs of a transaction to cache -- to avoid updating from chain or database
	CacheOutputs(txID string, outs []Output)
}

type IdIndexKey struct {
	string
	uint32
}

type BaseInputUpdater struct {
	Cache utils.Cache[string, []Output] // Map from transaction id to its outputs
}

func (iu *BaseInputUpdater) InitCache(maxSize int) {
	iu.Cache = utils.NewCache[string, []Output](maxSize)
}

func (iu *BaseInputUpdater) CacheOutputs(txID string, outs []Output) {
	iu.Cache.Add(txID, outs[:])
}

// Update inputs with addresses from outputs in cache
func (iu *BaseInputUpdater) UpdateInputsFromCache(notUpdated map[string][]Input) error {
	cachedOutputs := make([]Output, 0)
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
func UpdateInputsWithOutputs(inputs map[string][]Input, outputs []Output) error {
	txIdToOutput := make(map[IdIndexKey]Output)
	txIds := mapset.NewSet[string]()
	for _, out := range outputs {
		txIdToOutput[IdIndexKey{out.Tx(), out.Index()}] = out
		txIds.Add(out.Tx())
	}
	for txID := range txIds.Iterator().C {
		ins, ok := inputs[txID]
		if !ok {
			continue
		}
		for _, in := range ins {
			out, ok := txIdToOutput[IdIndexKey{in.OutTx(), in.OutIndex()}]
			if !ok {
				return fmt.Errorf("missing output with index %d of transaction %s", in.OutIndex, in.OutTx)
			}
			in.UpdateAddr(out.Addr())
		}
		delete(inputs, txID)
	}
	return nil
}
