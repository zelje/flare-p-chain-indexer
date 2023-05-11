package chain

import (
	"path"
	"runtime"
	"testing"
)

func PChainTestClient(t *testing.T) *RecordedIndexerClient {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := path.Split(filename)
	blocksFile := path.Join(dir, "../../resources/test/p_chain_blocks.json")
	client, err := NewRecordedIndexerClient(blocksFile)
	if err != nil {
		t.Fatal(err)
	}
	return client
}
