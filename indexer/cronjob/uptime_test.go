package cronjob

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/utils"
	"testing"
	"time"
)

func uptimeCronjobTestConfig() *config.Config {
	cfg := &config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "localflare",
			ChainID:         162,
		},
		UptimeCronjob: config.UptimeConfig{
			CronjobConfig: config.CronjobConfig{
				Enabled:        true,
				TimeoutSeconds: 60,
			},
		},
		DB: globalConfig.DBConfig{
			Username:   database.MysqlTestUser,
			Password:   database.MysqlTestPassword,
			Host:       database.MysqlTestHost,
			Port:       database.MysqlTestPort,
			Database:   "flare_indexer_indexer",
			LogQueries: false,
		},
	}
	return cfg

}

func createTestUptimeCronjob() (*uptimeCronjob, error) {
	ctx, err := context.BuildTestContext(uptimeCronjobTestConfig())
	if err != nil {
		return nil, err
	}
	return &uptimeCronjob{
		config: ctx.Config().UptimeCronjob,
		db:     ctx.DB(),
		client: testUptimeClient,
	}, nil
}

func TestUptime(t *testing.T) {
	cronjob, err := createTestUptimeCronjob()
	if err != nil {
		t.Fatal(err)
	}
	testUptimeClient.SetNow(utils.ParseTime("2023-02-02T14:29:50Z"))

	for i := 0; i < 100; i++ {
		cronjob.Call()
		testUptimeClient.Time.AdvanceNow(30 * time.Second)
	}
}
