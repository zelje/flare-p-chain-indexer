package chain

import (
	"context"
	"flare-indexer/logger"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
)

const (
	IndexerTimeout time.Duration = 3 * time.Minute
)

// Get range of indexed objects by calling "index.getContainerRange"
func FetchContainerRangeFromIndexer(client IndexerClient, from uint64, numToFetch int) ([]indexer.Container, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	return client.GetContainerRange(ctx, from, numToFetch)
}

// Get last accepted container by calling "index.getLastAccepted"
func FetchLastAcceptedContainer(client IndexerClient) (indexer.Container, uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	return client.GetLastAccepted(ctx)
}

// Get object by its id by calling "index.getIndex" and "index.getContainerByIndex" successively.
// Returns nil, nil if getIndex failed with an error.
func FetchContainerFromIndexer(client IndexerClient, id string) (*indexer.Container, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	txID, _ := ids.FromString(id)
	index, err := client.GetIndex(ctx, txID)
	if err != nil {
		// This can happen since some transactions (genesis) are not indexed
		// so we don't panic here with an error
		logger.Warn("Cannot fetch a container with id %s", id)
		return nil, nil
	}

	container, err := client.GetContainerByIndex(ctx, index)
	if err != nil {
		return nil, err
	}
	return &container, nil
}
