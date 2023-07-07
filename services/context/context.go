package context

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/services/config"

	"github.com/ethereum/go-ethereum/ethclient"

	"gorm.io/gorm"
)

type ServicesContext interface {
	Config() *config.Config
	DB() *gorm.DB
	EthRPCClient() *ethclient.Client
}

type servicesContext struct {
	config       *config.Config
	db           *gorm.DB
	ethRPCClient *ethclient.Client
}

func BuildContext() (ServicesContext, error) {
	ctx := servicesContext{}

	cfg, err := config.BuildConfig()
	if err != nil {
		return nil, err
	}
	ctx.config = cfg
	globalConfig.GlobalConfigCallback.Call(cfg)

	ctx.db, err = database.Connect(&cfg.DB)
	if err != nil {
		return nil, err
	}

	ctx.ethRPCClient, err = ethclient.Dial(cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}

	return &ctx, nil
}

func (c *servicesContext) Config() *config.Config { return c.config }

func (c *servicesContext) DB() *gorm.DB { return c.db }

func (c *servicesContext) EthRPCClient() *ethclient.Client { return c.ethRPCClient }
