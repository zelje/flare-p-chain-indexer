package xchain

import (
	"flare-indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
	"flare-indexer/utils"
	"flare-indexer/utils/chain"
)

const (
	StateName string = "x_chain_vtx"
)

type xChainTxIndexer struct {
	shared.ChainIndexerBase
}

func CreateXChainTxIndexer(ctx context.IndexerContext) *xChainTxIndexer {
	config := ctx.Config().XChainIndexer
	client := newClient(&ctx.Config().Chain)
	txClient := newTxClient(&ctx.Config().Chain)

	idxr := xChainTxIndexer{}
	idxr.StateName = StateName
	idxr.IndexerName = "X-chain Vertices"
	idxr.Client = client
	idxr.DB = ctx.DB()
	idxr.Config = config
	idxr.InitMetrics(StateName)

	idxr.BatchIndexer = NewXChainBatchIndexer(ctx, client, txClient)

	return &idxr
}

func (xi *xChainTxIndexer) Run() {
	xi.ChainIndexerBase.Run()
}

func newClient(cfg *config.ChainConfig) chain.IndexerClient {
	return chain.NewAvalancheIndexerClient(utils.JoinPaths(cfg.NodeURL, "ext/index/X/vtx"))
}

func newTxClient(cfg *config.ChainConfig) chain.IndexerClient {
	return chain.NewAvalancheIndexerClient(utils.JoinPaths(cfg.NodeURL, "ext/index/X/tx"))
}
