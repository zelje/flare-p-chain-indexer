package indexer

import (
	"flare-indexer/src/chain"
	"flare-indexer/src/dbmodel"
	"flare-indexer/src/logger"
	"time"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x"
	"gorm.io/gorm"
)

const (
	StateName      string        = "x_chain_tx"
	MaxBatch       int           = 100
	IndexerTimeout time.Duration = 3 * time.Minute
)

type XChainIndexer interface {
	Run() error
}

type xChainIndexer struct {
	db     *gorm.DB
	client indexer.Client
}

func CreateXChainIndexer(ctx IndexerContext) XChainIndexer {
	idxr := xChainIndexer{}

	idxr.client = ctx.Clients().XChainTxClient()
	idxr.db = ctx.DB()
	return &idxr
}

func (xi *xChainIndexer) Run() error {

	// Get current state of tx indexer from db
	currentState := dbmodel.FetchState(xi.db, StateName)

	// Get MaxBatch containers from the chain
	containers, err := fetchContainerRangeFromIndexer(xi.client, currentState.NextDBIndex, MaxBatch)
	if err != nil {
		return err
	}

	baseTxIndexer := NewBaseTxIndexer(len(containers))

	for i, container := range containers {
		index := currentState.NextDBIndex + uint64(i)

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

	// currentState.NextDBIndex += uint64(len(containers))
	// baseTxIndexer.UpdateIns(xi.db, xi.client)

	// tx := xi.db.Begin()
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		tx.Rollback()
	// 	}
	// }()

	return nil
}
