package pchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
	"flare-indexer/logger"
	"flare-indexer/utils"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"
	"gorm.io/gorm"
)

// Indexer for P-chain transactions. Implements ContainerBatchIndexer
type txBatchIndexer struct {
	db     *gorm.DB
	client indexer.Client

	inOutIndexer *shared.InputOutputIndexer
	newTxs       []*database.PChainTx
	newStakeOuts []*database.PChainStakeOutput
}

func NewPChainBatchIndexer(
	ctx context.IndexerContext,
	client indexer.Client,
) *txBatchIndexer {
	updater := newPChainInputUpdater(ctx, client)
	return &txBatchIndexer{
		db:     ctx.DB(),
		client: client,

		inOutIndexer: shared.NewInputOutputIndexer(updater),
		newTxs:       make([]*database.PChainTx, 0),
		newStakeOuts: make([]*database.PChainStakeOutput, 0),
	}
}

func (xi *txBatchIndexer) Reset(containerLen int) {
	xi.newTxs = make([]*database.PChainTx, 0, containerLen)
	xi.inOutIndexer.Reset()
}

func (xi *txBatchIndexer) AddContainer(index uint64, container indexer.Container) error {
	blk, err := block.Parse(container.Bytes)
	if err != nil {
		return err
	}
	innerBlk, err := blocks.Parse(blocks.GenesisCodec, blk.Block())
	if err != nil {
		return err
	}
	switch innerBlkType := innerBlk.(type) {
	case *blocks.ApricotProposalBlock:
		tx := innerBlkType.Tx
		xi.addTx(&container, tx, index)
	case *blocks.ApricotCommitBlock:
		logger.Info("Block %d is ApricotCommitBlock. Skipping indexing", index)
	case *blocks.ApricotStandardBlock:
		for _, tx := range innerBlkType.Txs() {
			xi.addTx(&container, tx, index)
		}
	default:
		return fmt.Errorf("block %d has unexpected type %T", index, innerBlkType)
	}
	return nil
}

func (xi *txBatchIndexer) ProcessBatch() error {
	return xi.inOutIndexer.ProcessBatch()
}

func (xi *txBatchIndexer) addTx(container *indexer.Container, tx *txs.Tx, index uint64) error {
	dbTx := &database.PChainTx{}
	dbTx.TxID = container.ID.String()
	dbTx.BlockIndex = index
	dbTx.Timestamp = time.Unix(container.Timestamp/1e9, container.Timestamp%1e9)
	dbTx.Bytes = container.Bytes

	var err error = nil
	switch unsignedTx := tx.Unsigned.(type) {
	case *txs.RewardValidatorTx:
		dbTx.Type = database.PChainRewardValidatorTx
	case *txs.AddValidatorTx:
		err = xi.updateAddValidatorTx(dbTx, unsignedTx)
	case *txs.AddDelegatorTx:
		err = xi.updateAddDelegatorTx(dbTx, unsignedTx)
	case *txs.ImportTx:
		err = xi.updateImportTx(dbTx, unsignedTx)
	case *txs.ExportTx:
		err = xi.updateExportTx(dbTx, unsignedTx)
	default:
		logger.Info("P-chain transaction %s with type %T in block %d is not indexed", dbTx.TxID, unsignedTx, index)
	}
	return err
}

func (xi *txBatchIndexer) updateAddValidatorTx(dbTx *database.PChainTx, tx *txs.AddValidatorTx) error {
	dbTx.Type = database.PChainAddValidatorTx

	err := xi.updateAddStakerTx(dbTx, tx)
	if err != nil {
		return err
	}

	ownerAddress, err := shared.RewardsOwnerAddress(tx.RewardsOwner)
	if err != nil {
		return err
	}
	dbTx.RewardsOwner = ownerAddress

	xi.newTxs = append(xi.newTxs, dbTx)
	return xi.inOutIndexer.AddTx(dbTx.TxID, &tx.BaseTx.BaseTx)
}

func (xi *txBatchIndexer) updateAddDelegatorTx(dbTx *database.PChainTx, tx *txs.AddDelegatorTx) error {
	dbTx.Type = database.PChainAddDelegatorTx

	err := xi.updateAddStakerTx(dbTx, tx)
	if err != nil {
		return err
	}
	ownerAddress, err := shared.RewardsOwnerAddress(tx.DelegationRewardsOwner)
	if err != nil {
		return err
	}
	dbTx.RewardsOwner = ownerAddress

	xi.newTxs = append(xi.newTxs, dbTx)
	return xi.inOutIndexer.AddTx(dbTx.TxID, &tx.BaseTx.BaseTx)
}

// Common code for AddDelegatorTx and AddValidatorTx
func (xi *txBatchIndexer) updateAddStakerTx(dbTx *database.PChainTx, tx txs.PermissionlessStaker) error {
	dbTx.NodeID = tx.NodeID().String()
	dbTx.StartTime = tx.StartTime()
	dbTx.EndTime = tx.EndTime()
	dbTx.Weight = tx.Weight()

	stakeOuts, err := shared.TxOutputsFromTxOuts(dbTx.TxID, tx.Stake())
	if err != nil {
		return err
	}
	xi.newStakeOuts = append(xi.newStakeOuts, utils.Map(stakeOuts, database.PChainStakeOutputFromTxOutput)...)
	return nil
}

func (xi *txBatchIndexer) updateImportTx(dbTx *database.PChainTx, tx *txs.ImportTx) error {
	dbTx.Type = database.PChainImportTx
	dbTx.ChainID = tx.SourceChain.String()
	xi.newTxs = append(xi.newTxs, dbTx)
	return xi.inOutIndexer.AddTx(dbTx.TxID, &tx.BaseTx.BaseTx)
}

func (xi *txBatchIndexer) updateExportTx(dbTx *database.PChainTx, tx *txs.ExportTx) error {
	dbTx.Type = database.PChainExportTx
	dbTx.ChainID = tx.DestinationChain.String()
	xi.newTxs = append(xi.newTxs, dbTx)
	return xi.inOutIndexer.AddTx(dbTx.TxID, &tx.BaseTx.BaseTx)
}

// Persist all entities
func (xi *txBatchIndexer) PersistEntities(db *gorm.DB) error {
	ins := xi.inOutIndexer.GetIns()
	dbIns := utils.Map(ins, database.PChainTxInputFromTxInput)

	outs := xi.inOutIndexer.GetOuts()
	dbOuts := utils.Map(outs, database.PChainTxOutputFromTxOutput)
	return database.CreatePChainEntities(db, xi.newTxs, dbIns, dbOuts, xi.newStakeOuts)
}
