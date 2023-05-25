package context

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/services/config"
)

func BuildTestContext(cfg *config.Config) (ServicesContext, error) {
	ctx := servicesContext{}
	var err error

	ctx.config = cfg
	globalConfig.GlobalConfigCallback.Call(cfg)

	ctx.db, err = database.ConnectTestDB(&cfg.DB)
	if err != nil {
		return nil, err
	}
	return &ctx, nil
}
