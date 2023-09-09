package cronjob

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/pchain"
	"flare-indexer/utils"
	"flare-indexer/utils/contracts/mirroring"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type testDB struct {
	epochs epochInfo
	states map[string]database.State
	txs    map[int64][]database.PChainTxData
}

func (db testDB) FetchState(name string) (database.State, error) {
	state, ok := db.states[name]
	if !ok {
		return state, errors.New("not found")
	}

	return state, nil
}

func (db testDB) UpdateJobState(epoch int64) error {
	return nil
}

func (db testDB) GetPChainTxsForEpoch(start, end time.Time) ([]database.PChainTxData, error) {
	epoch := db.epochs.getEpochIndex(start)
	return db.txs[epoch], nil
}

type testContracts struct {
	merkleRoots    map[int64][32]byte
	mirroredStakes []mirrorStakeInput
}

type mirrorStakeInput struct {
	stakeData   *mirroring.IPChainStakeMirrorVerifierPChainStake
	merkleProof [][32]byte
}

func (c testContracts) GetMerkleRoot(epoch int64) ([32]byte, error) {
	return c.merkleRoots[epoch], nil
}

func (c *testContracts) MirrorStake(
	stakeData *mirroring.IPChainStakeMirrorVerifierPChainStake,
	merkleProof [][32]byte,
) error {
	c.mirroredStakes = append(c.mirroredStakes, mirrorStakeInput{
		stakeData:   stakeData,
		merkleProof: merkleProof,
	})
	return nil
}

func TestMirror(t *testing.T) {
	cfg := config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "costwo",
		},
		Logger: globalConfig.LoggerConfig{
			Level:   "DEBUG",
			Console: true,
		},
	}
	globalConfig.GlobalConfigCallback.Call(cfg)

	epochCfg := config.EpochConfig{
		Period: 180 * time.Second,
		Start:  utils.Timestamp{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}
	epochs := newEpochInfo(&epochCfg)

	txid := "5uZETr5SUKqGJLzFP5BeGxbXU5CFcCBQYPu288eX9R1QDQMjn"
	startTime := epochs.getStartTime(3)
	endTime := epochs.getEndTime(999)

	tx := database.PChainTxData{
		PChainTx: database.PChainTx{
			ChainID:   "costwo",
			NodeID:    "NodeID-CZYx3on11wwYXFoHwZtAQZT5unZ9JHMf6",
			StartTime: &startTime,
			EndTime:   &endTime,
			TxID:      &txid,
			Type:      database.PChainAddDelegatorTx,
		},
		InputAddress: "costwo18atl0e95w5ym6t8u5yrjpz35vqqzxfzrrsnq8u",
	}

	db := testDB{
		epochs: epochs,
		states: map[string]database.State{
			pchain.StateName: {
				Updated: endTime,
			},
			mirrorStateName: {},
		},
		txs: map[int64][]database.PChainTxData{
			3: {tx},
		},
	}

	txHash, err := hashTransaction(&tx)
	require.NoError(t, err)

	contracts := testContracts{
		merkleRoots: map[int64][32]byte{
			3: txHash,
		},
	}

	j := mirrorCronJob{
		db:        db,
		contracts: &contracts,
		enabled:   true,
		epochs:    epochs,
	}

	err = j.Call()
	require.NoError(t, err)

	require.Len(t, contracts.mirroredStakes, 1)
	cupaloy.SnapshotT(t, contracts.mirroredStakes[0])
}
