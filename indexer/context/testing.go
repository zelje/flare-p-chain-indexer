package context

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/migrations"
)

func BuildTestContext(cfg *config.Config) (IndexerContext, error) {
	ctx := indexerContext{}
	var err error

	ctx.config = cfg
	globalConfig.GlobalConfigCallback.Call(cfg)

	ctx.db, err = database.ConnectAndInitializeTestDB(&cfg.DB, true)
	if err != nil {
		return nil, err
	}

	err = migrations.Container.ExecuteAll(ctx.db)
	if err != nil {
		return nil, err
	}

	return &ctx, nil
}
