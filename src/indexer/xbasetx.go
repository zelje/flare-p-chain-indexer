package indexer

import (
	"flare-indexer/src/chain"
	"flare-indexer/src/dbmodel"
	"flare-indexer/src/logger"
	"fmt"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x"
	mapset "github.com/deckarep/golang-set/v2"
	"gorm.io/gorm"
)

// Indexer for X-chain transactions of "type" baseTx

type keyType struct {
	string
	uint32
}

type baseTxIndexer struct {
	NewTxs  []*dbmodel.XChainTx
	NewOuts []*dbmodel.XChainTxOutput
	NewIns  []*dbmodel.XChainTxInput

	newInsBase []*chain.XChainTxInputBase
}

// Return new indexer; batch size is approximate and is used for
// the initialization of arrays
func NewBaseTxIndexer(batchSize int) baseTxIndexer {
	return baseTxIndexer{
		NewTxs:     make([]*dbmodel.XChainTx, 0, batchSize),
		NewOuts:    make([]*dbmodel.XChainTxOutput, 0, 4*batchSize),
		NewIns:     make([]*dbmodel.XChainTxInput, 0, 4*batchSize),
		newInsBase: make([]*chain.XChainTxInputBase, 0, 4*batchSize),
	}
}

func (i *baseTxIndexer) AddTx(data *chain.XChainTxData) {
	// New transaction goes db
	i.NewTxs = append(i.NewTxs, data.Tx)

	// New outs get saved to db
	i.NewOuts = append(i.NewOuts, data.TxOuts...)

	// New ins (not db objects)
	i.newInsBase = append(i.newInsBase, data.TxIns...)
}

// Persist new entities
func (i *baseTxIndexer) UpdateIns(db *gorm.DB, client indexer.Client) error {
	// Map of outs needed for ins; key is (txId, output index)
	outsMap := make(map[keyType]*dbmodel.XChainTxOutput)

	// First find all needed transactions for inputs
	missingTxIds := mapset.NewSet[string]()
	for _, in := range i.newInsBase {
		missingTxIds.Add(in.TxID)
	}

	updateOutsMapFromOuts(outsMap, i.NewOuts, missingTxIds)

	err := updateOutsMapFromDB(db, outsMap, missingTxIds)
	if err != nil {
		return err
	}

	err = updateOutsMapFromChain(client, outsMap, missingTxIds)
	if err != nil {
		return err
	}

	if missingTxIds.Cardinality() > 0 {
		return fmt.Errorf("unable to fetch transactions %v", missingTxIds)
	}

	for _, in := range i.newInsBase {
		out, ok := outsMap[keyType{in.TxID, in.OutputIndex}]
		if !ok {
			logger.Warn("unable to find output (%s, %d)", in.TxID, in.OutputIndex)
		} else {
			i.NewIns = append(i.NewIns, &dbmodel.XChainTxInput{
				TxID:    in.TxID,
				Address: out.Address,
			})
		}
	}

	return nil
}

// Update outsMap for missing transaction idxs from transactions fetched in this batch.
// Also updates missingTxIds set.
func updateOutsMapFromOuts(
	outsMap map[keyType]*dbmodel.XChainTxOutput,
	newOuts []*dbmodel.XChainTxOutput,
	missingTxIds mapset.Set[string],
) {
	for _, out := range newOuts {
		outsMap[keyType{out.TxID, out.Idx}] = out
		// if missingTxIds.Contains(out.TxID) {
		missingTxIds.Remove(out.TxID)
		// }
	}
}

// Update outsMap for missing transaction idxs. Also updates missingTxIds set.
func updateOutsMapFromDB(
	db *gorm.DB,
	outsMap map[keyType]*dbmodel.XChainTxOutput,
	missingTxIds mapset.Set[string],
) error {
	outs, err := dbmodel.FetchXChainTxOutputs(db, missingTxIds.ToSlice())
	if err != nil {
		return err
	}
	for _, out := range outs {
		outsMap[keyType{out.TxID, out.Idx}] = &out
		missingTxIds.Remove(out.TxID)
	}
	return nil
}

// Update outsMap for missing transaction idxs by fetching transactions from the chain.
// Also updates missingTxIds set.
func updateOutsMapFromChain(
	client indexer.Client,
	outsMap map[keyType]*dbmodel.XChainTxOutput,
	missingTxIds mapset.Set[string],
) error {
	for _, txId := range missingTxIds.ToSlice() {
		container, err := fetchContainerFromIndexer(client, txId)
		if err != nil {
			return err
		}
		if container == nil {
			missingTxIds.Remove(txId)
			continue
		}

		tx, err := x.Parser.ParseGenesisTx(container.Bytes)
		if err != nil {
			return err
		}

		var outs []*dbmodel.XChainTxOutput
		switch unsignedTx := tx.Unsigned.(type) {
		case *txs.BaseTx:
			outs, err = chain.XChainTxOutputsFromBaseTx(txId, unsignedTx)
		case *txs.ImportTx:
			outs, err = chain.XChainTxOutputsFromBaseTx(txId, &unsignedTx.BaseTx)
		default:
			return fmt.Errorf("transaction with id %s has unsupported type %T", container.ID.String(), unsignedTx)
		}

		if err != nil {
			return err
		}
		updateOutsMapFromOuts(outsMap, outs, missingTxIds)

	}
	return nil
}
