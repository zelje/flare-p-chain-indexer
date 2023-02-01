package shared

import (
	"flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/logger"
	"flare-indexer/utils/chain"
	"time"

	"github.com/ava-labs/avalanchego/indexer"
	"gorm.io/gorm"
)

type ChainIndexer interface {
	Run() error
}

type ContainerBatchIndexer interface {
	ProcessContainers(nextIndex uint64, containers []indexer.Container) (uint64, error)
	PersistEntities(db *gorm.DB) error
}

type ChainIndexerBase struct {
	StateName   string
	IndexerName string

	DB     *gorm.DB
	Client indexer.Client
	Config config.IndexerConfig
}

func (ci *ChainIndexerBase) IndexBatch(ch ContainerBatchIndexer) error {
	startTime := time.Now()

	// Get current state of tx indexer from db
	currentState, err := database.FetchState(ci.DB, ci.StateName)
	if err != nil {
		return err
	}

	var nextIndex uint64
	if currentState.NextDBIndex < ci.Config.StartIndex {
		nextIndex = ci.Config.StartIndex
	} else {
		nextIndex = currentState.NextDBIndex
	}

	// Fetch last accepted index on chain
	_, lastIndex, err := chain.FetchLastAcceptedContainer(ci.Client)
	if err != nil {
		return err
	}
	if lastIndex < nextIndex {
		// Nothing to do; no new containers
		logger.Debug("Nothing to do. Last index %d < next to process %d", lastIndex, nextIndex)
		return nil
	}

	// Get MaxBatch containers from the chain
	containers, err := chain.FetchContainerRangeFromIndexer(ci.Client, nextIndex, ci.Config.BatchSize)
	if err != nil {
		return err
	}

	lastProcessedIndex, err := ch.ProcessContainers(nextIndex, containers)
	if err != nil {
		return err
	}

	err = database.DoInTransaction(ci.DB,
		func(db *gorm.DB) error { return ch.PersistEntities(db) },
		func(db *gorm.DB) error {
			currentState.Update(lastProcessedIndex+1, lastIndex)
			return database.UpdateState(db, &currentState)
		},
	)
	if err != nil {
		return err
	}
	endTime := time.Now()
	logger.Info("Indexer '%s' processed to index %d, last accepted index is %d, duration %dms",
		ci.IndexerName,
		lastProcessedIndex, lastIndex, endTime.Sub(startTime).Milliseconds())
	return nil

}
