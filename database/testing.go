package database

import (
	"flare-indexer/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectTestDB(cfg *config.DBConfig) (*gorm.DB, error) {
	var gormLogLevel logger.LogLevel
	if cfg.LogQueries {
		gormLogLevel = logger.Info
	} else {
		gormLogLevel = logger.Silent
	}
	gormConfig := gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	}
	return gorm.Open(sqlite.Open(":memory:"), &gormConfig)
}

func ConnectAndInitializeTestDB(cfg *config.DBConfig) (*gorm.DB, error) {
	db, err := ConnectTestDB(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize - auto migrate
	err = db.AutoMigrate(entities...)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Queries for testing
/////////////////////////////////////////////////////////////////////////////////////////

// Fetch transactions by block heights
func FetchTransactionsByBlockHeights(db *gorm.DB, heights []uint64) ([]*PChainTx, error) {
	var transactions []*PChainTx
	err := db.Where("block_height IN ?", heights).Find(&transactions).Error
	return transactions, err
}
