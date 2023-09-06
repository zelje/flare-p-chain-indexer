package database

import (
	"flare-indexer/config"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	MysqlTestUser     string = "indexeruser"
	MysqlTestPassword string = "indexeruser"
	MysqlTestHost     string = "localhost"
	MysqlTestPort     int    = 3307
)

func ConnectTestDB(cfg *config.DBConfig) (*gorm.DB, error) {
	dbConfig := mysql.Config{
		User:                 cfg.Username,
		Passwd:               cfg.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:               cfg.Database,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	var gormLogLevel logger.LogLevel
	if cfg.LogQueries {
		gormLogLevel = logger.Info
	} else {
		gormLogLevel = logger.Silent
	}
	gormConfig := gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	}
	return gorm.Open(gormMysql.Open(dbConfig.FormatDSN()), &gormConfig)
}

func ConnectAndInitializeTestDB(cfg *config.DBConfig, dropTables bool) (*gorm.DB, error) {
	db, err := ConnectTestDB(cfg)
	if err != nil {
		return nil, err
	}

	if dropTables {
		err = db.Migrator().DropTable(entities...)
		if err != nil {
			return nil, err
		}
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

func FetchUptimes(db *gorm.DB, nodeIDs []string, start time.Time, end time.Time) ([]*UptimeCronjob, error) {
	var uptimes []*UptimeCronjob
	query := db.Table("uptime_cronjobs").
		Where("timestamp >= ?", start).
		Where("timestamp < ?", end)
	if len(nodeIDs) > 0 {
		query = query.Where("node_id IN ?", nodeIDs)
	}
	err := query.Find(&uptimes).Error
	return uptimes, err
}

func FetchAggregations(db *gorm.DB) ([]*UptimeAggregation, error) {
	var aggregations []*UptimeAggregation
	err := db.Find(&aggregations).Error
	return aggregations, err
}
