package shared

import (
	"flare-indexer/utils"
	"fmt"

	"github.com/ava-labs/avalanchego/vms/components/avax"
)

// Indexer for transactions of "type" baseTx (UTXO transactions)
type InputOutputIndexer struct {
	inUpdater InputUpdater
	outs      map[string][]Output // tx id -> outputs
	ins       []Input             // inputs
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
	iox.outs = make(map[string][]Output)
	iox.ins = make([]Input, 0, 100)
}

func (iox *InputOutputIndexer) AddFromBaseTx(
	txID string,
	tx *avax.BaseTx,
	creator InputOutputCreator,
) error {
	if _, ok := iox.outs[txID]; ok {
		return nil
	}
	outs, err := OutputsFromTxOuts(txID, tx.Outs, creator)
	if err != nil {
		return err
	}
	iox.outs[txID] = outs
	iox.inUpdater.CacheOutputs(txID, outs)

	iox.ins = append(iox.ins, InputsFromTxIns(txID, tx.Ins, creator)...)
	return nil
}

func (iox *InputOutputIndexer) Add(txID string, outs []Output, ins []Input) {
	if _, ok := iox.outs[txID]; ok {
		return
	}
	iox.outs[txID] = outs
	iox.inUpdater.CacheOutputs(txID, outs)

	iox.ins = append(iox.ins, ins...)
}

func (iox *InputOutputIndexer) UpdateInputs(inputs []Input) error {
	notUpdated := make(map[string][]Input)
	for _, in := range inputs {
		ins, ok := notUpdated[in.OutTx()]
		if !ok {
			ins = make([]Input, 0, 4)
		}
		notUpdated[in.OutTx()] = append(ins, in)
	}
	err := iox.inUpdater.UpdateInputs(notUpdated)
	if err != nil {
		return err
	}
	if len(notUpdated) > 0 {
		return fmt.Errorf("unable to fetch transactions with ids %v", utils.Keys(notUpdated))
	}
	return nil
}

func (iox *InputOutputIndexer) ProcessBatch() error {
	return iox.UpdateInputs(iox.ins)
}

func (iox *InputOutputIndexer) GetIns() []Input {
	return iox.ins
}

func (iox *InputOutputIndexer) GetOuts() []Output {
	result := make([]Output, 0, 4*len(iox.outs))
	for _, out := range iox.outs {
		result = append(result, out...)
	}
	return result
}
