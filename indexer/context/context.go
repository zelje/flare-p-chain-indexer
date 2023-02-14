package context

import (
	"flare-indexer/database"
	"flare-indexer/indexer/config"

	"gorm.io/gorm"
)

type IndexerContext interface {
	Config() *config.Config
	DB() *gorm.DB
}

type indexerContext struct {
	config *config.Config
	db     *gorm.DB
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
	return &ctx, nil
}

func (c *indexerContext) Config() *config.Config { return c.config }

func (c *indexerContext) DB() *gorm.DB { return c.db }
