package client

import (
	"flare-indexer/config"

	"github.com/ava-labs/avalanchego/indexer"
)

type Clients interface {
	XChainTxClient() indexer.Client
}

type clients struct {
	xChainTxClient indexer.Client
}

func NewClients(cfg *config.Config) Clients {
	cs := clients{}
	cs.xChainTxClient = indexer.NewClient("http://localhost:9650/ext/index/X/tx")
	return &cs
}

func (cs *clients) XChainTxClient() indexer.Client { return cs.xChainTxClient }
