package pchain

import (
	"flare-indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
	"flare-indexer/utils"
	"flare-indexer/utils/chain"

	"github.com/ybbus/jsonrpc/v3"
)

const (
	StateName string = "p_chain_block"
)

type pChainBlockIndexer struct {
	shared.ChainIndexerBase
}

func CreatePChainBlockIndexer(ctx context.IndexerContext) *pChainBlockIndexer {
	config := ctx.Config().PChainIndexer
	client := newIndexerClient(&ctx.Config().Chain)
	rpcClient := newJsonRpcClient(&ctx.Config().Chain)

	idxr := pChainBlockIndexer{}
	idxr.StateName = StateName
	idxr.IndexerName = "P-chain Blocks"
	idxr.Client = client
	idxr.DB = ctx.DB()
	idxr.Config = config
	idxr.InitMetrics(StateName)

	idxr.BatchIndexer = NewPChainBatchIndexer(ctx, client, rpcClient)

	return &idxr
}

func (xi *pChainBlockIndexer) Run() {
	xi.ChainIndexerBase.Run()
}

func newIndexerClient(cfg *config.ChainConfig) chain.IndexerClient {
	return chain.NewAvalancheIndexerClient(utils.JoinPaths(cfg.NodeURL, "ext/index/P/block"))
}

func newJsonRpcClient(cfg *config.ChainConfig) jsonrpc.RPCClient {
	return jsonrpc.NewClient(utils.JoinPaths(cfg.NodeURL, "ext/bc/P"))
}
