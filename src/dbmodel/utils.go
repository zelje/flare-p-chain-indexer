package dbmodel

import (
	"flare-indexer/src/config"
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
	}
)

func ConnectAndInitialize(cfg *config.Config) (*gorm.DB, error) {
	// Connect to the database
	dbConfig := mysql.Config{
		User:                 cfg.DB.Username,
		Passwd:               cfg.DB.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.DB.Host, cfg.DB.Port),
		DBName:               cfg.DB.Database,
		AllowNativePasswords: true,
		ParseTime:            true,
	}
	db, err := gorm.Open(gormMysql.Open(dbConfig.FormatDSN()), &gorm.Config{})
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
