package indexer

import (
	"flare-indexer/src/chain"
	"flare-indexer/src/config"
	"flare-indexer/src/dbmodel"

	"gorm.io/gorm"
)

type IndexerContext interface {
	Config() *config.Config
	DB() *gorm.DB
	Clients() chain.Clients
}

type indexerContext struct {
	config  *config.Config
	db      *gorm.DB
	clients chain.Clients
}

func BuildContext() (IndexerContext, error) {
	ctx := indexerContext{}

	cfg, err := config.BuildConfig()
	if err != nil {
		return nil, err
	}
	ctx.config = cfg

	ctx.db, err = dbmodel.ConnectAndInitialize(cfg)
	if err != nil {
		return nil, err
	}
	ctx.clients = chain.NewClients(cfg)

	return &ctx, nil
}

func (c *indexerContext) Config() *config.Config { return c.config }

func (c *indexerContext) DB() *gorm.DB { return c.db }

func (c *indexerContext) Clients() chain.Clients { return c.clients }
