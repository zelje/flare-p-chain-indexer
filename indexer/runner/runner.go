package runner

import (
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/xchain"
	"flare-indexer/logger"
	"time"
)

func Start(ctx context.IndexerContext) {
	xIndexer := xchain.CreateXChainIndexer(ctx)

	for {
		err := xIndexer.Run()
		if err != nil {
			logger.Error("Indexer error %v", err)
		}
		time.Sleep(time.Duration(ctx.Config().Indexer.TimeoutMillis) * time.Millisecond)
	}
}
