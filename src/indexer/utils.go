package indexer

import (
	"context"
	"flare-indexer/src/logger"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
)

// Get range of indexed objects by calling "index.getContainerRange"
func fetchContainerRangeFromIndexer(client indexer.Client, from uint64, to int) ([]indexer.Container, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	return client.GetContainerRange(ctx, from, to)
}

// Get object by its id by calling "index.getIndex" and "index.getContainerByIndex" successively.
// Returns nil, nil if getIndex failed with an error.
func fetchContainerFromIndexer(client indexer.Client, id string) (*indexer.Container, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	txID, _ := ids.FromString(id)
	index, err := client.GetIndex(ctx, txID)
	if err != nil {
		// This can happen since some transactions (genesis) are not indexed
		// so we don't panic here with an error
		logger.Warn("Cannot fetch a transaction with id %s", id)
		return nil, nil
	}
	fmt.Printf("index is %d\n", index)

	container, err := client.GetContainerByIndex(ctx, index)
	if err != nil {
		return nil, err
	}
	return &container, nil
}
