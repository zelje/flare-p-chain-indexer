package config

import (
	"flare-indexer/config"
)

var (
	IndexerConfigCallback config.ConfigCallback[IndexerApplicationConfig] = config.ConfigCallback[IndexerApplicationConfig]{}
)

type IndexerApplicationConfig interface {
	AddressHRP() string
}

type Config struct {
	DB            config.DBConfig     `toml:"db"`
	Logger        config.LoggerConfig `toml:"logger"`
	Metrics       MetricsConfig       `toml:"metrics"`
	Chain         ChainConfig         `toml:"chain"`
	XChainIndexer IndexerConfig       `toml:"x_chain_indexer"`
	PChainIndexer IndexerConfig       `toml:"p_chain_indexer"`
	UptimeCronjob CronjobConfig       `toml:"uptime_cronjob"`
}

type MetricsConfig struct {
	PrometheusAddress string `toml:"prometheus_address"`
}

type IndexerConfig struct {
	Enabled          bool   `toml:"enabled"`
	TimeoutMillis    int    `toml:"timeout_millis"`
	BatchSize        int    `toml:"batch_size"`
	StartIndex       uint64 `toml:"start_index"`
	OutputsCacheSize int    `toml:"outputs_cache_size"`
}

type CronjobConfig struct {
	Enabled        bool `toml:"enabled"`
	TimeoutSeconds int  `toml:"timeout_seconds"`
}

type ChainConfig struct {
	IndexerURL      string `toml:"indexer_url" envconfig:"CHAIN_INDEXER_URL"`
	ChainAddressHRP string `toml:"address_hrp" envconfig:"CHAIN_ADDRESS_HRP"`
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
		UptimeCronjob: CronjobConfig{
			Enabled:        false,
			TimeoutSeconds: 60,
		},
		Chain: ChainConfig{
			IndexerURL: "http://localhost:9650/",
		},
	}
}

func (c Config) AddressHRP() string {
	return c.Chain.ChainAddressHRP
}

func (c Config) LoggerConfig() config.LoggerConfig {
	return c.Logger
}

func BuildConfig() (*Config, error) {
	cfg := newConfig()
	err := config.ParseConfigFile(cfg, config.CONFIG_FILE, false)
	if err != nil {
		return nil, err
	}
	err = config.ParseConfigFile(cfg, config.LOCAL_CONFIG_FILE, true)
	if err != nil {
		return nil, err
	}
	err = config.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
