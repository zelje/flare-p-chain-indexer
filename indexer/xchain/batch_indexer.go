package xchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/shared"
	"flare-indexer/logger"
	"time"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x"
	"gorm.io/gorm"
)

// Indexer for X-chain transactions of "type" baseTx
// Implements ContainerBatchIndexer

type txBatchIndexer struct {
	db     *gorm.DB
	client indexer.Client

	inOutIndexer *shared.InputOutputIndexer
	newTxs       []*database.XChainTx
}

func NewXChainBatchIndexer(
	db *gorm.DB,
	client indexer.Client,
) *txBatchIndexer {
	updater := newXChainInputUpdater(db, client)
	return &txBatchIndexer{
		db:     db,
		client: client,

		inOutIndexer: shared.NewInputOutputIndexer(updater),
		newTxs:       make([]*database.XChainTx, 0),
	}
}

func (xi *txBatchIndexer) Reset(containerLen int) {
	xi.newTxs = make([]*database.XChainTx, 0, containerLen)
	xi.inOutIndexer.Reset()
}

func (xi *txBatchIndexer) AddContainer(index uint64, container indexer.Container) error {
	tx, err := x.Parser.ParseGenesisTx(container.Bytes)
	if err != nil {
		return err
	}

	switch unsignedTx := tx.Unsigned.(type) {
	case *txs.BaseTx:
		err := xi.addTx(&container, unsignedTx, database.XChainBaseTx, index)
		if err != nil {
			return err
		}
	case *txs.ImportTx:
		err := xi.addTx(&container, &unsignedTx.BaseTx, database.XChainImportTx, index)
		if err != nil {
			return err
		}
	default:
		logger.Warn("Transaction with id '%s' is NOT indexed, type is %T", container.ID, unsignedTx)
	}
	return nil
}

func (xi *txBatchIndexer) ProcessBatch() error {
	return xi.inOutIndexer.ProcessBatch()
}

func (xi *txBatchIndexer) addTx(container *indexer.Container, baseTx *txs.BaseTx, txType database.XChainTxType, index uint64) error {
	tx := &database.XChainTx{}
	tx.TxID = container.ID.String()
	tx.TxIndex = index
	tx.Type = txType
	tx.Timestamp = time.Unix(container.Timestamp/1e9, container.Timestamp%1e9)
	tx.Memo = string(baseTx.Memo)
	tx.Bytes = container.Bytes

	xi.newTxs = append(xi.newTxs, tx)

	return xi.inOutIndexer.AddTx(tx.TxID, baseTx)
}

// Persist all entities
func (i *txBatchIndexer) PersistEntities(db *gorm.DB) error {
	ins := i.inOutIndexer.GetIns()
	dbIns := make([]*database.XChainTxInput, 0, len(ins))
	for _, in := range i.inOutIndexer.GetIns() {
		dbIns = append(dbIns, &database.XChainTxInput{
			TxInput: *in,
		})
	}

	outs := i.inOutIndexer.GetOuts()
	dbOuts := make([]*database.XChainTxOutput, 0, len(outs))
	for _, out := range outs {
		dbOuts = append(dbOuts, &database.XChainTxOutput{
			TxOutput: *out,
		})
	}
	return database.CreateXChainEntities(db, i.newTxs, dbIns, dbOuts)
}
