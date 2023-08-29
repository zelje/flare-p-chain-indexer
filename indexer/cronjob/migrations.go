package cronjob

import (
	"flare-indexer/database"
	"flare-indexer/indexer/migrations"
	"time"

	"gorm.io/gorm"
)

func init() {
	migrations.Container.Add("2023-08-25-00-00", "Create initial state for voting cronjob", createVotingCronjobState)
}

func createVotingCronjobState(db *gorm.DB) error {
	return database.CreateState(db, &database.State{
		Name:           VotingStateName,
		NextDBIndex:    0,
		LastChainIndex: 0,
		Updated:        time.Now(),
	})
}
