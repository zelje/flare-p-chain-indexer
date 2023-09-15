//go:build integration
// +build integration

package cronjob

import (
	"flare-indexer/utils/chain"
	"log"
	"testing"
)

const (
	privateKey1 = "0xd49743deccbccc5dc7baa8e69e5be03298da8688a15dd202e20f15d5e0e9a9fb"
	privateKey2 = "0x23c601ae397441f3ef6f1075dcb0031ff17fb079837beadaf3c84d96c6f3e569"
)

var (
	testClient       *chain.RecordedIndexerClient //:= chain.PChainTestClient(t)
	testRPCClient    *chain.RecordedRPCClient     //:= chain.PChainTestRPCClient(t)
	testUptimeClient *chain.RecordedUptimeClient
)

func TestMain(m *testing.M) {
	var err error
	testUptimeClient, err = chain.UptimeTestClient()
	if err != nil {
		log.Fatal(err)
	}
	testClient, err = chain.PChainTestClient()
	if err != nil {
		log.Fatal(err)
	}

	testRPCClient, err = chain.PChainTestRPCClient()
	if err != nil {
		log.Fatal(err)
	}
	m.Run()
}
