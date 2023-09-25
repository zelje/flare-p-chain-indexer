package cronjob

import (
	"flare-indexer/database"
	indexerctx "flare-indexer/indexer/context"
	"flare-indexer/indexer/pchain"
	"flare-indexer/logger"
	"flare-indexer/utils"
	"flare-indexer/utils/chain"
	"flare-indexer/utils/contracts/mirroring"
	"flare-indexer/utils/merkle"
	"flare-indexer/utils/staking"
	"math/big"
	"strings"
	"time"

	"github.com/ava-labs/avalanchego/utils/crypto"
	"github.com/pkg/errors"
)

const mirrorStateName = "mirror_cronjob"

type mirrorCronJob struct {
	epochCronjob
	db        mirrorDB
	contracts mirrorContracts
	time      utils.ShiftedTime
}

type mirrorDB interface {
	FetchState(name string) (database.State, error)
	UpdateJobState(epoch int64, force bool) error
	GetPChainTxsForEpoch(start, end time.Time) ([]database.PChainTxData, error)
	GetPChainTx(txID string, address string) (*database.PChainTxData, error)
}

type mirrorContracts interface {
	GetMerkleRoot(epoch int64) ([32]byte, error)
	MirrorStake(
		stakeData *mirroring.IPChainStakeMirrorVerifierPChainStake,
		merkleProof [][32]byte,
	) error
	IsAddressRegistered(address string) (bool, error)
	RegisterPublicKey(publicKey crypto.PublicKey) error
}

func NewMirrorCronjob(ctx indexerctx.IndexerContext) (Cronjob, error) {
	cfg := ctx.Config()

	if !cfg.Mirror.Enabled {
		return &mirrorCronJob{}, nil
	}

	contracts, err := initMirrorJobContracts(cfg)
	if err != nil {
		return nil, err
	}

	mc := &mirrorCronJob{
		epochCronjob: newEpochCronjob(&cfg.Mirror.CronjobConfig, &cfg.Epochs),
		db:           NewMirrorDBGorm(ctx.DB()),
		contracts:    contracts,
	}
	mc.reset(ctx.Flags().ResetMirrorCronjob)
	return mc, nil
}

func (c *mirrorCronJob) Name() string {
	return "mirror"
}

func (c *mirrorCronJob) Timeout() time.Duration {
	return c.epochs.Period
}

func (c *mirrorCronJob) OnStart() error {
	return nil
}

func (c *mirrorCronJob) Call() error {
	epochRange, err := c.getEpochRange()
	if err != nil {
		if errors.Is(err, errNoEpochsToMirror) {
			logger.Debug("no epochs to mirror")
			return nil
		}

		return err
	}

	logger.Debug("mirroring epochs %d-%d", epochRange.start, epochRange.end)

	idxState, err := c.db.FetchState(pchain.StateName)
	if err != nil {
		return err
	}

	for epoch := epochRange.start; epoch <= epochRange.end; epoch++ {
		// Skip updating if indexer is behind
		if c.indexerBehind(&idxState, epoch) {
			logger.Debug("indexer is behind, skipping mirror")
			return nil
		}

		logger.Debug("mirroring epoch %d", epoch)
		if err := c.mirrorEpoch(epoch); err != nil {
			return err
		}
	}

	logger.Debug("successfully mirrored epochs %d-%d", epochRange.start, epochRange.end)

	if err := c.db.UpdateJobState(epochRange.end+1, false); err != nil {
		return err
	}

	return nil
}

var errNoEpochsToMirror = errors.New("no epochs to mirror")

func (c *mirrorCronJob) getEpochRange() (*epochRange, error) {
	startEpoch, err := c.getStartEpoch()
	if err != nil {
		return nil, err
	}

	logger.Debug("start epoch: %d", startEpoch)

	endEpoch, err := c.getEndEpoch(startEpoch)
	if err != nil {
		return nil, err
	}
	logger.Debug("Mirroring needed for epochs [%d, %d]", startEpoch, endEpoch)
	return c.getTrimmedEpochRange(startEpoch, endEpoch), nil
}

func (c *mirrorCronJob) getStartEpoch() (int64, error) {
	jobState, err := c.db.FetchState(mirrorStateName)
	if err != nil {
		return 0, err
	}

	return int64(jobState.NextDBIndex), nil
}

func (c *mirrorCronJob) getEndEpoch(startEpoch int64) (int64, error) {
	currEpoch := c.epochs.GetEpochIndex(c.time.Now())
	logger.Debug("current epoch: %d", currEpoch)

	for epoch := currEpoch; epoch > startEpoch; epoch-- {
		confirmed, err := c.isEpochConfirmed(epoch)
		if err != nil {
			return 0, err
		}

		if confirmed {
			return epoch, nil
		}
	}

	return 0, errNoEpochsToMirror
}

func (c *mirrorCronJob) isEpochConfirmed(epoch int64) (bool, error) {
	merkleRoot, err := c.contracts.GetMerkleRoot(epoch)
	if err != nil {
		return false, errors.Wrap(err, "votingContract.GetMerkleRoot")
	}

	return merkleRoot != [32]byte{}, nil
}

func (c *mirrorCronJob) mirrorEpoch(epoch int64) error {
	txs, err := c.getUnmirroredTxs(epoch)
	if err != nil {
		return err
	}

	if len(txs) == 0 {
		logger.Debug("no unmirrored txs found")
		return nil
	}

	logger.Info("mirroring %d txs", len(txs))
	if err := c.mirrorTxs(txs, epoch); err != nil {
		return err
	}

	return nil
}

func (c *mirrorCronJob) getUnmirroredTxs(epoch int64) ([]database.PChainTxData, error) {
	startTimestamp, endTimestamp := c.epochs.GetTimeRange(epoch)

	txs, err := c.db.GetPChainTxsForEpoch(startTimestamp, endTimestamp)
	if err != nil {
		return nil, err
	}

	return staking.DedupeTxs(txs), nil
}

func (c *mirrorCronJob) mirrorTxs(txs []database.PChainTxData, epochID int64) error {
	merkleTree, err := staking.BuildTree(txs)
	if err != nil {
		return err
	}

	if err := c.checkMerkleRoot(merkleTree, epochID); err != nil {
		return err
	}

	for i := range txs {
		in := mirrorTxInput{
			epochID:    big.NewInt(epochID),
			merkleTree: merkleTree,
			tx:         &txs[i],
		}

		if err := c.mirrorTx(&in); err != nil {
			return err
		}
	}

	return nil
}

func (c *mirrorCronJob) checkMerkleRoot(tree merkle.Tree, epoch int64) error {
	root, err := tree.Root()
	if err != nil {
		return err
	}

	contractRoot, err := c.contracts.GetMerkleRoot(epoch)
	if err != nil {
		return errors.Wrap(err, "votingContract.GetMerkleRoot")
	}

	if root != contractRoot {
		return errors.Errorf("merkle root mismatch: got %x, expected %x", root, contractRoot)
	}

	return nil
}

type mirrorTxInput struct {
	epochID    *big.Int
	merkleTree merkle.Tree
	tx         *database.PChainTxData
}

func (c *mirrorCronJob) mirrorTx(in *mirrorTxInput) error {
	stakeData, err := staking.ToStakeData(in.tx)
	if err != nil {
		return err
	}

	merkleProof, err := staking.GetMerkleProof(in.merkleTree, in.tx)
	if err != nil {
		return err
	}

	// Register addresses if needed, do not fail if not successful
	if err := c.registerAddress(*in.tx.TxID, in.tx.InputAddress); err != nil {
		logger.Error("error registering address: %s", err.Error())
	} else {
		logger.Info("registered address %s on address binder contract", in.tx.InputAddress)
	}

	logger.Debug("mirroring tx %s", *in.tx.TxID)
	err = c.contracts.MirrorStake(stakeData, merkleProof)
	if err != nil {
		if strings.Contains(err.Error(), "transaction already mirrored") {
			logger.Info("tx %s already mirrored", *in.tx.TxID)
			return nil
		}

		if strings.Contains(err.Error(), "staking already ended") {
			logger.Info("staking already ended for tx %s", *in.tx.TxID)
			return nil
		}

		if strings.Contains(err.Error(), "unknown staking address") {
			logger.Info("unknown staking address for tx %s", *in.tx.TxID)
			return nil
		}

		if strings.Contains(err.Error(), "Max node ids exceeded") {
			logger.Info("Max node ids exceeded for tx %s", *in.tx.TxID)
			return nil
		}

		return errors.Wrap(err, "mirroringContract.MirrorStake")
	}

	return nil
}

func (c *mirrorCronJob) registerAddress(txID string, address string) error {
	registered, err := c.contracts.IsAddressRegistered(address)
	if err != nil || registered {
		return err
	}
	tx, err := c.db.GetPChainTx(txID, address)
	if err != nil {
		return err
	}
	if tx == nil {
		return errors.New("tx not found")
	}
	publicKeys, err := chain.PublicKeysFromPChainBlock(tx.Bytes)
	if err != nil {
		return err
	}
	if tx.InputIndex >= uint32(len(publicKeys)) {
		return errors.New("input index out of range")
	}
	publicKey := publicKeys[tx.InputIndex]
	for _, k := range publicKey {
		err := c.contracts.RegisterPublicKey(k)
		if err != nil {
			return errors.Wrap(err, "mirroringContract.RegisterPublicKey")
		}
	}
	return nil
}

func (c *mirrorCronJob) reset(firstEpoch int64) error {
	if firstEpoch <= 0 {
		return nil
	}

	logger.Info("Resetting mirroring cronjob state to epoch %d", firstEpoch)
	err := c.db.UpdateJobState(firstEpoch, true)
	if err != nil {
		return err
	}
	c.epochs.First = firstEpoch
	return nil
}
