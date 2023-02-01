package context

import (
	"flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/indexer/client"

	"gorm.io/gorm"
)

type IndexerContext interface {
	Config() *config.Config
	DB() *gorm.DB
	Clients() client.Clients
}

type indexerContext struct {
	config  *config.Config
	db      *gorm.DB
	clients client.Clients
}

func BuildContext() (IndexerContext, error) {
	ctx := indexerContext{}

	cfg, err := config.BuildConfig()
	if err != nil {
		return nil, err
	}
	ctx.config = cfg

	ctx.db, err = database.ConnectAndInitialize(&cfg.DB)
	if err != nil {
		return nil, err
	}
	ctx.clients = client.NewClients(&cfg.Chain)

	return &ctx, nil
}

func (c *indexerContext) Config() *config.Config { return c.config }

func (c *indexerContext) DB() *gorm.DB { return c.db }

func (c *indexerContext) Clients() client.Clients { return c.clients }
