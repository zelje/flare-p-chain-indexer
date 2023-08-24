package runner

import (
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/cronjob"
	"flare-indexer/indexer/pchain"
	"flare-indexer/indexer/xchain"
)

func Start(ctx context.IndexerContext) {
	xIndexer := xchain.CreateXChainTxIndexer(ctx)
	pIndexer := pchain.CreatePChainBlockIndexer(ctx)

	go xIndexer.Run()
	go pIndexer.Run()

	uptimeCronjob := cronjob.NewUptimeCronjob(ctx)
	votingCronjob, _ := cronjob.NewVotingCronjob(ctx)

	go cronjob.RunCronjob(uptimeCronjob)
	go cronjob.RunCronjob(votingCronjob)
}
