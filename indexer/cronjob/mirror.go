package cronjob

import (
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	indexerctx "flare-indexer/indexer/context"
	"flare-indexer/indexer/pchain"
	"flare-indexer/logger"
	"flare-indexer/utils"
	"flare-indexer/utils/contracts/mirroring"
	"flare-indexer/utils/contracts/voting"
	"flare-indexer/utils/merkle"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const mirrorStateName = "mirror_cronjob"

type mirrorCronJob struct {
	db                *gorm.DB
	enabled           bool
	epochs            epochInfo
	mirroringContract *mirroring.Mirroring
	txOpts            *bind.TransactOpts
	votingContract    *voting.Voting
	time              utils.ShiftedTime
}

func NewMirrorCronjob(ctx indexerctx.IndexerContext) (*mirrorCronJob, error) {
	cfg := ctx.Config()

	if !cfg.Mirror.Enabled {
		return &mirrorCronJob{}, nil
	}

	contracts, err := initMirrorJobContracts(cfg)
	if err != nil {
		return nil, err
	}

	txOpts, err := TransactOptsFromPrivateKey(cfg.Chain.PrivateKey, cfg.Chain.ChainID)
	if err != nil {
		return nil, err
	}

	return &mirrorCronJob{
		db:                ctx.DB(),
		enabled:           cfg.Mirror.Enabled,
		epochs:            newEpochInfo(&cfg.Epochs),
		mirroringContract: contracts.mirroring,
		txOpts:            txOpts,
		votingContract:    contracts.voting,
	}, nil
}

type mirrorJobContracts struct {
	mirroring *mirroring.Mirroring
	voting    *voting.Voting
}

func initMirrorJobContracts(cfg *config.Config) (*mirrorJobContracts, error) {
	if cfg.Mirror.MirroringContract == (common.Address{}) {
		return nil, errors.New("mirroring contract address not set")
	}

	if cfg.VotingCronjob.ContractAddress == (common.Address{}) {
		return nil, errors.New("voting contract address not set")
	}

	eth, err := ethclient.Dial(cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}

	mirroringContract, err := mirroring.NewMirroring(cfg.Mirror.MirroringContract, eth)
	if err != nil {
		return nil, err
	}

	votingContract, err := voting.NewVoting(cfg.VotingCronjob.ContractAddress, eth)
	if err != nil {
		return nil, err
	}

	return &mirrorJobContracts{
		mirroring: mirroringContract,
		voting:    votingContract,
	}, nil
}

func (c *mirrorCronJob) Name() string {
	return "mirror"
}

func (c *mirrorCronJob) Enabled() bool {
	return c.enabled
}

func (c *mirrorCronJob) Timeout() time.Duration {
	return c.epochs.period
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

	idxState, err := database.FetchState(c.db, pchain.StateName)
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

	if err := c.updateJobState(epochRange.end); err != nil {
		return err
	}

	return nil
}

func (c *mirrorCronJob) indexerBehind(idxState *database.State, epoch int64) bool {
	epochEnd := c.epochs.getEndTime(epoch)
	return epochEnd.After(idxState.Updated)
}

func (c *mirrorCronJob) updateJobState(epoch int64) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		jobState, err := database.FetchState(tx, mirrorStateName)
		if err != nil {
			return errors.Wrap(err, "database.FetchState")
		}

		if jobState.NextDBIndex >= uint64(epoch) {
			logger.Debug("job state already up to date")
			return nil
		}

		jobState.NextDBIndex = uint64(epoch)

		return database.UpdateState(tx, &jobState)
	})
}

type epochRange struct {
	start int64
	end   int64
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

	return &epochRange{
		start: startEpoch,
		end:   endEpoch,
	}, nil
}

func (c *mirrorCronJob) getStartEpoch() (int64, error) {
	jobState, err := database.FetchState(c.db, mirrorStateName)
	if err != nil {
		return 0, err
	}

	return int64(utils.Max(jobState.NextDBIndex, c.epochs.first)), nil
}

func (c *mirrorCronJob) getEndEpoch(startEpoch int64) (int64, error) {
	currEpoch := c.epochs.getEpochIndex(c.time.Now())
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
	merkleRoot, err := c.votingContract.GetMerkleRoot(new(bind.CallOpts), big.NewInt(epoch))
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
	startTimestamp, endTimestamp := c.epochs.getTimeRange(epoch)

	txs, err := database.GetPChainTxsForEpoch(&database.GetPChainTxsForEpochInput{
		DB:             c.db,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	})
	if err != nil {
		return nil, err
	}

	return dedupeTxs(txs), nil
}

func (c *mirrorCronJob) mirrorTxs(txs []database.PChainTxData, epochID int64) error {
	merkleTree, err := buildTree(txs)
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

	contractRoot, err := c.votingContract.GetMerkleRoot(new(bind.CallOpts), big.NewInt(epoch))
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
	stakeData, err := toStakeData(in.tx)
	if err != nil {
		return err
	}

	merkleProof, err := getMerkleProof(in.merkleTree, in.tx)
	if err != nil {
		return err
	}

	_, err = c.mirroringContract.MirrorStake(c.txOpts, *stakeData, merkleProof)
	if err != nil {
		return errors.Wrap(err, "mirroringContract.MirrorStake")
	}

	return nil
}

func getTxType(txType database.PChainTxType) (uint8, error) {
	switch txType {
	case database.PChainAddValidatorTx:
		return 0, nil

	case database.PChainAddDelegatorTx:
		return 1, nil

	default:
		return 0, errors.New("invalid tx type")
	}
}

func getMerkleProof(merkleTree merkle.Tree, tx *database.PChainTxData) ([][32]byte, error) {
	hash, err := hashTransaction(tx)
	if err != nil {
		return nil, err
	}

	proof, err := merkleTree.GetProofFromHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "merkleTree.GetProof")
	}

	proofBytes := make([][32]byte, len(proof))
	for i := range proof {
		proofBytes[i] = [32]byte(proof[i])
	}

	return proofBytes, nil
}
