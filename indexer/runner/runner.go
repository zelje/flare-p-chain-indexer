package runner

import (
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/pchain"
	"flare-indexer/indexer/xchain"
	"flare-indexer/logger"
	"time"
)

func Start(ctx context.IndexerContext) {
	xIndexer := xchain.CreateXChainTxIndexer(ctx)
	pIndexer := pchain.CreatePChainBlockIndexer(ctx)
	for {
		err := xIndexer.Run()
		if err != nil {
			logger.Error("X-chain indexer error %v", err)
		}
		err = pIndexer.Run()
		if err != nil {
			logger.Error("P-chain indexer error %v", err)
		}
		time.Sleep(time.Duration(ctx.Config().Indexer.TimeoutMillis) * time.Millisecond)
	}
}
