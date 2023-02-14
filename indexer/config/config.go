package config

import (
	"flare-indexer/config"
)

const (
	LOCAL_CONFIG_FILE string = "config.local.yml"
	CONFIG_FILE       string = "config.yml"
)

type Config struct {
	DB            config.DBConfig `yaml:"db"`
	Chain         ChainConfig     `yaml:"chain"`
	XChainIndexer IndexerConfig   `yaml:"x_chain_indexer"`
	PChainIndexer IndexerConfig   `yaml:"p_chain_indexer"`
}

type IndexerConfig struct {
	Enabled          bool   `yaml:"enabled"`
	TimeoutMillis    int    `yaml:"timeout_millis"`
	BatchSize        int    `yaml:"batch_size"`
	StartIndex       uint64 `yaml:"start_index"`
	OutputsCacheSize int    `yaml:"outputs_cache_size"`
}

type ChainConfig struct {
	IndexerURL      string `yaml:"indexer_url" envconfig:"CHAIN_INDEXER_URL"`
	ChainAddressHRP string `yaml:"address_hrp" envconfig:"CHAIN_ADDRESS_HRP"`
}

func newConfig() *Config {
	return &Config{
		XChainIndexer: IndexerConfig{
			Enabled:       true,
			TimeoutMillis: 3000,
			BatchSize:     10,
			StartIndex:    0,
		},
		PChainIndexer: IndexerConfig{
			Enabled:       true,
			TimeoutMillis: 3000,
			BatchSize:     10,
			StartIndex:    0,
		},
		Chain: ChainConfig{
			IndexerURL: "http://localhost:9650/",
		},
	}
}

func BuildConfig() (*Config, error) {
	cfg := newConfig()
	err := config.ParseConfigFile(cfg, CONFIG_FILE)
	if err != nil {
		return nil, err
	}
	err = config.ParseConfigFile(cfg, LOCAL_CONFIG_FILE)
	if err != nil {
		return nil, err
	}
	err = config.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
