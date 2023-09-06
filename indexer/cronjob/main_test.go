package cronjob

import (
	"flare-indexer/utils/chain"
	"log"
	"testing"
)

var (
	testUptimeClient *chain.RecordedUptimeClient
)

func TestMain(m *testing.M) {
	var err error
	testUptimeClient, err = chain.UptimeTestClient()
	if err != nil {
		log.Fatal(err)
	}
	m.Run()
}
