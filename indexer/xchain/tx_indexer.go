package xchain

import (
	"flare-indexer/database"
	"flare-indexer/logger"
	"flare-indexer/utils/chain"
	"fmt"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x"
	mapset "github.com/deckarep/golang-set/v2"
	"gorm.io/gorm"
)

// Indexer for X-chain transactions of "type" baseTx
// Implements ContainerBatchIndexer

type txBatchIndexer struct {
	db     *gorm.DB
	client indexer.Client

	newTxs  []*database.XChainTx
	newOuts []*database.XChainTxOutput
	newIns  []*database.XChainTxInput

	newInsBase []*XChainTxInputBase
}

type keyType struct {
	string
	uint32
}

// Return new indexer; batch size is approximate and is used for
// the initialization of arrays
func NewTxBatchIndexer(
	db *gorm.DB,
	client indexer.Client,
	batchSize int,
) *txBatchIndexer {
	return &txBatchIndexer{
		db:     db,
		client: client,

		newTxs:     make([]*database.XChainTx, 0, batchSize),
		newOuts:    make([]*database.XChainTxOutput, 0, 4*batchSize),
		newIns:     make([]*database.XChainTxInput, 0, 4*batchSize),
		newInsBase: make([]*XChainTxInputBase, 0, 4*batchSize),
	}
}

func (i *txBatchIndexer) addTx(data *XChainTxData) {
	// New transaction goes db
	i.newTxs = append(i.newTxs, data.Tx)

	// New outs get saved to db
	i.newOuts = append(i.newOuts, data.TxOuts...)

	// New ins (not db objects)
	i.newInsBase = append(i.newInsBase, data.TxIns...)
}

// Persist new entities
func (i *txBatchIndexer) updateIns() error {
	// Map of outs needed for ins; key is (txId, output index)
	outsMap := make(map[keyType]*database.XChainTxOutput)

	// First find all needed transactions for inputs
	missingTxIds := mapset.NewSet[string]()
	for _, in := range i.newInsBase {
		missingTxIds.Add(in.TxID)
	}

	updateOutsMapFromOuts(outsMap, i.newOuts, missingTxIds)

	err := updateOutsMapFromDB(i.db, outsMap, missingTxIds)
	if err != nil {
		return err
	}

	err = updateOutsMapFromChain(i.client, outsMap, missingTxIds)
	if err != nil {
		return err
	}

	if missingTxIds.Cardinality() > 0 {
		return fmt.Errorf("unable to fetch transactions %v", missingTxIds)
	}

	for _, in := range i.newInsBase {
		out, ok := outsMap[keyType{in.TxID, in.OutputIndex}]
		if !ok {
			logger.Warn("Unable to find output (%s, %d)", in.TxID, in.OutputIndex)
		} else {
			i.newIns = append(i.newIns, &database.XChainTxInput{
				TxID:    in.TxID,
				Address: out.Address,
			})
		}
	}

	return nil
}

func (xi *txBatchIndexer) ProcessContainers(nextIndex uint64, containers []indexer.Container) (uint64, error) {

	var index uint64
	for i, container := range containers {
		index = nextIndex + uint64(i)

		tx, err := x.Parser.ParseGenesisTx(container.Bytes)
		if err != nil {
			return 0, err
		}

		switch unsignedTx := tx.Unsigned.(type) {
		case *txs.BaseTx:
			data, err := XChainTxDataFromBaseTx(&container, unsignedTx, database.BaseTx, index)
			if err != nil {
				return 0, nil
			}
			xi.addTx(data)
		case *txs.ImportTx:
			data, err := XChainTxDataFromBaseTx(&container, &unsignedTx.BaseTx, database.ImportTx, index)
			if err != nil {
				return 0, nil
			}
			xi.addTx(data)
		default:
			logger.Warn("Transaction with id '%s' is NOT indexed, type is %T", container.ID, unsignedTx)
		}
	}

	err := xi.updateIns()
	if err != nil {
		return 0, err
	}

	return index, nil
}

// Persist all entities
func (i *txBatchIndexer) PersistEntities(db *gorm.DB) error {
	return database.CreateXChainEntities(db, i.newTxs, i.newIns, i.newOuts)
}

// Update outsMap for missing transaction idxs from transactions fetched in this batch.
// Also updates missingTxIds set.
func updateOutsMapFromOuts(
	outsMap map[keyType]*database.XChainTxOutput,
	newOuts []*database.XChainTxOutput,
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
	outsMap map[keyType]*database.XChainTxOutput,
	missingTxIds mapset.Set[string],
) error {
	outs, err := database.FetchXChainTxOutputs(db, missingTxIds.ToSlice())
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
	outsMap map[keyType]*database.XChainTxOutput,
	missingTxIds mapset.Set[string],
) error {
	for _, txId := range missingTxIds.ToSlice() {
		container, err := chain.FetchContainerFromIndexer(client, txId)
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

		var outs []*database.XChainTxOutput
		switch unsignedTx := tx.Unsigned.(type) {
		case *txs.BaseTx:
			outs, err = XChainTxOutputsFromBaseTx(txId, unsignedTx)
		case *txs.ImportTx:
			outs, err = XChainTxOutputsFromBaseTx(txId, &unsignedTx.BaseTx)
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
