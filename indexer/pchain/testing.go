package pchain

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/indexer/config"
)

func pchainIndexerTestConfig(batchSize int, startIndex uint64) *config.Config {
	cfg := &config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "localflare",
			ChainID:         162,
		},
		PChainIndexer: config.IndexerConfig{
			Enabled:       true,
			TimeoutMillis: 3000,
			BatchSize:     batchSize,
			StartIndex:    startIndex,
		},
		UptimeCronjob: config.CronjobConfig{
			Enabled:        true,
			TimeoutSeconds: 60,
		},
		DB: globalConfig.DBConfig{
			LogQueries: false,
		},
	}
	return cfg
}
