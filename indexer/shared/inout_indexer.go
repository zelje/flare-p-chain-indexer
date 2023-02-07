package shared

import (
	"flare-indexer/database"

	"github.com/ava-labs/avalanchego/vms/avm/txs"
)

// Indexer for transactions of "type" baseTx (UTXO transactions)
type InputOutputIndexer struct {
	inUpdater InputUpdater
	outs      map[string][]*database.TxOutput // tx id -> outputs
	ins       []*database.TxInput             // tx id -> inputs
}

// Return new Input-Output indexer
func NewInputOutputIndexer(inUpdater InputUpdater) *InputOutputIndexer {
	indexer := InputOutputIndexer{
		inUpdater: inUpdater,
	}
	indexer.Reset()
	return &indexer
}

func (iox *InputOutputIndexer) Reset() {
	iox.outs = make(map[string][]*database.TxOutput)
	iox.ins = make([]*database.TxInput, 0, 100)
}

func (iox *InputOutputIndexer) AddTx(txID string, tx *txs.BaseTx) error {
	if _, ok := iox.outs[txID]; ok {
		return nil
	}
	outs, err := TxOutputsFromBaseTx(txID, tx)
	if err != nil {
		return err
	}
	iox.outs[txID] = outs
	iox.inUpdater.CacheOutputs(txID, outs)

	iox.ins = append(iox.ins, TxInputsFromBaseTx(txID, tx)...)
	return nil
}

func (iox *InputOutputIndexer) ProcessBatch() error {
	return iox.inUpdater.UpdateInputs(iox.ins)
}

func (iox *InputOutputIndexer) GetIns() []*database.TxInput {
	return iox.ins
}

func (iox *InputOutputIndexer) GetOuts() []*database.TxOutput {
	result := make([]*database.TxOutput, 0, 4*len(iox.outs))
	for _, out := range iox.outs {
		result = append(result, out...)
	}
	return result
}
