package config

import (
	"flare-indexer/config"
	"flare-indexer/utils"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

type Config struct {
	DB            config.DBConfig     `toml:"db"`
	Logger        config.LoggerConfig `toml:"logger"`
	Chain         config.ChainConfig  `toml:"chain"`
	Metrics       MetricsConfig       `toml:"metrics"`
	XChainIndexer IndexerConfig       `toml:"x_chain_indexer"`
	PChainIndexer IndexerConfig       `toml:"p_chain_indexer"`
	UptimeCronjob UptimeConfig        `toml:"uptime_cronjob"`
	Mirror        MirrorConfig        `toml:"mirroring_cronjob"`
	VotingCronjob VotingConfig        `toml:"voting_cronjob"`
	Epochs        EpochConfig         `toml:"epochs"`
}

type MetricsConfig struct {
	PrometheusAddress string `toml:"prometheus_address"`
}

type IndexerConfig struct {
	Enabled    bool          `toml:"enabled"`
	Timeout    time.Duration `toml:"timeout"`
	BatchSize  int           `toml:"batch_size"`
	StartIndex uint64        `toml:"start_index"`
}

type CronjobConfig struct {
	Enabled   bool          `toml:"enabled"`
	Timeout   time.Duration `toml:"timeout"`
	BatchSize int           `toml:"batch_size"`
}

type MirrorConfig struct {
	CronjobConfig
	MirroringContract common.Address `toml:"contract_address" envconfig:"MIRRORING_CONTRACT_ADDRESS"`
}

type VotingConfig struct {
	CronjobConfig
	ContractAddress common.Address `toml:"contract_address" envconfig:"VOTING_CONTRACT_ADDRESS"`
}

type EpochConfig struct {
	Period time.Duration   `toml:"period" envconfig:"EPOCH_PERIOD"`
	Start  utils.Timestamp `toml:"start" envconfig:"EPOCH_TIME"`
	First  int64           `toml:"first" envconfig:"EPOCH_FIRST"`
}

type UptimeConfig struct {
	CronjobConfig
	EpochConfig
	EnableVoting    bool          `toml:"enable_voting"`
	VotingInterval  time.Duration `toml:"voting_interval"`
	UptimeThreshold float64       `toml:"uptime_threshold"`
}

func newConfig() *Config {
	return &Config{
		XChainIndexer: IndexerConfig{
			Enabled:    true,
			Timeout:    3000 * time.Millisecond,
			BatchSize:  10,
			StartIndex: 0,
		},
		PChainIndexer: IndexerConfig{
			Enabled:    true,
			Timeout:    3000 * time.Millisecond,
			BatchSize:  10,
			StartIndex: 0,
		},
		UptimeCronjob: UptimeConfig{
			CronjobConfig: CronjobConfig{
				Enabled: false,
				Timeout: 60 * time.Second,
			},
		},
		Chain: config.ChainConfig{
			NodeURL: "http://localhost:9650/",
		},
		Epochs: EpochConfig{
			Period: 90 * time.Second,
		},
	}
}

func (c Config) LoggerConfig() config.LoggerConfig {
	return c.Logger
}

func (c Config) ChainConfig() config.ChainConfig {
	return c.Chain
}

func BuildConfig() (*Config, error) {
	cfgFileName := config.ConfigFileName()
	cfg := newConfig()
	err := config.ParseConfigFile(cfg, cfgFileName, false)
	if err != nil {
		return nil, err
	}
	err = config.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
