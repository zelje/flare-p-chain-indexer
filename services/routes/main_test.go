package routes

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/services/config"
	"flare-indexer/services/context"
	"log"
	"testing"
)

var (
	testContext context.ServicesContext
)

func TestMain(m *testing.M) {
	var err error
	cfg := testConfig()
	testContext, err = context.BuildTestContext(cfg)
	if err != nil {
		log.Fatal(err)
	}

	m.Run()
}

func testConfig() *config.Config {
	cfg := &config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "localflare",
			ChainID:         162,
		},
		DB: globalConfig.DBConfig{
			Username:   database.MysqlTestUser,
			Password:   database.MysqlTestPassword,
			Host:       database.MysqlTestHost,
			Port:       database.MysqlTestPort,
			Database:   "flare_indexer_services",
			LogQueries: true,
		},
	}
	return cfg
}
