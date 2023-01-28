package indexer

import (
	"flare-indexer/src/indexer/ctx"
	"flare-indexer/src/indexer/xchain"
	"flare-indexer/src/logger"
)

func Start(ctx ctx.IndexerContext) {
	xIndexer := xchain.CreateXChainIndexer(ctx)

	err := xIndexer.Run()
	if err != nil {
		logger.Error("Indexer error %v", err)
	}
	// for {
	// 	ind
	// 	time.Sleep(1 * time.Second)
	// }
}
