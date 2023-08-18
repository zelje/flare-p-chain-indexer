package main

import (
	"context"
	"errors"
	"flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/logger"
	"flare-indexer/mirror/contracts/mirroring"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	Chain    config.ChainConfig  `toml:"chain"`
	Database config.DBConfig     `toml:"db"`
	Logger   config.LoggerConfig `toml:"logger"`
	Mirror   MirrorConfig        `toml:"mirror"`
}

func (c *Config) ChainConfig() config.ChainConfig {
	return c.Chain
}

func (c *Config) LoggerConfig() config.LoggerConfig {
	return c.Logger
}

type MirrorConfig struct {
	EpochPeriod       time.Duration  `toml:"epoch_period" envconfig:"EPOCH_PERIOD"`
	MirroringContract common.Address `toml:"mirroring_contract" envconfig:"MIRRORING_CONTRACT"`
}

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	db, err := database.ConnectAndInitialize(&cfg.Database)
	if err != nil {
		return err
	}

	mirroringContract, err := newMirroringContract(ctx, cfg)
	if err != nil {
		return err
	}

	client := &Client{
		db:                db,
		mirroringContract: mirroringContract,
	}

	return client.run(ctx)
}

func loadConfig() (*Config, error) {
	cfg := defaultConfig()
	err := config.ParseConfigFile(cfg, config.ConfigFileName(), false)
	if err != nil {
		return nil, err
	}

	if err := config.ReadEnv(cfg); err != nil {
		return nil, err
	}

	config.GlobalConfigCallback.Call(cfg)

	logger.Info("Loaded config %v", cfg)

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Mirror: MirrorConfig{
			EpochPeriod: 90 * time.Second,
		},
	}
}

func newMirroringContract(ctx context.Context, cfg *Config) (*mirroring.Mirroring, error) {
	eth, err := ethclient.DialContext(ctx, cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}

	return mirroring.NewMirroring(cfg.Mirror.MirroringContract, eth)
}

type Client struct {
	db                *gorm.DB
	mirroringContract *mirroring.Mirroring
}

func (c *Client) run(ctx context.Context) error {
	// TODO
	logger.Info("Starting mirror client")
	return errors.New("not implemented")
}
