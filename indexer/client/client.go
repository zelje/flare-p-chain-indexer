package client

import (
	"flare-indexer/config"
	"flare-indexer/utils"

	"github.com/ava-labs/avalanchego/indexer"
)

type Clients interface {
	XChainTxClient() indexer.Client
}

type clients struct {
	xChainTxClient indexer.Client
}

func NewClients(cfg *config.ChainConfig) Clients {
	cs := clients{}
	cs.xChainTxClient = indexer.NewClient(utils.JoinPaths(cfg.IndexerURL, "ext/index/X/tx"))
	return &cs
}

func (cs *clients) XChainTxClient() indexer.Client { return cs.xChainTxClient }
