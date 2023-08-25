package cronjob

import (
	"context"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	idxCtx "flare-indexer/indexer/context"
	"flare-indexer/indexer/pchain"
	"flare-indexer/utils/contracts/voting"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

const (
	StateName string = "voting_cronjob"
)

type votingCronjob struct {
	enabled bool
	timeout int

	// Lock to prevent concurrent aggregation
	running bool

	// epoch start timestamp (unix seconds)
	start int64

	// Epoch duration in seconds
	interval int64

	db             *gorm.DB
	votingContract *voting.Voting
	txOpts         *bind.TransactOpts
}

func NewVotingCronjob(ctx idxCtx.IndexerContext) (Cronjob, error) {
	cfg := ctx.Config()

	if !cfg.VotingCronjob.Enabled {
		return &votingCronjob{}, nil
	}

	votingContract, err := newVotingContract(cfg)
	if err != nil {
		return nil, err
	}
	txOpts, err := TransactOptsFromPrivateKey(cfg.VotingCronjob.PrivateKey, cfg.Chain.ChainID)
	if err != nil {
		return nil, err
	}

	return &votingCronjob{
		enabled:        cfg.VotingCronjob.Enabled,
		timeout:        cfg.VotingCronjob.TimeoutSeconds,
		running:        false,
		db:             ctx.DB(),
		start:          cfg.VotingCronjob.EpochStart,
		interval:       cfg.VotingCronjob.EpochPeriod,
		votingContract: votingContract,
		txOpts:         txOpts,
	}, nil
}

func newVotingContract(cfg *config.Config) (*voting.Voting, error) {
	eth, err := ethclient.Dial(cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}
	return voting.NewVoting(cfg.VotingCronjob.ContractAddress, eth)
}

func (c *votingCronjob) Name() string {
	return "mirror"
}

func (c *votingCronjob) Enabled() bool {
	return c.enabled
}

func (c *votingCronjob) TimeoutSeconds() int {
	return c.timeout
}

func (c *votingCronjob) Call() error {
	idxState, err := database.FetchState(c.db, pchain.StateName)
	if err != nil {
		return err
	}

	now := time.Now()

	// Skip updating if indexer is behind
	if now.After(idxState.Updated) {
		return nil
	}

	state, err := database.FetchState(c.db, StateName)
	if err != nil {
		return err
	}

	callOpts := &bind.CallOpts{
		From:    c.txOpts.From,
		Context: context.Background(),
	}

	// Last epoch that was submitted to the contract
	nextEpochToSubmit := state.NextDBIndex
	lastEpochToSubmit := c.getEpochIndex(now) - 1
	for e := int64(nextEpochToSubmit); e <= lastEpochToSubmit; e++ {
		start, end := c.getEpochBounds(e)
		votingData, err := database.FetchPChainVotingData(c.db, start, end)
		if err != nil {
			return err
		}
		shouldVote, err := c.votingContract.ShouldVote(callOpts, big.NewInt(e), c.txOpts.From)
		if err != nil {
			return err
		}
		if shouldVote {
			merkleRoot, err := getMerkleRoot(votingData)
			if err != nil {
				return err
			}
			c.votingContract.SubmitVote(nil, big.NewInt(e), [32]byte(merkleRoot))
		}
		// Update state
		state.NextDBIndex = uint64(e + 1)
		database.UpdateState(c.db, &state)
	}
	return nil
}

func (c *votingCronjob) getEpochIndex(t time.Time) int64 {
	return (t.Unix() - c.start) / c.interval
}

func (c *votingCronjob) getEpochBounds(epoch int64) (start, end time.Time) {
	start = time.Unix(c.start+(epoch*c.interval), 0)
	end = start.Add(time.Duration(c.interval) * time.Second)
	return
}

func getMerkleRoot(votingData []database.PChainTxData) (common.Hash, error) {
	tree, err := buildTree(votingData)
	if err != nil {
		return [32]byte{}, err
	}
	hash, err := tree.GetHash(0)
	if err != nil {
		return [32]byte{}, err
	}
	return hash, nil
}
