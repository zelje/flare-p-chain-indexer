package chain

import (
	"path"
	"runtime"
)

func PChainTestClient() (*RecordedIndexerClient, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := path.Split(filename)
	blocksFile := path.Join(dir, "../../resources/test/p_chain_indexer_blocks.json")
	client, err := NewRecordedIndexerClient(blocksFile)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func PChainTestRPCClient() (*RecordedRPCClient, error) {
	_, filename, _, _ := runtime.Caller(0)
	dir, _ := path.Split(filename)
	blocksFile := path.Join(dir, "../../resources/test/p_chain_rpc_data.json")
	client, err := NewRecordedRPCClient(blocksFile)
	if err != nil {
		return nil, err
	}
	return client, nil
}
