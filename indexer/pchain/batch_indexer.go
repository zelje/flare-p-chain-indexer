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
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/blocks"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/proposervm/block"
	"github.com/ybbus/jsonrpc/v3"
	"gorm.io/gorm"
)

// Indexer for P-chain transactions. Implements ContainerBatchIndexer
type txBatchIndexer struct {
	db     *gorm.DB
	client indexer.Client

	inOutIndexer *shared.InputOutputIndexer
	newTxs       []*database.PChainTx
}

func NewPChainBatchIndexer(
	ctx context.IndexerContext,
	client indexer.Client,
	rpcClient jsonrpc.RPCClient,
) *txBatchIndexer {
	updater := newPChainInputUpdater(ctx, rpcClient)
	return &txBatchIndexer{
		db:     ctx.DB(),
		client: client,

		inOutIndexer: shared.NewInputOutputIndexer(updater),
		newTxs:       make([]*database.PChainTx, 0),
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
	dbTx.TxID = tx.ID().String()
	dbTx.BlockID = container.ID.String()
	dbTx.BlockIndex = index
	dbTx.Timestamp = time.Unix(container.Timestamp, 0)
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
	ownerAddress, err := shared.RewardsOwnerAddress(tx.RewardsOwner)
	if err != nil {
		return err
	}
	dbTx.RewardsOwner = ownerAddress
	return xi.updateAddStakerTx(dbTx, tx, tx.Ins)
}

func (xi *txBatchIndexer) updateAddDelegatorTx(dbTx *database.PChainTx, tx *txs.AddDelegatorTx) error {
	dbTx.Type = database.PChainAddDelegatorTx
	ownerAddress, err := shared.RewardsOwnerAddress(tx.DelegationRewardsOwner)
	if err != nil {
		return err
	}
	dbTx.RewardsOwner = ownerAddress
	return xi.updateAddStakerTx(dbTx, tx, tx.Ins)
}

func (xi *txBatchIndexer) updateImportTx(dbTx *database.PChainTx, tx *txs.ImportTx) error {
	dbTx.Type = database.PChainImportTx
	dbTx.ChainID = tx.SourceChain.String()
	xi.newTxs = append(xi.newTxs, dbTx)
	return xi.inOutIndexer.AddFromBaseTx(dbTx.TxID, &tx.BaseTx.BaseTx, PChainDefaultInputOutputCreator)
}

func (xi *txBatchIndexer) updateExportTx(dbTx *database.PChainTx, tx *txs.ExportTx) error {
	dbTx.Type = database.PChainExportTx
	dbTx.ChainID = tx.DestinationChain.String()
	xi.newTxs = append(xi.newTxs, dbTx)
	return xi.inOutIndexer.AddFromBaseTx(dbTx.TxID, &tx.BaseTx.BaseTx, PChainDefaultInputOutputCreator)
}

// Persist all entities
func (xi *txBatchIndexer) PersistEntities(db *gorm.DB) error {
	ins, err := utils.CastArray[*database.PChainTxInput](xi.inOutIndexer.GetIns())
	if err != nil {
		return err
	}
	outs, err := utils.CastArray[*database.PChainTxOutput](xi.inOutIndexer.GetOuts())
	if err != nil {
		return err
	}
	return database.CreatePChainEntities(db, xi.newTxs, ins, outs)
}

// Common code for AddDelegatorTx and AddValidatorTx
func (xi *txBatchIndexer) updateAddStakerTx(
	dbTx *database.PChainTx,
	tx txs.PermissionlessStaker,
	txIns []*avax.TransferableInput,
) error {
	dbTx.NodeID = tx.NodeID().String()
	dbTx.StartTime = tx.StartTime()
	dbTx.EndTime = tx.EndTime()
	dbTx.Weight = tx.Weight()

	outs, err := getAddStakerTxOutputs(dbTx.TxID, tx)
	if err != nil {
		return err
	}
	ins := shared.InputsFromTxIns(dbTx.TxID, txIns, PChainDefaultInputOutputCreator)

	xi.newTxs = append(xi.newTxs, dbTx)
	xi.inOutIndexer.Add(dbTx.TxID, outs, ins)
	return nil
}

func getAddStakerTxOutputs(txID string, tx txs.PermissionlessStaker) ([]shared.Output, error) {
	outs, err := shared.OutputsFromTxOuts(txID, tx.Outputs(), PChainDefaultInputOutputCreator)
	if err != nil {
		return nil, err
	}
	stakeOuts, err := shared.OutputsFromTxOutsI(txID, tx.Stake(), len(outs), PChainStakerInputOutputCreator)
	if err != nil {
		return nil, err
	}
	outs = append(outs, stakeOuts...)
	return outs, nil
}
