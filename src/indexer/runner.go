package indexer

import (
	"flare-indexer/src/indexer/ctx"
	"flare-indexer/src/indexer/xchain"
	"flare-indexer/src/logger"
	"time"
)

func Start(ctx ctx.IndexerContext) {
	xIndexer := xchain.CreateXChainIndexer(ctx)

	for {
		err := xIndexer.Run()
		if err != nil {
			logger.Error("Indexer error %v", err)
		}
		time.Sleep(time.Duration(ctx.Config().Indexer.TimeoutMillis) * time.Millisecond)
	}
}
