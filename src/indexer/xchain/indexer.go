package xchain

import (
	"flare-indexer/src/chain"
	"flare-indexer/src/config"
	"flare-indexer/src/dbmodel"
	"flare-indexer/src/indexer/ctx"
	"flare-indexer/src/logger"
	"time"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x"
	"gorm.io/gorm"
)

const (
	StateName string = "x_chain_tx"
)

type XChainIndexer interface {
	Run() error
}

type xChainIndexer struct {
	db     *gorm.DB
	client indexer.Client
	config config.IndexerConfig
}

func CreateXChainIndexer(ctx ctx.IndexerContext) XChainIndexer {
	idxr := xChainIndexer{}

	idxr.client = ctx.Clients().XChainTxClient()
	idxr.db = ctx.DB()
	idxr.config = ctx.Config().Indexer
	return &idxr
}

func (xi *xChainIndexer) Run() error {
	startTime := time.Now()

	// Get current state of tx indexer from db
	currentState, err := dbmodel.FetchState(xi.db, StateName)
	if err != nil {
		return err
	}

	_, lastIndex, err := chain.FetchLastAcceptedContainer(xi.client)
	if err != nil {
		return err
	}

	if lastIndex < currentState.NextDBIndex {
		// Nothing to do; no new containers
		logger.Debug("Nothing to do. Last index %d < next to process %d", lastIndex, currentState.NextDBIndex)
		return nil
	}

	// Get MaxBatch containers from the chain
	containers, err := chain.FetchContainerRangeFromIndexer(xi.client, currentState.NextDBIndex, xi.config.BatchSize)
	if err != nil {
		return err
	}

	baseTxIndexer := NewBaseTxIndexer(len(containers))

	var index uint64
	for i, container := range containers {
		index = currentState.NextDBIndex + uint64(i)

		tx, err := x.Parser.ParseGenesisTx(container.Bytes)
		if err != nil {
			return err
		}

		switch unsignedTx := tx.Unsigned.(type) {
		case *txs.BaseTx:
			data, err := chain.XChainTxDataFromBaseTx(&container, unsignedTx, dbmodel.BaseTx, index)
			if err != nil {
				return nil
			}
			baseTxIndexer.AddTx(data)
		case *txs.ImportTx:
			data, err := chain.XChainTxDataFromBaseTx(&container, &unsignedTx.BaseTx, dbmodel.ImportTx, index)
			if err != nil {
				return nil
			}
			baseTxIndexer.AddTx(data)
		default:
			logger.Warn("Transaction with id '%s' is NOT indexed, type is %T", container.ID, unsignedTx)
		}
	}

	err = baseTxIndexer.UpdateIns(xi.db, xi.client)
	if err != nil {
		return err
	}

	err = dbmodel.DoInTransaction(xi.db,
		func(db *gorm.DB) error { return baseTxIndexer.PersistEntities(db) },
		func(db *gorm.DB) error {
			currentState.Update(index+1, lastIndex)
			return dbmodel.UpdateState(db, &currentState)
		},
	)
	if err != nil {
		return err
	}
	endTime := time.Now()
	logger.Info("X-chain transactions processed to index %d (%d new), last accepted index is %d, duration %dms",
		index, len(baseTxIndexer.NewTxs), lastIndex, endTime.Sub(startTime).Milliseconds())
	return nil
}
