package xchain

import (
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
)

const (
	StateName string = "x_chain_tx"
)

type xChainTxIndexer struct {
	shared.ChainIndexerBase
}

func CreateXChainTxIndexer(ctx context.IndexerContext) shared.ChainIndexer {
	config := ctx.Config().Indexer
	client := ctx.Clients().XChainTxClient()

	idxr := xChainTxIndexer{}
	idxr.StateName = StateName
	idxr.IndexerName = "X-chain Transactions"
	idxr.Client = client
	idxr.DB = ctx.DB()
	idxr.Config = config

	return &idxr
}

func (xi *xChainTxIndexer) Run() error {
	batchHandler := NewTxBatchIndexer(xi.DB, xi.Client, xi.Config.BatchSize)
	return xi.IndexBatch(batchHandler)
}
