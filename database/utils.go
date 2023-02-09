package database

import (
	"flare-indexer/config"
	"fmt"

	"github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	// List entities to auto-migrate
	entities []interface{} = []interface{}{
		Migration{},
		State{},
		XChainTx{},
		XChainTxInput{},
		XChainTxOutput{},
		PChainTx{},
		PChainTxInput{},
		PChainTxOutput{},
		PChainStakeOutput{},
	}
)

func ConnectAndInitialize(cfg *config.DBConfig) (*gorm.DB, error) {
	// Connect to the database
	dbConfig := mysql.Config{
		User:                 cfg.Username,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:               cfg.Database,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	gormConfig := gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	}
	db, err := gorm.Open(gormMysql.Open(dbConfig.FormatDSN()), &gormConfig)
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

func DoInTransaction(db *gorm.DB, operations ...func(db *gorm.DB) error) error {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, f := range operations {
		if err := f(tx); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
