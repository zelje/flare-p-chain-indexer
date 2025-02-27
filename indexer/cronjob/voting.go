package cronjob

import (
	"flare-indexer/database"
	indexerctx "flare-indexer/indexer/context"
	"flare-indexer/indexer/pchain"
	"flare-indexer/logger"
	"flare-indexer/utils"
	"flare-indexer/utils/staking"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

const (
	votingStateName string = "voting_cronjob"
)

var (
	zeroBytes     [32]byte    = [32]byte{}
	zeroBytesHash common.Hash = crypto.Keccak256Hash(zeroBytes[:])

	ErrEpochConfig = errors.New("epoch config mismatch")
)

type votingCronjob struct {
	epochCronjob

	db       votingDB
	contract votingContract

	// For testing to set "now" to some past date
	time utils.ShiftedTime
}

type votingDB interface {
	FetchState(name string) (database.State, error)
	FetchPChainVotingData(start, end time.Time) ([]database.PChainTxData, error)
	UpdateState(state *database.State) error
}

type votingContract interface {
	ShouldVote(epoch *big.Int) (bool, error)
	SubmitVote(epoch *big.Int, merkleRoot [32]byte) error
	EpochConfig() (time.Time, time.Duration, error)
}

func NewVotingCronjob(ctx indexerctx.IndexerContext) (*votingCronjob, error) {
	cfg := ctx.Config()
	if !cfg.VotingCronjob.Enabled {
		return &votingCronjob{}, nil
	}

	db := &votingDBGorm{g: ctx.DB()}
	contract, err := newVotingContractCChain(cfg)
	if err != nil {
		return nil, err
	}

	start, period, err := contract.EpochConfig()
	if err != nil {
		return nil, err
	}

	epochs := staking.NewEpochInfo(&cfg.VotingCronjob.EpochConfig, start, period)

	vc := &votingCronjob{
		epochCronjob: newEpochCronjob(&cfg.VotingCronjob.CronjobConfig, epochs),
		db:           db,
		contract:     contract,
	}

	err = vc.reset(ctx.Flags().ResetVotingCronjob)
	if err != nil {
		return nil, err
	}
	return vc, nil
}

func (c *votingCronjob) Name() string {
	return "voting"
}

func (c *votingCronjob) OnStart() error {
	return nil
}

func (c *votingCronjob) Call() error {
	idxState, err := c.db.FetchState(pchain.StateName)
	if err != nil {
		return err
	}

	state, err := c.db.FetchState(votingStateName)
	if err != nil {
		return err
	}

	now := c.time.Now()

	// Last epoch that was submitted to the contract
	epochRange := c.getEpochRange(int64(state.NextDBIndex), now)
	logger.Debug("Voting needed for epochs [%d, %d]", epochRange.start, epochRange.end)
	for e := epochRange.start; e <= epochRange.end; e++ {
		start, end := c.epochs.GetTimeRange(e)

		if c.indexerBehind(&idxState, e) {
			logger.Debug("indexer is behind, skipping voting for epoch %d", e)
			return nil
		}

		votingData, err := c.db.FetchPChainVotingData(start, end)
		if err != nil {
			return err
		}
		err = c.submitVotes(e, votingData)
		if err != nil {
			return err
		}
		logger.Info("Submitted votes for epoch %d", e)

		// Update state
		state.NextDBIndex = uint64(e + 1)
		if err := c.db.UpdateState(&state); err != nil {
			return err
		}
	}
	return nil
}

func (c *votingCronjob) submitVotes(e int64, votingData []database.PChainTxData) error {
	votingData = staking.DedupeTxs(votingData)

	shouldVote, err := c.contract.ShouldVote(big.NewInt(e))
	if err != nil {
		return err
	}
	if !shouldVote {
		return nil
	}

	var merkleRoot common.Hash
	if len(votingData) == 0 {
		merkleRoot = zeroBytesHash
	} else {
		merkleRoot, err = staking.GetMerkleRoot(votingData)
		if err != nil {
			return err
		}
	}
	err = c.contract.SubmitVote(big.NewInt(e), [32]byte(merkleRoot))
	return err
}

func (c *votingCronjob) reset(firstEpoch int64) error {
	if firstEpoch <= 0 {
		return nil
	}

	logger.Info("Resetting voting cronjob state to epoch %d", firstEpoch)
	state, err := c.db.FetchState(votingStateName)
	if err != nil {
		return err
	}
	state.NextDBIndex = uint64(firstEpoch)
	err = c.db.UpdateState(&state)
	if err != nil {
		return err
	}
	c.epochs.First = firstEpoch
	return nil
}
