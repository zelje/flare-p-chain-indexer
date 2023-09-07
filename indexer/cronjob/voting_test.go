package cronjob

import (
	sysContext "context"
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/pchain"
	"flare-indexer/indexer/shared"
	"flare-indexer/utils"
	"flare-indexer/utils/contracts/voting"
	"math/big"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func votingCronjobTestConfig(epochStart time.Time, dbName string, privateKey string) *config.Config {
	cfg := &config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "localflare",
			ChainID:         31337,
			EthRPCURL:       "http://127.0.0.1:8545",
			PrivateKey:      privateKey,
		},
		VotingCronjob: config.VotingConfig{
			CronjobConfig: config.CronjobConfig{
				Enabled:        true,
				TimeoutSeconds: 30,
			},
			ContractAddress: common.HexToAddress("0x7c2C195CD6D34B8F845992d380aADB2730bB9C6F"),
		},
		Epochs: config.EpochConfig{
			Start:  utils.Timestamp{Time: epochStart},
			Period: 90 * time.Second,
		},
		PChainIndexer: config.IndexerConfig{
			Enabled:       true,
			TimeoutMillis: 3000,
			BatchSize:     200,
			StartIndex:    0,
		},
		DB: globalConfig.DBConfig{
			Username:   database.MysqlTestUser,
			Password:   database.MysqlTestPassword,
			Host:       database.MysqlTestHost,
			Port:       database.MysqlTestPort,
			Database:   dbName,
			LogQueries: false,
		},
		Logger: globalConfig.LoggerConfig{
			Level: "debug",
		},
	}
	return cfg
}

func createTestVotingCronjobs(epochStart time.Time) (*votingCronjob, *votingCronjob, *shared.ChainIndexerBase, *shared.ChainIndexerBase, error) {
	ctx1, err := context.BuildTestContext(votingCronjobTestConfig(epochStart, "flare_indexer_indexer", privateKey1))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cronjob1, err := NewVotingCronjob(ctx1)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx2, err := context.BuildTestContext(votingCronjobTestConfig(epochStart, "flare_indexer_indexer_2", privateKey2))
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cronjob2, err := NewVotingCronjob(ctx2)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	indexer1 := &shared.ChainIndexerBase{
		StateName:    pchain.StateName,
		IndexerName:  "P-chain Blocks Test",
		Client:       testClient,
		DB:           ctx1.DB(),
		Config:       ctx1.Config().PChainIndexer,
		BatchIndexer: pchain.NewPChainBatchIndexer(ctx1, testClient, testRPCClient),
	}
	indexer2 := &shared.ChainIndexerBase{
		StateName:    pchain.StateName,
		IndexerName:  "P-chain Blocks Test",
		Client:       testClient,
		DB:           ctx2.DB(),
		Config:       ctx2.Config().PChainIndexer,
		BatchIndexer: pchain.NewPChainBatchIndexer(ctx1, testClient, testRPCClient),
	}
	return cronjob1, cronjob2, indexer1, indexer2, nil
}

func getMerkleRootFromContract(votingContract *voting.Voting, epoch int64) ([32]byte, error) {
	ctx := sysContext.Background()
	opts := &bind.CallOpts{Context: ctx}
	merkleRoot, err := votingContract.GetMerkleRoot(opts, big.NewInt(epoch))
	if err != nil {
		return [32]byte{}, err
	}
	return merkleRoot, nil
}

func TestVoting(t *testing.T) {
	now := time.Unix(1675349340, 0) // 2023-02-02 14:49:00 UTC
	cronjob1, cronjob2, indexer1, indexer2, err := createTestVotingCronjobs(now)
	require.NoError(t, err)

	// Run indexer to allow voting client test to fetch validator data
	// We need two indexers, each one for a different voting client,
	// since the progress is stored in the DB
	err = indexer1.IndexBatch()
	require.NoError(t, err)
	err = indexer2.IndexBatch()
	require.NoError(t, err)

	cronjob1.time.SetNow(now)
	cronjob2.time.SetNow(now)
	for i := 0; i < 10; i++ {
		err := cronjob1.Call()
		require.NoError(t, err)
		err = cronjob2.Call()
		require.NoError(t, err)
		cronjob1.time.AdvanceNow(30 * time.Second)
		cronjob2.time.AdvanceNow(30 * time.Second)
	}
	root, err := getMerkleRootFromContract(cronjob1.votingContract, 0)
	require.NoError(t, err)
	cupaloy.SnapshotT(t, root)
}
