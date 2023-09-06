package cronjob

import (
	"flare-indexer/utils/chain"
	"log"
	"testing"
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
