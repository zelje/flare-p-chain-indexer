package staking

import (
	"flare-indexer/database"
	"flare-indexer/logger"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

const MirrorStateName = "mirror_cronjob"

type MirrorDB interface {
	FetchState(name string) (database.State, error)
	UpdateJobState(epoch int64) error
	GetPChainTxsForEpoch(start, end time.Time) ([]database.PChainTxData, error)
	GetPChainTx(txID string) (*database.PChainTx, error)
}

type mirrorDBGorm struct {
	db *gorm.DB
}

func NewMirrorDBGorm(db *gorm.DB) MirrorDB {
	return mirrorDBGorm{db: db}
}

func (m mirrorDBGorm) FetchState(name string) (database.State, error) {
	return database.FetchState(m.db, name)
}

func (m mirrorDBGorm) UpdateJobState(epoch int64) error {
	return m.db.Transaction(func(tx *gorm.DB) error {
		jobState, err := database.FetchState(tx, MirrorStateName)
		if err != nil {
			return errors.Wrap(err, "database.FetchState")
		}

		if jobState.NextDBIndex >= uint64(epoch) {
			logger.Debug("job state already up to date")
			return nil
		}

		jobState.NextDBIndex = uint64(epoch)

		return database.UpdateState(tx, &jobState)
	})
}

func (m mirrorDBGorm) GetPChainTxsForEpoch(start, end time.Time) ([]database.PChainTxData, error) {
	return database.GetPChainTxsForEpoch(&database.GetPChainTxsForEpochInput{
		DB:             m.db,
		StartTimestamp: start,
		EndTimestamp:   end,
	})
}

func (m mirrorDBGorm) GetPChainTx(txID string) (*database.PChainTx, error) {
	return database.FetchPChainTx(m.db, txID)
}
