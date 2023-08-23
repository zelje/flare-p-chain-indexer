package cronjob

import (
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/utils"
	"flare-indexer/utils/contracts/voting"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

const (
	StateName string = "voting_cronjob"
)

type votingCronJob struct {
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
	voterAddress   common.Address
}

func NewVotingCronJob(ctx context.IndexerContext) (Cronjob, error) {
	cfg := ctx.Config()
	votingContract, err := newVotingContract(cfg)
	if err != nil {
		return nil, err
	}
	voterAddress, err := newVoterAddress(cfg)
	if err != nil {
		return nil, err
	}

	return &votingCronJob{
		enabled:        cfg.VotingCronjob.Enabled,
		timeout:        cfg.VotingCronjob.TimeoutSeconds,
		running:        false,
		db:             ctx.DB(),
		start:          cfg.VotingCronjob.EpochStart,
		interval:       cfg.VotingCronjob.EpochPeriod,
		votingContract: votingContract,
		voterAddress:   voterAddress,
	}, nil
}

func newVotingContract(cfg *config.Config) (*voting.Voting, error) {
	eth, err := ethclient.Dial(cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}
	addr := common.HexToAddress(cfg.VotingCronjob.ContractAddress)
	return voting.NewVoting(addr, eth)
}

func newVoterAddress(cfg *config.Config) (common.Address, error) {
	_, addressBytes, err := utils.ParseAddress(cfg.VotingCronjob.VoterAddress)
	if err != nil {
		return common.Address{}, err
	}
	return common.BytesToAddress(addressBytes), nil
}

func (c *votingCronJob) Name() string {
	return "mirror"
}

func (c *votingCronJob) Enabled() bool {
	return c.enabled
}

func (c *votingCronJob) TimeoutSeconds() int {
	return c.timeout
}

func (c *votingCronJob) Call() error {
	state, err := database.FetchState(c.db, StateName)
	if err != nil {
		return err
	}

	// Last epoch that was submitted to the contract
	nextEpochToSubmit := state.NextDBIndex
	lastEpochToSubmit := c.getEpochIndex(time.Now()) - 1
	for e := int64(nextEpochToSubmit); e <= lastEpochToSubmit; e++ {
		start, end := c.getEpochBounds(e)
		votingData, err := database.FetchPChainVotingData(c.db, start, end)
		if err != nil {
			return err
		}
		shouldVote, err := c.votingContract.ShouldVote(nil, big.NewInt(e), c.voterAddress)
		if err != nil {
			return err
		}
		if shouldVote {
			merkleRoot, err := getMerkleRoot(votingData)
			if err != nil {
				return err
			}
			c.votingContract.SubmitVote(nil, big.NewInt(e), [32]byte(merkleRoot))
			state.NextDBIndex = uint64(e + 1)

			// Update state
			database.UpdateState(c.db, &state)
		}
	}
	return nil
}

func (c *votingCronJob) getEpochIndex(t time.Time) int64 {
	return (t.Unix() - c.start) / c.interval
}

func (c *votingCronJob) getEpochBounds(epoch int64) (start, end time.Time) {
	start = time.Unix(c.start+(epoch*c.interval), 0)
	end = start.Add(time.Duration(c.interval) * time.Second)
	return
}

func getMerkleRoot(votingData []database.PChainVotingData) ([32]byte, error) {
	return [32]byte{}, nil
}
