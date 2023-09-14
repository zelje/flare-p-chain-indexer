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

	"github.com/ava-labs/avalanchego/ids"
	"github.com/bradleyjkemp/cupaloy"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// TODO this could go in TestMain but conflicts with the existing TestMain
// defined in the same pkg.
func initMirrorTest() {
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
}

func TestOneTransaction(t *testing.T) {
	initMirrorTest()

	epochs := initEpochs()

	startTime := epochs.getStartTime(3)
	endTime := epochs.getEndTime(999)

	txid := "5uZETr5SUKqGJLzFP5BeGxbXU5CFcCBQYPu288eX9R1QDQMjn"
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

	txs := map[int64][]database.PChainTxData{
		3: {tx},
	}

	txHash, err := hashTransaction(&tx)
	require.NoError(t, err)

	merkleRoots := map[int64][32]byte{
		3: txHash,
	}

	contracts := testContracts{
		merkleRoots: merkleRoots,
	}

	testMirror(t, txs, contracts, epochs)
}

func TestMultipleTransactionsInEpoch(t *testing.T) {
	initMirrorTest()

	epochs := initEpochs()

	startTime := epochs.getStartTime(3)
	endTime := epochs.getEndTime(999)

	txs := make([]database.PChainTxData, 3)
	txIDs := []string{
		"XnfV79XVMyuXbTw8iNreQ9FrUgy9csYBJp1xRscay3oDzhyq8",
		"nsPmyQbm4oo77jyykxbjf7s4Zp4urNptkyAouxVWZ2EB2kw1z",
		"2p32tpqNrfzP3SStbP9bQGHZtJkCxjV3iHNssVnkcpUWxHMSuj",
	}

	for i := 0; i < 3; i++ {
		txs[i] = database.PChainTxData{
			PChainTx: database.PChainTx{
				ChainID:   "costwo",
				NodeID:    "NodeID-CZYx3on11wwYXFoHwZtAQZT5unZ9JHMf6",
				StartTime: &startTime,
				EndTime:   &endTime,
				TxID:      &txIDs[i],
				Type:      database.PChainAddDelegatorTx,
			},
			InputAddress: "costwo18atl0e95w5ym6t8u5yrjpz35vqqzxfzrrsnq8u",
		}
	}

	txsMap := map[int64][]database.PChainTxData{
		3: txs,
	}

	root := common.HexToHash("b3ec965b802c71f9058d2ed4d80bdf5af902a3741a75221992c5eb2f879a116c")

	merkleRoots := map[int64][32]byte{
		3: root,
	}

	contracts := testContracts{
		merkleRoots: merkleRoots,
	}

	testMirror(t, txsMap, contracts, epochs)
}

func TestMultipleTransactionsInSeparateEpochs(t *testing.T) {
	initMirrorTest()

	epochs := initEpochs()

	startTime := epochs.getStartTime(3)
	endTime := epochs.getEndTime(999)

	txs := make([]database.PChainTxData, 3)
	txIDs := []string{
		"XnfV79XVMyuXbTw8iNreQ9FrUgy9csYBJp1xRscay3oDzhyq8",
		"nsPmyQbm4oo77jyykxbjf7s4Zp4urNptkyAouxVWZ2EB2kw1z",
		"2p32tpqNrfzP3SStbP9bQGHZtJkCxjV3iHNssVnkcpUWxHMSuj",
	}

	for i := 0; i < 3; i++ {
		txs[i] = database.PChainTxData{
			PChainTx: database.PChainTx{
				ChainID:   "costwo",
				NodeID:    "NodeID-CZYx3on11wwYXFoHwZtAQZT5unZ9JHMf6",
				StartTime: &startTime,
				EndTime:   &endTime,
				TxID:      &txIDs[i],
				Type:      database.PChainAddDelegatorTx,
			},
			InputAddress: "costwo18atl0e95w5ym6t8u5yrjpz35vqqzxfzrrsnq8u",
		}
	}

	txsMap := make(map[int64][]database.PChainTxData, 3)
	for i := 0; i < 3; i++ {
		txsMap[int64(i)] = []database.PChainTxData{txs[i]}
	}

	merkleRoots := make(map[int64][32]byte, 3)
	for i := 0; i < 3; i++ {
		txHash, err := hashTransaction(&txs[i])
		require.NoError(t, err)

		merkleRoots[int64(i)] = txHash
	}

	contracts := testContracts{
		merkleRoots: merkleRoots,
	}

	testMirror(t, txsMap, contracts, epochs)
}

func TestAlreadyMirrored(t *testing.T) {
	testMirrorErrors(t, "transaction already mirrored")
}

func TestStakingEnded(t *testing.T) {
	testMirrorErrors(t, "staking already ended")
}

func testMirrorErrors(t *testing.T, errorMsg string) {
	initMirrorTest()

	epochs := initEpochs()

	startTime := epochs.getStartTime(3)
	endTime := epochs.getEndTime(999)

	txid := "5uZETr5SUKqGJLzFP5BeGxbXU5CFcCBQYPu288eX9R1QDQMjn"
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

	txs := map[int64][]database.PChainTxData{
		3: {tx},
	}

	txHash, err := hashTransaction(&tx)
	require.NoError(t, err)

	merkleRoots := map[int64][32]byte{
		3: txHash,
	}

	txidBytes, err := ids.FromString(*tx.TxID)
	require.NoError(t, err)

	contracts := testContracts{
		merkleRoots: merkleRoots,
		mirrorErrors: map[[32]byte]error{
			txidBytes: errors.New(errorMsg),
		},
	}

	testMirror(t, txs, contracts, epochs)
}

func initEpochs() epochInfo {
	epochCfg := config.EpochConfig{
		Period: 180 * time.Second,
		Start:  utils.Timestamp{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
	}

	return newEpochInfo(&epochCfg)
}

func testMirror(
	t *testing.T,
	txs map[int64][]database.PChainTxData,
	contracts testContracts,
	epochs epochInfo,
) {
	db := testDB{
		epochs: epochs,
		states: map[string]database.State{
			pchain.StateName: {
				Updated: epochs.getEndTime(999),
			},
			mirrorStateName: {},
		},
		txs: txs,
	}

	j := mirrorCronJob{
		db:        db,
		contracts: &contracts,
		epochCronjob: epochCronjob{
			enabled: true,
			epochs:  epochs,
		},
	}

	err := j.Call()
	require.NoError(t, err)

	cupaloy.SnapshotT(t, contracts.mirroredStakes)
}

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
	mirrorErrors   map[[32]byte]error
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
	if err := c.mirrorErrors[stakeData.TxId]; err != nil {
		return err
	}

	c.mirroredStakes = append(c.mirroredStakes, mirrorStakeInput{
		stakeData:   stakeData,
		merkleProof: merkleProof,
	})
	return nil
}
