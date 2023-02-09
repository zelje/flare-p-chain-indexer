package pchain

import (
	"flare-indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
	"flare-indexer/utils"

	"github.com/ava-labs/avalanchego/indexer"
)

const (
	StateName string = "p_chain_block"
)

type pChainBlockIndexer struct {
	shared.ChainIndexerBase
}

func CreatePChainBlockIndexer(ctx context.IndexerContext) shared.ChainIndexer {
	config := ctx.Config().Indexer
	client := newIndexerClient(&ctx.Config().Chain)

	idxr := pChainBlockIndexer{}
	idxr.StateName = StateName
	idxr.IndexerName = "P-chain Blocks"
	idxr.Client = client
	idxr.DB = ctx.DB()
	idxr.Config = config

	idxr.BatchIndexer := NewTxBatchIndexer(xi.DB, xi.Client, xi.Config.BatchSize)

	return &idxr
}

func (xi *pChainBlockIndexer) Run() error {
	return xi.IndexBatch()
}

func newIndexerClient(cfg *config.ChainConfig) indexer.Client {
	return indexer.NewClient(utils.JoinPaths(cfg.IndexerURL, "ext/index/P/block"))
}
