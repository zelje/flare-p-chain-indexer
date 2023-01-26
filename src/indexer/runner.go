package indexer

import (
	"flare-indexer/src/logger"
)

func Start(ctx IndexerContext) {
	xIndexer := CreateXChainIndexer(ctx)

	err := xIndexer.Run()
	if err != nil {
		logger.Error("Indexer error %v", err)
	}
	// for {
	// 	ind
	// 	time.Sleep(1 * time.Second)
	// }
}
