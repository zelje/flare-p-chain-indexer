package chain

import (
	"context"
	"flare-indexer/logger"
	"fmt"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/indexer"
)

const (
	IndexerTimeout time.Duration = 3 * time.Minute
)

// Get range of indexed objects by calling "index.getContainerRange"
func FetchContainerRangeFromIndexer(client indexer.Client, from uint64, to int) ([]indexer.Container, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	return client.GetContainerRange(ctx, from, to)
}

// Get last accepted container by calling "index.getLastAccepted"
func FetchLastAcceptedContainer(client indexer.Client) (indexer.Container, uint64, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), IndexerTimeout)
	defer cancelCtx()

	return client.GetLastAccepted(ctx)
}

// Get object by its id by calling "index.getIndex" and "index.getContainerByIndex" successively.
// Returns nil, nil if getIndex failed with an error.
func FetchContainerFromIndexer(client indexer.Client, id string) (*indexer.Container, error) {
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
	fmt.Printf("index is %d\n", index)

	container, err := client.GetContainerByIndex(ctx, index)
	if err != nil {
		return nil, err
	}
	return &container, nil
}
